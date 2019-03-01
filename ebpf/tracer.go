// +build linux_bpf

package ebpf

import (
	"bytes"
	"fmt"
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	log "github.com/cihub/seelog"
	bpflib "github.com/iovisor/gobpf/elf"
)

var (
	// Feature versions sourced from: https://github.com/iovisor/bcc/blob/master/docs/kernel-versions.md
	// Minimum kernel version -> max(3.15 - eBPF,
	//                               3.18 - tables/maps,
	//                               4.1 - kprobes,
	//                               4.3 - perf events)
	// 	                      -> 4.3
	minRequiredKernelCode = linuxKernelVersionCode(4, 3, 0)
)

type Tracer struct {
	m *bpflib.Module

	config *Config

	state          NetworkState
	portMapping    *PortMapping
	localAddresses map[string]struct{}

	perfMap      *bpflib.PerfMap
	perfReceived uint64
	perfLost     uint64
}

// maxActive configures the maximum number of instances of the kretprobe-probed functions handled simultaneously.
// This value should be enough for typical workloads (e.g. some amount of processes blocked on the accept syscall).
const (
	maxActive = 128
)

// CurrentKernelVersion exposes calculated kernel version - exposed in LINUX_VERSION_CODE format
// That is, for kernel "a.b.c", the version number will be (a<<16 + b<<8 + c)
func CurrentKernelVersion() (uint32, error) {
	return bpflib.CurrentKernelVersion()
}

// IsTracerSupportedByOS returns whether or not the current kernel version supports tracer functionality
func IsTracerSupportedByOS() (bool, error) {
	currentKernelCode, err := bpflib.CurrentKernelVersion()
	if err != nil {
		return false, fmt.Errorf("could not get kernel version: %s", err)
	}

	if currentKernelCode < minRequiredKernelCode {
		return false, fmt.Errorf("incompatible linux version. at least %d required, got %d", minRequiredKernelCode, currentKernelCode)
	}
	return true, nil
}

func NewTracer(config *Config) (*Tracer, error) {
	m, err := readBPFModule()
	if err != nil {
		return nil, fmt.Errorf("could not read bpf module: %s", err)
	}

	err = m.Load(SectionsFromConfig(config))
	if err != nil {
		return nil, fmt.Errorf("could not load bpf module: %s", err)
	}

	// Use the config to determine what kernel probes should be enabled
	enabledProbes := config.EnabledKProbes()
	for k := range m.IterKprobes() {
		if _, ok := enabledProbes[KProbeName(k.Name)]; ok {
			if err = m.EnableKprobe(k.Name, maxActive); err != nil {
				return nil, fmt.Errorf("could not enable kprobe(%s): %s", k.Name, err)
			}
		}
	}

	// TODO: Disable TCPv{4,6} connect kernel probes once offsets have been figured out.
	if err := guess(m, config); err != nil {
		return nil, fmt.Errorf("failed to init module: error guessing offsets: %v", err)
	}

	portMapping := NewPortMapping(config.ProcRoot, config)
	if err := portMapping.ReadInitialState(); err != nil {
		return nil, fmt.Errorf("failed to read initial pid->port mapping: %s", err)
	}

	localAddresses := readLocalAddresses()

	tr := &Tracer{
		m:              m,
		config:         config,
		state:          NewDefaultNetworkState(),
		portMapping:    portMapping,
		localAddresses: localAddresses,
	}

	tr.perfMap, err = tr.initPerfPolling()
	if err != nil {
		return nil, fmt.Errorf("could not start polling bpf events: %s", err)
	}

	return tr, nil
}

// initPerfPolling starts the listening on perf buffer events to grab closed connections
func (t *Tracer) initPerfPolling() (*bpflib.PerfMap, error) {
	closedChannel := make(chan []byte, 100)
	lostChannel := make(chan uint64, 10)

	pm, err := bpflib.InitPerfMap(t.m, string(tcpCloseEventMap), closedChannel, lostChannel)
	if err != nil {
		return nil, fmt.Errorf("error initializing perf map: %s", err)
	}

	pm.PollStart()

	go func() {
		// Stats about how much connections have been closed / lost
		ticker := time.NewTicker(5 * time.Minute)

		for {
			select {
			case conn, ok := <-closedChannel:
				if !ok {
					return
				}
				atomic.AddUint64(&t.perfReceived, 1)
				t.state.StoreClosedConnection(decodeRawTCPConn(conn))
			case lostCount, ok := <-lostChannel:
				if !ok {
					return
				}
				atomic.AddUint64(&t.perfLost, lostCount)
			case <-ticker.C:
				recv := atomic.SwapUint64(&t.perfReceived, 0)
				lost := atomic.SwapUint64(&t.perfLost, 0)
				if lost > 0 {
					log.Debugf("closed connection polling: %d received, %d lost", recv, lost)
				}
			}
		}
	}()

	return pm, nil
}

func (t *Tracer) Stop() {
	t.m.Close()
	t.perfMap.PollStop()
}

func (t *Tracer) GetActiveConnections(clientID string) (*Connections, error) {
	latestConns, latestTime, err := t.getConnections()
	if err != nil {
		return nil, fmt.Errorf("error retrieving connections: %s", err)
	}

	return &Connections{Conns: t.state.Connections(clientID, latestTime, latestConns)}, nil
}

func (t *Tracer) getConnections() ([]ConnectionStats, uint64, error) {
	mp, err := t.getMap(connMap)
	if err != nil {
		return nil, 0, fmt.Errorf("error retrieving the bpf %s map: %s", connMap, err)
	}

	tcpMp, err := t.getMap(tcpStatsMap)
	if err != nil {
		return nil, 0, fmt.Errorf("error retrieving the bpf %s map: %s", tcpStatsMap, err)
	}

	portMp, err := t.getMap(portBindingsMap)
	if err != nil {
		return nil, 0, fmt.Errorf("error retrieving the bpf %s map: %s", portBindingsMap, err)
	}

	latestTime, ok, err := t.getLatestTimestamp()
	if err != nil {
		return nil, 0, fmt.Errorf("error retrieving latest timestamp: %s", err)
	}

	if !ok { // if no timestamps have been captured, there can be no packets
		return nil, 0, nil
	}

	closedPortBindings, err := t.populatePortMapping(portMp)
	if err != nil {
		return nil, 0, fmt.Errorf("error populating port mapping: %s", err)
	}

	// Iterate through all key-value pairs in map
	key, nextKey, stats := &ConnTuple{}, &ConnTuple{}, &ConnStatsWithTimestamp{}
	active := make([]ConnectionStats, 0)
	expired := make([]*ConnTuple, 0)
	for {
		hasNext, _ := t.m.LookupNextElement(mp, unsafe.Pointer(key), unsafe.Pointer(nextKey), unsafe.Pointer(stats))
		if !hasNext {
			break
		} else if stats.isExpired(latestTime, t.timeoutForConn(nextKey)) {
			expired = append(expired, nextKey.copy())
		} else {
			conn := connStats(nextKey, stats, t.getTCPStats(tcpMp, nextKey))

			conn.Direction = t.determineConnectionDirection(&conn)

			active = append(active, conn)
		}
		key = nextKey
	}

	// Remove expired entries
	t.removeEntries(mp, tcpMp, expired)

	for _, key := range closedPortBindings {
		t.portMapping.RemoveMapping(key)

		_ = t.m.DeleteElement(portMp, unsafe.Pointer(&key))
	}

	return active, latestTime, nil
}

func (t *Tracer) removeEntries(mp, tcpMp *bpflib.Map, entries []*ConnTuple) {
	for i := range entries {
		err := t.m.DeleteElement(mp, unsafe.Pointer(entries[i]))
		if err != nil {
			// It's possible some other process deleted this entry already (e.g. tcp_close)
			log.Warnf("failed to remove entry from connections map: %s", err)
		}

		// We have to remove the PID to remove the element from the TCP Map since we don't use the pid there
		entries[i].pid = 0
		// We can ignore the error for this map since it will not always contain the entry
		_ = t.m.DeleteElement(tcpMp, unsafe.Pointer(entries[i]))
	}
}

// getTCPStats reads tcp related stats for the given ConnTuple
func (t *Tracer) getTCPStats(mp *bpflib.Map, tuple *ConnTuple) *TCPStats {
	// Remove the PID since we don't use it in the TCP Stats map
	tup := tuple.copy()
	tup.pid = 0

	stats := &TCPStats{retransmits: 0}
	if err := t.m.LookupElement(mp, unsafe.Pointer(tup), unsafe.Pointer(stats)); err != nil {
		return stats
	}

	return stats
}

// getLatestTimestamp reads the most recent timestamp captured by the eBPF
// module.  if the eBFP module has not yet captured a timestamp (as will be the
// case if the eBPF module has just started), the second return value will be
// false.
func (t *Tracer) getLatestTimestamp() (uint64, bool, error) {
	tsMp, err := t.getMap(latestTimestampMap)
	if err != nil {
		return 0, false, fmt.Errorf("error retrieving latest timestamp map: %s", err)
	}

	var latestTime uint64
	if err := t.m.LookupElement(tsMp, unsafe.Pointer(&zero), unsafe.Pointer(&latestTime)); err != nil {
		// If we can't find latest timestamp, there probably haven't been any messages yet
		return 0, false, nil
	}

	return latestTime, true, nil
}

func (t *Tracer) getMap(name bpfMapName) (*bpflib.Map, error) {
	mp := t.m.Map(string(name))
	if mp == nil {
		return nil, fmt.Errorf("no map with name %s", name)
	}
	return mp, nil
}

func readBPFModule() (*bpflib.Module, error) {
	buf, err := Asset("tracer-ebpf.o")
	if err != nil {
		return nil, fmt.Errorf("couldn't find asset: %s", err)
	}

	m := bpflib.NewModuleFromReader(bytes.NewReader(buf))
	if m == nil {
		return nil, fmt.Errorf("BPF not supported")
	}
	return m, nil
}

func (t *Tracer) timeoutForConn(c *ConnTuple) uint64 {
	if connType(uint(c.metadata)) == TCP {
		return uint64(t.config.TCPConnTimeout.Nanoseconds())
	}
	return uint64(t.config.UDPConnTimeout.Nanoseconds())
}

// GetStats returns a map of statistics about the current tracer's internal state
func (t *Tracer) GetStats() (map[string]interface{}, error) {
	if t.state == nil {
		return nil, fmt.Errorf("internal state not yet initialized")
	}
	return t.state.GetStats(atomic.LoadUint64(&t.perfLost), atomic.LoadUint64(&t.perfReceived)), nil
}

// populatePortMapping reads the entire portBinding bpf map and populates the local port/address map.  A list of
// closed ports will be returned
func (t *Tracer) populatePortMapping(mp *bpflib.Map) ([]uint16, error) {
	var key, nextKey uint16
	var state uint8

	closedPortBindings := make([]uint16, 0)

	for {
		hasNext, _ := t.m.LookupNextElement(mp, unsafe.Pointer(&key), unsafe.Pointer(&nextKey), unsafe.Pointer(&state))
		if !hasNext {
			break
		}

		port := nextKey

		t.portMapping.AddMapping(port)

		if isPortClosed(state) {
			closedPortBindings = append(closedPortBindings, port)
		}

		key = nextKey
	}

	return closedPortBindings, nil
}

func (t *Tracer) determineConnectionDirection(conn *ConnectionStats) ConnectionDirection {
	sourceLocal := t.isLocalAddress(conn.Source)
	destLocal := t.isLocalAddress(conn.Dest)

	if sourceLocal && destLocal {
		return LOCAL
	}

	if sourceLocal && t.portMapping.IsListening(conn.SPort) {
		return INCOMING
	}

	return OUTGOING
}

func (t *Tracer) isLocalAddress(address string) bool {
	_, ok := t.localAddresses[address]
	return ok
}

func readLocalAddresses() map[string]struct{} {
	addresses := make(map[string]struct{}, 0)

	interfaces, err := net.Interfaces()
	if err != nil {
		log.Errorf("error reading network interfaces: %s", err)
		return addresses
	}

	for _, intf := range interfaces {
		addrs, err := intf.Addrs()

		if err != nil {
			log.Errorf("error reading interface %s addresses", intf.Name, err)
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				addresses[v.IP.String()] = struct{}{}
			case *net.IPAddr:
				addresses[v.IP.String()] = struct{}{}
			}
		}

	}

	return addresses
}

// SectionsFromConfig returns a map of string -> gobpf.SectionParams used to configure the way we load the BPF program (bpf map sizes)
func SectionsFromConfig(c *Config) map[string]bpflib.SectionParams {
	return map[string]bpflib.SectionParams{
		connMap.sectionName(): {
			MapMaxEntries: int(c.MaxTrackedConnections),
		},
		tcpStatsMap.sectionName(): {
			MapMaxEntries: int(c.MaxTrackedConnections),
		},
		portBindingsMap.sectionName(): {
			MapMaxEntries: int(c.MaxTrackedConnections),
		},
		tcpCloseEventMap.sectionName(): {
			MapMaxEntries: 1024,
		},
	}
}
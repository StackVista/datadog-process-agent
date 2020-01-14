// +build !windows

package config

const (
	defaultLogFilePath = "/var/log/datadog/process-agent.log"

	// Agent 5
	defaultDDAgentPy    = "/opt/datadog-agent/embedded/bin/python"
	defaultDDAgentPyEnv = "PYTHONPATH=/opt/datadog-agent/agent"

	// Agent 6
	defaultDDAgentBin = "/opt/datadog-agent/bin/agent/agent"
)

// Process blacklist
var defaultBlacklistPatterns = []string{
	"stress",
	"^-bash",
	"^su$",
	"^/lib/systemd/",
	"^/usr/lib/systemd/",
	"^pickup",
	"^/sbin/",
	"^/usr/sbin/",
	"^qmgr",
	"^sshd:",
	"^/usr/bin/vi(?:m|m.basic)?$",
	"^/usr/bin/tail",
	"^\\(sd-pam\\)",
	"^containerd-shim",
	"^pause",
	"^/pause",
	// Kubernetes / Containerized environments
	"^/usr/lib/accountsservice/accounts-daemon$",
	"^/usr/sbin/acpid$",
	"^/adapter$",
	"^agent$",
	"^/opt/stackstate-agent/bin/agent/agent$",
	"^/sbin/agetty$",
	"^/bin/alertmanager*$",
	"^/usr/bin/amazon-ssm-agent$",
	"^/usr/lib/at-spi2-core/at-spi-bus-launcher$",
	"^/usr/lib/at-spi2-core/at-spi2-registryd$",
	"^atlantis$",
	"^avahi-daemon.$",
	"^awk$",
	"^/app/aws-k8s-agent$",
	"^bash$",
	"^/bin/bash$",
	"^/bin/busybox$",
	"^/usr/bin/bash$",
	"^/usr/lib/bluetooth/bluetoothd$",
	"^/app/cmd/cainjector/cainjector$",
	"^calico-node$",
	"^calico-typha$",
	"^/bin/chaoskube$",
	"^./cluster-autoscaler$",
	"^/usr/lib/colord/colord$",
	"^/configmap-reload$",
	"^/usr/bin/containerd$",
	"^./controller$",
	"^/coredns$",
	"^/usr/sbin/cron$",
	"^crond$",
	"^/usr/sbin/cupsd$",
	"^/usr/bin/dbus-daemon$",
	"^/usr/lib/dconf/dconf-service$",
	"^/sbin/dhclient$",
	"^/usr/sbin/dnsmasq$",
	"^/usr/bin/dockerd$",
	"^dotnet$",
	"^.*dumb-init$",
	"^/usr/bin/exporter$",
	"^external-dns$",
	"^/usr/lib/x86_64-linux-gnu/fwupd/fwupd$",
	"^/usr/bin/ghostunnel$",
	"^/usr/lib/gvfs/gvfsd-fuse$",
	"^/usr/lib/gvfs/gvfsd$",
	"^.*indicator-application/indicator-application-service$",
	"^.*indicator-bluetooth/indicator-bluetooth-service$",
	"^.*indicator-datetime/indicator-datetime-service$",
	"^.*indicator-keyboard/indicator-keyboard-service$",
	"^.*indicator-messages/indicator-messages-service$",
	"^.*indicator-power/indicator-power-service$",
	"^.*indicator-session/indicator-session-service$",
	"^.indicator-sound/indicator-sound-service$",
	"^/bin/IngressMonitorController$",
	"^/sbin/init$",
	"^/usr/sbin/irqbalance$",
	"^/go/bin/kube-eagle$",
	"^kube-proxy$",
	"^/kube-state-metrics$",
	"^kube2iam$",
	"^lightdm$",
	"^/usr/bin/lsmd$",
	"^/sbin/lvmetad$",
	"^/usr/libexec/postfix/master$",
	"^/metrics-server$",
	"^/usr/sbin/ModemManager$",
	"^/usr/sbin/NetworkManager$",
	"^/nginx-ingress-controller$",
	"^nm-applet$",
	"^/usr/bin/kubelet$",
	"^/usr/sbin/lightdm$",
	"^/bin/node_exporter$",
	"^/usr/lib/x86_64-linux-gnu/notify-osd$",
	"^/bin/oauth2_proxy$",
	"^/usr/lib/policykit-1/polkitd",
	"^.process-agent$",
	"^/usr/bin/pulseaudio$",
	"^python$",
	"^/usr/bin/python$",
	"^rescheduler$",
	"^/usr/sbin/rsyslogd$",
	"^/usr/lib/rtkit/rtkit-daemon$",
	"^ruby$",
	"^/usr/local/bin/ruby$",
	"^runsv$",
	"^/usr/bin/runsvdir$",
	"^/home/weave/runsvinit$",
	"^/bin/operator$",
	"^s6-format-filter$",
	"^s6-supervise$",
	"^s6-svscan$",
	"^s6-",
	"^sleep$",
	"^/usr/sbin/sshd$",
	"^./lib/systemd/systemd-journald$",
	"^/lib/systemd/systemd-logind$",
	"^/lib/systemd/systemd-timesyncd$",
	"^./lib/systemd/systemd-udevd$",
	"^/lib/systemd/systemd$",
	"^tail$",
	"^/usr/sbin/thermald$",
	"^/tiller$",
	"^/usr/bin/tini$",
	"^/opt/stackstate-agent/embedded/bin/trace-agent$",
	"tini$",
	"^/usr/lib/udisks2/udisksd$",
	"^/usr/sbin/unity-greeter$",
	"^.*unity-settings-daemon$",
	"^/usr/bin/unpigz$",
	"^/usr/lib/upower/upowerd$",
	"^upstart$",
	"^/app/cmd/webhook/webhook$",
	"^/usr/bin/whoopsie$",
	"^/sbin/wpa_supplicant$",
	"^/usr/lib/xorg/Xorg$",
}

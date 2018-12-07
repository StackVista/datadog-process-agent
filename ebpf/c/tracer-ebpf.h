#ifndef __TRACER_BPF_H
#define __TRACER_BPF_H

#include <linux/types.h>

static const __u8 GUESS_SADDR      = 0;
static const __u8 GUESS_DADDR      = 1;
static const __u8 GUESS_FAMILY     = 2;
static const __u8 GUESS_SPORT      = 3;
static const __u8 GUESS_DPORT      = 4;
static const __u8 GUESS_NETNS      = 5;
static const __u8 GUESS_DADDR_IPV6 = 6;

#ifndef TASK_COMM_LEN
#define TASK_COMM_LEN 16
#endif

typedef struct {
    char comm[TASK_COMM_LEN];
} proc_t ;

typedef struct {
	__u64 send_bytes;
	__u64 recv_bytes;
	__u64 timestamp;
} conn_stats_ts_t;


// Metadata bit masks
static const __u8 CONN_TYPE_UDP = 0;
static const __u8 CONN_TYPE_TCP = 1;

// tcp_set_state doesn't run in the context of the process that initiated the
// connection so we need to store a map TUPLE -> PID to send the right PID on
// the event
typedef struct {
	__u32 saddr;
	__u32 daddr;
	__u16 sport;
	__u16 dport;
	__u32 netns;
	__u32 pid;
	// Metadata description:
	// First bit indicates if the connection is TCP (1) or UDP (0)
	__u32 metadata; // This is that big because it seems that we atleast need a 32-bit aligned struct
} ipv4_tuple_t;

typedef struct {
	/* Using the type unsigned __int128 generates an error in the ebpf verifier */
	__u64 saddr_h;
	__u64 saddr_l;
	__u64 daddr_h;
	__u64 daddr_l;
	__u16 sport;
	__u16 dport;
	__u32 netns;
	__u32 pid;
	// Metadata description:
	// First bit indicates if the connection is TCP (1) or UDP (0)
	__u32 metadata; // This is that big because it seems that we atleast need a 32-bit aligned struct
} ipv6_tuple_t;

static const __u8 TRACER_STATE_UNINITIALIZED = 0;
static const __u8 TRACER_STATE_CHECKING      = 1;
static const __u8 TRACER_STATE_CHECKED       = 2;
static const __u8 TRACER_STATE_READY         = 3;

static const __u8 TRACER_IPV6_DISABLED = 0;
static const __u8 TRACER_IPV6_ENABLED  = 1;

typedef struct {
	__u64 state;

	/* checking */
	proc_t proc;
	__u64 what;
	__u64 offset_saddr;
	__u64 offset_daddr;
	__u64 offset_sport;
	__u64 offset_dport;
	__u64 offset_netns;
	__u64 offset_ino;
	__u64 offset_family;
	__u64 offset_daddr_ipv6;

	__u64 err;

	__u32 daddr_ipv6[4];
	__u32 netns;
	__u32 saddr;
	__u32 daddr;
	__u16 sport;
	__u16 dport;
	__u16 family;

	__u8 ipv6_enabled;
	__u8 padding;
} tracer_status_t;

#endif

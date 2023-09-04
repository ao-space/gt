package request

import psNet "github.com/shirou/gopsutil/v3/net"

// SimplifiedConnection is a simplified version of gopsutil ConnectionStat
type SimplifiedConnection struct {
	Family uint32     `json:"family"`
	Type   uint32     `json:"type"`
	Laddr  psNet.Addr `json:"localaddr"`
	Raddr  psNet.Addr `json:"remoteaddr"`
	Status string     `json:"status"`
}

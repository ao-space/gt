package request

import "github.com/shirou/gopsutil/v3/net"

// SimplifiedConnection is a simplified version of gopsutil ConnectionStat
type SimplifiedConnection struct {
	Family uint32   `json:"family"`
	Type   uint32   `json:"type"`
	Laddr  net.Addr `json:"localaddr"`
	Raddr  net.Addr `json:"remoteaddr"`
	Status string   `json:"status"`
}
type SimplifiedConnectionWithID struct {
	ID     string   `json:"id"`
	Family uint32   `json:"family"`
	Type   uint32   `json:"type"`
	Laddr  net.Addr `json:"localaddr"`
	Raddr  net.Addr `json:"remoteaddr"`
	Status string   `json:"status"`
}

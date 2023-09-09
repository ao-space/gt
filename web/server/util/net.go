package util

import (
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/shirou/gopsutil/v3/net"
	"strconv"
)

func ConvertToNetAddrString(addr net.Addr) string {
	return addr.IP + ":" + strconv.Itoa(int(addr.Port))
}

func FilterOutMatchingConnections(source []net.ConnectionStat, filter []client.PoolInfo) []net.ConnectionStat {
	var filteredConns []net.ConnectionStat

	filterMap := make(map[string]struct{})
	for _, i := range filter {
		key := i.LocalAddr.String() + "-" + i.RemoteAddr.String()
		filterMap[key] = struct{}{}
	}

	for _, conn := range source {
		key := ConvertToNetAddrString(conn.Laddr) + "-" + ConvertToNetAddrString(conn.Raddr)
		if _, exist := filterMap[key]; !exist {
			filteredConns = append(filteredConns, conn)
		}
	}

	return filteredConns
}

func SimplifyConnections(conns []net.ConnectionStat) []request.SimplifiedConnection {
	var simplifiedConns []request.SimplifiedConnection

	for _, conn := range conns {
		simplifiedConns = append(simplifiedConns, request.SimplifiedConnection{
			Family: conn.Family,
			Type:   conn.Type,
			Laddr:  conn.Laddr,
			Raddr:  conn.Raddr,
			Status: conn.Status,
		})
	}

	return simplifiedConns
}

func SwitchToPoolInfo(conns []server.ConnectionInfo) []client.PoolInfo {
	var poolInfos []client.PoolInfo

	for _, conn := range conns {
		poolInfos = append(poolInfos, client.PoolInfo{
			LocalAddr:  conn.LocalAddr,
			RemoteAddr: conn.RemoteAddr,
		})
	}

	return poolInfos
}

func SelectedMatchingConnections(source []net.ConnectionStat, filter []server.ConnectionInfo) map[string][]net.ConnectionStat {
	filteredConns := make(map[string][]net.ConnectionStat)

	filterMap := make(map[string]string) // store id
	for _, i := range filter {
		key := i.LocalAddr.String() + "-" + i.RemoteAddr.String()
		filterMap[key] = i.ID
	}

	for _, conn := range source {
		key := ConvertToNetAddrString(conn.Laddr) + "-" + ConvertToNetAddrString(conn.Raddr)
		if id, exist := filterMap[key]; exist {
			filteredConns[id] = append(filteredConns[id], conn)
		}
	}

	return filteredConns
}

func SimplifyConnectionsWithID(conns map[string][]net.ConnectionStat) []request.SimplifiedConnectionWithID {
	var simplifiedConns []request.SimplifiedConnectionWithID

	for id, connSlice := range conns {
		for _, conn := range connSlice {
			simplifiedConns = append(simplifiedConns, request.SimplifiedConnectionWithID{
				ID:     id,
				Family: conn.Family,
				Type:   conn.Type,
				Laddr:  conn.Laddr,
				Raddr:  conn.Raddr,
				Status: conn.Status,
			})
		}
	}

	return simplifiedConns
}

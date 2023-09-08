package util

import (
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/web/server/model/request"
	psNet "github.com/shirou/gopsutil/v3/net"
	"strconv"
)

func ConvertToNetAddrString(addr psNet.Addr) string {
	return addr.IP + ":" + strconv.Itoa(int(addr.Port))
}

func FilterOutMatchingConnections(source []psNet.ConnectionStat, filter []client.PoolInfo) []psNet.ConnectionStat {
	var filteredConns []psNet.ConnectionStat

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

func SimplifyConnections(conns []psNet.ConnectionStat) []request.SimplifiedConnection {
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

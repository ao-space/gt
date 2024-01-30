package util

import (
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/web/server/model/request"
	psNet "github.com/shirou/gopsutil/v3/net"
	"net"
	"reflect"
	"testing"
)

func TestConvertToNetAddrString(t *testing.T) {
	tests := []struct {
		addr     psNet.Addr
		expected string
	}{
		{
			addr: psNet.Addr{
				IP:   "127.0.0.1",
				Port: 8080,
			},
			expected: "127.0.0.1:8080",
		},
		{
			addr: psNet.Addr{
				IP:   "192.168.1.1",
				Port: 22,
			},
			expected: "192.168.1.1:22",
		},
	}

	for _, tt := range tests {
		result := ConvertToNetAddrString(tt.addr)
		if result != tt.expected {
			t.Errorf("Expected %s, got %s for IP %s and Port %d", tt.expected, result, tt.addr.IP, tt.addr.Port)
		}
	}
}

func TestFilterOutMatchingConnections(t *testing.T) {
	source := []psNet.ConnectionStat{
		{
			Laddr: psNet.Addr{IP: "127.0.0.1", Port: 8080},
			Raddr: psNet.Addr{IP: "192.168.1.1", Port: 22},
		},
		{
			Laddr: psNet.Addr{IP: "127.0.0.1", Port: 9090},
			Raddr: psNet.Addr{IP: "192.168.1.2", Port: 22},
		},
		{
			Laddr: psNet.Addr{IP: "127.0.0.1", Port: 9091},
			Raddr: psNet.Addr{IP: "192.168.1.3", Port: 23},
		},
	}

	tests := []struct {
		name     string
		source   []psNet.ConnectionStat
		filter   []client.PoolInfo
		expected int
	}{
		{
			"All Matched",
			source,
			[]client.PoolInfo{
				{LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22}},
				{LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9090}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 22}},
				{LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9091}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.3"), Port: 23}},
			},
			0,
		},
		{
			"None Matched",
			source,
			[]client.PoolInfo{
				{LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.2"), Port: 8080}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 22}},
			},
			3,
		},
		{
			"Partially Matched",
			source,
			[]client.PoolInfo{
				{LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22}},
				{LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9090}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 22}},
			},
			1,
		},
		{
			"Empty Filter",
			source,
			[]client.PoolInfo{},
			3,
		},
		{
			"Empty Source",
			[]psNet.ConnectionStat{},
			[]client.PoolInfo{
				{LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22}},
			},
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FilterOutMatchingConnections(test.source, test.filter)
			if len(result) != test.expected {
				t.Errorf("For test %v, expected %d but got %d", test.name, test.expected, len(result))
			}
		})
	}
}

func TestSelectedMatchingConnections(t *testing.T) {
	source := []psNet.ConnectionStat{
		{
			Laddr: psNet.Addr{IP: "127.0.0.1", Port: 8080},
			Raddr: psNet.Addr{IP: "192.168.1.1", Port: 22},
		},
		{
			Laddr: psNet.Addr{IP: "127.0.0.1", Port: 9090},
			Raddr: psNet.Addr{IP: "192.168.1.2", Port: 22},
		},
		{
			Laddr: psNet.Addr{IP: "127.0.0.1", Port: 9091},
			Raddr: psNet.Addr{IP: "192.168.1.3", Port: 23},
		},
	}

	tests := []struct {
		name     string
		source   []psNet.ConnectionStat
		filter   []server.ConnectionInfo
		expected int
	}{
		{
			"All Matched",
			source,
			[]server.ConnectionInfo{
				{ID: "1", LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22}},
				{ID: "2", LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9090}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 22}},
				{ID: "3", LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9091}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.3"), Port: 23}},
			},
			3,
		},
		{
			"None Matched",
			source,
			[]server.ConnectionInfo{
				{ID: "4", LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.2"), Port: 8081}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 24}},
			},
			0,
		},
		{
			"Partially Matched",
			source,
			[]server.ConnectionInfo{
				{ID: "5", LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22}},
			},
			1,
		},
		{
			"Empty Filter",
			source,
			[]server.ConnectionInfo{},
			0,
		},
		{
			"Empty Source",
			[]psNet.ConnectionStat{},
			[]server.ConnectionInfo{
				{ID: "6", LocalAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22}},
			},
			0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := SelectedMatchingConnections(test.source, test.filter)
			if len(result) != test.expected {
				t.Errorf("For test %v, expected %d matched connections but got %d", test.name, test.expected, len(result))
			}
		})
	}
}

func TestSimplifyConnections(t *testing.T) {
	tests := []struct {
		name                    string
		inputConns              []psNet.ConnectionStat
		expectedSimplifiedConns []request.SimplifiedConnection
	}{
		{
			name: "Basic Scenario",
			inputConns: []psNet.ConnectionStat{
				{
					Fd:     1,
					Family: 1,
					Type:   1,
					Laddr:  psNet.Addr{IP: "127.0.0.1", Port: 8080},
					Raddr:  psNet.Addr{IP: "192.168.1.1", Port: 22},
					Status: "ESTABLISHED",
					Uids:   []int32{1001},
					Pid:    12345,
				},
				{
					Fd:     2,
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.2", Port: 9090},
					Raddr:  psNet.Addr{IP: "192.168.1.2", Port: 23},
					Status: "LISTEN",
					Uids:   []int32{1002},
					Pid:    12346,
				},
			},
			expectedSimplifiedConns: []request.SimplifiedConnection{
				{
					Family: 1,
					Type:   1,
					Laddr:  psNet.Addr{IP: "127.0.0.1", Port: 8080},
					Raddr:  psNet.Addr{IP: "192.168.1.1", Port: 22},
					Status: "ESTABLISHED",
				},
				{
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.2", Port: 9090},
					Raddr:  psNet.Addr{IP: "192.168.1.2", Port: 23},
					Status: "LISTEN",
				},
			},
		},
		{
			name:                    "Empty Connection List",
			inputConns:              []psNet.ConnectionStat{},
			expectedSimplifiedConns: []request.SimplifiedConnection{},
		},
		{
			name: "Various Connection States",
			inputConns: []psNet.ConnectionStat{
				{
					Fd:     3,
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.3", Port: 9091},
					Raddr:  psNet.Addr{IP: "192.168.1.3", Port: 24},
					Status: "CLOSE_WAIT",
					Uids:   []int32{1003},
					Pid:    12347,
				},
			},
			expectedSimplifiedConns: []request.SimplifiedConnection{
				{
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.3", Port: 9091},
					Raddr:  psNet.Addr{IP: "192.168.1.3", Port: 24},
					Status: "CLOSE_WAIT",
				},
			},
		},
		{
			name: "Connection Without Remote Address",
			inputConns: []psNet.ConnectionStat{
				{
					Fd:     4,
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.4", Port: 9092},
					Status: "LISTEN",
					Uids:   []int32{1004},
					Pid:    12348,
				},
			},
			expectedSimplifiedConns: []request.SimplifiedConnection{
				{
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.4", Port: 9092},
					Status: "LISTEN",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SimplifyConnections(tt.inputConns)
			if !reflect.DeepEqual(result, tt.expectedSimplifiedConns) {
				t.Errorf("Expected %v but got %v", tt.expectedSimplifiedConns, result)
			}
		})
	}
}

func TestSimplifyConnectionsWithID(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string][]psNet.ConnectionStat
		expected []request.SimplifiedConnectionWithID
	}{
		{
			name: "Typical Usage",
			input: map[string][]psNet.ConnectionStat{
				"1": {
					{
						Family: 1,
						Type:   1,
						Laddr:  psNet.Addr{IP: "127.0.0.1", Port: 8080},
						Raddr:  psNet.Addr{IP: "192.168.1.1", Port: 22},
						Status: "ESTABLISHED",
					},
				},
				"2": {
					{
						Family: 2,
						Type:   2,
						Laddr:  psNet.Addr{IP: "127.0.0.2", Port: 9090},
						Raddr:  psNet.Addr{IP: "192.168.1.2", Port: 23},
						Status: "LISTEN",
					},
				},
			},
			expected: []request.SimplifiedConnectionWithID{
				{
					ID:     "1",
					Family: 1,
					Type:   1,
					Laddr:  psNet.Addr{IP: "127.0.0.1", Port: 8080},
					Raddr:  psNet.Addr{IP: "192.168.1.1", Port: 22},
					Status: "ESTABLISHED",
				},
				{
					ID:     "2",
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.2", Port: 9090},
					Raddr:  psNet.Addr{IP: "192.168.1.2", Port: 23},
					Status: "LISTEN",
				},
			},
		},
		{
			name:     "Empty Input",
			input:    map[string][]psNet.ConnectionStat{},
			expected: []request.SimplifiedConnectionWithID{},
		},
		{
			name: "Multiple Connections for Same ID",
			input: map[string][]psNet.ConnectionStat{
				"3": {
					{
						Family: 1,
						Type:   1,
						Laddr:  psNet.Addr{IP: "127.0.0.1", Port: 8081},
						Raddr:  psNet.Addr{IP: "192.168.1.3", Port: 24},
						Status: "ESTABLISHED",
					},
					{
						Family: 2,
						Type:   2,
						Laddr:  psNet.Addr{IP: "127.0.0.2", Port: 9091},
						Raddr:  psNet.Addr{IP: "192.168.1.4", Port: 25},
						Status: "LISTEN",
					},
				},
			},
			expected: []request.SimplifiedConnectionWithID{
				{
					ID:     "3",
					Family: 1,
					Type:   1,
					Laddr:  psNet.Addr{IP: "127.0.0.1", Port: 8081},
					Raddr:  psNet.Addr{IP: "192.168.1.3", Port: 24},
					Status: "ESTABLISHED",
				},
				{
					ID:     "3",
					Family: 2,
					Type:   2,
					Laddr:  psNet.Addr{IP: "127.0.0.2", Port: 9091},
					Raddr:  psNet.Addr{IP: "192.168.1.4", Port: 25},
					Status: "LISTEN",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SimplifyConnectionsWithID(tt.input)
			if !slicesEqual(result, tt.expected) {
				t.Errorf("For %s, expected %v but got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func contains(slice []request.SimplifiedConnectionWithID, elem request.SimplifiedConnectionWithID) bool {
	for _, e := range slice {
		if reflect.DeepEqual(e, elem) {
			return true
		}
	}
	return false
}

func slicesEqual(slice1, slice2 []request.SimplifiedConnectionWithID) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for _, e := range slice1 {
		if !contains(slice2, e) {
			return false
		}
	}
	for _, e := range slice2 {
		if !contains(slice1, e) {
			return false
		}
	}
	return true
}

func TestSwitchToPoolInfo(t *testing.T) {
	tests := []struct {
		name   string
		conns  []server.ConnectionInfo
		result []client.PoolInfo
	}{
		{
			name:   "Empty Connections",
			conns:  []server.ConnectionInfo{},
			result: []client.PoolInfo{},
		},
		{
			name: "Single Connection",
			conns: []server.ConnectionInfo{
				{
					ID:         "id1",
					LocalAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
					RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22},
				},
			},
			result: []client.PoolInfo{
				{
					LocalAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
					RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22},
				},
			},
		},
		{
			name: "Multiple Connections",
			conns: []server.ConnectionInfo{
				{
					LocalAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
					RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22},
				},
				{
					ID:         "id2",
					LocalAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.2"), Port: 9090},
					RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 23},
				},
			},
			result: []client.PoolInfo{
				{
					LocalAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080},
					RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.1"), Port: 22},
				},
				{
					LocalAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.2"), Port: 9090},
					RemoteAddr: &net.TCPAddr{IP: net.ParseIP("192.168.1.2"), Port: 23},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SwitchToPoolInfo(tt.conns)
			if !reflect.DeepEqual(result, tt.result) {
				t.Errorf("Expected %v but got %v", tt.result, result)
			}
		})
	}
}

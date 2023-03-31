// Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	connection "github.com/isrc-cas/gt/conn"
	"github.com/isrc-cas/gt/predef"
)

type client struct {
	id           string
	tunnels      map[*conn]struct{}
	tunnelsRWMtx sync.RWMutex
	taskIDSeed   uint32
	closeOnce    sync.Once

	tcps                   []tcp
	associatedTCPListeners sync.Map // key: serverIndex value: net.Listener

	speedMutex    sync.Mutex
	speedNum      uint32
	uploadCount   uint32
	downloadCount uint32

	connections uint32

	hostPrefixMap map[string]uint16 // key: hostPrefix value: serverIndex
	host          host
}

func newClient() interface{} {
	return &client{}
}

// 这一步不在 newClient() 中进行，因为 newClient() 时有锁的存在
func (c *client) init(id string, u user) {
	c.tunnelsRWMtx.Lock()
	c.id = id
	c.host = u.Host
	c.hostPrefixMap = make(map[string]uint16)
	c.tunnels = make(map[*conn]struct{})
	c.tunnelsRWMtx.Unlock()

	c.tcps = u.TCPs

	c.speedMutex.Lock()
	c.speedNum = u.Speed
	c.speedMutex.Unlock()

	c.connections = u.Connections
}

func (c *client) getHostPrefix(hostPrefix string) (uint16, bool) {
	serviceIndex, ok := c.hostPrefixMap[hostPrefix]
	return serviceIndex, ok
}

func (c *client) addHostPrefix(hostPrefix string, serviceIndex uint16) (err error) {
	// 数量限制
	if atomic.LoadUint32(&c.host.usedHost) >= *c.host.Number {
		return connection.ErrHostNumberLimited
	}
	atomic.AddUint32(&c.host.usedHost, 1)

	c.hostPrefixMap[hostPrefix] = serviceIndex
	return
}

func (c *client) process(task *conn) (err error) {
	taskID := atomic.AddUint32(&c.taskIDSeed, 1)
	if taskID >= connection.PreservedSignal {
		atomic.StoreUint32(&c.taskIDSeed, 1)
		taskID = 1
	}
	var tunnel *conn
	for i := 0; i < 3; i++ {
		tunnel = c.getTunnel()
		if tunnel != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if tunnel == nil {
		return ErrNoTunnelExists
	}
	tunnel.process(taskID, task, c)
	return nil
}

func (c *client) canAddTunnel() (ok bool) {
	c.tunnelsRWMtx.RLock()
	defer c.tunnelsRWMtx.RUnlock()
	return uint32(len(c.tunnels)) < c.connections
}

func (c *client) addTunnel(conn *conn) (ok bool) {
	c.tunnelsRWMtx.Lock()
	defer c.tunnelsRWMtx.Unlock()

	if c.tunnels == nil {
		return false
	}
	c.tunnels[conn] = struct{}{}
	return true
}

func (c *client) removeTunnel(tunnel *conn) {
	c.tunnelsRWMtx.Lock()
	if _, ok := c.tunnels[tunnel]; ok {
		delete(c.tunnels, tunnel)
		if len(c.tunnels) < 1 {
			c.tunnels = nil
			tunnel.server.removeClient(c.id)
			for hostPrefix := range c.hostPrefixMap {
				tunnel.Logger.Info().Msgf("remove associated host prefix: %v", hostPrefix)
				tunnel.server.removeHostPrefix(hostPrefix)
			}
		}
	}
	c.tunnelsRWMtx.Unlock()
}

func (c *client) getTunnel() (conn *conn) {
	c.tunnelsRWMtx.RLock()
	defer c.tunnelsRWMtx.RUnlock()
	if len(c.tunnels) == 1 {
		for t := range c.tunnels {
			conn = t
			conn.TasksCount.Add(1)
			return
		}
	}
	var min uint32
	for t := range c.tunnels {
		count := t.TasksCount.Load()
		if count == 0 {
			conn = t
			conn.TasksCount.Add(1)
			return
		}
		if min > count || conn == nil {
			min = count
			conn = t
		}
	}
	if conn != nil {
		conn.TasksCount.Add(1)
	}
	return
}

func (c *client) close() {
	c.closeOnce.Do(func() {
		c.tunnelsRWMtx.Lock()
		for t := range c.tunnels {
			t.SendForceCloseSignal()
			t.Close()
		}
		c.tunnelsRWMtx.Unlock()

		c.associatedTCPListeners.Range(func(key, value interface{}) bool {
			_ = value.(net.Listener).Close()
			return true
		})
	})
}

func (c *client) shutdown() {
	c.tunnelsRWMtx.Lock()
	for t := range c.tunnels {
		t.Shutdown()
	}
	c.tunnelsRWMtx.Unlock()

	c.associatedTCPListeners.Range(func(key, value interface{}) bool {
		_ = value.(net.Listener).Close()
		return true
	})
}

func (c *client) openTCPPort(serviceIndex uint16, tcpPort uint16, random bool, tunnel *conn) (openedTCPPort uint16, err error) {
	if len(c.tcps) == 0 {
		err = errors.New("no permission to open tcp port")
		return
	}

	for i := 0; i < len(c.tcps); i++ {
		if tcpPort < c.tcps[i].PortRange.Min || tcpPort > c.tcps[i].PortRange.Max {
			continue
		}

		err = c.openSpecifiedTCPPort(serviceIndex, tcpPort, tunnel, &c.tcps[i])
		if err != nil {
			if predef.Debug {
				tunnel.Logger.Debug().Err(err).Msg("failed to open tcp port")
			}
			continue
		}
		openedTCPPort = tcpPort
		return
	}

	if !random {
		err = errors.New("user disable random tcp port when specified tcp port failed to open")
		return
	}
	for i := 0; i < len(c.tcps); i++ {
		for openedTCPPort = uint16(c.tcps[i].PortRange.Min); openedTCPPort <= uint16(c.tcps[i].PortRange.Max); openedTCPPort++ {
			err = c.openSpecifiedTCPPort(serviceIndex, openedTCPPort, tunnel, &c.tcps[i])
			if err != nil {
				if predef.Debug {
					tunnel.Logger.Debug().Err(err).Msg("failed to open tcp port")
				}
				continue
			}
			return
		}
	}
	err = errors.New("the number of the tcp ports has reached the upper limit")
	return
}

func (c *client) openSpecifiedTCPPort(serviceIndex uint16, tcpPort uint16, tunnel *conn, tcp *tcp) error {
	listener, err := tcp.openTCPPort(tcpPort)
	if err != nil {
		return err
	}
	tunnel.Logger.Info().Msgf("tcp port %v opened", tcpPort)

	c.associatedTCPListeners.Store(serviceIndex, listener)

	// 启动 goroutine 处理 tcp 连接
	go tunnel.server.acceptLoop(listener, func(conn *conn) {
		conn.serviceIndex = serviceIndex
		conn.handle(func() bool {
			err = c.process(conn)
			if err != nil {
				conn.Logger.Error().Err(err).Msg("openSpecifiedTCPPort")
				return false
			}
			return true
		})
	})

	return nil
}

func (c *client) needOpenTCPPort(serverIndex uint16) bool {
	_, ok := c.associatedTCPListeners.Load(serverIndex)
	return !ok
}

func (c *client) needSpeedLimit() (ok bool) {
	return c.speedNum > 0
}

func (c *client) speedLimit(bufLen uint32, isUpload bool) {
	c.speedMutex.Lock()
	defer c.speedMutex.Unlock()

	// 上下行分开限速，限制的速度都是 Speed
	count := &c.downloadCount
	if isUpload {
		count = &c.uploadCount
	}

	// 乐观思想，假设数据包可以立即到达客户端，仅控制服务端的发包速度
	*count += bufLen
	if *count < c.speedNum {
		return
	}
	sleepSeconds := *count / c.speedNum
	*count -= sleepSeconds * c.speedNum
	time.Sleep(time.Duration(sleepSeconds) * time.Second)
}

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
	"fmt"
	"github.com/libp2p/go-reuseport"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	connection "github.com/isrc-cas/gt/conn"
	ssync "github.com/isrc-cas/gt/server/sync"
	"github.com/rs/zerolog"
)

type client struct {
	id           string
	logger       zerolog.Logger
	tunnels      map[*conn]struct{}
	tunnelsRWMtx sync.RWMutex
	taskIDSeed   uint32
	closeOnce    sync.Once

	portsManager *portsManager
	tcpListeners ssync.Map // key: serverIndex value: net.Listener

	speedMutex    sync.Mutex
	speedNum      uint32
	uploadCount   uint32
	downloadCount uint32

	connections uint32

	host host

	checksumBlacklist     *lru.Cache[[32]byte, any]
	lastProcessedChecksum [32]byte
}

func newClient() interface{} {
	return &client{}
}

// 这一步不在 newClient() 中进行，因为 newClient() 时有锁的存在
func (c *client) init(id string, u user, s *Server) {
	c.host = u.Host
	c.portsManager = u.portsManager
	c.speedNum = u.Speed
	c.connections = u.Connections
	c.checksumBlacklist, _ = lru.New[[32]byte, any](3)
	c.logger = s.Logger.With().
		Str("client", id).
		Logger()

	c.tunnelsRWMtx.Lock()
	c.id = id
	c.tunnels = make(map[*conn]struct{})
	c.tunnelsRWMtx.Unlock()
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
		time.Sleep(100 * time.Millisecond)
	}
	if tunnel == nil {
		return ErrNoTunnelExists
	}
	tunnel.process(taskID, task, c)
	return nil
}

type hostPrefixChanges struct {
	oldServiceIndex hostPrefixOption
	serviceIndex    hostPrefixOption
	remove          bool
}

func (c *client) addTunnel(t *conn, reload bool, o options) (ok bool, err error) {
	c.tunnelsRWMtx.Lock()
	defer c.tunnelsRWMtx.Unlock()
	t.Logger = t.Logger.With().Str("client", c.id).Logger()

	if uint32(len(c.tunnels)) >= c.connections {
		err = connection.ErrReachedMaxConnections
		if e := t.SendErrorSignalReachedMaxConnections(); e != nil {
			t.Logger.Error().Err(e).Msg("failed to SendErrorSignalReachedMaxConnections")
		}
		return
	}

	if c.tunnels == nil {
		return
	}

	if c.checksumBlacklist.Contains(o.configChecksum) {
		t.Logger.Info().
			Hex("checksum", o.configChecksum[:]).
			Msg("old config connecting")
		conflict := false
		for tunnel := range c.tunnels {
			tunnel.Logger.Info().
				Hex("newChecksum", o.configChecksum[:]).
				Hex("oldChecksum", tunnel.configChecksum[:]).
				Msg("connection in pool")
			if o.configChecksum != tunnel.configChecksum {
				tunnel.Logger.Info().
					Hex("newChecksum", o.configChecksum[:]).
					Hex("oldChecksum", tunnel.configChecksum[:]).
					Msg("connection in pool is different")
				conflict = true
			}
		}
		if conflict {
			err = connection.ErrDifferentConfigClientConnected
			if e := t.SendErrorSignalDifferentConfigClientConnected(); e != nil {
				t.Logger.Error().Err(e).Msg("failed to SendErrorSignalDifferentConfigClientConnected")
			}
			return
		}
	}

	if c.lastProcessedChecksum == o.configChecksum {
		t.ids = o.ids
		t.configChecksum = o.configChecksum
		c.tunnels[t] = struct{}{}
		ok = true
		t.Logger.Info().Msg("checksum is same with last processed checksum")
		return
	}

	c.tunnels[t] = struct{}{}

	err = t.processHostPrefixes(o, c)
	if err != nil {
		return
	}

	if c.portsManager != nil {
		err = t.processTCPOptions(o, c)
		if err != nil {
			return
		}
	} else {
		if len(o.ports) > 0 {
			err = connection.ErrTCPNumberLimited
			if e := t.SendErrorSignalTCPNumberLimited(); e != nil {
				t.Logger.Error().Err(e).Msg("failed to SendErrorSignalTCPNumberLimited")
			}
			return
		}
	}

	c.lastProcessedChecksum = o.configChecksum

	if reload {
		ok = true
		return
	}
	// remove old tunnels that checksum is old and its host prefixes does not need anymore
	// for new tunnel connection from another client process
	var oldTunnels []*conn
	for tunnel := range c.tunnels {
		if o.configChecksum != tunnel.configChecksum {
			tunnel.Logger.Info().
				Hex("newChecksum", o.configChecksum[:]).
				Hex("oldChecksum", tunnel.configChecksum[:]).
				Msg("old connection closing")
			oldTunnels = append(oldTunnels, tunnel)
		}
	}
	if len(oldTunnels) > 0 {
		// remove host prefixes from old connections
		// checksum -> id -> hostPrefixChanges
		oldIds := make(map[[32]byte]map[string]hostPrefixChanges)
		for _, tunnel := range oldTunnels {
			delete(c.tunnels, tunnel)
			tunnel.SendCloseSignal()

			m, ok := oldIds[tunnel.configChecksum]
			if ok {
				for id, osi := range tunnel.ids {
					if si, e := o.ids[id]; !e {
						m[id] = hostPrefixChanges{remove: true, oldServiceIndex: osi}
					} else if osi != si {
						m[id] = hostPrefixChanges{oldServiceIndex: osi, serviceIndex: si}
					}
				}
			} else {
				m = make(map[string]hostPrefixChanges)
				for id, osi := range tunnel.ids {
					if si, e := o.ids[id]; !e {
						m[id] = hostPrefixChanges{remove: true, oldServiceIndex: osi}
					} else if osi != si {
						m[id] = hostPrefixChanges{oldServiceIndex: osi, serviceIndex: si}
					}
				}
				oldIds[tunnel.configChecksum] = m
			}
		}
		for checksum, ids := range oldIds {
			c.checksumBlacklist.Add(checksum, struct{}{})
			time.AfterFunc(60*time.Second, func() {
				c.checksumBlacklist.RemoveOldest()
			})
			t.Logger.Info().
				Str("id", c.id).
				Hex("oldChecksum", checksum[:]).
				Hex("newChecksum", o.configChecksum[:]).
				Msg("added old checksum to blacklist")
			for id, changes := range ids {
				if changes.remove || changes.oldServiceIndex.tls != changes.serviceIndex.tls {
					t.server.removeHostPrefix(id, changes.oldServiceIndex.tls)
					t.Logger.Info().
						Str("id", c.id).
						Hex("oldChecksum", checksum[:]).
						Hex("newChecksum", o.configChecksum[:]).
						Str("prefix", id).
						Str("oldServiceIndex", changes.oldServiceIndex.String()).
						Msg("removed associated host prefix")
				} else {
					t.server.storeHostPrefix(id, changes.serviceIndex.tls,
						clientWithServiceIndex{
							client:       c,
							serviceIndex: changes.serviceIndex.serviceIndex,
						})
					t.Logger.Info().
						Str("id", c.id).
						Hex("oldChecksum", checksum[:]).
						Hex("newChecksum", o.configChecksum[:]).
						Str("prefix", id).
						Str("oldServiceIndex", changes.oldServiceIndex.String()).
						Str("newServiceIndex", changes.serviceIndex.String()).
						Msg("updated associated host prefix")
				}
			}
		}
	}
	ok = true
	return
}

func (c *client) removeTunnel(tunnel *conn) {
	c.tunnelsRWMtx.Lock()
	defer c.tunnelsRWMtx.Unlock()
	if _, ok := c.tunnels[tunnel]; ok {
		delete(c.tunnels, tunnel)
		if len(c.tunnels) < 1 {
			c.tunnels = nil
			tunnel.server.removeClient(c.id)
			for hostPrefix, o := range tunnel.ids {
				tunnel.Logger.Info().
					Hex("checksum", tunnel.configChecksum[:]).
					Str("prefix", hostPrefix).
					Str("serviceIndex", o.String()).
					Msg("remove associated host prefix")
				tunnel.server.removeHostPrefix(hostPrefix, o.tls)
			}
			c.closeTCPListeners()
		}
	}
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
		c.closeTCPListeners()
		c.tunnelsRWMtx.Unlock()
	})
}

func (c *client) shutdown() {
	c.tunnelsRWMtx.Lock()
	for t := range c.tunnels {
		t.Shutdown()
	}
	c.closeTCPListeners()
	c.tunnelsRWMtx.Unlock()
}

func (c *client) closeTCPListeners() {
	c.tcpListeners.Range(func(key, value interface{}) bool {
		l, ok := value.(*tcpListener)
		if ok {
			if l.l != nil {
				port := uint16(l.l.Addr().(*net.TCPAddr).Port)
				c.logger.Info().
					Interface("serviceIndex", key).
					Uint16("port", port).
					Msg("close associated tcp listener")
				c.portsManager.portsMtx.Lock()
				c.portsManager.ports[port] = struct{}{}
				c.portsManager.portsMtx.Unlock()
				_ = l.l.Close()
			}
		}
		return true
	})
}

func (c *client) deleteTCPListener(si uint16) {
	value, loaded := c.tcpListeners.LoadAndDelete(si)
	if loaded {
		l, ok := value.(*tcpListener)
		if ok {
			if l.l != nil {
				port := uint16(l.l.Addr().(*net.TCPAddr).Port)
				c.logger.Info().
					Uint16("serviceIndex", si).
					Uint16("port", port).
					Msg("close associated tcp listener")
				c.portsManager.portsMtx.Lock()
				c.portsManager.ports[port] = struct{}{}
				c.portsManager.portsMtx.Unlock()
				_ = l.l.Close()
			}
		}
	}
}

type portsManager struct {
	ports    map[uint16]struct{}
	portsMtx sync.Mutex
}

type tcpListener struct {
	l    net.Listener
	port openTCPOption
}

func (c *client) openTCPPort(serviceIndex uint16, l *tcpListener, tunnel *conn) (openedTCPPort uint16, err error) {
	tcpPort := l.port.port
	random := l.port.random

	c.portsManager.portsMtx.Lock()
	defer c.portsManager.portsMtx.Unlock()
	if len(c.portsManager.ports) == 0 {
		err = errors.New("no available tcp port")
		return
	}

	if _, ok := c.portsManager.ports[tcpPort]; ok {
		err = c.openSpecifiedTCPPort(serviceIndex, l, tcpPort, tunnel)
		if err == nil {
			openedTCPPort = tcpPort
			delete(c.portsManager.ports, tcpPort)
			return
		}
	}
	tunnel.Logger.Warn().Err(err).Uint16("port", tcpPort).Msg("failed to open the tcp port user asked")
	if !random {
		err = fmt.Errorf("user disable random tcp port when %w", err)
		return
	}

	retry := 0
	for tcpPort := range c.portsManager.ports {
		err = c.openSpecifiedTCPPort(serviceIndex, l, tcpPort, tunnel)
		if err == nil {
			openedTCPPort = tcpPort
			delete(c.portsManager.ports, tcpPort)
			return
		}
		tunnel.Logger.Warn().Err(err).Msg("failed to open tcp port")
		retry++
		if retry >= 3 {
			break
		}
	}
	err = errors.New("failed to open random tcp port")
	return
}

func (c *client) openSpecifiedTCPPort(serviceIndex uint16, l *tcpListener, tcpPort uint16, tunnel *conn) error {
	listener, err := reuseport.Listen("tcp", ":"+strconv.Itoa(int(tcpPort)))
	if err != nil {
		return err
	}
	tunnel.Logger.Info().Uint16("port", tcpPort).Msg("tcp port opened")
	l.l = listener

	// 启动 goroutine 处理 tcp 连接
	go tunnel.server.acceptLoop(listener, func(conn *conn) {
		defer func() {
			tunnel.Logger.Info().Uint16("serviceIndex", serviceIndex).Uint16("tcpPort", tcpPort).Msg("tcp forward stop")
		}()
		tunnel.Logger.Info().Uint16("serviceIndex", serviceIndex).Uint16("tcpPort", tcpPort).Msg("tcp forward start")
		conn.serviceIndex = serviceIndex
		conn.handle(func() bool {
			err = c.process(conn)
			if err != nil {
				conn.Logger.Error().Err(err).Msg("tcp handle")
				return false
			}
			return true
		})
	})

	return nil
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

type ConnectionInfo struct {
	ID         string
	LocalAddr  net.Addr
	RemoteAddr net.Addr
}

func (c *client) GetConnectionInfo() (info []ConnectionInfo) {
	c.tunnelsRWMtx.RLock()
	defer c.tunnelsRWMtx.RUnlock()
	for conn := range c.tunnels {
		info = append(info, ConnectionInfo{
			ID:         c.id,
			RemoteAddr: conn.RemoteAddr(),
			LocalAddr:  conn.LocalAddr(),
		})
	}
	return
}

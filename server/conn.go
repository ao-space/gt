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
	"bytes"
	"errors"
	"io"
	"net"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/isrc-cas/gt/bufio"
	connection "github.com/isrc-cas/gt/conn"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
)

var (
	// ErrInvalidID is an error returned when id is invalid
	ErrInvalidID = errors.New("invalid id")
	// ErrIDNotFound is an error returned when id is not in the url
	ErrIDNotFound = errors.New("id not found")
	// ErrNoTunnelExists is an error returned when there is no tunnel
	ErrNoTunnelExists = errors.New("no tunnel exists")
)

type conn struct {
	connection.Connection
	server       *Server
	tasks        map[uint32]*conn
	tasksRWMtx   sync.RWMutex
	serviceIndex uint16 // 0 表示客户端只有一个 Local，使用 predef.Data，兼容老客户端；大于 0 使用 predef.ServicesData
}

func newConn(c net.Conn, s *Server) *conn {
	nc := &conn{
		Connection: connection.Connection{
			Conn:         c,
			WriteTimeout: s.config.Timeout,
		},
		server: s,
		tasks:  make(map[uint32]*conn, 100),
	}
	nc.Logger = s.Logger.With().
		Str("serverConn", strconv.FormatUint(uint64(uintptr(unsafe.Pointer(nc))), 16)).
		Str("ip", c.RemoteAddr().String()).
		Logger()
	nc.Logger.Info().Msg("accepted")
	return nc
}

func (c *conn) addTask(taskID uint32, conn *conn) {
	c.tasksRWMtx.Lock()
	c.tasks[taskID] = conn
	c.tasksRWMtx.Unlock()
}

func (c *conn) removeTask(id uint32) {
	c.tasksRWMtx.Lock()
	delete(c.tasks, id)
	c.tasksRWMtx.Unlock()
}

func (c *conn) getTask(taskID uint32) (conn *conn, ok bool) {
	c.tasksRWMtx.RLock()
	conn, ok = c.tasks[taskID]
	c.tasksRWMtx.RUnlock()
	return
}

func (c *conn) handle(handleFunc func() bool) {
	startTime := time.Now()
	reader := pool.GetReader(c.Conn)
	c.Reader = reader
	handled := false
	defer func() {
		c.Close()
		pool.PutReader(reader)
		endTime := time.Now()
		if !predef.Debug {
			if e := recover(); e != nil {
				c.Logger.Error().Msgf("recovered panic: %#v\n%s", e, debug.Stack())
			}
		}
		c.Logger.Info().Dur("cost", endTime.Sub(startTime)).Msg("closed")
		if !handled {
			atomic.AddUint64(&c.server.failed, 1)
		}
	}()
	if c.server.config.Timeout > 0 {
		dl := startTime.Add(c.server.config.Timeout)
		err := c.SetReadDeadline(dl)
		if err != nil {
			c.Logger.Debug().Err(err).Msg("handle set deadline failed")
			return
		}
	}

	version, err := reader.Peek(2)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			c.Logger.Warn().Err(err).Msg("failed to peek version field")
		}
		return
	}
	if version[0] == predef.MagicNumber {
		switch version[1] {
		case 0x01:
			_, err = reader.Discard(2)
			if err != nil {
				c.Logger.Warn().Err(err).Msg("failed to discard version field")
				return
			}

			// 判断 IP 是否处于被限制状态
			remoteAddr, ok := c.RemoteAddr().(*net.TCPAddr)
			if !ok {
				c.Logger.Warn().Msg("conn is not tcp conn")
				return
			}
			remoteIP := remoteAddr.IP.String()
			c.server.reconnectRWMutex.RLock()
			reconnectTimes := c.server.reconnect[remoteIP]
			c.server.reconnectRWMutex.RUnlock()
			if reconnectTimes > c.server.config.ReconnectTimes {
				c.Logger.Warn().Msgf("IP: '%v' is limited", remoteIP)
				return
			}

			// 不能将 reconnectTimes 传参，多线程环境下这个值应该实时获取
			handled = c.handleTunnel(remoteIP)
			return
		}
	}
	handled = handleFunc()
}

func (c *conn) handleSNI() (handled bool) {
	var err error
	var host []byte
	defer func() {
		if err != nil {
			c.Logger.Error().Bytes("host", host).Err(err).Msg("handleSNI")
		}
		if !predef.Debug {
			if e := recover(); e != nil {
				c.Logger.Error().Msgf("recovered panic: %#v\n%s", e, debug.Stack())
			}
		}
		atomic.AddUint64(&c.server.served, 1)
		handled = true
	}()

	host, err = peekTLSHost(c.Reader)
	if err != nil {
		return
	}
	if len(host) < 1 {
		err = ErrInvalidHTTPProtocol
		return
	}
	id, err := parseIDFromHost(host)
	if err != nil {
		return
	}
	if len(id) < predef.MinIDSize {
		err = ErrInvalidID
		return
	}
	client, ok := c.server.getHostPrefix(string(id))
	if ok {
		c.serviceIndex = client.hostPrefixMap[string(id)]
		err = client.process(c)
	} else {
		err = ErrIDNotFound
	}

	return
}

func (c *conn) handleHTTP() (handled bool) {
	var err error
	var host []byte
	var id []byte
	defer func() {
		if err != nil {
			c.Logger.Error().Bytes("host", host).Bytes("id", id).Err(err).Msg("handleHTTP")
		}
		if !predef.Debug {
			if e := recover(); e != nil {
				c.Logger.Error().Msgf("recovered panic: %#v\n%s", e, debug.Stack())
			}
		}
		atomic.AddUint64(&c.server.served, 1)
		handled = true
	}()
	if c.server.config.HTTPMUXHeader == "Host" {
		host, err = peekHost(c.Reader)
		if err != nil {
			return
		}
		if len(host) < 1 {
			err = ErrInvalidHTTPProtocol
			return
		}
		id, err = parseIDFromHost(host)
		if err != nil {
			return
		}
	} else {
		id, err = peekHeader(c.Reader, c.server.config.HTTPMUXHeader+":")
		if err != nil {
			return
		}
	}
	if len(id) < predef.MinIDSize {
		err = ErrInvalidID
		return
	}
	client, ok := c.server.getHostPrefix(string(id))
	if ok {
		c.serviceIndex = client.hostPrefixMap[string(id)]
		err = client.process(c)
	} else {
		err = ErrIDNotFound
	}
	return
}

func (c *conn) handleTunnel(remoteIP string) (handled bool) {
	reader := c.Reader

	var err error
	defer func() {
		switch err {
		case connection.ErrFailedToOpenTCPPort:
			err = c.SendErrorSignalFailedToOpenTCPPort()
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to send FailedToOpenTCPPort signal")
				return
			}
		}
	}()

	// id
	idLen, err := reader.ReadByte()
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to read id len")
		return
	}
	if idLen < predef.MinIDSize && idLen > predef.MaxIDSize {
		c.Logger.Error().Err(err).Msg("invalid id len")
		return
	}
	id, err := reader.Peek(int(idLen))
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to read id")
		return
	}
	idStr := string(id)
	_, err = reader.Discard(int(idLen))
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to discard id")
		return
	}

	// secret
	secretLen, err := reader.ReadByte()
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to read secret len")
		return
	}
	if secretLen < predef.MinSecretSize && secretLen > predef.MaxSecretSize {
		c.Logger.Error().Err(err).Msg("invalid secret len")
		return
	}
	secret, err := reader.Peek(int(secretLen))
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to read secret")
		return
	}
	secretStr := string(secret)
	_, err = reader.Discard(int(secretLen))
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to discard secret")
		return
	}

	// 验证 id secret
	if err := c.server.authUser(idStr, secretStr); err != nil {
		// 使用局部锁而不是全局锁可以明显提高并发性能，但少数情况下会降低限制效果
		c.server.reconnectRWMutex.Lock()
		reconnectTimes := c.server.reconnect[remoteIP]
		reconnectTimes++
		c.server.reconnect[remoteIP] = reconnectTimes
		c.server.reconnectRWMutex.Unlock()

		if reconnectTimes > c.server.config.ReconnectTimes {
			go func() {
				// ReconnectDuration 为 0 表示不进行限制解除
				if c.server.config.ReconnectDuration == 0 {
					return
				}
				time.Sleep(c.server.config.ReconnectDuration)
				c.server.reconnectRWMutex.Lock()
				c.server.reconnect[remoteIP] = 0
				c.server.reconnectRWMutex.Unlock()
				c.Logger.Debug().Msgf("unlimit IP: '%v'", remoteIP)
			}()
		}

		e := c.SendErrorSignalInvalidIDAndSecret()
		c.Logger.Debug().Err(err).AnErr("respErr", e).Msg("invalid id and secret")
		return
	}

	// 获取或创建 client
	var cli *client
	var ok bool
	var exists bool
	for i := 0; i < 5; i++ {
		cli, exists = c.server.getOrCreateClient(idStr, newClient)
		if !exists {
			// 默认使用全局的配置，如果用户单独配置了则使用用户配置的
			value, ok := c.server.users.Load(idStr)
			if !ok {
				value = user{
					// Secret:      secretStr, // 这个字段暂时用不到
					TCPs:        append([]tcp(nil), c.server.config.TCPs...),
					Speed:       c.server.config.Speed,
					Connections: c.server.config.Connections,
					Host:        c.server.config.Host,
				}
			}
			cli.init(idStr, value.(user))
		}

		if !cli.canAddTunnel() {
			e := c.SendErrorSignalReachedTheMaxConnections()
			c.Logger.Error().AnErr("respErr", e).Msg("client has reached the max number of tunnel connections")
			return
		}
		ok = cli.addTunnel(c)
		if ok {
			break
		}
	}
	if !ok || cli == nil {
		c.Logger.Error().Msg("failed to create client")
		return
	}
	defer cli.removeTunnel(c)

	// options
	var serviceIndex uint16
	for leftOptions := 1; leftOptions > 0; leftOptions-- {
		optionFirst, err := reader.ReadByte()
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to read option")
			return
		}
		optionLeftLen := optionFirst >> 6
		optionLeft, err := reader.Peek(int(optionLeftLen))
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to read option left")
			return
		}
		option := append([]byte{optionFirst}, optionLeft...)
		_, err = reader.Discard(int(optionLeftLen))
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to discard option left")
			return
		}
		// 按照正常的逻辑，多个 tunnel 只有第一个 tunnel 的 option 需要执行，
		// 但是这样需要加上一个层次比较高的锁，会影响并发性能，
		// 所以重复的 option 需要各个 option 自己处理
		switch {
		case bytes.Equal(option, predef.IDAsHostPrefix):
			// 如果这个 host prefix 已经在这个 client 的中添加过了，那么跳过
			if _, ok := cli.getHostPrefix(idStr); ok {
				break
			}

			_, ok = c.server.getHostPrefix(idStr)
			if ok {
				c.Logger.Error().Err(err).Msg("failed to add host prefix")
				err = c.SendErrorSignalHostConflict()
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to send error signal")
				}
				return
			}
			c.server.addHostPrefix(idStr, cli)
			_, ok := cli.getHostPrefix(idStr)
			if ok {
				c.Logger.Error().Err(connection.ErrHostConflict).Msg("failed to add host prefix")
				err = c.SendErrorSignalHostConflict()
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to send error signal")
				}
				return
			}
			err = cli.addHostPrefix(idStr, serviceIndex)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to add host prefix")
				if err == connection.ErrHostNumberLimited {
					err = c.SendErrorSignalHostNumberLimited()
					if err != nil {
						c.Logger.Error().Err(err).Msg("failed to send error signal")
					}
				}
				return
			}
		case bytes.Equal(option, predef.OpenTCPPort):
			random, err := reader.ReadByte()
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to read random byte")
				return
			}
			var peekBytes []byte
			peekBytes, err = reader.Peek(2)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to peek tcp port range")
				return
			}
			tcpPort := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
			_, err = reader.Discard(2)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to discard tcp port range")
				return
			}

			// 将 serverIndex 与 tcp 端口绑定，如果这个 serverIndex 已经绑定了，那么跳过
			if !cli.needOpenTCPPort(serviceIndex) {
				break
			}
			openedPort, err := cli.openTCPPort(serviceIndex, tcpPort, random != 0, c)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to open tcp port")

				err = c.SendErrorSignalFailedToOpenTCPPort()
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to send FailedToOpenTCPPort signal")
				}

				return
			}
			err = c.SendInfoTCPPortOpened(openedPort)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to send Info signal")
				return
			}
		case bytes.Equal(option, predef.OptionAndNextOption):
			leftOptions += 2
			continue // 跳过 serverIndex++
		case bytes.Equal(option, predef.OpenHost):
			hostPrefixLen, err := reader.ReadByte()
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to read host prefix length")
				return
			}
			hostPrefix, err := reader.Peek(int(hostPrefixLen))
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to peek host prefix")
				return
			}
			hostPrefixStr := string(hostPrefix)
			_, err = reader.Discard(int(hostPrefixLen))
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to discard host prefix")
				return
			}

			// 如果这个 host prefix 已经在这个 client 的中添加过了，那么跳过
			if _, ok := cli.getHostPrefix(hostPrefixStr); ok {
				break
			}

			// 检查 host prefix 是否符合要求
			if len(*cli.host.Regex) > 0 {
				matched := false
				for _, regex := range *cli.host.Regex {
					if regex.Match([]byte(hostPrefix)) {
						matched = true
						break
					}
				}
				if !matched {
					c.Logger.Error().Err(connection.ErrHostRegexMismatch).Msg("failed when check host prefix")
					err = c.SendErrorSignalHostRegexMismatch()
					if err != nil {
						c.Logger.Error().Err(err).Msg("failed when check host prefix")
					}
					return
				}
			}

			if *cli.host.WithID {
				hostPrefixStr = cli.id + "-" + hostPrefixStr
			}

			_, ok = c.server.getHostPrefix(hostPrefixStr)
			if ok {
				c.Logger.Error().Err(err).Msg("failed to add host prefix")
				err = c.SendErrorSignalHostConflict()
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to send error signal")
				}
				return
			}
			c.server.addHostPrefix(hostPrefixStr, cli)
			_, ok := cli.getHostPrefix(hostPrefixStr)
			if ok {
				c.Logger.Error().Err(connection.ErrHostConflict).Msg("failed to add host prefix")
				err = c.SendErrorSignalHostConflict()
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to send error signal")
				}
				return
			}
			err = cli.addHostPrefix(hostPrefixStr, serviceIndex)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to add host prefix")
				if err == connection.ErrHostNumberLimited {
					err = c.SendErrorSignalHostNumberLimited()
					if err != nil {
						c.Logger.Error().Err(err).Msg("failed to send error signal")
					}
				}
				return
			}
		default:
			c.Logger.Error().Msgf("invalid option: %v", optionFirst)
			return
		}
		serviceIndex++
	}

	atomic.AddUint64(&c.server.tunneling, 1)
	handled = true
	c.readLoop(cli)
	return
}

func (c *conn) readLoop(cli *client) {
	var err error
	defer func() {
		c.Logger.Debug().Err(err).Msg("readLoop ended")
		c.tasksRWMtx.RLock()
		for _, t := range c.tasks {
			t.Close()
		}
		c.tasksRWMtx.RUnlock()
	}()
	err = c.SendReadySignal()
	if err != nil {
		return
	}
	r := &bufio.LimitedReader{}
	isClosing := false
	for {
		if c.server.config.Timeout > 0 {
			dl := time.Now().Add(c.server.config.Timeout)
			err = c.SetReadDeadline(dl)
			if err != nil {
				return
			}
		}
		var peekBytes []byte
		peekBytes, err = c.Reader.Peek(4)
		if err != nil {
			return
		}
		signal := uint32(peekBytes[3]) | uint32(peekBytes[2])<<8 | uint32(peekBytes[1])<<16 | uint32(peekBytes[0])<<24
		_, err = c.Reader.Discard(4)
		if err != nil {
			return
		}
		switch signal {
		case connection.PingSignal:
			if predef.Debug {
				c.Logger.Trace().Msg("readLoop read ping signal")
			}
			err = c.SendPingSignal()
			if err != nil {
				c.Logger.Debug().Err(err).Msg("readLoop resp ping signal failed")
				return
			}
			continue
		case connection.CloseSignal:
			if predef.Debug {
				c.Logger.Trace().Msg("readLoop read close signal")
			}
			if isClosing {
				return
			}
			isClosing = true
			cli.removeTunnel(c)
			c.SendCloseSignal()
			continue
		}
		taskID := signal
		if predef.Debug {
			c.Logger.Trace().Uint32("taskID", taskID).Msg("readLoop read taskID")
		}
		peekBytes, err = c.Reader.Peek(2)
		if err != nil {
			return
		}
		taskOption := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
		_, err = c.Reader.Discard(2)
		if err != nil {
			return
		}
		task, ok := c.getTask(taskID)
		switch taskOption {
		case predef.Data:
			if predef.Debug {
				c.Logger.Trace().Uint32("taskID", taskID).Msg("read data op")
			}
			peekBytes, err = c.Reader.Peek(4)
			if err != nil {
				return
			}
			l := uint32(peekBytes[3]) | uint32(peekBytes[2])<<8 | uint32(peekBytes[1])<<16 | uint32(peekBytes[0])<<24
			_, err = c.Reader.Discard(4)
			if err != nil {
				return
			}
			if cli.needSpeedLimit() {
				cli.speedLimit(l, true) // 对客户端上行进行限速
			}
			if predef.Debug {
				c.Logger.Trace().Uint32("len", l).Msg("readLoop read len")
			}
			r.Reader = c.Reader
			r.N = int64(l)
			if !ok && r.N > 0 {
				if !predef.Debug {
					_, err = r.Discard(int(r.N))
					if err != nil {
						return
					}
				} else {
					event := c.Logger.Trace().Int64("N", r.N)
					bs, err := io.ReadAll(r)
					event.Uint16("taskOption", taskOption).Hex("content", bs).Err(err).Uint32("taskID", taskID).Msg("orphan resp")
				}
				continue
			}
			_, err = r.WriteTo(task)
			if r.N > 0 {
				if !predef.Debug {
					_, err = r.Discard(int(r.N))
					if err != nil {
						return
					}
				} else {
					event := c.Logger.Trace().Int64("N", r.N)
					bs, err := io.ReadAll(r)
					event.Uint16("taskOption", taskOption).Hex("content", bs).Err(err).Uint32("taskID", taskID).Msg("orphan resp")
				}
			}
			if err != nil {
				switch e := err.(type) {
				case *net.OpError:
					switch e.Op {
					case "write":
						c.Logger.Debug().Err(err).Uint32("taskID", taskID).Msg("remote req resp writer closed")
						continue
					}
				case *bufio.WriteErr:
					c.Logger.Debug().Err(err).Uint32("taskID", taskID).Msg("remote req resp writer closed")
					continue
				}
				return
			}
			if c.server.config.Timeout > 0 && !c.server.config.TimeoutOnUnidirectionalTraffic {
				dl := time.Now().Add(c.server.config.Timeout)
				err = task.SetReadDeadline(dl)
				if err != nil {
					c.Logger.Debug().Err(err).Uint32("taskID", taskID).Msg("update read deadline failed")
				}
			}
		case predef.Close:
			if predef.Debug {
				c.Logger.Trace().Uint32("taskID", taskID).Msg("read close op")
			}
			if ok {
				task.CloseByRemote()
			}
		}
	}
}

func (c *conn) process(taskID uint32, task *conn, cli *client) {
	var rErr error
	var wErr error
	c.addTask(taskID, task)
	buf := pool.BytesPool.Get().([]byte)
	defer func() {
		c.removeTask(taskID)
		if wErr == nil && !task.IsClosingByRemote() {
			buf[4] = byte(predef.Close >> 8)
			buf[5] = byte(predef.Close)
			_, wErr = c.Write(buf[:6])
		}
		pool.BytesPool.Put(buf)
		if rErr != nil || wErr != nil {
			c.Logger.Debug().AnErr("read err", rErr).AnErr("write err", wErr).Uint32("taskID", taskID).Msg("process err")
		}
		if c.TasksCount.Add(^uint32(0)) == 0 && c.IsClosing() {
			c.SendForceCloseSignal()
			c.Close()
		} else if wErr != nil {
			c.Close()
		}
	}()
	buf[0] = byte(taskID >> 24)
	buf[1] = byte(taskID >> 16)
	buf[2] = byte(taskID >> 8)
	buf[3] = byte(taskID)
	bufIndex := 4
	if task.serviceIndex == 0 {
		buf[bufIndex] = byte(predef.Data >> 8)
		bufIndex++
		buf[bufIndex] = byte(predef.Data)
		bufIndex++
	} else {
		buf[bufIndex] = byte(predef.ServicesData >> 8)
		bufIndex++
		buf[bufIndex] = byte(predef.ServicesData)
		bufIndex++
		buf[bufIndex] = byte(task.serviceIndex >> 8)
		bufIndex++
		buf[bufIndex] = byte(task.serviceIndex)
		bufIndex++
	}
	for {
		if c.server.config.Timeout > 0 {
			dl := time.Now().Add(c.server.config.Timeout)
			rErr = task.SetReadDeadline(dl)
			if rErr != nil {
				return
			}
		}
		var l int
		l, rErr = task.Reader.Read(buf[bufIndex+4:])
		if cli.needSpeedLimit() {
			cli.speedLimit(uint32(l), false) // 对客户端下行进行限速
		}
		if l > 0 {
			buf[bufIndex] = byte(l >> 24)
			buf[bufIndex+1] = byte(l >> 16)
			buf[bufIndex+2] = byte(l >> 8)
			buf[bufIndex+3] = byte(l)
			l += bufIndex + 4

			if predef.Debug {
				c.Logger.Trace().Hex("data", buf[:l]).Msg("write")
			}
			_, wErr = c.Write(buf[:l])
			if wErr != nil {
				return
			}
			if c.server.config.Timeout > 0 && !c.server.config.TimeoutOnUnidirectionalTraffic {
				dl := time.Now().Add(c.server.config.Timeout)
				wErr = c.SetReadDeadline(dl)
				if wErr != nil {
					return
				}
			}
		}
		if rErr != nil {
			return
		}
	}
}

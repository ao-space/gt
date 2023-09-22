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
	"crypto/sha256"
	"errors"
	"io"
	"net"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/emirpasic/gods/trees/btree"
	"github.com/emirpasic/gods/utils"
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
	server         *Server
	tasks          map[uint32]*conn
	tasksRWMtx     sync.RWMutex
	serviceIndex   uint16 // 0 表示客户端只有一个 Local，使用 predef.Data，兼容老客户端；大于 0 使用 predef.ServicesData
	ids            hostPrefixOptions
	configChecksum [32]byte
}

func newConn(c net.Conn, s *Server) *conn {
	nc := &conn{
		Connection: connection.Connection{
			Conn:         c,
			WriteTimeout: s.config.Timeout.Duration,
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
	if c.server.config.Timeout.Duration > 0 {
		dl := startTime.Add(c.server.config.Timeout.Duration)
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

			c.Logger = c.Logger.With().Time("tunnel", time.Now()).Logger()

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
			handled = c.handleTunnelLoop(remoteIP)
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
	client, ok := c.server.getTLSHostPrefix(string(id))
	if ok {
		c.serviceIndex = client.serviceIndex
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
		c.serviceIndex = client.serviceIndex
		err = client.process(c)
	} else {
		err = ErrIDNotFound
	}
	return
}

func (c *conn) handleTunnelLoop(remoteIP string) (handled bool) {
	var reload bool
	var cli *client
	clients := make(map[*client]struct{})
	defer func() {
		for cli := range clients {
			cli.removeTunnel(c)
		}
	}()
	for {
		handled, reload, cli = c.handleTunnel(remoteIP, reload)
		if cli != nil {
			clients[cli] = struct{}{}
		}
		if !reload {
			break
		}
	}
	return
}

func (c *conn) handleTunnel(remoteIP string, r bool) (handled, reload bool, cli *client) {
	reader := c.Reader

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

	var options options
	var u user
	if c.server.authUser != nil {
		// 验证 id secret
		u, err = c.server.authUser(idStr, secretStr)
		if err != nil {
			// 使用局部锁而不是全局锁可以明显提高并发性能，但少数情况下会降低限制效果
			c.server.reconnectRWMutex.Lock()
			reconnectTimes := c.server.reconnect[remoteIP]
			reconnectTimes++
			c.server.reconnect[remoteIP] = reconnectTimes
			c.server.reconnectRWMutex.Unlock()

			// ReconnectDuration 为 0 表示不进行限制解除
			if c.server.config.ReconnectDuration.Duration > 0 && reconnectTimes > c.server.config.ReconnectTimes {
				time.AfterFunc(c.server.config.ReconnectDuration.Duration, func() {
					c.server.reconnectRWMutex.Lock()
					c.server.reconnect[remoteIP] = 0
					c.server.reconnectRWMutex.Unlock()
					c.Logger.Info().Msgf("release blocked IP: '%v'", remoteIP)
				})
			}

			e := c.SendErrorSignalInvalidIDAndSecret()
			c.Logger.Info().Err(err).Str("id", idStr).AnErr("respErr", e).Msg("invalid id and secret")
			return
		}

		options, err = c.parseOptions(reader, idStr, u)
		if err != nil {
			c.Logger.Info().Err(err).Msg("failed to parse options")
			return
		}
	} else {
		u = user{
			TCPNumber:   &c.server.config.TCPNumber,
			Speed:       c.server.config.Speed,
			Connections: c.server.config.Connections,
			Host:        c.server.config.Host,
		}
		options, err = c.parseOptions(reader, idStr, u)
		if err != nil {
			c.Logger.Info().Err(err).Msg("failed to parse options")
			return
		}
		prefixes := make([]string, 0, len(options.ids))
		for s := range options.ids {
			prefixes = append(prefixes, s)
		}
		u, err = c.server.authUserWithAPI(idStr, secretStr, prefixes)
		if err != nil {
			// 使用局部锁而不是全局锁可以明显提高并发性能，但少数情况下会降低限制效果
			c.server.reconnectRWMutex.Lock()
			reconnectTimes := c.server.reconnect[remoteIP]
			reconnectTimes++
			c.server.reconnect[remoteIP] = reconnectTimes
			c.server.reconnectRWMutex.Unlock()

			// ReconnectDuration 为 0 表示不进行限制解除
			if c.server.config.ReconnectDuration.Duration > 0 && reconnectTimes > c.server.config.ReconnectTimes {
				time.AfterFunc(c.server.config.ReconnectDuration.Duration, func() {
					c.server.reconnectRWMutex.Lock()
					c.server.reconnect[remoteIP] = 0
					c.server.reconnectRWMutex.Unlock()
					c.Logger.Debug().Msgf("release blocked IP: '%v'", remoteIP)
				})
			}

			e := c.SendErrorSignalInvalidIDAndSecret()
			c.Logger.Info().Err(err).Str("id", idStr).AnErr("respErr", e).Msg("invalid id and secret")
			return
		}
		if len(u.Host.Prefixes) > 0 {
			for id := range options.ids {
				if _, ok := u.Host.Prefixes[id]; !ok {
					c.Logger.Info().Str("id", idStr).Str("prefix", id).Msg("prefix not exists on platform")
					delete(options.ids, id)
				}
			}
		} else {
			for id := range options.ids {
				if id != idStr {
					c.Logger.Info().Str("id", idStr).Str("prefix", id).Msg("prefix not exists on platform")
					delete(options.ids, id)
				}
			}
		}
	}

	c.Logger.Info().Hex("checksum", options.configChecksum[:]).Bool("reload", r).Msg("handling tunnel")

	// 获取或创建 client
	var ok bool
	var exists bool
	for i := 0; i < 5; i++ {
		cli, exists = c.server.getOrCreateClient(idStr, newClient)
		if !exists {
			cli.init(idStr, u, c.server)
		}

		ok, err = cli.addTunnel(c, r, options)
		if ok {
			break
		}
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to add tunnels")
			return
		}
	}
	if !ok || cli == nil {
		c.Logger.Error().Msg("failed to create client")
		return
	}

	if !r {
		atomic.AddUint64(&c.server.tunneling, 1)
		err = c.SendReadySignal()
	} else {
		err = c.SendServicesSignal()
	}
	if err != nil {
		c.Logger.Error().Err(err).Bool("reload", r).Msg("failed to send ready/services signal")
		return
	}

	handled = true
	reload = c.readLoop(cli)
	return
}

func (c *conn) processHostPrefixes(options options, cli *client) (err error) {
	rollbackIds := make(map[string]bool)
	// add host prefixes
	for id, o := range options.ids {
		v, ok := c.server.getOrCreateHostPrefix(id, o.tls, func() interface{} {
			return clientWithServiceIndex{client: cli, serviceIndex: o.serviceIndex}
		})
		if ok {
			if v.client == cli {
				continue
			} else {
				c.Logger.Error().
					Str("id", cli.id).
					Str("prefix", id).
					Bool("tls", o.tls).
					Err(connection.ErrHostConflict).
					Msg("failed to add host prefix")
				for id, tls := range rollbackIds {
					c.server.removeHostPrefix(id, tls)
					c.Logger.Info().
						Hex("checksum", options.configChecksum[:]).
						Str("id", cli.id).
						Str("prefix", id).
						Bool("tls", tls).
						Msg("rollback added associated host prefix because host prefixes conflict")
				}
				err = c.SendErrorSignalHostConflict()
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to SendErrorSignalHostConflict")
				}
				return connection.ErrHostConflict
			}
		}
		c.Logger.Info().
			Hex("checksum", options.configChecksum[:]).
			Str("id", cli.id).
			Str("prefix", id).
			Str("newServiceIndex", o.String()).
			Msg("added associated host prefix")
		rollbackIds[id] = o.tls
	}
	// remove host prefixes that are no longer used
	for id, oo := range c.ids {
		o, ok := options.ids[id]
		if !ok || oo.tls != o.tls {
			c.server.removeHostPrefix(id, oo.tls)
			c.Logger.Info().
				Str("id", cli.id).
				Hex("last checksum", c.configChecksum[:]).
				Hex("checksum", options.configChecksum[:]).
				Str("oldServiceIndex", oo.String()).
				Str("prefix", id).Msg("removed associated host prefix no longer needed")
		} else if oo.serviceIndex != o.serviceIndex {
			c.server.storeHostPrefix(id, oo.tls, clientWithServiceIndex{client: cli, serviceIndex: o.serviceIndex})
			c.Logger.Info().
				Str("id", cli.id).
				Hex("last checksum", c.configChecksum[:]).
				Hex("checksum", options.configChecksum[:]).
				Str("oldServiceIndex", oo.String()).
				Str("newServiceIndex", o.String()).
				Str("prefix", id).Msg("updated associated host prefix with different service index")
		}
	}
	c.Logger.Info().
		Str("old prefixes", c.ids.String()).
		Str("prefixes", options.ids.String()).
		Msg("updated tunnel host prefixes")
	c.ids = options.ids
	c.configChecksum = options.configChecksum
	return
}

type options struct {
	ids            hostPrefixOptions
	ports          map[uint16]openTCPOption
	configChecksum [32]byte
}

type openTCPOption struct {
	port   uint16
	random bool
}

type hostPrefixOption struct {
	serviceIndex uint16
	tls          bool
}

func (h *hostPrefixOption) String() string {
	if h.tls {
		return strconv.FormatUint(uint64(h.serviceIndex), 10) + "tls"
	}
	return strconv.FormatUint(uint64(h.serviceIndex), 10)
}

type hostPrefixOptions map[string]hostPrefixOption

func (h hostPrefixOptions) String() string {
	var sb strings.Builder
	for id, option := range h {
		sb.WriteString(id)
		sb.WriteString(":")
		sb.WriteString(option.String())
		sb.WriteByte(',')
	}
	return sb.String()
}

func (c *conn) parseOptions(reader *bufio.Reader, idStr string, u user) (options options, err error) {
	var optionsCount uint16
	var serviceIndex uint16
	ids := make(hostPrefixOptions)
	ports := make(map[uint16]openTCPOption)
	num := *u.Host.Number
	tcpNum := *u.TCPNumber
	for leftOptions := 1; leftOptions > 0; leftOptions-- {
		if optionsCount+1 > c.server.config.MaxHandShakeOptions {
			c.Logger.Error().
				AnErr("SendError", c.SendErrorSignalReachedMaxOptions()).
				Msg("client has reached the max number of options")
			return options, connection.ErrReachedMaxOptions
		}
		optionsCount++
		var optionFirst byte
		optionFirst, err = reader.ReadByte()
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to read option")
			return options, err
		}
		optionLeftLen := optionFirst >> 6
		var optionLeft []byte
		optionLeft, err = reader.Peek(int(optionLeftLen))
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to read option left")
			return options, err
		}
		option := append([]byte{optionFirst}, optionLeft...)
		_, err = reader.Discard(int(optionLeftLen))
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to discard option left")
			return options, err
		}

		tls := false
		switch {
		case bytes.Equal(option, predef.IDAsTLSHostPrefix):
			tls = true
			fallthrough
		case bytes.Equal(option, predef.IDAsHostPrefix):
			if num != 0 && uint32(len(ids))+1 > num {
				err = connection.ErrHostNumberLimited
				e := c.SendErrorSignalHostNumberLimited()
				c.Logger.Error().Err(err).AnErr("SendError", e).Msg("client has reached the max number of host prefixes")
				return options, err
			}
			c.Logger.Info().
				Str("prefix", idStr).
				Uint16("serviceIndex", serviceIndex).
				Str("id", idStr).
				Msg("adding associated host prefix")
			ids[idStr] = hostPrefixOption{serviceIndex: serviceIndex, tls: tls}
			serviceIndex++
		case bytes.Equal(option, predef.OpenTCPPort):
			if tcpNum != 0 && uint16(len(ports))+1 > tcpNum {
				err = connection.ErrTCPNumberLimited
				e := c.SendErrorSignalTCPNumberLimited()
				c.Logger.Error().Err(err).AnErr("SendError", e).Msg("client has reached the max number of tcp ports")
				return options, err
			}
			var random byte
			random, err = reader.ReadByte()
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to read random byte")
				return options, err
			}
			var peekBytes []byte
			peekBytes, err = reader.Peek(2)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to peek tcp port range")
				return options, err
			}
			tcpPort := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
			_, err = reader.Discard(2)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to discard tcp port range")
				return options, err
			}

			ports[serviceIndex] = openTCPOption{port: tcpPort, random: random != 0}
			serviceIndex++
		case bytes.Equal(option, predef.OptionAndNextOption):
			leftOptions += 2
			continue // 跳过 serverIndex++
		case bytes.Equal(option, predef.OpenTLSHost):
			tls = true
			fallthrough
		case bytes.Equal(option, predef.OpenHost):
			if num != 0 && uint32(len(ids))+1 > num {
				err = connection.ErrHostNumberLimited
				e := c.SendErrorSignalHostNumberLimited()
				c.Logger.Error().Err(err).AnErr("SendError", e).Msg("client has reached the max number of host prefixes")
				return options, err
			}
			var hostPrefixLen byte
			hostPrefixLen, err = reader.ReadByte()
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to read host prefix length")
				return options, err
			}
			hostPrefix, err := reader.Peek(int(hostPrefixLen))
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to peek host prefix")
				return options, err
			}
			hostPrefixStr := string(hostPrefix)
			_, err = reader.Discard(int(hostPrefixLen))
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to discard host prefix")
				return options, err
			}

			if len(*u.Host.Regex) > 0 {
				match := false
				for _, r := range *u.Host.Regex {
					if r.MatchString(hostPrefixStr) {
						match = true
						break
					}
				}
				if !match {
					c.Logger.Info().Err(err).
						AnErr("sendSignalError", c.SendErrorSignalHostRegexMismatch()).
						Msg("invalid host prefixes")
					return options, connection.ErrHostRegexMismatch
				}
			}
			if *u.Host.WithID {
				hostPrefixStr = idStr + "-" + hostPrefixStr
			}
			c.Logger.Info().
				Str("prefix", hostPrefixStr).
				Uint16("serviceIndex", serviceIndex).
				Str("id", idStr).
				Msg("adding associated host prefix")
			ids[hostPrefixStr] = hostPrefixOption{serviceIndex: serviceIndex, tls: tls}
			serviceIndex++
		default:
			c.Logger.Error().Msgf("invalid option: %v", optionFirst)
			return options, errors.New("invalid option")
		}
	}
	sum := calChecksum(ids, ports)
	options.ids = ids
	options.ports = ports
	options.configChecksum = sum
	return
}

func calChecksum(ids hostPrefixOptions, ports map[uint16]openTCPOption) (result [32]byte) {
	tree := btree.NewWith(3, utils.UInt16Comparator)
	for id, o := range ids {
		si := o.serviceIndex
		if o.tls {
			id = id + "-tls"
		}
		tree.Put(si, id)
	}
	for si, port := range ports {
		tree.Put(si, port)
	}
	h := sha256.New()
	it := tree.Iterator()
	k := []byte{0x0, 0x0}
	for it.Next() {
		key := it.Key().(uint16)
		k[0], k[1] = byte(key>>8), byte(key)
		h.Write(k)
		value := it.Value()
		switch v := value.(type) {
		case string:
			h.Write([]byte(v))
		case openTCPOption:
			k[0], k[1] = byte(v.port>>8), byte(v.port)
			h.Write(k)
			if v.random {
				h.Write([]byte{0x1})
			} else {
				h.Write([]byte{0x0})
			}
		}
	}
	h.Sum(result[:0])
	return
}

func (c *conn) readLoop(cli *client) (reload bool) {
	var err error
	c.Logger.Info().Msg("readLoop begin")
	defer func() {
		c.Logger.Info().Err(err).Msg("readLoop ended")
		c.tasksRWMtx.RLock()
		for _, t := range c.tasks {
			t.Close()
		}
		c.tasksRWMtx.RUnlock()
	}()
	r := &bufio.LimitedReader{}
	isClosing := false
	for {
		if c.server.config.Timeout.Duration > 0 {
			dl := time.Now().Add(c.server.config.Timeout.Duration)
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
			cli.tunnelsRWMtx.RLock()
			_, ok := cli.tunnels[c]
			cli.tunnelsRWMtx.RUnlock()
			if !ok {
				c.SendCloseSignal()
				return
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
		case connection.ServicesSignal:
			if predef.Debug {
				c.Logger.Trace().Msg("readLoop read services signal")
			}
			return true
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
			if c.server.config.Timeout.Duration > 0 && !c.server.config.TimeoutOnUnidirectionalTraffic {
				dl := time.Now().Add(c.server.config.Timeout.Duration)
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
	buf[bufIndex] = byte(predef.ServicesData >> 8)
	buf[bufIndex+1] = byte(predef.ServicesData)
	buf[bufIndex+2] = byte(task.serviceIndex >> 8)
	buf[bufIndex+3] = byte(task.serviceIndex)
	bufIndex += 4

	buffered := task.Reader.Buffered()
	var l int
	if buffered > 0 {
		var peek []byte
		peek, rErr = task.Reader.Peek(buffered)
		if rErr != nil {
			return
		}
		l = copy(buf[bufIndex+4:], peek)
		_, rErr = task.Reader.Discard(l)
		if rErr != nil {
			return
		}
	}
	if cli.needSpeedLimit() {
		cli.speedLimit(uint32(l), false) // 对客户端下行进行限速
	}
	buf[bufIndex] = byte(l >> 24)
	buf[bufIndex+1] = byte(l >> 16)
	buf[bufIndex+2] = byte(l >> 8)
	buf[bufIndex+3] = byte(l)
	l += bufIndex + 4
	_, wErr = c.Write(buf[:l])
	if wErr != nil {
		return
	}

	bufIndex -= 4
	buf[bufIndex] = byte(predef.Data >> 8)
	buf[bufIndex+1] = byte(predef.Data)
	bufIndex += 2
	for {
		if c.server.config.Timeout.Duration > 0 {
			dl := time.Now().Add(c.server.config.Timeout.Duration)
			rErr = task.SetReadDeadline(dl)
			if rErr != nil {
				return
			}
		}
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
			if c.server.config.Timeout.Duration > 0 && !c.server.config.TimeoutOnUnidirectionalTraffic {
				dl := time.Now().Add(c.server.config.Timeout.Duration)
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

type clientWithServiceIndex struct {
	*client
	serviceIndex uint16
}

func (c *conn) processTCPOptions(o options, cli *client) (err error) {
	var success []uint16
	defer func() {
		if err != nil {
			for _, si := range success {
				cli.deleteTCPListener(si)
			}
		}
	}()
	for si, portOption := range o.ports {
		v, ok := cli.tcpListeners.LoadOrCreate(si, func() interface{} {
			return &tcpListener{
				port: portOption,
			}
		})
		vl := v.(*tcpListener)
		if ok && vl.l != nil {
			port := vl.l.Addr().(*net.TCPAddr).Port
			if port == int(portOption.port) || portOption.random {
				if err := c.SendInfoTCPPortOpened(si, uint16(port)); err != nil {
					c.Logger.Error().Err(err).Msg("failed to send InfoTCPPortOpened signal")
				}
				continue
			} else {
				if vl.l != nil {
					port := uint16(vl.l.Addr().(*net.TCPAddr).Port)
					cli.logger.Info().
						Uint16("serviceIndex", si).
						Uint16("port", port).
						Msg("close associated tcp listener")
					cli.portsManager.portsMtx.Lock()
					cli.portsManager.ports[port] = struct{}{}
					cli.portsManager.portsMtx.Unlock()
					_ = vl.l.Close()
				}
			}
		}

		var openedPort uint16
		openedPort, err = cli.openTCPPort(si, vl, c)
		if err == nil {
			success = append(success, si)
		} else {
			c.Logger.Error().Err(err).
				Uint16("port", portOption.port).
				Bool("random", portOption.random).
				AnErr("respErr", c.SendErrorSignalFailedToOpenTCPPort(si)).
				Msg("failed to open tcp port")
			return err
		}
		if err := c.SendInfoTCPPortOpened(si, openedPort); err != nil {
			c.Logger.Error().Err(err).Msg("failed to send InfoTCPPortOpened signal")
		}
	}

	// remove old tcp listeners that are no longer needed
	var oldServiceIndexes []uint16
	cli.tcpListeners.Range(func(key, value any) bool {
		si := key.(uint16)
		_, ok := o.ports[si]
		if !ok {
			oldServiceIndexes = append(oldServiceIndexes, si)
		}
		return true
	})
	for _, si := range oldServiceIndexes {
		cli.deleteTCPListener(si)
	}
	return
}

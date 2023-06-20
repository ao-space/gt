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

package client

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isrc-cas/gt/bufio"
	"github.com/isrc-cas/gt/client/api"
	connection "github.com/isrc-cas/gt/conn"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
)

type conn struct {
	connection.Connection
	client        *Client
	tasks         map[uint32]*httpTask
	finishedTasks atomic.Uint64
	tasksRWMtx    sync.RWMutex
	stuns         []string
	services      atomic.Pointer[services]
}

func newConn(c net.Conn, client *Client) *conn {
	nc := &conn{
		Connection: connection.Connection{
			Conn:         c,
			Reader:       pool.GetReader(c),
			WriteTimeout: client.Config().RemoteTimeout,
		},
		client: client,
		tasks:  make(map[uint32]*httpTask, 100),
	}
	return nc
}

func (c *conn) init() (err error) {
	buf := c.Connection.Reader.GetBuf()
	var n int
	buf[n] = predef.MagicNumber
	n++
	buf[n] = 0x01 // version
	n++

	bufIndex := gen(*c.client.config.Load(), *c.client.services.Load(), buf[n:])
	_, err = c.Conn.Write(buf[:bufIndex+n])
	return
}

func gen(config Config, services services, buf []byte) (n int) {
	// id
	buf[n] = byte(len(config.ID))
	n++
	idLen := copy(buf[n:], config.ID)
	n += idLen

	// secret
	buf[n] = byte(len(config.Secret))
	n++
	secretLen := copy(buf[n:], config.Secret)
	n += secretLen

	// services
	for i, service := range services {
		if i != len(services)-1 {
			optionLen := copy(buf[n:], predef.OptionAndNextOption)
			n += optionLen
		}
		switch service.LocalURL.Scheme {
		case "tcp":
			optionLen := copy(buf[n:], predef.OpenTCPPort)
			n += optionLen

			if *service.RemoteTCPRandom {
				buf[n] = 1
			} else {
				buf[n] = 0
			}
			n++

			buf[n] = byte(service.RemoteTCPPort >> 8)
			buf[n+1] = byte(service.RemoteTCPPort)
			n += 2
		case "http":
			if service.HostPrefix == config.ID {
				optionLen := copy(buf[n:], predef.IDAsHostPrefix)
				n += optionLen
			} else {
				optionLen := copy(buf[n:], predef.OpenHost)
				n += optionLen

				buf[n] = byte(len(service.HostPrefix))
				n++
				hostPrefixLen := copy(buf[n:], service.HostPrefix)
				n += hostPrefixLen
			}
		case "https":
			if service.HostPrefix == config.ID {
				optionLen := copy(buf[n:], predef.IDAsTLSHostPrefix)
				n += optionLen
			} else {
				optionLen := copy(buf[n:], predef.OpenTLSHost)
				n += optionLen

				buf[n] = byte(len(service.HostPrefix))
				n++
				hostPrefixLen := copy(buf[n:], service.HostPrefix)
				n += hostPrefixLen
			}
		}
	}
	return
}

func (c *conn) IsTimeout(e error) (result bool) {
	if ne, ok := e.(*net.OpError); ok && ne.Timeout() {
		err := c.Connection.SendPingSignal()
		if err == nil {
			result = true
			return
		}
		c.Logger.Debug().Err(err).Msg("failed to send ping signal")
	}
	return
}

func (c *conn) Close() {
	if !c.Closing.CompareAndSwap(0, 1) {
		return
	}
	c.tasksRWMtx.Lock()
	for _, task := range c.tasks {
		task.Close()
	}
	c.tasksRWMtx.Unlock()
	c.Connection.CloseOnce()
}

func (c *conn) tasksLen() (n int) {
	c.tasksRWMtx.RLock()
	n = len(c.tasks)
	c.tasksRWMtx.RUnlock()
	return n
}

func (c *conn) readLoop(connID uint) {
	var err error
	var pings int
	var lastPing int
	var isClosing bool
	defer func() {
		c.client.removeTunnel(c)
		c.Close()
		c.Logger.Info().Err(err).Bool("isClosing", isClosing).Uint64("finishedTasks", c.finishedTasks.Load()).
			Int("tasksCount", c.tasksLen()).Int("pings", pings).Msg("tunnel closed")
		c.onTunnelClose()
		pool.PutReader(c.Reader)
	}()

	r := &bufio.LimitedReader{}
	r.Reader = c.Reader
	var timeout time.Duration
	if c.client.Config().RemoteTimeout > 0 {
		timeout = c.client.Config().RemoteTimeout / 2
		if timeout <= 0 {
			timeout = c.client.Config().RemoteTimeout
		}
	}
	for pings <= 3 {
		if timeout > 0 {
			err = c.Conn.SetReadDeadline(time.Now().Add(timeout))
			if err != nil {
				return
			}
		}
		var peekBytes []byte
		peekBytes, err = c.Reader.Peek(4)
		if err != nil {
			if c.IsTimeout(err) {
				pings++
				c.Logger.Info().Bool("isClosing", isClosing).Uint64("finishedTasks", c.finishedTasks.Load()).
					Int("tasksCount", c.tasksLen()).Int("pings", pings).Msg("sent ping")
				err = nil
				continue
			}
			return
		}
		signal := uint32(peekBytes[3]) | uint32(peekBytes[2])<<8 | uint32(peekBytes[1])<<16 | uint32(peekBytes[0])<<24
		_, err = c.Reader.Discard(4)
		if err != nil {
			return
		}
		switch signal {
		case connection.PingSignal:
			pings--
			lastPing++
			if isClosing && lastPing >= 3 {
				if c.tasksLen() == 0 {
					return
				}
			}
			if lastPing >= 6 {
				lastPing = 0
				if c.client.idleManager.ChangeToWait(connID) {
					c.SendCloseSignal()
					c.Logger.Info().Msg("sent close signal")
				}
			}
			continue
		case connection.CloseSignal:
			c.Logger.Info().Msg("read close signal")
			if isClosing {
				return
			}
			if c.tasksLen() == 0 {
				return
			}
			isClosing = true
			continue
		case connection.ReadySignal:
			c.client.addTunnel(c)
			c.Logger.Info().Msg("tunnel started")
			continue
		case connection.ServicesSignal:
			c.services.Store(c.client.services.Load())
			c.Logger.Info().Msg("tunnel updated")
			c.client.reloadWaitGroup.Done()
			c.Logger.Info().Msg("client reload wait group done")
			continue
		case connection.ErrorSignal:
			peekBytes, err = c.Reader.Peek(2)
			if err != nil {
				return
			}
			errCode := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
			c.Logger.Error().Err(connection.Error(errCode)).Msg("read error signal")
			return
		case connection.InfoSignal:
			peekBytes, err = c.Reader.Peek(2)
			if err != nil {
				return
			}
			infoCode := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
			_, err = c.Reader.Discard(2)
			if err != nil {
				return
			}
			info, err := connection.Info(infoCode).ReadInfo(c.Reader)
			if err != nil {
				return
			}
			c.Logger.Info().Msgf("receive server information: %s", info)
			continue
		}
		lastPing = 0
		taskID := signal
		peekBytes, err = c.Reader.Peek(2)
		if err != nil {
			return
		}
		taskOption := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
		_, err = c.Reader.Discard(2)
		if err != nil {
			return
		}
		switch taskOption {
		case predef.ServicesData:
			serviceIndex := uint16(0)
			peekBytes, err = c.Reader.Peek(2)
			if err != nil {
				return
			}
			serviceIndex = uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
			_, err = c.Reader.Discard(2)
			if err != nil {
				return
			}
			if serviceIndex >= uint16(len(*c.services.Load())) {
				c.Logger.Error().Uint16("serviceIndex", serviceIndex).Msg("invalid service index")
				return
			}
			service := &(*c.services.Load())[serviceIndex]

			peekBytes, err = c.Reader.Peek(4)
			if err != nil {
				return
			}
			l := uint32(peekBytes[3]) | uint32(peekBytes[2])<<8 | uint32(peekBytes[1])<<16 | uint32(peekBytes[0])<<24
			_, err = c.Reader.Discard(4)
			if err != nil {
				return
			}
			r.N = int64(l)
			rErr, wErr := c.processServiceData(connID, taskID, service, r)
			if rErr != nil {
				err = wErr
				if !errors.Is(rErr, net.ErrClosed) {
					c.Logger.Warn().Err(rErr).Msg("failed to read data in processData")
				}
				return
			}
			if r.N > 0 {
				_, err = r.Discard(int(r.N))
				if err != nil {
					return
				}
			}
			if wErr != nil {
				if !errors.Is(wErr, net.ErrClosed) {
					c.Logger.Warn().Err(wErr).Msg("failed to write data in processData")
				}
				continue
			}
		case predef.Data:
			peekBytes, err = c.Reader.Peek(4)
			if err != nil {
				return
			}
			l := uint32(peekBytes[3]) | uint32(peekBytes[2])<<8 | uint32(peekBytes[1])<<16 | uint32(peekBytes[0])<<24
			_, err = c.Reader.Discard(4)
			if err != nil {
				return
			}
			r.N = int64(l)
			rErr, wErr := c.processData(taskID, r)
			if rErr != nil {
				err = wErr
				if !errors.Is(rErr, net.ErrClosed) {
					c.Logger.Warn().Err(rErr).Msg("failed to read data in processData")
				}
				return
			}
			if r.N > 0 {
				_, err = r.Discard(int(r.N))
				if err != nil {
					return
				}
			}
			if wErr != nil {
				if !errors.Is(wErr, net.ErrClosed) {
					c.Logger.Warn().Err(wErr).Msg("failed to write data in processData")
				}
				continue
			}
		case predef.Close:
			c.tasksRWMtx.RLock()
			t, ok := c.tasks[taskID]
			c.tasksRWMtx.RUnlock()
			if ok {
				t.CloseByRemote()
			}
		}
	}
}

func (c *conn) dial(s *service) (task *httpTask, err error) {
	conn, err := net.Dial("tcp", s.LocalURL.Host)
	if err != nil {
		return
	}
	task = newHTTPTask(conn)
	task.service = s
	if s.UseLocalAsHTTPHost {
		err = task.setHost(s.LocalURL.Host)
	}
	return
}

func (c *conn) processServiceData(connID uint, taskID uint32, s *service, r *bufio.LimitedReader) (readErr, writeErr error) {
	var peekBytes []byte
	peekBytes, readErr = r.Peek(2)
	if readErr != nil {
		return
	}
	// first 2 bytes of p2p sdp request is "XP"(0x5850)
	isP2P := (uint16(peekBytes[1]) | uint16(peekBytes[0])<<8) == 0x5850
	if isP2P {
		if len(c.stuns) < 1 {
			respAndClose(taskID, c, [][]byte{
				[]byte("HTTP/1.1 403 Forbidden\r\nConnection: Closed\r\n\r\n"),
			})
			return
		}
		c.processP2P(taskID, r)
		return
	}

	var task *httpTask
	for i := 0; i < 3; i++ {
		task, writeErr = c.dial(s)
		if writeErr == nil {
			break
		}
	}
	if writeErr != nil {
		return
	}
	task.Logger = c.Logger.With().
		Uint32("task", taskID).
		Logger()
	task.Logger.Info().Msg("task started")
	c.tasksRWMtx.Lock()
	ot, ok := c.tasks[taskID]
	if ok && ot != nil {
		ot.Close()
		ot.Logger.Info().Msg("got closed because task with same id is received")
	}
	c.tasks[taskID] = task
	c.tasksRWMtx.Unlock()
	go task.process(connID, taskID, c)

	_, err := r.WriteTo(task)
	if err != nil {
		switch e := err.(type) {
		case *net.OpError:
			switch e.Op {
			case "write":
				writeErr = err
			}
		case *bufio.WriteErr:
			writeErr = err
		default:
			readErr = err
		}
	}
	if task.service.LocalTimeout > 0 {
		dl := time.Now().Add(task.service.LocalTimeout)
		writeErr = task.conn.SetReadDeadline(dl)
		if writeErr != nil {
			return
		}
	}
	return
}

func (c *conn) processData(taskID uint32, r *bufio.LimitedReader) (readErr, writeErr error) {
	c.tasksRWMtx.RLock()
	task, ok := c.tasks[taskID]
	c.tasksRWMtx.RUnlock()
	if !ok {
		c.client.peersRWMtx.RLock()
		pt, ok := c.client.peers[taskID]
		c.client.peersRWMtx.RUnlock()
		if ok && pt != nil {
			if len(c.stuns) < 1 {
				respAndClose(taskID, c, [][]byte{
					[]byte("HTTP/1.1 403 Forbidden\r\nConnection: Closed\r\n\r\n"),
				})
				return
			}
			_, err := r.WriteTo(pt.apiConn.PipeWriter)
			if err != nil {
				pt.Logger.Error().Err(err).Msg("processP2P WriteTo failed")
			}
			return
		}
		return nil, errors.New("task not exists")
	}
	_, err := r.WriteTo(task)
	if err != nil {
		switch e := err.(type) {
		case *net.OpError:
			switch e.Op {
			case "write":
				writeErr = err
			}
		case *bufio.WriteErr:
			writeErr = err
		default:
			readErr = err
		}
	}
	if task.service.LocalTimeout > 0 {
		dl := time.Now().Add(task.service.LocalTimeout)
		writeErr = task.conn.SetReadDeadline(dl)
		if writeErr != nil {
			return
		}
	}
	return
}

func (c *conn) processP2P(id uint32, r *bufio.LimitedReader) {
	var t = &peerTask{}
	t.id = id
	t.tunnel = c
	t.apiConn = api.NewConn(id, "", c)
	t.apiConn.ProcessOffer = t.processOffer
	t.apiConn.GetOffer = t.getOffer
	t.apiConn.ProcessAnswer = t.processAnswer
	t.data = pool.BytesPool.Get().([]byte)
	t.candidateOutChan = make(chan string, 16)
	t.closeChan = make(chan struct{})
	t.waitNegotiationNeeded = make(chan struct{})
	t.Logger = c.Logger.With().
		Uint32("peerTask", id).
		Logger()
	t.timer = time.AfterFunc(120*time.Second, func() {
		t.Logger.Info().Msg("peer task timeout")
		t.CloseWithLock()
	})

	c.client.peersRWMtx.Lock()
	ot, ok := c.client.peers[id]
	if ok && ot != nil {
		ot.CloseWithLock()
		ot.Logger.Info().Msg("got closed because task with same id is received")
	}
	c.client.peers[id] = t
	c.client.peersRWMtx.Unlock()

	c.client.apiServer.Listener.AcceptCh() <- t.apiConn
	t.Logger.Info().Msg("peer task started")
	_, err := r.WriteTo(t.apiConn.PipeWriter)
	if err != nil {
		t.Logger.Error().Err(err).Msg("processP2P WriteTo failed")
	}
}

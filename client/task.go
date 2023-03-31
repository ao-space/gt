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
	"bytes"
	"errors"
	"net"
	"sync/atomic"
	"time"

	connection "github.com/isrc-cas/gt/conn"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
	"github.com/rs/zerolog"
)

var (
	// ErrHostIsTooLong is an error returned when host is too long
	ErrHostIsTooLong = errors.New("host is too long")

	host = []byte("Host:")
)

type httpTask struct {
	conn     net.Conn
	buf      []byte
	tempBuf  *bytes.Buffer
	Logger   zerolog.Logger
	skipping bool
	passing  bool
	closing  uint32
}

func newHTTPTask(c net.Conn) (t *httpTask) {
	t = &httpTask{
		conn: c,
	}
	return
}

func (t *httpTask) setHost(host string) (err error) {
	if len(host) > 200 {
		return ErrHostIsTooLong
	}
	t.buf = pool.BytesPool.Get().([]byte)
	n := copy(t.buf, "Host: ")
	n += copy(t.buf[n:], host)
	n += copy(t.buf[n:], "\r\n")
	t.tempBuf = bytes.NewBuffer(t.buf[n:][:0])
	t.buf = t.buf[:n]
	return
}

func (t *httpTask) Write(p []byte) (n int, err error) {
	if t.tempBuf == nil {
		return t.conn.Write(p)
	} else if t.skipping {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			return len(p), nil
		}
		t.skipping = false
		t.tempBuf = nil
		if len(p) <= i+1 {
			return len(p), nil
		}
		if predef.Debug {
			t.Logger.Debug().Bytes("data", p[i+1:]).Msg("write")
		}
		n, err = t.conn.Write(p[i+1:])
		n += i + 1
		return
	} else if t.passing {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			if predef.Debug {
				t.Logger.Debug().Bytes("data", p).Msg("write")
			}
			return t.conn.Write(p)
		}
		t.passing = false
		if predef.Debug {
			t.Logger.Debug().Bytes("data", p[:i+1]).Msg("write")
		}
		n, err = t.conn.Write(p[:i+1])
		if err != nil {
			return
		}
		p = p[i+1:]
		if len(p) <= 0 {
			return
		}
	}

	var nw int
	var updated bool
	if t.tempBuf != nil {
		if len(p)+t.tempBuf.Len() < 5 {
			nw, err = t.tempBuf.Write(p)
			n += nw
			return
		}

		if t.tempBuf.Len() > 0 {
			nw, err = t.tempBuf.Write(p)
			n += nw
			if err != nil {
				return
			}
			p = t.tempBuf.Bytes()
			t.tempBuf.Reset()
			updated = true
		}
	}

	s := 0
	for len(p[s:]) >= 1 {
		switch t.isGoodToWrite(p[s:]) {
		case good:
			i := bytes.IndexByte(p[s:], '\n')
			if i < 0 {
				t.passing = true
				if predef.Debug {
					t.Logger.Debug().Bytes("data", p[s:]).Msg("write")
				}
				nw, err = t.conn.Write(p[s:])
				if !updated {
					n += nw
				}
				return
			}
			if predef.Debug {
				t.Logger.Debug().Bytes("data", p[s:s+i+1]).Msg("write")
			}
			nw, err = t.conn.Write(p[s : s+i+1])
			if !updated {
				n += nw
			}
			if err != nil {
				return
			}
			s += i + 1
		case unsure:
			nw, err = t.tempBuf.Write(p[s:])
			if !updated {
				n += nw
			}
			return
		case replace:
			if predef.Debug {
				t.Logger.Debug().Bytes("data", t.buf).Msg("write")
			}
			_, err = t.conn.Write(t.buf)
			if err != nil {
				return
			}
			i := bytes.IndexByte(p[s:], '\n')
			if i < 0 {
				t.skipping = true
				nw = len(p[s:])
				if !updated {
					n += nw
				}
				return
			}
			s += i + 1
			if !updated {
				n += i + 1
			}
			t.tempBuf = nil
			pool.BytesPool.Put(t.buf[:cap(t.buf)])
			if len(p[s:]) > 0 {
				if predef.Debug {
					t.Logger.Debug().Bytes("data", p[s:]).Msg("write")
				}
				nw, err = t.conn.Write(p[s:])
				if !updated {
					n += nw
				}
			}
			return
		}
	}
	return
}

const (
	good = iota
	unsure
	replace
)

func (t *httpTask) isGoodToWrite(p []byte) int {
	l := len(p)
	if l > 5 {
		l = 5
	}
	if bytes.Equal(p[:l], host[:l]) {
		if l < 5 {
			return unsure
		}
		return replace
	}
	return good
}

func (t *httpTask) Close() {
	t.CloseWithValue(connection.Close)
}

func (t *httpTask) CloseByRemote() {
	t.CloseWithValue(connection.CloseByRemote)
}

func (t *httpTask) CloseWithValue(value uint32) {
	if !atomic.CompareAndSwapUint32(&t.closing, 0, value) {
		return
	}
	var err error
	if t.conn != nil {
		err = t.conn.Close()
	}
	t.Logger.Info().Uint32("by", atomic.LoadUint32(&t.closing)).Err(err).Msg("task closed")
}

// IsClosingByRemote tells is the connection closing by remote
func (t *httpTask) IsClosingByRemote() (closingByRemote bool) {
	return atomic.LoadUint32(&t.closing) == connection.CloseByRemote
}

func (t *httpTask) process(connID uint, taskID uint32, c *conn, timeout time.Duration) {
	count := c.TasksCount.Add(1)
	c.client.idleManager.SetRunningWithTaskCount(connID, count)
	var rErr error
	var wErr error
	buf := pool.BytesPool.Get().([]byte)
	defer func() {
		if wErr == nil && !t.IsClosingByRemote() {
			buf[4] = byte(predef.Close >> 8)
			buf[5] = byte(predef.Close)
			_, wErr = c.Write(buf[:6])
		}
		pool.BytesPool.Put(buf)
		t.Logger.Info().AnErr("read err", rErr).AnErr("write err", wErr).Msg("http task read loop returned")
		c.tasksRWMtx.Lock()
		delete(c.tasks, taskID)
		c.tasksRWMtx.Unlock()
		c.finishedTasks.Add(1)
		t.Close()
		if c.TasksCount.Add(^uint32(0)) == 0 {
			c.client.idleManager.SetIdle(connID)
			if c.IsClosing() {
				c.SendForceCloseSignal()
				c.Close()
				return
			}
		}
		if wErr != nil {
			c.Close()
		}
	}()
	buf[0] = byte(taskID >> 24)
	buf[1] = byte(taskID >> 16)
	buf[2] = byte(taskID >> 8)
	buf[3] = byte(taskID)
	buf[4] = byte(predef.Data >> 8)
	buf[5] = byte(predef.Data)
	for {
		if timeout > 0 {
			dl := time.Now().Add(timeout)
			rErr = t.conn.SetReadDeadline(dl)
			if rErr != nil {
				return
			}
		}
		var l int
		l, rErr = t.conn.Read(buf[10:])
		if l > 0 {
			buf[6] = byte(l >> 24)
			buf[7] = byte(l >> 16)
			buf[8] = byte(l >> 8)
			buf[9] = byte(l)
			l += 10

			if predef.Debug {
				c.Logger.Trace().Hex("data", buf[:l]).Msg("write")
			}
			_, wErr = c.Write(buf[:l])
			if wErr != nil {
				return
			}
			if c.client.config.RemoteTimeout > 0 {
				dl := time.Now().Add(c.client.config.RemoteTimeout)
				wErr = c.Conn.SetReadDeadline(dl)
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

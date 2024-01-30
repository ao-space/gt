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

package api

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/isrc-cas/gt/client/std"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
)

type Conn struct {
	PipeReader    *std.PipeReader
	PipeWriter    *std.PipeWriter
	writer        io.Writer
	closeCallback func()
	addr          string
	timer         *time.Timer
	once          sync.Once
	id            uint32
	buf           []byte
	mtx           sync.Mutex

	ProcessOffer  func(r *http.Request, writer http.ResponseWriter)
	GetOffer      func(r *http.Request, writer http.ResponseWriter)
	ProcessAnswer func(r *http.Request, writer http.ResponseWriter)
}

func (c *Conn) SetCloseCallback(closeCallback func()) {
	c.closeCallback = closeCallback
}

func NewConn(id uint32, addr string, writer io.Writer) (conn *Conn) {
	conn = &Conn{
		addr:   addr,
		id:     id,
		buf:    pool.BytesPool.Get().([]byte),
		writer: writer,
	}
	conn.PipeReader, conn.PipeWriter = std.Pipe()
	return
}

func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c.PipeReader.Read(b)
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	return
}

func (c *Conn) Write(b []byte) (n int, err error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	binary.BigEndian.PutUint32(c.buf[0:], c.id)
	binary.BigEndian.PutUint16(c.buf[4:], predef.Data)

	var i int
	for {
		j := copy(c.buf[10:], b[i:])
		i += j
		binary.BigEndian.PutUint32(c.buf[6:], uint32(j))
		var nw int
		nw, err = c.writer.Write(c.buf[:10+j])
		if err != nil {
			return
		}
		n += nw
		if len(b[i:]) == 0 {
			break
		}
	}
	return
}

func (c *Conn) writeClose() (n int, err error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	binary.BigEndian.PutUint32(c.buf[0:], c.id)
	binary.BigEndian.PutUint16(c.buf[4:], predef.Close)
	n, err = c.writer.Write(c.buf[:6])
	return
}

func (c *Conn) Close() (err error) {
	c.once.Do(func() {
		defer pool.BytesPool.Put(c.buf)
		if c.closeCallback != nil {
			defer c.closeCallback()
		}
		e1 := c.PipeReader.Close()
		_, e2 := c.writeClose()
		if e1 != nil && e2 != nil {
			err = fmt.Errorf("reader err: %v, writer err: %v", e1, e2)
			return
		}
		if e1 != nil {
			err = e1
			return
		}
		err = e2
	})
	return
}

func (c *Conn) LocalAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.2:1")
	return addr
}

func (c *Conn) RemoteAddr() net.Addr {
	var addr net.Addr
	var err error
	if len(c.addr) > 0 {
		addr, err = net.ResolveTCPAddr("tcp", c.addr)
		if err == nil {
			return addr
		}
	}
	addr, _ = net.ResolveTCPAddr("tcp", "127.0.0.3:1")
	return addr
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.SetReadDeadline(t)
}

var aLongTimeAgo = time.Unix(1, 0)

func (c *Conn) SetReadDeadline(t time.Time) error {
	if t == aLongTimeAgo {
		c.PipeReader.StopPending(&net.OpError{Op: "set", Net: "virtual", Source: nil, Addr: nil, Err: context.DeadlineExceeded})
		return nil
	}
	dur := time.Until(t)
	if dur <= 0 {
		return errors.New("deadline has already passed")
	}
	c.timer = time.AfterFunc(dur, func() {
		c.PipeReader.StopPending(&net.OpError{Op: "set", Net: "virtual", Source: nil, Addr: nil, Err: context.DeadlineExceeded})
	})
	return nil
}

func (c *Conn) SetWriteDeadline(_ time.Time) error {
	return nil
}

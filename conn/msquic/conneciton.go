package msquic

/*
#include <stdlib.h>

#include "connection.h"
#include "stream.h"
*/
import "C"

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/mattn/go-pointer"
)

type Connection struct {
	cppConn             unsafe.Pointer
	pointerID           unsafe.Pointer
	onConnected         chan struct{}
	onPeerStreamStarted chan net.Conn
	onClose             chan struct{}
	closeOnce           sync.Once
}

func NewConnection(
	server string,
	idleTimeoutMs uint64,
	certFile string,
	unsecure bool,
) (conn *Connection, err error) {
	conn = &Connection{
		onConnected:         make(chan struct{}, 1),
		onPeerStreamStarted: make(chan net.Conn, 1),
		onClose:             make(chan struct{}),
	}
	conn.pointerID = pointer.Save(conn)
	serverName, serverPortStr, err := net.SplitHostPort(server)
	if err != nil {
		return nil, err
	}
	cServerName := C.CString(serverName)
	defer C.free(unsafe.Pointer(cServerName))
	serverPort, err := strconv.Atoi(serverPortStr)
	if err != nil {
		return nil, err
	}
	cCertFile := C.CString(certFile)
	defer C.free(unsafe.Pointer(cCertFile))
	conn.cppConn = C.NewConnection(
		conn.pointerID,
		cServerName,
		C.uint16_t(serverPort),
		C.uint64_t(idleTimeoutMs),
		cCertFile,
		C.bool(unsecure),
	)
	if conn.cppConn == nil {
		return nil, errors.New("msquic NewConnection failed")
	}
	select {
	case <-conn.onConnected:
		return
	case <-conn.onClose:
		return nil, errors.New("msquic connection closed")
	}
}

// TODO quic connection层面的read和write主要指DATAGRAM扩展，有待实现
func (c *Connection) Read(b []byte) (n int, err error) {
	panic("not implemented")
}

func (c *Connection) Write(b []byte) (n int, err error) {
	panic("not implemented")
}

func (c *Connection) LocalAddr() (addr net.Addr) {
	cAddr := C.GetConnectionAddr(c.cppConn, C.bool(true))
	if cAddr == nil {
		return
	}

	addr, err := net.ResolveUDPAddr("udp", C.GoString(cAddr))
	if err != nil {
		addr = nil
	}
	return
}

func (c *Connection) RemoteAddr() (addr net.Addr) {
	cAddr := C.GetConnectionAddr(c.cppConn, C.bool(false))
	if cAddr == nil {
		return
	}

	addr, err := net.ResolveUDPAddr("udp", C.GoString(cAddr))
	if err != nil {
		addr = nil
	}
	return
}

func (c *Connection) SetDeadline(t time.Time) error {
	// TODO 这里实际上设置的是连接的空闲时间
	timeout := time.Since(t) / time.Millisecond
	C.SetConnectionIdleTimeout(c.cppConn, C.uint64_t(timeout))
	return nil
}

func (c *Connection) SetReadDeadline(t time.Time) error {
	return c.SetDeadline(t)
}

func (c *Connection) SetWriteDeadline(t time.Time) error {
	return c.SetDeadline(t)
}

func (c *Connection) Close() error {
	c.closeOnce.Do(func() {
		close(c.onClose)
	})
	pointer.Unref(c.pointerID)
	C.DeleteConnection(c.cppConn)
	return nil
}

func (c *Connection) OpenStream() (conn net.Conn, err error) {
	s := &stream{
		onStarted: make(chan struct{}, 1),
		onSend:    make(chan struct{}, 1),
		onClose:   make(chan struct{}),
		onReceive: make(chan []byte, 1),
		conn:      c,
	}
	s.pointerID = pointer.Save(s)
	s.cppStream = C.OpenStream(c.cppConn, s.pointerID)
	if s.cppStream == nil {
		return nil, errors.New("msquic OpenStream failed")
	}
	select {
	case <-s.onStarted:
		return s, nil
	case <-s.onClose:
		return nil, errors.New("msquic stream closed")
	}
}

func (c *Connection) PeerStreamStarted() (conn net.Conn, err error) {
	select {
	case conn = <-c.onPeerStreamStarted:
	case <-c.onClose:
		err = errors.New("msquic connection closed")
	}
	return
}

//export OnConnectionConnected
func OnConnectionConnected(cppConn, context unsafe.Pointer) {
	conn, ok := pointer.Restore(context).(*Connection)
	if !ok || conn == nil {
		return
	}
	select {
	case conn.onConnected <- struct{}{}:
	case <-conn.onClose:
	}
}

//export OnConnectionShutdownComplete
func OnConnectionShutdownComplete(cppConn, context unsafe.Pointer) {
	conn, ok := pointer.Restore(context).(*Connection)
	if !ok || conn == nil {
		return
	}
	conn.closeOnce.Do(func() {
		close(conn.onClose)
	})
}

//export OnPeerStreamStarted
func OnPeerStreamStarted(cppConn, cppStream, context unsafe.Pointer) {
	conn, ok := pointer.Restore(context).(*Connection)
	if !ok || conn == nil {
		return
	}
	s := &stream{
		cppStream: cppStream,
		onStarted: make(chan struct{}, 1),
		onSend:    make(chan struct{}, 1),
		onClose:   make(chan struct{}, 1),
		onReceive: make(chan []byte, 1),
		conn:      conn,
	}
	s.pointerID = pointer.Save(s)
	C.SetStreamContext(cppStream, s.pointerID)
	select {
	case conn.onPeerStreamStarted <- s:
	case <-conn.onClose:
	}
}

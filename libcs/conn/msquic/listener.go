package msquic

/*
#include <stdlib.h>

#include "connection.h"
#include "listener.h"
*/
import "C"

import (
	"errors"
	"net"
	"unsafe"

	"github.com/mattn/go-pointer"
)

type Listener struct {
	cppListener     unsafe.Pointer
	pointerID       unsafe.Pointer
	onClose         chan struct{}
	onNewConnection chan net.Conn
}

func NewListenr(
	addr string,
	idleTimeoutMs uint64,
	keyFile string,
	certFile string,
	password string,
) (listener *Listener, err error) {
	listener = &Listener{
		onClose:         make(chan struct{}),
		onNewConnection: make(chan net.Conn, 1),
	}
	listener.pointerID = pointer.Save(listener)
	cAddr := C.CString(addr)
	defer C.free(unsafe.Pointer(cAddr))
	cKeyFile := C.CString(keyFile)
	defer C.free(unsafe.Pointer(cKeyFile))
	cCertFile := C.CString(certFile)
	defer C.free(unsafe.Pointer(cCertFile))
	cPassword := C.CString(password)
	defer C.free(unsafe.Pointer(cPassword))
	listener.cppListener = C.NewListener(
		cAddr,
		C.uint64_t(idleTimeoutMs),
		cKeyFile,
		cCertFile,
		cPassword,
		listener.pointerID,
	)
	if listener.cppListener == nil {
		err = errors.New("failed to create listener")
	}
	return
}

func (l *Listener) Accept() (conn net.Conn, err error) {
	select {
	case quicConn := <-l.onNewConnection:
		newQuicConn, ok := quicConn.(*Connection)
		if !ok {
			return nil, errors.New("msquic not renturn a msquic connection")
		}
		streamConn, err := newQuicConn.PeerStreamStarted()
		if streamConn == nil {
			return nil, errors.New("msquic AcceptStream failed")
		}
		msquicConn := &MsquicConn{
			Conn:             streamConn,
			MsquicConnection: newQuicConn,
		}
		return msquicConn, err
	case <-l.onClose:
		err = errors.New("listener closed")
	}
	return
}

func (l *Listener) Close() error {
	pointer.Unref(l.pointerID)
	C.DeleteListener(l.cppListener)
	return nil
}

func (l *Listener) Addr() (addr net.Addr) {
	cAddr := C.GetListenerAddr(l.cppListener)
	if cAddr == nil {
		return
	}

	addr, err := net.ResolveUDPAddr("udp", C.GoString(cAddr))
	if err != nil {
		addr = nil
	}
	return
}

//export OnNewConnection
func OnNewConnection(cppListener, cppConn, context unsafe.Pointer) {
	listener, ok := pointer.Restore(context).(*Listener)
	if !ok || listener == nil {
		return
	}
	c := &Connection{
		cppConn:             cppConn,
		onConnected:         make(chan struct{}, 1),
		onPeerStreamStarted: make(chan net.Conn, 1),
		onClose:             make(chan struct{}),
	}
	c.pointerID = pointer.Save(c)
	C.SetConnectionContext(cppConn, c.pointerID)
	select {
	case listener.onNewConnection <- c:
	case <-listener.onClose:
	}
}

//export OnListenerStopComplete
func OnListenerStopComplete(cppListener, context unsafe.Pointer) {
	listener, ok := pointer.Restore(context).(*Listener)
	if !ok || listener == nil {
		return
	}
	close(listener.onClose)
}

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

package conn

import (
	"errors"
	"math"
	"net"
	"sync/atomic"
	"time"

	"github.com/isrc-cas/gt/bufio"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
	"github.com/rs/zerolog"
)

// ErrInvalidWrite is an error of write operation, number of wrote data is less than passed
var ErrInvalidWrite = errors.New("invalid write")

// Connection is an extended net.Conn
type Connection struct {
	net.Conn
	Logger       zerolog.Logger
	Reader       *bufio.Reader
	WriteTimeout time.Duration
	TasksCount   atomic.Uint32
	Closing      atomic.Uint32
}

func (c *Connection) Write(b []byte) (n int, err error) {
	l := len(b)
	if c.WriteTimeout > 0 {
		dl := time.Now().Add(c.WriteTimeout)
		err = c.Conn.SetWriteDeadline(dl)
		if err != nil {
			return
		}
	}
	n, err = c.Conn.Write(b)
	if l != n && err == nil {
		err = ErrInvalidWrite
	}
	return
}

const (
	_ = iota
	// Close indicates the connection is closed at local
	Close
	// CloseByRemote indicates the connection is closed by remote operation
	CloseByRemote
)

// Close closes Connection
func (c *Connection) Close() {
	c.CloseWithValue(Close)
}

// CloseByRemote closes Connection with closeByRemote flag
func (c *Connection) CloseByRemote() {
	c.CloseWithValue(CloseByRemote)
}

// CloseWithValue closes Connection with specified reason
func (c *Connection) CloseWithValue(value uint32) {
	if c.Closing.CompareAndSwap(0, value) {
		c.CloseOnce()
	}
}

// CloseOnce closes Connection
func (c *Connection) CloseOnce() {
	err := c.Conn.Close()
	c.Logger.Info().Uint32("by", c.Closing.Load()).Err(err).Msg("conn close")
}

// IsClosing tells is the connection closing.
func (c *Connection) IsClosing() (closing bool) {
	return c.Closing.Load() > 0
}

// IsClosingByRemote tells is the connection closing by remote
func (c *Connection) IsClosingByRemote() (closingByRemote bool) {
	return c.Closing.Load() == CloseByRemote
}

// Signal is alias type of uint32 for specify signal in the protocol
type Signal = uint32

const (
	// PingSignal is a signal used for ping
	PingSignal Signal = math.MaxUint32 - iota
	// CloseSignal is a signal used for close
	CloseSignal
	// ReadySignal is a signal used for ready
	ReadySignal
	// ErrorSignal is a signal used for errors
	ErrorSignal
	// InfoSignal is a signal used for information
	InfoSignal
	// ServicesSignal is a signal used for services changes
	ServicesSignal
	// ReconnectSignal is a signal used for services changes
	ReconnectSignal

	// PreservedSignal is a signal used for preserved signals
	PreservedSignal Signal = math.MaxUint32 - 3000
)

var (
	pingBytes                              = []byte{0xFF, 0xFF, 0xFF, 0xFF}
	closeBytes                             = []byte{0xFF, 0xFF, 0xFF, 0xFE}
	forceCloseBytes                        = []byte{0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFF, 0xFE}
	readyBytes                             = []byte{0xFF, 0xFF, 0xFF, 0xFD}
	errInvalidIDAndSecretBytes             = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x01}
	errFailedToOpenTCPPortBytes            = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x02}
	errReachedTheMaxConnectionsBytes       = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x03}
	errHostNumberLimitedBytes              = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x04}
	errHostConflictBytes                   = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x05}
	errHostRegexMismatchBytes              = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x06}
	errDifferentConfigClientConnectedBytes = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x07}
	errReachedMaxOptionsBytes              = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x08}
	errTCPNumberLimited                    = []byte{0xFF, 0xFF, 0xFF, 0xFC, 0x00, 0x09}
	infoTCPPortOpened                      = []byte{0xFF, 0xFF, 0xFF, 0xFB, 0x00, 0x01}
	ServicesBytes                          = []byte{0xFF, 0xFF, 0xFF, 0xFA}
	reconnectBytes                         = []byte{0xFF, 0xFF, 0xFF, 0xF9}
)

// Error represents a specific error signal
type Error uint16

func (e Error) Error() string {
	switch e {
	case ErrInvalidIDAndSecret:
		return "invalid id and secret"
	case ErrFailedToOpenTCPPort:
		return "failed to open tcp port"
	case ErrReachedMaxConnections:
		return "reached the max connections"
	case ErrHostNumberLimited:
		return "host number limited"
	case ErrHostConflict:
		return "host conflict"
	case ErrHostRegexMismatch:
		return "host regex mismatch"
	case ErrDifferentConfigClientConnected:
		return "another client that with different config already connected"
	case ErrReachedMaxOptions:
		return "reached the max options"
	case ErrTCPNumberLimited:
		return "tcp number limited"
	}
	return "unknown error"
}

const (
	_ Error = iota
	// ErrInvalidIDAndSecret represents an invalid ID and secret
	ErrInvalidIDAndSecret
	// ErrFailedToOpenTCPPort represents failed to open tcp port
	ErrFailedToOpenTCPPort
	// ErrReachedMaxConnections represents reached the max connections
	ErrReachedMaxConnections
	// ErrHostNumberLimited represents host number limited
	ErrHostNumberLimited
	// ErrHostConflict represents host conflict
	ErrHostConflict
	// ErrHostRegexMismatch represents host regex mismatch
	ErrHostRegexMismatch
	// ErrDifferentConfigClientConnected represents a client that with different config already connected
	ErrDifferentConfigClientConnected
	// ErrReachedMaxOptions represents reached the max options
	ErrReachedMaxOptions
	// ErrTCPNumberLimited represents tcp number limited
	ErrTCPNumberLimited
)

// Info represents a specific information signal
type Info uint16

const (
	_ Info = iota
	// InfoTCPPortOpened represents TCP port opened successfully
	InfoTCPPortOpened
)

// SendPingSignal sends ping signal to the other side
func (c *Connection) SendPingSignal() (err error) {
	_, err = c.Write(pingBytes)
	return
}

// SendForceCloseSignal sends close signal to the other side
func (c *Connection) SendForceCloseSignal() {
	_, err := c.Write(forceCloseBytes)
	if predef.Debug {
		if err != nil {
			c.Logger.Trace().Err(err).Msg("failed to send close signal")
		}
	}
}

// SendCloseSignal sends close signal to the other side
func (c *Connection) SendCloseSignal() {
	_, err := c.Write(closeBytes)
	if predef.Debug {
		if err != nil {
			c.Logger.Trace().Err(err).Msg("failed to send close signal")
		}
	}
}

// SendReconnectSignal sends reconnect signal to the other side
func (c *Connection) SendReconnectSignal() {
	c.Logger.Info().Msg("sending reconnect signal")
	_, err := c.Write(reconnectBytes)
	if predef.Debug {
		if err != nil {
			c.Logger.Trace().Err(err).Msg("failed to send reconnect signal")
		}
	}
}

// SendReadySignal sends ready signal to the other side
func (c *Connection) SendReadySignal() (err error) {
	_, err = c.Write(readyBytes)
	return
}

// SendServicesSignal sends services signal to the other side
func (c *Connection) SendServicesSignal() (err error) {
	_, err = c.Write(ServicesBytes)
	return
}

// SendErrorSignalInvalidIDAndSecret sends InvalidIDAndSecret signal to the other side
func (c *Connection) SendErrorSignalInvalidIDAndSecret() (err error) {
	_, err = c.Write(errInvalidIDAndSecretBytes)
	return
}

// SendErrorSignalFailedToOpenTCPPort sends FailedToOpenTCPPort signal to the other side
func (c *Connection) SendErrorSignalFailedToOpenTCPPort(si uint16) (err error) {
	buf := pool.BytesPool.Get().([]byte)
	defer pool.BytesPool.Put(buf)
	n := copy(buf, errFailedToOpenTCPPortBytes)
	buf[n] = byte(si >> 8)
	buf[n+1] = byte(si)
	_, err = c.Write(buf[:n+2])
	return
}

// SendInfoTCPPortOpened sends InfoTCPPortOpened signal to the other side
func (c *Connection) SendInfoTCPPortOpened(si uint16, tcpPort uint16) (err error) {
	buf := pool.BytesPool.Get().([]byte)
	defer pool.BytesPool.Put(buf)
	n := copy(buf, infoTCPPortOpened)
	buf[n] = byte(si >> 8)
	buf[n+1] = byte(si)
	buf[n+2] = byte(tcpPort >> 8)
	buf[n+3] = byte(tcpPort)
	_, err = c.Write(buf[:n+4])
	return
}

// SendErrorSignalReachedMaxConnections sends ReachedMaxConnections signal to the other side
func (c *Connection) SendErrorSignalReachedMaxConnections() (err error) {
	_, err = c.Write(errReachedTheMaxConnectionsBytes)
	return
}

// SendErrorSignalHostNumberLimited sends HostNumberLimited signal to the other side
func (c *Connection) SendErrorSignalHostNumberLimited() (err error) {
	_, err = c.Write(errHostNumberLimitedBytes)
	return
}

// SendErrorSignalTCPNumberLimited sends TCPNumberLimited signal to the other side
func (c *Connection) SendErrorSignalTCPNumberLimited() (err error) {
	_, err = c.Write(errTCPNumberLimited)
	return
}

// SendErrorSignalHostConflict sends HostConflict signal to the other side
func (c *Connection) SendErrorSignalHostConflict() (err error) {
	_, err = c.Write(errHostConflictBytes)
	return
}

// SendErrorSignalHostRegexMismatch sends HostRegexMismatch signal to the other side
func (c *Connection) SendErrorSignalHostRegexMismatch() (err error) {
	_, err = c.Write(errHostRegexMismatchBytes)
	return
}

// SendErrorSignalDifferentConfigClientConnected sends DifferentConfigClientConnected signal to the other side
func (c *Connection) SendErrorSignalDifferentConfigClientConnected() (err error) {
	_, err = c.Write(errDifferentConfigClientConnectedBytes)
	return
}

// SendErrorSignalReachedMaxOptions sends ReachedMaxOptions signal to the other side
func (c *Connection) SendErrorSignalReachedMaxOptions() (err error) {
	_, err = c.Write(errReachedMaxOptionsBytes)
	return
}

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

package webrtc

/*
#include <stdlib.h>
#include "datachannel.h"
*/
import "C"

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/mattn/go-pointer"
)

type DataState int

const (
	DataStateConnecting DataState = iota
	DataStateOpen
	DataStateClosing
	DataStateClose
)

func (d DataState) String() string {
	switch d {
	case DataStateConnecting:
		return "connecting"
	case DataStateOpen:
		return "open"
	case DataStateClosing:
		return "closing"
	case DataStateClose:
		return "close"
	}
	panic("unreachable")
}

type DataChannelConfig struct {
	OnStateChange func(state DataState)
	OnMessage     func(message []byte)
}

type DataChannel struct {
	Label string
	ID    int

	dataChannel unsafe.Pointer
	pointerID   unsafe.Pointer
	config      *DataChannelConfig

	bufferedAmountMtx        sync.Mutex
	bufferedAmountChangeCond *sync.Cond

	waiting atomic.Bool
	closed  atomic.Bool
}

//export onDataChannelStateChange
func onDataChannelStateChange(state C.int, id C.int, dataChannel, userData unsafe.Pointer) {
	d, ok := pointer.Restore(userData).(*DataChannel)
	if !ok || d == nil {
		return
	}

	if d.config != nil && d.config.OnStateChange != nil {
		d.ID = int(id)
		d.config.OnStateChange(DataState(state))
	}
}

//export onDataChannelMessage
func onDataChannelMessage(bufC unsafe.Pointer, bufLen C.int, dataChannel, userData unsafe.Pointer) {
	d, ok := pointer.Restore(userData).(*DataChannel)
	if !ok || d == nil {
		return
	}

	if d.dataChannel == nil || d.dataChannel != dataChannel {
		panic("unreachable")
	}

	if d.config != nil && d.config.OnMessage != nil {
		d.config.OnMessage(C.GoBytes(bufC, bufLen))
	}
}

//export onBufferedAmountChange
func onBufferedAmountChange(sentDataSize C.uint64_t, userData unsafe.Pointer) {
	d, ok := pointer.Restore(userData).(*DataChannel)
	if !ok || d == nil {
		return
	}
	if d.waiting.Load() {
		d.bufferedAmountChangeCond.Signal()
	}
}

func (d *DataChannel) SendOnce(message []byte) (sent bool, closed bool) {
	// 等待缓冲区可用空间
	if uint64(len(message)) > d.MaxSendQueueSize()-d.BufferedAmount() {
		d.bufferedAmountMtx.Lock()
		defer d.bufferedAmountMtx.Unlock()
		for {
			d.waiting.Store(true)
			d.bufferedAmountChangeCond.Wait()
			if d.closed.Load() {
				closed = true
				return
			}
			if uint64(len(message)) <= d.MaxSendQueueSize()-d.BufferedAmount() {
				d.waiting.Store(false)
				break
			}
		}
	}

	b := C.DataChannelSend(unsafe.Pointer(&message[0]), C.int(len(message)), d.dataChannel)
	sent = bool(b)
	return
}

func (d *DataChannel) Send(message []byte) bool {
	for i := 0; i < 10; i++ {
		sent, closed := d.SendOnce(message)
		if sent {
			return true
		}
		if closed {
			return false
		}
		if !sent && d.State() != DataStateOpen {
			return false
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func (d *DataChannel) Close() {
	if d.closed.CompareAndSwap(false, true) {
		d.bufferedAmountChangeCond.Signal()
		C.DeleteDataChannel(d.dataChannel)
		pointer.Unref(d.pointerID)
	}
}

func (d *DataChannel) Reliable() bool {
	return bool(C.GetDataChannelReliable(d.dataChannel))
}

func (d *DataChannel) Ordered() bool {
	return bool(C.GetDataChannelOrdered(d.dataChannel))
}

func (d *DataChannel) Protocol() string {
	protocolC := C.GetDataChannelProtocol(d.dataChannel)
	protocol := C.GoString(protocolC)
	C.free(unsafe.Pointer(protocolC))
	return protocol
}

func (d *DataChannel) Negotiated() bool {
	return bool(C.GetDataChannelNegotiated(d.dataChannel))
}

func (d *DataChannel) State() DataState {
	return DataState(C.GetDataChannelState(d.dataChannel))
}

func (d *DataChannel) Error() string {
	errorC := C.GetDataChannelError(d.dataChannel)
	error := C.GoString(errorC)
	C.free(unsafe.Pointer(errorC))
	return error
}

func (d *DataChannel) MessageSent() uint32 {
	return uint32(C.GetDataChannelMessageSent(d.dataChannel))
}

func (d *DataChannel) MessageReceived() uint32 {
	return uint32(C.GetDataChannelMessageReceived(d.dataChannel))
}

func (d *DataChannel) BytesSent() uint64 {
	return uint64(C.GetDataChannelBytesSent(d.dataChannel))
}

func (d *DataChannel) BytesReceived() uint64 {
	return uint64(C.GetDataChannelBytesReceived(d.dataChannel))
}

func (d *DataChannel) BufferedAmount() uint64 {
	return uint64(C.GetDataChannelBufferedAmount(d.dataChannel))
}

func (d *DataChannel) MaxSendQueueSize() uint64 {
	return uint64(C.GetDataChannelMaxSendQueueSize(d.dataChannel))
}

type DataChannelWithoutCallback struct {
	label   string
	id      int
	pointer unsafe.Pointer
}

// SetCallback 使用 dataChannelPointer 是为了防止回调时值还没有被设置，因为是不同的线程
func (d *DataChannelWithoutCallback) SetCallback(config *DataChannelConfig, dataChannelPointer **DataChannel) {
	channel := &DataChannel{
		Label:  d.label,
		ID:     d.id,
		config: config,
	}
	channel.bufferedAmountChangeCond = sync.NewCond(&channel.bufferedAmountMtx)
	*dataChannelPointer = channel
	(*dataChannelPointer).pointerID = pointer.Save(*dataChannelPointer)
	C.SetDataChannelCallback(d.pointer, &(*dataChannelPointer).dataChannel, (*dataChannelPointer).pointerID)
}

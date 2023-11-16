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
#include <stdbool.h>
#include "peerconnection.h"
#include "datachannel.h"
*/
import "C"

import (
	"errors"
	"sync"
	"unsafe"

	"github.com/mattn/go-pointer"
)

type SignalingState int

const (
	SignalingStateStable SignalingState = iota
	SignalingStateHaveLocalOffer
	SignalingStateHaveLocalPrAnswer
	SignalingStateHaveRemoteOffer
	SignalingStateHaveRemotePrAnswer
	SignalingStateClosed
)

func (s SignalingState) String() string {
	switch s {
	case SignalingStateStable:
		return "stable"
	case SignalingStateHaveLocalOffer:
		return "haveLocalOffer"
	case SignalingStateHaveLocalPrAnswer:
		return "haveLocalPrAnswer"
	case SignalingStateHaveRemoteOffer:
		return "haveRemoteOffer"
	case SignalingStateHaveRemotePrAnswer:
		return "haveRemotePrAnswer"
	case SignalingStateClosed:
		return "closed"
	}
	panic("unreachable")
}

type ICEGatheringState int

const (
	ICEGatheringStateNew ICEGatheringState = iota
	ICEGatheringStateGathering
	ICEGatheringStateComplete
)

func (i ICEGatheringState) String() string {
	switch i {
	case ICEGatheringStateNew:
		return "New"
	case ICEGatheringStateGathering:
		return "Gathering"
	case ICEGatheringStateComplete:
		return "Complete"
	}
	panic("unreachable")
}

type ICEConnectionState int

const (
	ICEConnectionStateNew ICEConnectionState = iota
	ICEConnectionStateChecking
	ICEConnectionStateConnected
	ICEConnectionStateCompleted
	ICEConnectionStateFailed
	ICEConnectionStateDisconnected
	ICEConnectionStateClosed
	ICEConnectionStateMax
)

func (i ICEConnectionState) String() string {
	switch i {
	case ICEConnectionStateNew:
		return "New"
	case ICEConnectionStateChecking:
		return "Checking"
	case ICEConnectionStateConnected:
		return "Connected"
	case ICEConnectionStateCompleted:
		return "Completed"
	case ICEConnectionStateFailed:
		return "Failed"
	case ICEConnectionStateDisconnected:
		return "Disconnected"
	case ICEConnectionStateClosed:
		return "Closed"
	case ICEConnectionStateMax:
		return "Max"
	}
	panic("unreachable")
}

type PeerConnectionState int

const (
	PeerConnectionStateNew PeerConnectionState = iota
	PeerConnectionStateConnecting
	PeerConnectionStateConnected
	PeerConnectionStateDisconnected
	PeerConnectionStateFailed
	PeerConnectionStateClosed
)

func (p PeerConnectionState) String() string {
	switch p {
	case PeerConnectionStateNew:
		return "new"
	case PeerConnectionStateConnecting:
		return "connecting"
	case PeerConnectionStateConnected:
		return "connected"
	case PeerConnectionStateDisconnected:
		return "disconnected"
	case PeerConnectionStateFailed:
		return "failed"
	case PeerConnectionStateClosed:
		return "closed"
	}
	panic("unreachable")
}

type PeerConnectionConfig struct {
	ICEServers                        []string
	MinPort                           *uint16
	MaxPort                           *uint16
	OnSignalingChange                 func(state SignalingState)
	OnDataChannel                     func(dataChannelWithoutCallback *DataChannelWithoutCallback)
	OnRenegotiationNeeded             func()
	OnNegotiationNeeded               func()
	OnICEConnectionChange             func(state ICEConnectionState)
	OnStandardizedICEConnectionChange func(state ICEConnectionState)
	OnConnectionChange                func(state PeerConnectionState)
	OnICEGatheringChange              func(state ICEGatheringState)
	OnICECandidate                    func(iceCandidate *ICECandidate)
	OnICECandidateError               func(addrss string, port int, url string, errorCode int, errorText string)
}

// NewPeerConnection 使用 peerConnectionPointer 是为了防止回调时值还没有被设置，因为是不同的线程
func NewPeerConnection(
	config *PeerConnectionConfig,
	peerConnectionPointer **PeerConnection,
	signalingThread unsafe.Pointer,
	networkThread unsafe.Pointer,
	workerThread unsafe.Pointer,
) (err error) {
	iceServers := C.malloc(C.size_t(len(config.ICEServers)) * C.size_t(unsafe.Sizeof(uintptr(0))))
	iceServersPointer := (*[1<<30 - 1]*C.char)(iceServers)
	for i, stun := range config.ICEServers {
		iceServersPointer[i] = C.CString(stun)
	}
	defer func() {
		for i := range config.ICEServers {
			C.free(unsafe.Pointer(iceServersPointer[i]))
		}
		C.free(iceServers)
	}()
	*peerConnectionPointer = &PeerConnection{
		config: config,
		offerChan: make(chan struct {
			offer *SessionDescription
			err   error
		}, 1),
		answerChan: make(chan struct {
			answer *SessionDescription
			err    error
		}, 1),
		localDescriptionErrChan:  make(chan error, 1),
		remoteDescriptionErrChan: make(chan error, 1),
	}
	(*peerConnectionPointer).pointerID = pointer.Save(*peerConnectionPointer)
	errC := C.NewPeerConnection(
		&(*peerConnectionPointer).peerConnection,
		(**C.char)(iceServers),
		C.int(len(config.ICEServers)),
		(*C.uint16_t)(config.MinPort),
		(*C.uint16_t)(config.MaxPort),
		signalingThread,
		networkThread,
		workerThread,
		(*peerConnectionPointer).pointerID,
	)
	if errC != nil {
		err = errors.New(C.GoString(errC))
		C.free(unsafe.Pointer(errC))
	}
	return
}

type PeerConnection struct {
	pointerID      unsafe.Pointer
	peerConnection unsafe.Pointer

	config *PeerConnectionConfig

	offerChan chan struct {
		offer *SessionDescription
		err   error
	}
	answerChan chan struct {
		answer *SessionDescription
		err    error
	}
	localDescriptionErrChan  chan error
	remoteDescriptionErrChan chan error
}

//export onSignalingChange
func onSignalingChange(state C.int, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnSignalingChange != nil {
		p.config.OnSignalingChange(SignalingState(state))
	}
}

//export onDataChannel
func onDataChannel(label *C.char, id C.int, dataChannelWithoutCallback, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnDataChannel != nil {
		d := DataChannelWithoutCallback{
			label:   C.GoString(label),
			id:      int(id),
			pointer: dataChannelWithoutCallback,
		}
		p.config.OnDataChannel(&d)
	}
}

//export onRenegotiationNeeded
func onRenegotiationNeeded(userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnRenegotiationNeeded != nil {
		p.config.OnRenegotiationNeeded()
	}
}

//export onNegotiationNeeded
func onNegotiationNeeded(userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnNegotiationNeeded != nil {
		p.config.OnNegotiationNeeded()
	}
}

//export onICEConnectionChange
func onICEConnectionChange(state C.int, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnICEConnectionChange != nil {
		p.config.OnICEConnectionChange(ICEConnectionState(state))
	}
}

//export onStandardizedICEConnectionChange
func onStandardizedICEConnectionChange(state C.int, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnStandardizedICEConnectionChange != nil {
		p.config.OnStandardizedICEConnectionChange(ICEConnectionState(state))
	}
}

//export onConnectionChange
func onConnectionChange(state C.int, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnConnectionChange != nil {
		p.config.OnConnectionChange(PeerConnectionState(state))
	}
}

//export onICEGatheringChange
func onICEGatheringChange(state C.int, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnICEGatheringChange != nil {
		p.config.OnICEGatheringChange(ICEGatheringState(state))
	}
}

//export onICECandidate
func onICECandidate(sdpMid *C.char, sdpMLineIndex C.int, sdp *C.char, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	i := ICECandidate{
		SDPMid:        C.GoString(sdpMid),
		SDPMLineIndex: int(sdpMLineIndex),
		SDP:           C.GoString(sdp),
	}

	if p.config != nil && p.config.OnICECandidate != nil {
		p.config.OnICECandidate(&i)
	}
}

//export onICECandidateError
func onICECandidateError(
	address *C.char,
	port C.int,
	url *C.char,
	errorCode C.int,
	errorText *C.char,
	userData unsafe.Pointer,
) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	if p.config != nil && p.config.OnICECandidateError != nil {
		p.config.OnICECandidateError(
			C.GoString(address),
			int(port),
			C.GoString(url),
			int(errorCode),
			C.GoString(errorText),
		)
	}
}

func (p *PeerConnection) CreateOffer() (offer *SessionDescription, err error) {
	C.CreateOffer(p.peerConnection)
	v := <-p.offerChan
	return v.offer, v.err
}

//export onOffer
func onOffer(sdp *C.char, errC *C.char, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	s := SessionDescription{
		Type: SDPType.Offer,
		SDP:  C.GoString(sdp),
	}
	var err error
	if errC != nil {
		err = errors.New(C.GoString(errC))
	}
	p.offerChan <- struct {
		offer *SessionDescription
		err   error
	}{
		offer: &s,
		err:   err,
	}
}

func (p *PeerConnection) CreateAnswer() (offer *SessionDescription, err error) {
	C.CreateAnswer(p.peerConnection)
	v := <-p.answerChan
	return v.answer, v.err
}

//export onAnswer
func onAnswer(sdp *C.char, errC *C.char, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	s := SessionDescription{
		Type: SDPType.Answer,
		SDP:  C.GoString(sdp),
	}
	var err error
	if errC != nil {
		err = errors.New(C.GoString(errC))
	}
	p.answerChan <- struct {
		answer *SessionDescription
		err    error
	}{
		answer: &s,
		err:    err,
	}
}

func (p *PeerConnection) SetRemoteDescription(description *SessionDescription) (err error) {
	sdp := C.CString(description.SDP)
	C.SetRemoteDescription(C.int(sdpType2IntMap[description.Type]), sdp, p.peerConnection)
	C.free(unsafe.Pointer(sdp))
	return <-p.remoteDescriptionErrChan
}

//export onSetRemoteDescription
func onSetRemoteDescription(errC *C.char, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	var err error
	if errC != nil {
		err = errors.New(C.GoString(errC))
	}
	p.remoteDescriptionErrChan <- err
}

func (p *PeerConnection) SetLocalDescription(description *SessionDescription) (err error) {
	sdp := C.CString(description.SDP)
	C.SetLocalDescription(C.int(sdpType2IntMap[description.Type]), sdp, p.peerConnection)
	C.free(unsafe.Pointer(sdp))
	return <-p.localDescriptionErrChan
}

//export onSetLocalDescription
func onSetLocalDescription(errC *C.char, userData unsafe.Pointer) {
	p, ok := pointer.Restore(userData).(*PeerConnection)
	if !ok || p == nil {
		return
	}

	var err error
	if errC != nil {
		err = errors.New(C.GoString(errC))
	}
	p.localDescriptionErrChan <- err
}

func (p *PeerConnection) GetRemoteDescription() (description *SessionDescription) {
	var sdpType C.int
	var sdp *C.char
	C.GetRemoteDescription(&sdpType, &sdp, p.peerConnection)
	description = &SessionDescription{
		Type: sdpType2StringMap[int(sdpType)],
		SDP:  C.GoString(sdp),
	}
	C.free(unsafe.Pointer(sdp))
	return
}

func (p *PeerConnection) GetLocalDescription() (description *SessionDescription) {
	var sdpType C.int
	var sdp *C.char
	C.GetLocalDescription(&sdpType, &sdp, p.peerConnection)
	description = &SessionDescription{
		Type: sdpType2StringMap[int(sdpType)],
		SDP:  C.GoString(sdp),
	}
	C.free(unsafe.Pointer(sdp))
	return
}

func (p *PeerConnection) AddICECandidate(candidate *ICECandidate) (err error) {
	sdpMid := C.CString(candidate.SDPMid)
	sdp := C.CString(candidate.SDP)
	errC := C.AddICECandidate(sdpMid, C.int(candidate.SDPMLineIndex), sdp, p.peerConnection)
	C.free(unsafe.Pointer(sdpMid))
	C.free(unsafe.Pointer(sdp))
	if errC != nil {
		err = errors.New(C.GoString(errC))
		C.free(unsafe.Pointer(errC))
	}
	return
}

// CreateDataChannel 使用 dataChannelPointer 是为了防止回调时值还没有被设置，因为是不同的线程
func (p *PeerConnection) CreateDataChannel(label string, negotiated bool, config *DataChannelConfig, dataChannelPointer **DataChannel) (err error) {
	channel := &DataChannel{
		Label:  label,
		config: config,
	}
	channel.bufferedAmountChangeCond = sync.NewCond(&channel.bufferedAmountMtx)
	*dataChannelPointer = channel
	(*dataChannelPointer).pointerID = pointer.Save(*dataChannelPointer)

	labelC := C.CString(label)
	errC := C.CreateDataChannel(
		&(*dataChannelPointer).dataChannel,
		labelC,
		C.bool(negotiated),
		(*dataChannelPointer).pointerID,
		p.peerConnection,
	)
	C.free(unsafe.Pointer(labelC))
	if errC != nil {
		err = errors.New(C.GoString(errC))
		C.free(unsafe.Pointer(errC))
	}
	return
}

func (p *PeerConnection) Close() {
	C.DeletePeerConnection(p.peerConnection)
	pointer.Unref(p.pointerID)
}

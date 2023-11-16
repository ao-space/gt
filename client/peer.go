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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/isrc-cas/gt/client/api"
	"github.com/isrc-cas/gt/client/std"
	"github.com/isrc-cas/gt/client/webrtc"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/util"
	"github.com/rs/zerolog"
)

type peerTask struct {
	Logger                zerolog.Logger
	id                    uint32
	closing               uint32
	initDone              uint32
	state                 uint8
	dataLen               [2]byte
	n                     int
	data                  []byte
	conn                  *webrtc.PeerConnection
	candidateOutChan      chan string
	closeChan             chan struct{}
	tunnel                *conn
	timer                 *time.Timer
	channelID             atomic.Uint32
	channelCount          atomic.Uint32
	apiConn               *api.Conn
	waitNegotiationNeeded chan struct{}
}

func (pt *peerTask) OnSignalingChange(state webrtc.SignalingState) {
	pt.Logger.Info().Str("state", state.String()).Msg("signaling state changed")
}

func (pt *peerTask) OnDataChannel(dataChannelWithoutCallback *webrtc.DataChannelWithoutCallback) {
	observer := &dataChannelObserver{
		peerTask: pt,
		httpTask: util.NewBlockValue[httpTask](),
	}
	config := webrtc.DataChannelConfig{
		OnStateChange: observer.OnStateChange,
		OnMessage:     observer.OnMessage,
	}
	dataChannelWithoutCallback.SetCallback(&config, &observer.dataChannel)
}

func (pt *peerTask) OnRenegotiationNeeded() {
}

func (pt *peerTask) OnNegotiationNeeded() {
	close(pt.waitNegotiationNeeded)
}

func (pt *peerTask) OnICEConnectionChange(state webrtc.ICEConnectionState) {
	pt.Logger.Info().Str("state", state.String()).Msg("ice connection state changed")
}

func (pt *peerTask) OnStandardizedICEConnectionChange(state webrtc.ICEConnectionState) {
	pt.Logger.Info().Str("state", state.String()).Msg("standardized ice connection state changed")
}

func (pt *peerTask) OnConnectionChange(state webrtc.PeerConnectionState) {
	pt.Logger.Info().Str("state", state.String()).Msg("peer connection state changed")
	switch state {
	case webrtc.PeerConnectionStateConnected:
		pt.timer.Stop()
	case webrtc.PeerConnectionStateClosed, webrtc.PeerConnectionStateDisconnected, webrtc.PeerConnectionStateFailed:
		pt.CloseWithLock()
	}
}

func (pt *peerTask) OnICEGatheringChange(state webrtc.ICEGatheringState) {
	pt.Logger.Info().Str("state", state.String()).Msg("ice gathering state changed")
	if state == webrtc.ICEGatheringStateComplete {
		pt.candidateOutChan <- ""
	}
}

func (pt *peerTask) OnICECandidate(iceCandidate *webrtc.ICECandidate) {
	pt.Logger.Info().Interface("candidate", iceCandidate).Msg("OnICECandidate")
	iceCandidateBytes, err := json.Marshal(iceCandidate)
	if err != nil {
		pt.Logger.Error().Err(err).Msg("failed to marshal ice candidate")
		return
	}
	pt.candidateOutChan <- string(iceCandidateBytes)
}

func (pt *peerTask) OnICECandidateError(address string, port int, url string, errorCode int, errorText string) {
	pt.Logger.Error().Str("address", address).Int("port", port).Str("url", url).Int("errorCode", errorCode).Str("errorText", errorText).Msg("failed to handle ice candidate")
}

func (pt *peerTask) init(c *conn) (err error) {
	config := c.client.Config()
	peerConnectionConfig := webrtc.PeerConnectionConfig{
		ICEServers:                        c.stuns,
		MinPort:                           &config.WebRTCMinPort,
		MaxPort:                           &config.WebRTCMaxPort,
		OnSignalingChange:                 pt.OnSignalingChange,
		OnDataChannel:                     pt.OnDataChannel,
		OnRenegotiationNeeded:             pt.OnRenegotiationNeeded,
		OnNegotiationNeeded:               pt.OnNegotiationNeeded,
		OnICEConnectionChange:             pt.OnICEConnectionChange,
		OnStandardizedICEConnectionChange: pt.OnStandardizedICEConnectionChange,
		OnConnectionChange:                pt.OnConnectionChange,
		OnICEGatheringChange:              pt.OnICEGatheringChange,
		OnICECandidate:                    pt.OnICECandidate,
		OnICECandidateError:               pt.OnICECandidateError,
	}
	signalingThread := pt.tunnel.client.webrtcThreadPool.GetThread()
	networkThread := pt.tunnel.client.webrtcThreadPool.GetSocketThread()
	workerThread := pt.tunnel.client.webrtcThreadPool.GetThread()
	err = webrtc.NewPeerConnection(&peerConnectionConfig, &pt.conn, signalingThread, networkThread, workerThread)
	return
}

func (pt *peerTask) Close() {
	if !atomic.CompareAndSwapUint32(&pt.closing, 0, 1) {
		return
	}
	defer pool.BytesPool.Put(pt.data)
	close(pt.closeChan)
	client := pt.tunnel.client
	delete(client.peers, pt.id)
	if pt.conn != nil {
		pt.conn.Close()
	}
	if pt.timer != nil {
		pt.timer.Stop()
	}
	pt.Logger.Info().Msg("peer task closed")
}

func (pt *peerTask) CloseWithLock() {
	if !atomic.CompareAndSwapUint32(&pt.closing, 0, 1) {
		return
	}
	defer pool.BytesPool.Put(pt.data)
	close(pt.closeChan)
	client := pt.tunnel.client
	client.peersRWMtx.Lock()
	delete(client.peers, pt.id)
	client.peersRWMtx.Unlock()
	if pt.conn != nil {
		pt.conn.Close()
	}
	if pt.timer != nil {
		pt.timer.Stop()
	}
	pt.Logger.Info().Msg("peer task closed")
}

func respAndClose(id uint32, c *conn, data [][]byte) {
	var wErr error
	buf := pool.BytesPool.Get().([]byte)
	defer func() {
		if wErr == nil {
			binary.BigEndian.PutUint32(buf[0:], id)
			binary.BigEndian.PutUint16(buf[4:], predef.Close)
			_, wErr = c.Write(buf[:6])
		}
		pool.BytesPool.Put(buf)
		if wErr != nil {
			c.Logger.Debug().AnErr("write err", wErr).Uint32("peerTask", id).Msg("respAndClose err")
			c.Close()
		}
	}()
	binary.BigEndian.PutUint32(buf[0:], id)
	binary.BigEndian.PutUint16(buf[4:], predef.Data)
	l := 0
	s := 0
	for _, d := range data {
		s += len(d)
	}
	if s > len(buf) {
		buf = make([]byte, s)
	}
	dataBuf := buf[10:]
	for _, d := range data {
		l += copy(dataBuf[l:], d)
	}
	if l > 0 {
		binary.BigEndian.PutUint32(buf[6:], uint32(l))
		l += 10

		if predef.Debug {
			c.Logger.Trace().Hex("data", buf[:l]).Msg("write")
		}
		_, wErr = c.Write(buf[:l])
		if wErr != nil {
			return
		}
		remoteTimeout := c.client.Config().RemoteTimeout
		if remoteTimeout.Duration > 0 {
			dl := time.Now().Add(remoteTimeout.Duration)
			wErr = c.Conn.SetReadDeadline(dl)
			if wErr != nil {
				return
			}
		}
	}
}

const (
	dataLength = iota
	dataBody
	processData
)

func (pt *peerTask) process(r io.Reader, writer http.ResponseWriter, initFn func() error) {
	var err error
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		if e := recover(); e != nil {
			pt.Logger.Info().Interface("panic", e).Msg("processOffer panic")
		}
		pt.Logger.Info().Err(err).Msg("processOffer done")
	}()
	reader := std.NewChunkedReader(r)
	for {
		switch pt.state {
		case dataLength:
			var n int
			n, err = reader.Read(pt.dataLen[pt.n:2])
			if n > 0 {
				pt.n += n
			}
			if pt.n != 2 {
				if err != nil {
					if err == io.EOF || err == io.ErrUnexpectedEOF {
						err = nil
						return
					}
					pt.Logger.Error().Err(err).Hex("data", pt.dataLen[:pt.n]).Int("read n", n).Msg("failed to read data")
					return
				}
				continue
			}
			en := uint(pt.dataLen[0])<<8 | uint(pt.dataLen[1])
			if en > pool.MaxBufferSize {
				err = errors.New("dataLen is too long")
				return
			}
			pt.n = 0
			pt.state = dataBody
			pt.Logger.Debug().Int("read n", n).Uint("len", en).Msg("read data length")
			fallthrough
		case dataBody:
			var n int
			en := uint16(pt.dataLen[0])<<8 | uint16(pt.dataLen[1])
			n, err = reader.Read(pt.data[pt.n:en])
			if n > 0 {
				pt.n += n
			}
			if pt.n != int(en) {
				if err != nil {
					pt.Logger.Error().Err(err).Hex("data", pt.data[:pt.n]).Int("read n", n).Uint16("len", en).Msg("failed to read data")
					return
				}
				continue
			}
			pt.state = processData
			fallthrough
		case processData:
			pt.Logger.Debug().Str("data", string(pt.data[:pt.n])).Msg("read json")
			if atomic.CompareAndSwapUint32(&pt.initDone, 0, 1) {
				err = initFn()
				if err != nil {
					pt.Logger.Error().Err(err).Msg("initFn failed")
					return
				}
			} else {
				var candidate webrtc.ICECandidate
				err := json.Unmarshal(pt.data[:pt.n], &candidate)
				if err != nil {
					pt.Logger.Error().Err(err).Msg("failed to Unmarshal candidate")
					return
				}
				err = pt.conn.AddICECandidate(&candidate)
				if err != nil {
					pt.Logger.Error().Err(err).Msg("failed to add ice candidate")
					return
				}
				pt.Logger.Debug().Interface("candidate", candidate).Msg("AddICECandidate")
			}

			pt.n = 0
			pt.state = dataLength
			if err != nil {
				if err == io.EOF {
					err = nil
					return
				}
				pt.Logger.Error().Err(err).Hex("data", pt.data[:pt.n]).Msg("err after processData")
				return
			}
			continue
		}
	}
}

func (pt *peerTask) response(writer http.ResponseWriter, sdp []byte) {
	var err error
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		pt.Logger.Info().Err(err).Msg("response done")
	}()

	pt.Logger.Debug().Bytes("sdp", sdp).Msg("local sdp")
	writer.Header().Add("Transfer-Encoding", "chunked")
	n := uint16(len(sdp))
	l := copy(pt.data, []byte{byte(n >> 8), byte(n)})
	l += copy(pt.data[l:], sdp)
	_, err = writer.Write(pt.data[:l])
	if err != nil {
		return
	}

outLoop:
	for {
		select {
		case candidate, ok := <-pt.candidateOutChan:
			if !ok || len(candidate) == 0 {
				break outLoop
			}
			n := uint16(len(candidate))
			l := copy(pt.data, []byte{byte(n >> 8), byte(n)})
			l += copy(pt.data[l:], candidate)
			_, err = writer.Write(pt.data[:l])
			if err != nil {
				return
			}
			pt.Logger.Debug().Interface("candidate", candidate).Msg("local candidate")
		case <-pt.closeChan:
			break outLoop
		}
	}
}

func (pt *peerTask) processOffer(r *http.Request, writer http.ResponseWriter) {
	pt.process(r.Body, writer, func() (err error) {
		err = pt.init(pt.tunnel)
		if err != nil {
			return
		}

		var offer webrtc.SessionDescription
		err = json.Unmarshal(pt.data[:pt.n], &offer)
		if err != nil {
			return
		}
		err = pt.conn.SetRemoteDescription(&offer)
		if err != nil {
			return
		}
		answer, err := pt.conn.CreateAnswer()
		if err != nil {
			return
		}
		err = pt.conn.SetLocalDescription(answer)
		if err != nil {
			return
		}
		answer = pt.conn.GetLocalDescription()
		answerJSON, err := json.Marshal(&answer)
		if err != nil {
			return
		}
		pt.response(writer, answerJSON)
		return
	})
}

func (pt *peerTask) getOffer(_ *http.Request, writer http.ResponseWriter) {
	c := pt.tunnel

	var err error
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		pt.Logger.Info().Err(err).Msg("getOffer done")
	}()
	err = pt.init(c)
	if err != nil {
		return
	}

	var dataChannelUnused *webrtc.DataChannel
	err = pt.conn.CreateDataChannel("only", true, nil, &dataChannelUnused)
	if err != nil {
		pt.Logger.Error().Err(err).Msg("failed to create only data channel")
		return
	}
	pt.Logger.Info().
		Int("id", dataChannelUnused.ID).
		Str("label", dataChannelUnused.Label).
		Str("state", dataChannelUnused.State().String()).
		Str("error", dataChannelUnused.Error()).
		Uint32("messageSent", dataChannelUnused.MessageSent()).
		Uint32("messageReceived", dataChannelUnused.MessageReceived()).
		Uint64("bytesSent", dataChannelUnused.BytesSent()).
		Uint64("bytesReceived", dataChannelUnused.BytesReceived()).
		Uint64("bufferedAmount", dataChannelUnused.BufferedAmount()).
		Msg("close data channel")
	dataChannelUnused.Close()

	select {
	case <-pt.waitNegotiationNeeded:
	case <-pt.closeChan:
		err = errors.New("closed when <-pt.waitNegotiationNeeded")
		return
	}
	offer, err := pt.conn.CreateOffer()
	if err != nil {
		pt.Logger.Error().Err(err).Msg("failed to create offer")
		return
	}
	err = pt.conn.SetLocalDescription(offer)
	if err != nil {
		pt.Logger.Error().Err(err).Msg("failed to set local description")
		return
	}
	offer = pt.conn.GetLocalDescription()
	offerBytes, err := json.Marshal(&offer)
	if err != nil {
		return
	}
	pt.response(writer, offerBytes)
	_, err = writer.Write([]byte(fmt.Sprintf(`{"id":%d}`, pt.id)))
}

func (pt *peerTask) processAnswer(r *http.Request, writer http.ResponseWriter) {
	c := pt.tunnel
	var err error
	var shouldClose bool
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		pt.Logger.Info().Err(err).Msg("processAnswer done")
		if shouldClose {
			pt.CloseWithLock()
		}
	}()
	task := pt
	idValue := r.Header.Get("WebRTC-OP-ID")
	id, err := strconv.ParseUint(idValue, 10, 32)
	if err != nil {
		return
	}
	if uint32(id) != pt.id {
		shouldClose = true
		client := c.client
		client.peersRWMtx.RLock()
		pt, ok := client.peers[uint32(id)]
		client.peersRWMtx.RUnlock()
		if !ok {
			err = errors.New("invalid task id")
			return
		}
		task = pt
	}
	task.process(r.Body, writer, func() (err error) {
		var answer webrtc.SessionDescription
		err = json.Unmarshal(task.data[:task.n], &answer)
		if err != nil {
			return
		}
		err = task.conn.SetRemoteDescription(&answer)
		if err != nil {
			return
		}
		return
	})
}

type dataChannelObserver struct {
	dataChannel *webrtc.DataChannel
	peerTask    *peerTask
	channelID   uint32
	httpTask    util.BlockValue[httpTask]
}

func (dco *dataChannelObserver) OnOpen() {
	logger := dco.peerTask.Logger.With().Str("label", dco.dataChannel.Label).Uint32("id", dco.channelID).Logger()
	logger.Info().Uint32("channelCount", dco.peerTask.channelCount.Add(1)).Msg("data channel on open")
	dco.peerTask.timer.Stop()

	var err error
	defer func() {
		count := dco.peerTask.channelCount.Add(^uint32(0))
		if count == 0 {
			dco.peerTask.timer.Reset(dco.peerTask.tunnel.client.Config().WebRTCConnectionIdleTimeout.Duration)
		}
		logger.Info().Err(err).Uint32("channelCount", count).
			Str("state", dco.dataChannel.State().String()).
			Str("error", dco.dataChannel.Error()).
			Uint32("messageSent", dco.dataChannel.MessageSent()).
			Uint32("messageReceived", dco.dataChannel.MessageReceived()).
			Uint64("bytesSent", dco.dataChannel.BytesSent()).
			Uint64("bytesReceived", dco.dataChannel.BytesReceived()).
			Uint64("bufferedAmount", dco.dataChannel.BufferedAmount()).
			Msg("data channel closed")
		dco.dataChannel.Close()
	}()

	tunnel := dco.peerTask.tunnel
	service := (*tunnel.services.Load())[0]
	task, err := dco.peerTask.tunnel.dial(&service)
	if err != nil {
		return
	}
	task.Logger = logger
	dco.httpTask.Set(task)
	task.Logger.Info().Msg("p2p http task started")

	var rErr error
	var wErr error
	defer func() {
		task.Logger.Info().AnErr("read err", rErr).AnErr("write err", wErr).Msg("p2p http task read loop returned")
		task.Close()
	}()

	buf := pool.BytesPool.Get().([]byte)
	defer pool.BytesPool.Put(buf)

	for {
		if service.LocalTimeout.Duration > 0 {
			dl := time.Now().Add(service.LocalTimeout.Duration)
			rErr = task.conn.SetReadDeadline(dl)
			if rErr != nil {
				return
			}
		}
		var l int
		l, rErr = task.conn.Read(buf)
		if l > 0 {
			if predef.Debug {
				dco.peerTask.Logger.Trace().Hex("data", buf[:l]).Msg("write")
			}
			if !dco.dataChannel.Send(buf[:l]) {
				dco.peerTask.Logger.Error().Msg("failed to send message")
				return
			}
		}
		if rErr != nil {
			return
		}
	}
}

func (dco *dataChannelObserver) OnStateChange(state webrtc.DataState) {
	dco.peerTask.Logger.Info().Str("state", state.String()).Msg("data channel state changed")
	switch state {
	case webrtc.DataStateOpen:
		dco.channelID = dco.peerTask.channelID.Add(1) // 回调函数是单线程调用的，on open 先于 on message，所以这里不需要同步措施
		go dco.OnOpen()                               // 这里使用 goroutine 是为了避免阻塞 google-webrtc
	case webrtc.DataStateClose:
		task := dco.httpTask.Get()
		if task != nil {
			task.Close()
		}
	}
}

func (dco *dataChannelObserver) OnMessage(message []byte) {
	if predef.Debug {
		logger := dco.peerTask.Logger.With().Str("label", dco.dataChannel.Label).Uint32("id", dco.channelID).Logger()
		logger.Trace().Hex("data", message).Str("text", string(message)).Msg("data channel on msg")
	}
	task := dco.httpTask.Get()
	if task == nil {
		dco.peerTask.Logger.Error().Str("label", dco.dataChannel.Label).Uint32("id", dco.channelID).Msg("OnMessage task is nil")
		return
	}
	_, err := task.Write(message)
	if err != nil && !errors.Is(err, net.ErrClosed) {
		task.Logger.Error().Err(err).Msg("failed to write task conn")
	}
}

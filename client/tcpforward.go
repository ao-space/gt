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
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/isrc-cas/gt/client/std"
	"github.com/isrc-cas/gt/client/webrtc"
	"github.com/isrc-cas/gt/pool"
)

func (c *Client) tcpForwardStart(dialer dialer) {
	// 通过函数的形式隐藏 peerConnection 并发的过程
	getPeerConnection, err := c.createPeerConnections(dialer)
	if err != nil {
		c.Logger.Error().Err(err).Msg("failed to create peer connections")
		return
	}

	// 转发 tcp 的数据
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		conn, err := c.tcpForwardListener.Accept()
		if err != nil {
			if atomic.LoadUint32(&c.closing) > 0 {
				return
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				c.Logger.Error().Err(err).Dur("delay", tempDelay).Msg("Client tcp forward accept error")
				time.Sleep(tempDelay)
				continue
			}
			return
		}

		go func() {
			c.Logger.Info().Msg("tcp forward started")
			defer func() {
				c.Logger.Info().Msg("tcp forward stopped")
				_ = conn.Close()
			}()

			// 创建 dataChannel
			waitStateChange := make(chan struct{}, 1)
			var dataChannel *webrtc.DataChannel
			dataChannelConfig := webrtc.DataChannelConfig{
				OnStateChange: func(state webrtc.DataState) {
					c.Logger.Debug().Str("state", state.String()).Msg("data channel state change")
					select {
					case waitStateChange <- struct{}{}:
					default:
					}
				},
				OnMessage: func(message []byte) {
					_, err := conn.Write(message)
					if err != nil {
						c.Logger.Error().Err(err).Msg("failed to write to conn")
						return
					}
				},
			}
			err = getPeerConnection().CreateDataChannel(conn.RemoteAddr().String(), false, &dataChannelConfig, &dataChannel)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to create data channel")
				return
			}
			defer func() {
				c.Logger.Debug().
					Int("id", dataChannel.ID).
					Str("label", dataChannel.Label).
					Str("state", dataChannel.State().String()).
					Str("error", dataChannel.Error()).
					Uint32("messageSent", dataChannel.MessageSent()).
					Uint32("messageReceived", dataChannel.MessageReceived()).
					Uint64("bytesSent", dataChannel.BytesSent()).
					Uint64("bytesReceived", dataChannel.BytesReceived()).
					Uint64("bufferedAmount", dataChannel.BufferedAmount()).
					Msg("close data channel")
				dataChannel.Close()
			}()

			<-waitStateChange
			buf := pool.BytesPool.Get().([]byte)
			defer pool.BytesPool.Put(buf)
			for {
				nread, err := conn.Read(buf)
				if nread > 0 {
					if !dataChannel.Send(buf[:nread]) {
						c.Logger.Error().Msg("failed to send message with data channel")
						return
					}
				}
				if err != nil {
					if err == io.EOF {
						return
					}
					c.Logger.Error().Err(err).Msg("failed to read from conn")
					return
				}
			}
		}()
	}
}

func (c *Client) createPeerConnections(dialer dialer) (getPeerConnection func() *webrtc.PeerConnection, err error) {
	var peerConnections []*webrtc.PeerConnection
	for i := uint(0); i < c.config.TCPForwardConnections; i++ {
		var peerConnection *webrtc.PeerConnection
		peerConnection, err = c.createPeerConnection(dialer)
		if err != nil {
			return
		}
		peerConnections = append(peerConnections, peerConnection)
	}

	var i atomic.Uint32
	getPeerConnection = func() *webrtc.PeerConnection {
		iLoad := i.Load()
		iLoad %= uint32(len(peerConnections))
		i.Store(iLoad + 1)
		return peerConnections[iLoad]
	}
	return
}

func (c *Client) createPeerConnection(dialer dialer) (peerConnection *webrtc.PeerConnection, err error) {
	// 设置 peerConnection
	candidateDoneChan := make(chan struct{}, 1)
	waitNegotiationNeeded := make(chan struct{}, 1)
	peerConnectionConfig := webrtc.PeerConnectionConfig{
		ICEServers: []string{},
		OnSignalingChange: func(state webrtc.SignalingState) {
			c.Logger.Debug().Str("state", state.String()).Msg("peer connection signaling state change")
			if state == webrtc.SignalingStateClosed {
				select {
				case candidateDoneChan <- struct{}{}:
				default:
				}
				select {
				case waitNegotiationNeeded <- struct{}{}:
				default:
				}
			}
		},
		OnDataChannel: func(dataChannel *webrtc.DataChannelWithoutCallback) {
		},
		OnRenegotiationNeeded: func() {
		},
		OnNegotiationNeeded: func() {
			select {
			case waitNegotiationNeeded <- struct{}{}:
			default:
			}
		},
		OnICEConnectionChange: func(state webrtc.ICEConnectionState) {
		},
		OnStandardizedICEConnectionChange: func(state webrtc.ICEConnectionState) {
		},
		OnConnectionChange: func(state webrtc.PeerConnectionState) {
			c.Logger.Debug().Str("state", state.String()).Msg("peer connection state change")
		},
		OnICEGatheringChange: func(state webrtc.ICEGatheringState) {
			if state == webrtc.ICEGatheringStateComplete {
				select {
				case candidateDoneChan <- struct{}{}:
				default:
				}
			}
			c.Logger.Debug().Str("state", state.String()).Msg("peer connection ice gathering state change")
		},
		OnICECandidate: func(iceCandidate *webrtc.ICECandidate) {
			c.Logger.Debug().Msgf("get peer connection ice candidate: '%v'", iceCandidate)
		},
		OnICECandidateError: func(addrss string, port int, url string, errorCode int, errorText string) {
		},
	}
	err = webrtc.NewPeerConnection(&peerConnectionConfig, &peerConnection)
	if err != nil {
		return
	}
	var dataChannelUnused *webrtc.DataChannel
	err = peerConnection.CreateDataChannel("only", true, nil, &dataChannelUnused)
	if err != nil {
		return
	}
	c.Logger.Debug().
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

	// 发送 offer
	conn, err := dialer.dialFn()
	if err != nil {
		return
	}
	dialFn := func(ctx context.Context, network string, address string) (net.Conn, error) {
		return conn, nil
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialFn,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       5 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	req, err := http.NewRequest("XP", "http://"+c.config.TCPForwardHostPrefix+".example.com", nil)
	if err != nil {
		return
	}
	<-waitNegotiationNeeded
	offer, err := peerConnection.CreateOffer()
	if err != nil {
		return
	}
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		return
	}
	<-candidateDoneChan
	offerBuf := &bytes.Buffer{}
	chunkedWriter := std.NewChunkedWriter(offerBuf)
	offerBytes, err := json.Marshal(peerConnection.GetLocalDescription())
	if err != nil {
		return
	}
	c.Logger.Debug().Msgf("send offer: %s\n", string(offerBytes))
	_, err = chunkedWriter.Write(append([]byte{byte(len(offerBytes) >> 8), byte(len(offerBytes))}, offerBytes...))
	if err != nil {
		return
	}
	req.Body = io.NopCloser(offerBuf)
	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("invalid status code")
		return
	}

	// 获取 answer
	var answerLen uint16
	err = binary.Read(resp.Body, binary.BigEndian, &answerLen)
	if err != nil {
		return
	}
	answerJSON := make([]byte, answerLen)
	_, err = io.ReadFull(resp.Body, answerJSON)
	if err != nil {
		return
	}
	var answer webrtc.SessionDescription
	err = json.Unmarshal(answerJSON, &answer)
	if err != nil {
		return
	}
	c.Logger.Debug().Msgf("get answer: %s\n", string(answerJSON))
	err = peerConnection.SetRemoteDescription(&answer)
	if err != nil {
		return
	}

	// 获取 candidate
	for {
		var candidateLen uint16
		err = binary.Read(resp.Body, binary.BigEndian, &candidateLen)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			err = resp.Body.Close()
			return
		}
		candidateJSON := make([]byte, candidateLen)
		_, err = io.ReadFull(resp.Body, candidateJSON)
		if err != nil {
			err = resp.Body.Close()
			return
		}
		c.Logger.Debug().Msgf("get candidate: %s\n", string(candidateJSON))
		var candidate webrtc.ICECandidate
		err = json.Unmarshal(candidateJSON, &candidate)
		if err != nil {
			return
		}
		err = peerConnection.AddICECandidate(&candidate)
		if err != nil {
			return
		}
	}
}

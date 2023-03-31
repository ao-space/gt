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

package test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/isrc-cas/gt/client/std"
	"github.com/isrc-cas/gt/client/webrtc"
)

func TestP2PGetOffer(t *testing.T) {
	t.Parallel()

	// 创建 HTTP echo 服务
	httpListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	httpEchoServerAddr := httpListener.Addr().String()
	go httpEchoServer(httpListener)

	// 创建客户端、服务端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-stunAddr", "127.0.0.1:0",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "abc",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", fmt.Sprintf("http://%s", httpEchoServerAddr),
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteSTUN", "stun:" + s.GetTURNListenerAddrPort().String(),
		"-logLevel", "debug",
		"-webrtcLogLevel", "warning",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// 仅建立连接
	var peerConnection *webrtc.PeerConnection
	candidateDoneChan := make(chan struct{})
	peerConnectionConfig := webrtc.PeerConnectionConfig{
		ICEServers: []string{},
		OnSignalingChange: func(state webrtc.SignalingState) {
		},
		OnDataChannel: func(dataChannelWithoutCallback *webrtc.DataChannelWithoutCallback) {
		},
		OnRenegotiationNeeded: func() {
		},
		OnNegotiationNeeded: func() {
		},
		OnICEConnectionChange: func(state webrtc.ICEConnectionState) {
		},
		OnStandardizedICEConnectionChange: func(state webrtc.ICEConnectionState) {
		},
		OnConnectionChange: func(state webrtc.PeerConnectionState) {
		},
		OnICEGatheringChange: func(state webrtc.ICEGatheringState) {
			if state == webrtc.ICEGatheringStateComplete {
				close(candidateDoneChan)
			}
		},
		OnICECandidate: func(iceCandidate *webrtc.ICECandidate) {
		},
		OnICECandidateError: func(addrss string, port int, url string, errorCode int, errorText string) {
		},
	}
	err = webrtc.NewPeerConnection(&peerConnectionConfig, &peerConnection)
	if err != nil {
		t.Fatal(err)
	}
	var dataChannelUnused *webrtc.DataChannel
	err = peerConnection.CreateDataChannel("only", true, nil, &dataChannelUnused)
	if err != nil {
		t.Fatal(err)
	}
	dataChannelUnused.Close()

	// 获取 offer
	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	req, err := http.NewRequest("XP", "http://abc.p2p.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("WebRTC-OP", "get-offer")
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	offerLen := make([]byte, 2)
	_, err = io.ReadFull(resp.Body, offerLen)
	if err != nil {
		t.Fatal(err)
	}
	offerBytes := make([]byte, uint16(offerLen[0])<<8|uint16(offerLen[1]))
	_, err = io.ReadFull(resp.Body, offerBytes)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("get offer: %s", string(offerBytes))
	var offer webrtc.SessionDescription
	err = json.Unmarshal(offerBytes, &offer)
	if err != nil {
		t.Fatal(err)
	}
	err = peerConnection.SetRemoteDescription(&offer)
	if err != nil {
		t.Fatal(err)
	}

	// 获取 candidate
	bufReader := bufio.NewReader(resp.Body)
	for {
		candidateLen, err := bufReader.Peek(2)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Equal(candidateLen, []byte("{\"")) {
			break
		}
		_, err = bufReader.Discard(2)
		if err != nil {
			t.Fatal(err)
		}
		candidateBytes := make([]byte, uint16(candidateLen[0])<<8|uint16(candidateLen[1]))
		_, err = io.ReadFull(bufReader, candidateBytes)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("get candidate: %s", string(candidateBytes))
		var candidate webrtc.ICECandidate
		err = json.Unmarshal(candidateBytes, &candidate)
		if err != nil {
			t.Fatal(err)
		}
		err = peerConnection.AddICECandidate(&candidate)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 获取 id
	var id struct {
		ID uint32
	}
	idBytes, err := io.ReadAll(bufReader)
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(idBytes, &id)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("get id: %d", id.ID)
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 发送 answer
	answer, err := peerConnection.CreateAnswer()
	if err != nil {
		t.Fatal(err)
	}
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		t.Fatal(err)
	}
	httpClient = setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	req, err = http.NewRequest("XP", "http://abc.p2p.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("WebRTC-OP", "resp-answer")
	req.Header.Set("WebRTC-OP-ID", strconv.FormatUint(uint64(id.ID), 10))
	<-candidateDoneChan
	answerBuf := &bytes.Buffer{}
	chunkedWriter := std.NewChunkedWriter(answerBuf)
	answerBytes, err := json.Marshal(peerConnection.GetLocalDescription())
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("send answer: %s", string(answerBytes))
	_, err = chunkedWriter.Write([]byte{byte(len(answerBytes) >> 8), byte(len(answerBytes))})
	if err != nil {
		t.Fatal(err)
	}
	_, err = chunkedWriter.Write(answerBytes)
	if err != nil {
		t.Fatal(err)
	}
	req.Body = io.NopCloser(answerBuf)
	resp, err = httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 传输随机数据，验证是否工作正常
	var dataChannel *webrtc.DataChannel
	done := make(chan struct{})
	randomBuf := make([]byte, 1024)
	_, err = rand.Read(randomBuf)
	if err != nil {
		t.Fatal(err)
	}
	dataChannelConfig := webrtc.DataChannelConfig{
		OnStateChange: func(state webrtc.DataState) {
			if state == webrtc.DataStateOpen {
				buf := []byte(fmt.Sprintf("POST / HTTP/1.1\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\nHost: abc.p2p.com\r\n\r\n", len(randomBuf)))
				buf = append(buf, randomBuf...)
				if !dataChannel.Send(buf) {
					t.Fatal("failed to send message with data channel")
				}
				t.Logf("send data: %s", string(buf))

			}
		},
		OnMessage: func(message []byte) {
			t.Logf("OnMessage: %s", string(message))
			bytesReader := bytes.NewReader(message)
			bufReader := bufio.NewReader(bytesReader)
			resp, err := http.ReadResponse(bufReader, nil)
			if err != nil {
				t.Fatal(err)
			}
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(randomBuf, respBody) {
				t.Fatal("invalid http echo server response data")
			}
			close(done)
		},
	}
	err = peerConnection.CreateDataChannel("label", false, &dataChannelConfig, &dataChannel)
	if err != nil {
		t.Fatal(err)
	}
	<-done
}

func httpEchoServer(l net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		_, err = w.Write(buf)
		if err != nil {
			panic(err)
		}
	})
	err := http.Serve(l, mux)
	if err != nil {
		panic(err)
	}
}

func TestP2PSetOffer(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			panic(err)
		}
	})
	httpServer := &http.Server{Handler: mux}
	httpListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := httpServer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		err := httpServer.Serve(httpListener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-stunAddr", "127.0.0.1:0",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "abc",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", fmt.Sprintf("http://%s", httpListener.Addr().String()),
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteSTUN", "stun:" + s.GetTURNListenerAddrPort().String(),
		"-logLevel", "trace",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)

	pc, ctx, offer := initOffer(t, s.GetTURNListenerAddrPort().String())

	req, err := http.NewRequest("XP", "http://abc.p2p.com/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	raw := &bytes.Buffer{}
	b := std.NewChunkedWriter(raw)
	req.Body = io.NopCloser(raw)
	req.ContentLength = -1
	sdp, err := json.Marshal(offer)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("XP sdp: %s", sdp)
	sdpLen := uint16(len(sdp))
	_, err = b.Write(append([]byte{byte(sdpLen >> 8), byte(sdpLen)}, sdp...))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid status code")
	}

	var answer webrtc.SessionDescription
	var dataLen uint16
	err = binary.Read(resp.Body, binary.BigEndian, &dataLen)
	if err != nil {
		t.Fatal(err)
	}
	data := make([]byte, 4096)
	_, err = resp.Body.Read(data[:dataLen])
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("XP sdp: %s", data[:dataLen])
	err = json.Unmarshal(data[:dataLen], &answer)
	if err != nil {
		t.Fatal(err)
	}

	err = pc.SetRemoteDescription(&answer)
	if err != nil {
		t.Fatal(err)
	}

	for {
		err = binary.Read(resp.Body, binary.BigEndian, &dataLen)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		_, err = io.ReadFull(resp.Body, data[:dataLen])
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		t.Logf("XP candidate: %s", data[:dataLen])
		var candidate webrtc.ICECandidate
		err = json.Unmarshal(data[:dataLen], &candidate)
		if err != nil {
			t.Fatal(err)
		}
		err = pc.AddICECandidate(&candidate)
		if err != nil {
			t.Fatal(err)
		}
	}

	<-ctx.Done()
	if ctx.Err() != context.Canceled {
		t.Fatal("invalid context")
	}
	t.Log("XP done")
	s.Shutdown()
}

func initOffer(t *testing.T, addr string) (*webrtc.PeerConnection, context.Context, *webrtc.SessionDescription) {
	waitNegotiationNeeded := make(chan struct{})
	var peerConnection *webrtc.PeerConnection
	candidateDoneChan := make(chan struct{})
	config := webrtc.PeerConnectionConfig{
		ICEServers: []string{
			fmt.Sprintf("stun:%s", addr),
		},
		OnSignalingChange: func(state webrtc.SignalingState) {
		},
		OnDataChannel: func(dataChannelWithoutCallback *webrtc.DataChannelWithoutCallback) {
		},
		OnRenegotiationNeeded: func() {
		},
		OnNegotiationNeeded: func() {
			close(waitNegotiationNeeded)
		},
		OnICEConnectionChange: func(state webrtc.ICEConnectionState) {
			fmt.Println("ice connection", state.String())
		},
		OnStandardizedICEConnectionChange: func(state webrtc.ICEConnectionState) {
		},
		OnConnectionChange: func(state webrtc.PeerConnectionState) {
		},
		OnICEGatheringChange: func(state webrtc.ICEGatheringState) {
			if state == webrtc.ICEGatheringStateComplete {
				close(candidateDoneChan)
			}
		},
		OnICECandidate: func(iceCandidate *webrtc.ICECandidate) {
		},
		OnICECandidateError: func(addrss string, port int, url string, errorCode int, errorText string) {
		},
	}
	var err error
	err = webrtc.NewPeerConnection(&config, &peerConnection)
	if err != nil {
		t.Fatal(err)
	}

	var dataChannelUnused *webrtc.DataChannel
	err = peerConnection.CreateDataChannel("only", true, nil, &dataChannelUnused)
	if err != nil {
		t.Fatal(err)
	}
	dataChannelUnused.Close()
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	go func() {
		var num uint16
		var wg sync.WaitGroup
		wg.Add(10)
		for i := 0; i < 10; i++ {
			num++
			time.Sleep(time.Second)
			var dataChannel *webrtc.DataChannel
			reader, writer := io.Pipe()
			config := webrtc.DataChannelConfig{
				OnStateChange: func(state webrtc.DataState) {
					fmt.Println("data channel", state.String())
					if state == webrtc.DataStateOpen {
						if !dataChannel.Send([]byte("GET / HTTP/1.1\r\nHost: abc.p2p.com\r\n\r\n")) {
							panic("failed to send message with data channel")
						}
					}
				},
				OnMessage: func(message []byte) {
					fmt.Printf("Message from DataChannel %s payload %s\n", dataChannel.Label, string(message))
					_, err := writer.Write(message)
					if err != nil {
						panic(err)
					}
				},
			}
			err = peerConnection.CreateDataChannel(fmt.Sprintf("test%d", num), false, &config, &dataChannel)
			if err != nil {
				panic(err)
			}
			go func() {
				resp, err := http.ReadResponse(bufio.NewReader(reader), nil)
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					panic("invalid http status")
				}
				data, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(err)
				}
				if string(data) != "ok" {
					panic("resp body != ok")
				}
				fmt.Println("sendChannel has received ok")
				wg.Done()
			}()
		}
		wg.Wait()
		cancelFunc()
	}()

	<-waitNegotiationNeeded
	offer, err := peerConnection.CreateOffer()
	if err != nil {
		t.Fatal(err)
	}
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		t.Fatal(err)
	}
	<-candidateDoneChan

	return peerConnection, ctx, peerConnection.GetLocalDescription()
}

func TestTCPForward(t *testing.T) {
	t.Parallel()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-stunAddr", "127.0.0.1:0",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cSlice, err := setupClients(clientOption{
		args: []string{
			"client",
			"-id", "id1",
			"-secret", "secret1",
			"-remote", s.GetListenerAddrPort().String(),
			"-local", "http://www.baidu.com/",
			"-remoteTimeout", "5s",
			"-useLocalAsHTTPHost",
			"-remoteSTUN", "stun:" + s.GetTURNListenerAddrPort().String(),
		},
	}, clientOption{
		args: []string{
			"client",
			"-id", "id2",
			"-secret", "secret2",
			"-remote", s.GetListenerAddrPort().String(),
			"-local", "http://www.baidu.com/",
			"-remoteTimeout", "5s",
			"-useLocalAsHTTPHost",
			"-remoteSTUN", "stun:" + s.GetTURNListenerAddrPort().String(),
			"-tcpForwardAddr", "127.0.0.1:0",
			"-tcpForwardHostPrefix", "id1",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, c := range cSlice {
			c.Close()
		}
	}()

	// 向 client2 的 tcpforward 地址发送 http 请求
	client2TCPForwardAddrPort := cSlice[1].GetTCPForwardListenerAddrPort()
	resp, err := http.Get("http://" + client2TCPForwardAddrPort.String() + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid http status")
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) > 100 {
		all = all[:100]
	}
	t.Logf("%s", all)
}

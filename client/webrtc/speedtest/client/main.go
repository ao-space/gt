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

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/isrc-cas/gt/client/webrtc"
)

func main() {
	webrtc.SetLog(webrtc.LoggingSeverityWarning, func(severity webrtc.LoggingSeverity, message, tag string) {
		fmt.Println("severity", severity.String(), "message", message, "tag", tag)
	})
	go client(100 * 1024 * 1024)
	select {}
}

func client(dataNumber int) {
	waitNegotiationNeeded := make(chan struct{})
	var peerConnection *webrtc.PeerConnection
	var err error
	peerConnectionConfig := webrtc.PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
		OnSignalingChange: func(state webrtc.SignalingState) {
			fmt.Println(state.String())
		},
		OnDataChannel: func(dataChannel *webrtc.DataChannelWithoutCallback) {
		},
		OnRenegotiationNeeded: func() {
		},
		OnNegotiationNeeded: func() {
			close(waitNegotiationNeeded)
		},
		OnICEConnectionChange: func(state webrtc.ICEConnectionState) {
		},
		OnStandardizedICEConnectionChange: func(state webrtc.ICEConnectionState) {
		},
		OnConnectionChange: func(state webrtc.PeerConnectionState) {
			fmt.Println("peer connection", state.String())
		},
		OnICEGatheringChange: func(state webrtc.ICEGatheringState) {
			fmt.Println("ice gathering", state.String())
		},
		OnICECandidate: func(iceCandidate *webrtc.ICECandidate) {
			fmt.Printf("get ice candidate:'%#v'\n", iceCandidate)
		},
		OnICECandidateError: func(addrss string, port int, url string, errorCode int, errorText string) {
		},
	}
	err = webrtc.NewPeerConnection(&peerConnectionConfig, &peerConnection)
	if err != nil {
		panic(err)
	}

	var dataChannel *webrtc.DataChannel
	dataChannelConfig := webrtc.DataChannelConfig{
		OnStateChange: func(state webrtc.DataState) {
			fmt.Println("data channel", state.String())
			if state == webrtc.DataStateOpen {
				buf := make([]byte, 1024)
				_, err := rand.Read(buf)
				if err != nil {
					panic(err)
				}

				startTime := time.Now()
				dataNumberCopy := dataNumber
				for dataNumberCopy > 0 {
					bufLen := len(buf)
					if dataNumberCopy < bufLen {
						bufLen = dataNumberCopy
					}
					if !dataChannel.Send(buf[:bufLen]) {
						panic("data channel send failed")
					}
					dataNumberCopy -= bufLen
				}
				fmt.Printf("all data send done, took %v to send %v bytes, %.2f MB/s\n", time.Since(startTime), dataNumber, float64(dataNumber)/1024/1024/float64(time.Since(startTime)/time.Second))
			}
		},
		OnMessage: func(message []byte) {
		},
	}
	err = peerConnection.CreateDataChannel("speed test", false, &dataChannelConfig, &dataChannel)
	if err != nil {
		panic(err)
	}

	<-waitNegotiationNeeded
	offer, err := peerConnection.CreateOffer()
	if err != nil {
		panic(err)
	}
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		panic(err)
	}
	offer = peerConnection.GetLocalDescription()
	offerJSON, err := json.Marshal(offer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("offer: '%v'\n", string(offerJSON))

	fmt.Print("please enter answer: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answerJSON := scanner.Text()
	var answer webrtc.SessionDescription
	err = json.Unmarshal([]byte(answerJSON), &answer)
	if err != nil {
		panic(err)
	}
	err = peerConnection.SetRemoteDescription(&answer)
	if err != nil {
		panic(err)
	}
}

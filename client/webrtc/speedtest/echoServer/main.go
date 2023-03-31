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
	"os"

	"github.com/isrc-cas/gt/client/webrtc"
)

func main() {
	webrtc.SetLog(webrtc.LoggingSeverityWarning, func(severity webrtc.LoggingSeverity, message, tag string) {
		fmt.Println("severity", severity.String(), "message", message, "tag", tag)
	})
	go echoServer()
	select {}
}

func echoServer() {
	waitICEGatheringComplete := make(chan struct{})
	var peerConnection *webrtc.PeerConnection
	var err error
	peerConnectionConfig := webrtc.PeerConnectionConfig{
		ICEServers: []string{"stun:stun.l.google.com:19302"},
		OnSignalingChange: func(state webrtc.SignalingState) {
			fmt.Println("signaling", state.String())
		},
		OnDataChannel: func(dataChannelWithoutCallback *webrtc.DataChannelWithoutCallback) {
			var dataChannel *webrtc.DataChannel
			dataChannelConfig := webrtc.DataChannelConfig{
				OnStateChange: func(state webrtc.DataState) {
					fmt.Println("data channel", state.String())
				},
				OnMessage: func(message []byte) {
					if !dataChannel.Send(message) {
						panic("data channel send failed")
					}
				},
			}
			dataChannelWithoutCallback.SetCallback(&dataChannelConfig, &dataChannel)
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
			fmt.Println("peer connection", state.String())
		},
		OnICEGatheringChange: func(state webrtc.ICEGatheringState) {
			fmt.Println("ice gathering", state.String())
			if state == webrtc.ICEGatheringStateComplete {
				close(waitICEGatheringComplete)
			}
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

	fmt.Print("please entern offer: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	offerJSON := scanner.Text()
	fmt.Println(offerJSON)
	var offer webrtc.SessionDescription
	err = json.Unmarshal([]byte(offerJSON), &offer)
	if err != nil {
		panic(err)
	}
	err = peerConnection.SetRemoteDescription(&offer)
	if err != nil {
		panic(err)
	}

	answer, err := peerConnection.CreateAnswer()
	if err != nil {
		panic(err)
	}
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}
	<-waitICEGatheringComplete
	answer = peerConnection.GetLocalDescription()
	answerJSON, err := json.Marshal(answer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("answer: '%v'\n", string(answerJSON))
}

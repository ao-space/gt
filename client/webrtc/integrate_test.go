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

package webrtc_test

import (
	"fmt"
	"testing"

	"github.com/isrc-cas/gt/client/webrtc"
)

func TestWebRTC(t *testing.T) {
	webrtc.SetLog(webrtc.LoggingSeverityWarning, func(severity webrtc.LoggingSeverity, message, tag string) {
		fmt.Println("severity", severity.String(), "message", message, "tag", tag)
	})
	offerChan := make(chan *webrtc.SessionDescription)
	answerChan := make(chan *webrtc.SessionDescription)
	exitChan := make(chan struct{})
	go client(offerChan, answerChan)
	go server(offerChan, answerChan, exitChan)
	<-exitChan
}

func client(offerChan, answerChan chan *webrtc.SessionDescription) {
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

	{
		var dataChannel *webrtc.DataChannel
		dataChannelConfig := webrtc.DataChannelConfig{
			OnStateChange: func(state webrtc.DataState) {
				fmt.Println("data channel", state.String())
				if state == webrtc.DataStateOpen {
					if !dataChannel.Send([]byte("test data channel send")) {
						panic("data channel send failed")
					}
				}
			},
			OnMessage: func(message []byte) {
			},
		}
		err = peerConnection.CreateDataChannel("test 1", false, &dataChannelConfig, &dataChannel)
		if err != nil {
			panic(err)
		}
	}
	{
		var dataChannel *webrtc.DataChannel
		dataChannelConfig := webrtc.DataChannelConfig{
			OnStateChange: func(state webrtc.DataState) {
				fmt.Println("data channel", state.String())
				if state == webrtc.DataStateOpen {
					if !dataChannel.Send([]byte("test data channel send")) {
						panic("data channel send failed")
					}
				}
			},
			OnMessage: func(message []byte) {
			},
		}
		err = peerConnection.CreateDataChannel("test 2", false, &dataChannelConfig, &dataChannel)
		if err != nil {
			panic(err)
		}
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
	fmt.Printf("send offer:'%#v'\n", offer)
	offerChan <- offer

	answer := <-answerChan
	err = peerConnection.SetRemoteDescription(answer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("receive answer:'%#v'\n", offer)
}

func server(offerChan, answerChan chan *webrtc.SessionDescription, exitChan chan struct{}) {
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
					if state == webrtc.DataStateOpen {
						if !dataChannel.Send([]byte("test data channel send")) {
							panic("data channel send failed")
						}
					}
				},
				OnMessage: func(message []byte) {
					fmt.Println("server recieve data channel message:", string(message))
					if dataChannel.Label == "test 2" {
						close(exitChan)
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

	offer := <-offerChan
	fmt.Printf("receive offer:'%#v'\n", offer)
	err = peerConnection.SetRemoteDescription(offer)
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
	answerChan <- answer
	fmt.Printf("send answer:'%#v'\n", answer)
}

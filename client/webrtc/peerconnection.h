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

#ifndef PEERCONNECTION_H
#define PEERCONNECTION_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

char *NewPeerConnection(void **peerConnectionOutside, char **iceServers, int iceServersLen,
                        uint16_t *minPort, uint16_t *maxPort, void *userData);
void DeletePeerConnection(void *peerConnection);

void onSignalingChange(int new_state, void *userData);
void onDataChannel(char *label, int id, void *dataChannel, void *userData);
void onRenegotiationNeeded(void *userData);
void onNegotiationNeeded(void *userData);
void onStandardizedICEConnectionChange(int state, void *userData);
void onICECandidateError(char *address, int port, char *url, int errorCode, char *errorText,
                         void *userData);
void onConnectionChange(int state, void *userData);
void onICEGatheringChange(int state, void *userData);
void onICEConnectionChange(int state, void *userData);
void onICECandidate(char *sdpMid, int sdpMLineIndex, char *sdp, void *userData);
char *AddICECandidate(char *sdpMid, int sdpMLineIndex, char *sdp, void *peerConnection);
void CreateOffer(void *peerConnection);
void onOffer(char *offer, char *err, void *userData);
void CreateAnswer(void *peerConnection);
void onAnswer(char *answer, char *err, void *userData);
void SetRemoteDescription(int sdpType, char *sdp, void *peerConnection);
void SetLocalDescription(int sdpType, char *sdp, void *peerConnection);
void GetRemoteDescription(int *sdpType, char **sdp, void *peerConnection);
void GetLocalDescription(int *sdpType, char **sdp, void *peerConnection);
char *CreateDataChannel(void **dataChannel, char *label, bool negotiated, void *dataChannelUserData,
                        void *peerConnectionOutside);
void onSetLocalDescription(char *err, void *userData);
void onSetRemoteDescription(char *err, void *userData);

#ifdef __cplusplus
}
#endif

#endif /* PEERCONNECTION_H */

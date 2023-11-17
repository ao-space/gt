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

#include <iostream>
#include <sstream>

#include <api/create_peerconnection_factory.h>
#include <api/data_channel_interface.h>

#include "datachannel.hpp"
#include "peerconnection.h"

class SetLocalDescriptionObserver : public webrtc::SetLocalDescriptionObserverInterface {
  public:
    SetLocalDescriptionObserver(void *userData) : userData(userData) {}

    void OnSetLocalDescriptionComplete(webrtc::RTCError error) override {
        char *err = nullptr;
        if (!error.ok()) {
            std::stringstream ss;
            ss << "type:'" << ToString(error.type()) << "' message:'" << error.message()
               << "' error_detail:'" << ToString(error.error_detail()) << "'";
            err = (char *)ss.str().c_str();
        }
        ::onSetLocalDescription(err, userData);
    }

  private:
    void *userData;
};

class SetRemoteDescriptionObserver : public webrtc::SetRemoteDescriptionObserverInterface {
  public:
    SetRemoteDescriptionObserver(void *userData) : userData(userData) {}

    void OnSetRemoteDescriptionComplete(webrtc::RTCError error) override {
        char *err = nullptr;
        if (!error.ok()) {
            std::stringstream ss;
            ss << "type:'" << ToString(error.type()) << "' message:'" << error.message()
               << "' error_detail:'" << ToString(error.error_detail()) << "'";
            err = (char *)ss.str().c_str();
        }
        ::onSetRemoteDescription(err, userData);
    }

  private:
    void *userData;
};

class CreateOfferObserver : public webrtc::CreateSessionDescriptionObserver {
  public:
    CreateOfferObserver(void *userData) : userData(userData) {}

  protected:
    void OnSuccess(webrtc::SessionDescriptionInterface *desc) {
        std::string descStr;
        desc->ToString(&descStr);
        onOffer((char *)descStr.c_str(), nullptr, userData);
    }

    void OnFailure(webrtc::RTCError error) {
        if (!error.ok()) {
            std::stringstream ss;
            ss << "type:'" << ToString(error.type()) << "' message:'" << error.message()
               << "' error_detail:'" << ToString(error.error_detail()) << "'";
            onOffer(nullptr, (char *)ss.str().c_str(), userData);
        }
    }

  private:
    void *userData;
};

class CreateAnswerObserver : public webrtc::CreateSessionDescriptionObserver {
  public:
    CreateAnswerObserver(void *userData) : userData(userData) {}

  protected:
    void OnSuccess(webrtc::SessionDescriptionInterface *desc) {
        std::string descStr;
        desc->ToString(&descStr);
        onAnswer((char *)descStr.c_str(), nullptr, userData);
    }

    void OnFailure(webrtc::RTCError error) {
        if (!error.ok()) {
            std::stringstream ss;
            ss << "type:'" << ToString(error.type()) << "' message:'" << error.message()
               << "' error_detail:'" << ToString(error.error_detail()) << "'";
            onAnswer(nullptr, (char *)ss.str().c_str(), userData);
        }
    }

  private:
    void *userData;
};

class PeerConnectionObserver : public webrtc::PeerConnectionObserver {
  public:
    PeerConnectionObserver(void *userData) : userData(userData) {
        createOfferObserver = rtc::make_ref_counted<CreateOfferObserver>(userData);
        createAnswerObserver = rtc::make_ref_counted<CreateAnswerObserver>(userData);
    }

    void Delete() {
        peerConnection->Close();
        peerConnection->Release(); // 在构造 peerConnection 时传递了 this 指针，释放 peerConnection
                                   // 的同时也会释放 this
    }

    char *Start(char **iceServers, int iceServersLen, uint16_t *minPort, uint16_t *maxPort,
                void *signalingThreadOutside, void *networkThreadOutside,
                void *workerThreadOutside) {
        if (signalingThreadOutside == nullptr) {
            ownedSignalingThread = rtc::Thread::Create();
            auto ok = ownedSignalingThread->Start();
            if (!ok) {
                return (char *)"signalingThread start failed";
            }
            signalingThread = ownedSignalingThread.get();
        } else {
            signalingThread = (rtc::Thread *)signalingThreadOutside;
        }

        webrtc::PeerConnectionFactoryDependencies dependencies;
        dependencies.signaling_thread = signalingThread;
        dependencies.network_thread = (rtc::Thread *)networkThreadOutside;
        dependencies.worker_thread = (rtc::Thread *)workerThreadOutside;
        auto peerConnectionFactory =
            webrtc::CreateModularPeerConnectionFactory(std::move(dependencies));

        webrtc::PeerConnectionInterface::RTCConfiguration configuration;
        if (iceServersLen > 0) {
            webrtc::PeerConnectionInterface::IceServer iceServer;
            for (int i = 0; i < iceServersLen; i++) {
                iceServer.urls.push_back(iceServers[i]);
            }
            configuration.servers.push_back(iceServer);
        }
        if (minPort != nullptr && *minPort != 0) {
            configuration.set_min_port(*minPort);
        }
        if (maxPort != nullptr && *maxPort != 0) {
            configuration.set_max_port(*maxPort);
        }

        webrtc::PeerConnectionDependencies connectionDependencies(this);

        auto peerConnectionOrError = peerConnectionFactory->CreatePeerConnectionOrError(
            configuration, std::move(connectionDependencies));
        if (!peerConnectionOrError.ok()) {
            std::stringstream ss;
            ss << "type:'" << ToString(peerConnectionOrError.error().type()) << "' message:'"
               << peerConnectionOrError.error().message() << "' error_detail:'"
               << ToString(peerConnectionOrError.error().error_detail()) << "'";
            auto str = ss.str();
            auto buf = calloc(str.size() + 1, 1);
            memcpy(buf, str.data(), str.size());
            return (char *)buf;
        }
        peerConnection = peerConnectionOrError.MoveValue();

        return nullptr;
    }

    char *CreateDataChannel(void **dataChannelOutside, char *label, bool negotiated,
                            void *dataChannelUserData) {
        char *err = nullptr;
        signalingThread->BlockingCall([&] {
            webrtc::DataChannelInit config;
            config.negotiated = negotiated;
            auto dataChannelOrError = peerConnection->CreateDataChannelOrError(label, &config);
            if (!dataChannelOrError.ok()) {
                std::stringstream ss;
                ss << "type:'" << webrtc::ToString(dataChannelOrError.error().type())
                   << "' message:'" << dataChannelOrError.error().message() << "' error_detail:'"
                   << webrtc::ToString(dataChannelOrError.error().error_detail()) << "'";
                auto str = ss.str();
                auto buf = calloc(str.size() + 1, 1);
                memcpy(buf, str.data(), str.size());
                err = (char *)buf;
                return;
            }
            auto dataChannel = dataChannelOrError.MoveValue();
            auto dataChannelReleased = dataChannel.release();
            auto dataChannelObserver =
                new ::DataChannelObserver(dataChannelReleased, dataChannelUserData);
            *dataChannelOutside = (void *)dataChannelObserver;
            dataChannelReleased->RegisterObserver(dataChannelObserver);
        });
        return err;
    }

    void CreateOffer() {
        signalingThread->BlockingCall([&] {
            webrtc::PeerConnectionInterface::RTCOfferAnswerOptions options;
            peerConnection->CreateOffer(createOfferObserver.get(), options);
        });
    }

    void CreateAnswer() {
        signalingThread->BlockingCall([&] {
            webrtc::PeerConnectionInterface::RTCOfferAnswerOptions options;
            peerConnection->CreateAnswer(createAnswerObserver.get(), options);
        });
    }

    void SetDescription(int isLocal, int sdpType, char *sdp) {
        signalingThread->BlockingCall([&] {
            webrtc::SdpParseError error;
            auto desc = webrtc::CreateSessionDescription((webrtc::SdpType)sdpType, sdp, &error);
            if (desc == nullptr) {
                std::stringstream ss;
                ss << "line:'" << error.line << "' description:'" << error.description << "'";
                if (isLocal) {
                    ::onSetLocalDescription((char *)ss.str().c_str(), userData);
                } else {
                    ::onSetRemoteDescription((char *)ss.str().c_str(), userData);
                }
                return;
            }
            if (isLocal) {
                peerConnection->SetLocalDescription(
                    std::move(desc),
                    rtc::make_ref_counted<::SetLocalDescriptionObserver>(userData));
            } else {
                peerConnection->SetRemoteDescription(
                    std::move(desc),
                    rtc::make_ref_counted<::SetRemoteDescriptionObserver>(userData));
            }
        });
    }

    void GetDescription(int isLocal, int *sdpType, char **sdp) {
        signalingThread->BlockingCall([&] {
            const webrtc::SessionDescriptionInterface *desc;
            if (isLocal) {
                desc = peerConnection->local_description();
            } else {
                desc = peerConnection->remote_description();
            }
            std::string descStr;
            *sdpType = (int)desc->GetType();
            desc->ToString(&descStr);
            *sdp = (char *)calloc(1, descStr.size() + 1);
            memcpy((void *)*sdp, (const void *)descStr.data(), descStr.size());
        });
    }

    char *AddICECandidate(char *sdpMid, int sdpMLineIndex, char *sdp) {
        char *err = nullptr;
        signalingThread->BlockingCall([&] {
            webrtc::SdpParseError sdpParseError;
            auto candidate = CreateIceCandidate(sdpMid, sdpMLineIndex, sdp, &sdpParseError);
            if (candidate == nullptr) {
                std::stringstream ss;
                ss << "line:'" << sdpParseError.line << "' description:'"
                   << sdpParseError.description << "'";
                auto str = ss.str();
                auto buf = calloc(str.size() + 1, 1);
                memcpy(buf, str.data(), str.size());
                err = (char *)buf;
                return;
            }
            peerConnection->AddIceCandidate(candidate);
        });
        return err;
    }

  protected:
    void OnSignalingChange(webrtc::PeerConnectionInterface::SignalingState new_state) {
        ::onSignalingChange((int)new_state, userData);
    }

    void OnDataChannel(rtc::scoped_refptr<webrtc::DataChannelInterface> data_channel) {
        auto dataChannelReleased = data_channel.release();
        ::onDataChannel((char *)dataChannelReleased->label().c_str(), dataChannelReleased->id(),
                        (void *)dataChannelReleased, userData);
    }

    void OnRenegotiationNeeded() { ::onRenegotiationNeeded(userData); }

    void OnNegotiationNeededEvent(uint32_t event_id) {
        if (peerConnection->ShouldFireNegotiationNeededEvent(event_id)) {
            ::onNegotiationNeeded(userData);
        }
    }

    void OnStandardizedIceConnectionChange(
        webrtc::PeerConnectionInterface::IceConnectionState new_state) {
        ::onStandardizedICEConnectionChange((int)new_state, userData);
    }

    void OnIceCandidateError(const std::string &address, int port, const std::string &url,
                             int error_code, const std::string &error_text) {
        ::onICECandidateError((char *)address.c_str(), port, (char *)url.c_str(), error_code,
                              (char *)error_text.c_str(), userData);
    }

    void OnIceConnectionChange(webrtc::PeerConnectionInterface::IceConnectionState new_state) {
        ::onICEConnectionChange(new_state, userData);
    }

    void OnConnectionChange(webrtc::PeerConnectionInterface::PeerConnectionState new_state) {
        ::onConnectionChange(int(new_state), userData);
    }

    void OnIceGatheringChange(webrtc::PeerConnectionInterface::IceGatheringState new_state) {
        ::onICEGatheringChange((int)new_state, userData);
    }

    void OnIceCandidate(const webrtc::IceCandidateInterface *candidate) {
        std::string sdp;
        candidate->ToString(&sdp);
        ::onICECandidate((char *)candidate->sdp_mid().c_str(), candidate->sdp_mline_index(),
                         (char *)sdp.c_str(), userData);
    }

  private:
    rtc::scoped_refptr<webrtc::PeerConnectionInterface> peerConnection;
    rtc::Thread *signalingThread;
    std::unique_ptr<rtc::Thread> ownedSignalingThread;
    rtc::scoped_refptr<CreateOfferObserver> createOfferObserver;
    rtc::scoped_refptr<CreateAnswerObserver> createAnswerObserver;
    void *userData;
};

char *NewPeerConnection(void **peerConnectionOutside, char **iceServers, int iceServersLen,
                        uint16_t *minPort, uint16_t *maxPort, void *signalingThread,
                        void *networkThread, void *workerThread, void *userData) {
    auto peerConnectionObserver = rtc::make_ref_counted<::PeerConnectionObserver>(userData);
    *peerConnectionOutside = (void *)peerConnectionObserver.release();
    auto err = (*(::PeerConnectionObserver **)peerConnectionOutside)
                   ->Start(iceServers, iceServersLen, minPort, maxPort, signalingThread,
                           networkThread, workerThread);
    return err;
}

void DeletePeerConnection(void *peerConnectionOutside) {
    if (peerConnectionOutside == nullptr)
        return;
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    peerConnectionObserver->Delete();
}

char *AddICECandidate(char *sdpMid, int sdpMLineIndex, char *sdp, void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    auto err = peerConnectionObserver->AddICECandidate(sdpMid, sdpMLineIndex, sdp);
    return err;
}

void CreateOffer(void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    peerConnectionObserver->CreateOffer();
}

void CreateAnswer(void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    peerConnectionObserver->CreateAnswer();
}

void SetRemoteDescription(int sdpType, char *sdp, void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    peerConnectionObserver->SetDescription(false, sdpType, sdp);
}

void SetLocalDescription(int sdpType, char *sdp, void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    peerConnectionObserver->SetDescription(true, sdpType, sdp);
}

void GetRemoteDescription(int *sdpType, char **sdp, void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    peerConnectionObserver->GetDescription(false, sdpType, sdp);
}
void GetLocalDescription(int *sdpType, char **sdp, void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    peerConnectionObserver->GetDescription(true, sdpType, sdp);
}

char *CreateDataChannel(void **dataChannel, char *label, bool negotiated, void *dataChannelUserData,
                        void *peerConnectionOutside) {
    auto peerConnectionObserver = (::PeerConnectionObserver *)peerConnectionOutside;
    auto err = peerConnectionObserver->CreateDataChannel(dataChannel, label, negotiated,
                                                         dataChannelUserData);
    return err;
}

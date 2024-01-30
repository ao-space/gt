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

#include <api/data_channel_interface.h>

#include "datachannel.h"
#include "datachannel.hpp"

::DataChannelObserver::DataChannelObserver(webrtc::DataChannelInterface *dataChannel,
                                           void *userData)
    : dataChannel(dataChannel), userData(userData) {}

::DataChannelObserver::~DataChannelObserver() {
    dataChannel->UnregisterObserver();
    dataChannel->Close();
    dataChannel->Release();
}

void ::DataChannelObserver::OnStateChange() {
    onDataChannelStateChange((int)dataChannel->state(), dataChannel->id(), (void *)this, userData);
}

void ::DataChannelObserver::OnMessage(const webrtc::DataBuffer &buffer) {
    onDataChannelMessage((void *)buffer.data.data(), buffer.size(), (void *)this, userData);
}

void ::DataChannelObserver::OnBufferedAmountChange(uint64_t sent_data_size) {
    onBufferedAmountChange(sent_data_size, userData);
}

void DeleteDataChannel(void *dataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)dataChannel;
    delete dataChannelObserverInternal;
}

bool DataChannelSend(void *buf, int bufLen, void *dataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)dataChannel;
    return dataChannelObserverInternal->dataChannel->Send(
        webrtc::DataBuffer(rtc::CopyOnWriteBuffer((char *)buf, (size_t)bufLen), true));
}

void SetDataChannelCallback(void *dataChannelWithoutCallback, void **dataChannelOutside,
                            void *userData) {
    auto dataChannel = (webrtc::DataChannelInterface *)dataChannelWithoutCallback;
    auto dataChannelObserver = new ::DataChannelObserver(dataChannel, userData);
    *dataChannelOutside = (void *)dataChannelObserver;
    dataChannel->RegisterObserver(dataChannelObserver);
}

bool GetDataChannelReliable(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->reliable();
}

bool GetDataChannelOrdered(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->ordered();
}

char *GetDataChannelProtocol(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    char *protocol =
        (char *)calloc(dataChannelObserverInternal->dataChannel->protocol().size() + 1, 1);
    memcpy(protocol, dataChannelObserverInternal->dataChannel->protocol().data(),
           dataChannelObserverInternal->dataChannel->protocol().size());
    return protocol;
}

bool GetDataChannelNegotiated(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->negotiated();
}

int GetDataChannelState(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return (int)dataChannelObserverInternal->dataChannel->state();
}

char *GetDataChannelError(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    char *err = nullptr;
    if (!dataChannelObserverInternal->dataChannel->error().ok()) {
        std::stringstream ss;
        ss << "type:'" << ToString(dataChannelObserverInternal->dataChannel->error().type())
           << "' message:'" << dataChannelObserverInternal->dataChannel->error().message()
           << "' error_detail:'"
           << ToString(dataChannelObserverInternal->dataChannel->error().error_detail()) << "'";
        err = (char *)calloc(ss.str().size() + 1, 1);
        memcpy(err, ss.str().data(), ss.str().size());
    }
    return err;
}

uint32_t GetDataChannelMessageSent(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->messages_sent();
}

uint32_t GetDataChannelMessageReceived(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->messages_received();
}

uint64_t GetDataChannelBytesSent(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->bytes_sent();
}

uint64_t GetDataChannelBytesReceived(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->bytes_received();
}

uint64_t GetDataChannelBufferedAmount(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->buffered_amount();
}

uint64_t GetDataChannelMaxSendQueueSize(void *DataChannel) {
    auto dataChannelObserverInternal = (::DataChannelObserver *)DataChannel;
    return dataChannelObserverInternal->dataChannel->MaxSendQueueSize();
}

#include "stream.h"
#include "quic.hpp"
#include "stream.hpp"

#ifdef _WIN64
#include "stdlib.h"
#endif

Stream::Stream(void *context) : context(context) {}

Stream::Stream(HQUIC stream) : stream(stream) {
    MsQuic->SetCallbackHandler(stream, (void *)redirectCallback, this);
}

Stream::~Stream() {
    if (stream != nullptr) {
        MsQuic->StreamClose(stream);
        MsQuic->SetCallbackHandler(stream, nullptr, nullptr);
        stream = nullptr;
    }
}

bool Stream::Start(HQUIC connection) {
    QUIC_STATUS status = MsQuic->StreamOpen(connection, QUIC_STREAM_OPEN_FLAG_0_RTT,
                                            redirectCallback, this, &stream);
    if (QUIC_FAILED(status)) {
        return false;
    }

    status = MsQuic->StreamStart(stream, QUIC_STREAM_START_FLAG_NONE);
    if (QUIC_FAILED(status)) {
        return false;
    }

    return true;
}

bool Stream::Send(void *data, uint32_t dataLen) {
    sendBuffer.Buffer = (uint8_t *)data;
    sendBuffer.Length = dataLen;
    QUIC_STATUS status =
        MsQuic->StreamSend(stream, &sendBuffer, 1, QUIC_SEND_FLAG_ALLOW_0_RTT, data);
    return QUIC_SUCCEEDED(status);
}

QUIC_STATUS Stream::callback(QUIC_STREAM_EVENT *event) {
    uint32_t bufLen = 0;
    void *buf = nullptr;
    // fprintf(stderr, "stream(%p) event: %s\n", this, StringEvent(event->Type));

    auto status = QUIC_STATUS_SUCCESS;
    switch (event->Type) {
    case QUIC_STREAM_EVENT_START_COMPLETE:
        OnStreamStartComplete(this, context);
        break;
    case QUIC_STREAM_EVENT_SEND_COMPLETE:
        free(event->SEND_COMPLETE.ClientContext);
        OnStreamSendComplete(this, context);
        break;
    case QUIC_STREAM_EVENT_SHUTDOWN_COMPLETE:
        OnStreamShutdownComplete(this, context);
        break;
    case QUIC_STREAM_EVENT_RECEIVE:
        for (uint32_t i = 0; i < event->RECEIVE.BufferCount; ++i) {
            bufLen += event->RECEIVE.Buffers[i].Length;
        }
        buf = malloc(bufLen);
        for (uint32_t i = 0, offset = 0; i < event->RECEIVE.BufferCount; ++i) {
            memcpy((uint8_t *)buf + offset, event->RECEIVE.Buffers[i].Buffer,
                   event->RECEIVE.Buffers[i].Length);
            offset += event->RECEIVE.Buffers[i].Length;
        }
        OnStreamReceive(this, context, buf, bufLen);
        status = QUIC_STATUS_PENDING;
        break;
    default:
        break;
    }
    return status;
}

void Stream::ReceiveComplete(uint64_t bufferLength) {
    MsQuic->StreamReceiveComplete(stream, bufferLength);
}

QUIC_STATUS Stream::redirectCallback(HQUIC stream, void *context, QUIC_STREAM_EVENT *event) {
    return ((Stream *)context)->callback(event);
};

void Stream::SetContext(void *context) { this->context = context; }

void DeleteStream(void *stream) { delete (Stream *)stream; }

void StreamSend(void *stream, void *data, size_t length) { ((Stream *)stream)->Send(data, length); }

void StreamReceiveComplete(void *stream, uint64_t bufferLength) {
    ((Stream *)stream)->ReceiveComplete(bufferLength);
}

void SetStreamContext(void *stream, void *context) { ((Stream *)stream)->SetContext(context); }

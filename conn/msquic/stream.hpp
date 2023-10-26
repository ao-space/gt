#ifndef STREAM_HPP
#define STREAM_HPP

#include <msquic.h>

class Stream {
  public:
    Stream(void *context);
    Stream(HQUIC stream);
    ~Stream();
    bool Start(HQUIC connection);
    bool Send(void *data, uint32_t dataLen);
    void ReceiveComplete(uint64_t bufferLength);
    void SetContext(void *context);

  private:
    HQUIC stream = {};
    void *context = {};
    QUIC_BUFFER sendBuffer = {};

    QUIC_STATUS callback(QUIC_STREAM_EVENT *event);
    static QUIC_STATUS redirectCallback(HQUIC stream, void *context, QUIC_STREAM_EVENT *event);
};

#endif /* STREAM_HPP */
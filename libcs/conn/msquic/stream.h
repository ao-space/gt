#ifndef STREAM_H
#define STREAM_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stddef.h>
#include <stdint.h>

void DeleteStream(void *stream);
void StreamSend(void *stream, void *data, size_t length);
void StreamReceiveComplete(void *stream, uint64_t bufferLength);
void SetStreamContext(void *stream, void *context);

void OnStreamStartComplete(void *stream, void *context);
void OnStreamShutdownComplete(void *stream, void *context);
void OnStreamReceive(void *stream, void *context, void *data, size_t length);
void OnStreamSendComplete(void *stream, void *context);

#ifdef __cplusplus
}
#endif

#endif /* STREAM_H */
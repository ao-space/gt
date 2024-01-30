#ifndef CONNECTION_H
#define CONNECTION_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

void *NewConnection(void *context, char *serverName, uint16_t serverPort, uint64_t idleTimeoutMs,
                    char *certFile, bool unsecure);
void DeleteConnection(void *conn);
void *OpenStream(void *conn, void *context);
void *AcceptStream(void *conn, void *context);
char *GetConnectionAddr(void *conn, bool local);
void SetConnectionContext(void *conn, void *context);
void SetConnectionIdleTimeout(void *conn, uint64_t idleTimeoutMs);

void OnConnectionConnected(void *conn, void *context);
void OnConnectionShutdownComplete(void *conn, void *context);
void OnPeerStreamStarted(void *conn, void *stream, void *context);

#ifdef __cplusplus
}
#endif

#endif /* CONNECTION_H */
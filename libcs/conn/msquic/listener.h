#ifndef LISTENER_H
#define LISTENER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

void *NewListener(char *addr, uint64_t idleTimeoutMs, char *keyFile, char *certFile, char *password,
                  void *context);
void DeleteListener(void *listener);
char *GetListenerAddr(void *listener);

void OnNewConnection(void *listener, void *conn, void *context);
void OnListenerStopComplete(void *listener, void *context);

#ifdef __cplusplus
}
#endif

#endif /* LISTENER_H */
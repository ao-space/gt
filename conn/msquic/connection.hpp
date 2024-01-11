#ifndef CONNECTION_HPP
#define CONNECTION_HPP

#include <msquic.h>

#include "stream.hpp"

class Connection {
  public:
    Connection(void *context);
    Connection(HQUIC connection);
    ~Connection();
    bool Start(char *serverName, uint16_t serverPort, uint64_t IdleTimeoutMs, char *certFile,
               bool unsecure);
    Stream *OpenStream(void *context);
    Stream *AcceptStream(void *context);
    char *GetAddr(bool local);
    void SetContext(void *context);
    void SetIdleTimeout(uint64_t idleTimeoutMs);

  private:
    HQUIC configuration = {};
    HQUIC connection = {};
    void *context = {};
    QUIC_ADDR_STR localAddrStr = {};
    QUIC_ADDR_STR remoteAddrStr = {};
    QUIC_SETTINGS settings = {};

    static QUIC_STATUS redirectCallback(HQUIC connection, void *context,
                                        QUIC_CONNECTION_EVENT *event);
    QUIC_STATUS callback(QUIC_CONNECTION_EVENT *event);
};

#endif /* CONNECTION_HPP */
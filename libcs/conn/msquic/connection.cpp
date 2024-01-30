#include "connection.h"
#include "connection.hpp"
#include "quic.hpp"
#include "stream.hpp"

Connection::Connection(void *context) : context(context) {}

Connection::Connection(HQUIC connection) : connection(connection) {
    MsQuic->SetCallbackHandler(connection, (void *)redirectCallback, this);
}

Connection::~Connection() {
    if (connection != nullptr) {
        MsQuic->ConnectionClose(connection);
        MsQuic->SetCallbackHandler(connection, nullptr, nullptr);
        connection = nullptr;
    }
    if (configuration != nullptr) {
        MsQuic->ConfigurationClose(configuration);
        configuration = nullptr;
    }
}

bool Connection::Start(char *serverName, uint16_t serverPort, uint64_t IdleTimeoutMs,
                       char *certFile, bool unsecure) {
    settings.IdleTimeoutMs = IdleTimeoutMs;
    settings.IsSet.IdleTimeoutMs = true;
    QUIC_STATUS status = MsQuic->ConfigurationOpen(Registration, &ALPN, 1, &settings,
                                                   sizeof(settings), nullptr, &configuration);
    if (QUIC_FAILED(status)) {
        return false;
    }

    QUIC_CREDENTIAL_CONFIG credConfig = {};
    credConfig.Type = QUIC_CREDENTIAL_TYPE_NONE;
    credConfig.Flags = QUIC_CREDENTIAL_FLAG_CLIENT;
    if (strlen(certFile) != 0) {
        // FIXME 按照 msquic 的 docs 描述，应该是这样用的，但是不知道为什么不行
        credConfig.Flags |= QUIC_CREDENTIAL_FLAG_SET_CA_CERTIFICATE_FILE;
        credConfig.CaCertificateFile = certFile;
    }
    if (unsecure) {
        credConfig.Flags |= QUIC_CREDENTIAL_FLAG_NO_CERTIFICATE_VALIDATION;
    }
    status = MsQuic->ConfigurationLoadCredential(configuration, &credConfig);
    if (QUIC_FAILED(status)) {
        return false;
    }

    status = MsQuic->ConnectionOpen(Registration, redirectCallback, this, &connection);
    if (QUIC_FAILED(status)) {
        return false;
    }

    status = MsQuic->ConnectionStart(connection, configuration, QUIC_ADDRESS_FAMILY_UNSPEC,
                                     serverName, serverPort);
    if (QUIC_FAILED(status)) {
        return false;
    }

    return true;
}

Stream *Connection::OpenStream(void *context) {
    auto stream = new Stream(context);
    auto ok = stream->Start(connection);
    if (!ok) {
        delete stream;
        stream = nullptr;
    }
    return stream;
}

Stream *Connection::AcceptStream(void *context) {
    auto stream = new Stream(context);
    // TODO implement stream accept
    auto ok = stream->Start(connection);
    if (!ok) {
        delete stream;
        stream = nullptr;
    }
    return stream;
}

char *Connection::GetAddr(bool local) {
    QUIC_ADDR addr;
    uint32_t addrLen = sizeof(addr);
    auto flag = local ? QUIC_PARAM_CONN_LOCAL_ADDRESS : QUIC_PARAM_CONN_REMOTE_ADDRESS;
    auto addrStr = local ? &localAddrStr : &remoteAddrStr;
    QUIC_STATUS status = MsQuic->GetParam(connection, flag, &addrLen, &addr);
    if (QUIC_FAILED(status)) {
        return nullptr;
    }
    auto ok = QuicAddrToString(&addr, addrStr);
    if (!ok) {
        return nullptr;
    }
    return addrStr->Address;
}

QUIC_STATUS Connection::redirectCallback(HQUIC connection, void *context,
                                         QUIC_CONNECTION_EVENT *event) {
    return ((Connection *)context)->callback(event);
}

QUIC_STATUS Connection::callback(QUIC_CONNECTION_EVENT *event) {
    Stream *stream;
    // fprintf(stderr, "connection(%p) event: %s\n", this, StringEvent(event->Type));

    switch (event->Type) {
    case QUIC_CONNECTION_EVENT_CONNECTED:
        OnConnectionConnected(this, context);
        break;
    case QUIC_CONNECTION_EVENT_SHUTDOWN_INITIATED_BY_TRANSPORT:
        break;
    case QUIC_CONNECTION_EVENT_SHUTDOWN_INITIATED_BY_PEER:
        break;
    case QUIC_CONNECTION_EVENT_SHUTDOWN_COMPLETE:
        OnConnectionShutdownComplete(this, context);
        break;
    case QUIC_CONNECTION_EVENT_PEER_STREAM_STARTED:
        stream = new Stream(event->PEER_STREAM_STARTED.Stream);
        OnPeerStreamStarted(this, stream, context);
        break;
    default:
        break;
    }
    return QUIC_STATUS_SUCCESS;
}

void Connection::SetContext(void *context) { this->context = context; }

void Connection::SetIdleTimeout(uint64_t idleTimeoutMs) {
    settings.IdleTimeoutMs = idleTimeoutMs;
    settings.IsSet.IdleTimeoutMs = true;
    MsQuic->SetParam(connection, QUIC_PARAM_CONN_SETTINGS, sizeof(settings), &settings);
}

void *NewConnection(void *context, char *serverName, uint16_t serverPort, uint64_t idleTimeoutMs,
                    char *certFile, bool unsecure) {
    auto conn = new Connection(context);
    auto ok = conn->Start(serverName, serverPort, idleTimeoutMs, certFile, unsecure);
    if (!ok) {
        delete conn;
        conn = nullptr;
    }
    return conn;
}

void DeleteConnection(void *conn) { delete (Connection *)conn; }

void *OpenStream(void *conn, void *context) { return ((Connection *)conn)->OpenStream(context); }

void *AcceptStream(void *conn, void *context) { return ((Connection *)conn)->AcceptStream(context); }

char *GetConnectionAddr(void *conn, bool local) { return ((Connection *)conn)->GetAddr(local); }

void SetConnectionContext(void *conn, void *context) { ((Connection *)conn)->SetContext(context); }

void SetConnectionIdleTimeout(void *conn, uint64_t idleTimeoutMs) {
    ((Connection *)conn)->SetIdleTimeout(idleTimeoutMs);
}

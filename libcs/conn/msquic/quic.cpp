#include "quic.h"
#include "quic.hpp"

const QUIC_API_TABLE *MsQuic = {};
HQUIC Registration = {};
const QUIC_BUFFER ALPN = {sizeof("default") - 1, (uint8_t *)"default"};

bool Init() {
    QUIC_STATUS status = QUIC_STATUS_SUCCESS;

    if (MsQuic == nullptr) {
        status = MsQuicOpen2((const QUIC_API_TABLE **)&MsQuic);
        if (QUIC_FAILED(status)) {
            goto Error;
        }

        status = MsQuic->RegistrationOpen(nullptr, &Registration);
        if (QUIC_FAILED(status)) {
            goto Error;
        }
    }
    return true;

Error:
    if (MsQuic != nullptr) {
        if (Registration != nullptr) {
            MsQuic->RegistrationClose(Registration);
            Registration = nullptr;
        }
        MsQuicClose(MsQuic);
        MsQuic = nullptr;
    }
    return false;
}

const char *StringStatus(QUIC_STATUS status) {
    switch (status) {
    case QUIC_STATUS_SUCCESS:
        return "success";
    case QUIC_STATUS_PENDING:
        return "pending";
    case QUIC_STATUS_CONTINUE:
        return "continue";
    case QUIC_STATUS_OUT_OF_MEMORY:
        return "out of memory";
    case QUIC_STATUS_INVALID_PARAMETER:
        return "invalid parameter";
    case QUIC_STATUS_INVALID_STATE:
        return "invalid state";
    case QUIC_STATUS_NOT_SUPPORTED:
        return "not supported";
    case QUIC_STATUS_NOT_FOUND:
        return "not found";
    case QUIC_STATUS_BUFFER_TOO_SMALL:
        return "buffer too small";
    case QUIC_STATUS_HANDSHAKE_FAILURE:
        return "handshake failure";
    case QUIC_STATUS_ABORTED:
        return "aborted";
    case QUIC_STATUS_ADDRESS_IN_USE:
        return "address in use";
    case QUIC_STATUS_INVALID_ADDRESS:
        return "invalid address";
    case QUIC_STATUS_CONNECTION_TIMEOUT:
        return "connection timeout";
    case QUIC_STATUS_CONNECTION_IDLE:
        return "connection idle";
    case QUIC_STATUS_INTERNAL_ERROR:
        return "internal error";
    case QUIC_STATUS_CONNECTION_REFUSED:
        return "connection refused";
    case QUIC_STATUS_PROTOCOL_ERROR:
        return "protocol error";
    case QUIC_STATUS_VER_NEG_ERROR:
        return "ver neg error";
    case QUIC_STATUS_UNREACHABLE:
        return "unreachable";
    case QUIC_STATUS_TLS_ERROR:
        return "tls error";
    case QUIC_STATUS_USER_CANCELED:
        return "user canceled";
    case QUIC_STATUS_ALPN_NEG_FAILURE:
        return "alpn neg failure";
    case QUIC_STATUS_STREAM_LIMIT_REACHED:
        return "stream limit reached";
    case QUIC_STATUS_ALPN_IN_USE:
        return "alpn in use";
    case QUIC_STATUS_ADDRESS_NOT_AVAILABLE:
        return "address not available";
    case QUIC_STATUS_CLOSE_NOTIFY:
        return "close notify";
    case QUIC_STATUS_BAD_CERTIFICATE:
        return "bad certificate";
    case QUIC_STATUS_UNSUPPORTED_CERTIFICATE:
        return "unsupported certificate";
    case QUIC_STATUS_REVOKED_CERTIFICATE:
        return "revoked certificate";
    case QUIC_STATUS_EXPIRED_CERTIFICATE:
        return "expired certificate";
    case QUIC_STATUS_UNKNOWN_CERTIFICATE:
        return "unknown certificate";
    case QUIC_STATUS_REQUIRED_CERTIFICATE:
        return "required certificate";
    case QUIC_STATUS_CERT_EXPIRED:
        return "cert expired";
    case QUIC_STATUS_CERT_UNTRUSTED_ROOT:
        return "cert untrusted root";
    case QUIC_STATUS_CERT_NO_CERT:
        return "cert no cert";
    }
    return "unknown";
}

const char *StringEvent(QUIC_STREAM_EVENT_TYPE type) {
    switch (type) {
    case QUIC_STREAM_EVENT_START_COMPLETE:
        return "start complete";
    case QUIC_STREAM_EVENT_RECEIVE:
        return "receive";
    case QUIC_STREAM_EVENT_SEND_COMPLETE:
        return "send complete";
    case QUIC_STREAM_EVENT_PEER_SEND_SHUTDOWN:
        return "peer send shutdown";
    case QUIC_STREAM_EVENT_PEER_SEND_ABORTED:
        return "peer send aborted";
    case QUIC_STREAM_EVENT_PEER_RECEIVE_ABORTED:
        return "peer receive aborted";
    case QUIC_STREAM_EVENT_SEND_SHUTDOWN_COMPLETE:
        return "send shutdown complete";
    case QUIC_STREAM_EVENT_SHUTDOWN_COMPLETE:
        return "shutdown complete";
    case QUIC_STREAM_EVENT_IDEAL_SEND_BUFFER_SIZE:
        return "ideal send buffer size";
    case QUIC_STREAM_EVENT_PEER_ACCEPTED:
        return "peer accepted";
    }
    return "unknown";
}

const char *StringEvent(QUIC_CONNECTION_EVENT_TYPE type) {
    switch (type) {
    case QUIC_CONNECTION_EVENT_CONNECTED:
        return "connected";
    case QUIC_CONNECTION_EVENT_SHUTDOWN_INITIATED_BY_TRANSPORT:
        return "shutdown initiated by transport";
    case QUIC_CONNECTION_EVENT_SHUTDOWN_INITIATED_BY_PEER:
        return "shutdown initiated by peer";
    case QUIC_CONNECTION_EVENT_SHUTDOWN_COMPLETE:
        return "shutdown complete";
    case QUIC_CONNECTION_EVENT_LOCAL_ADDRESS_CHANGED:
        return "local address changed";
    case QUIC_CONNECTION_EVENT_PEER_ADDRESS_CHANGED:
        return "peer address changed";
    case QUIC_CONNECTION_EVENT_PEER_STREAM_STARTED:
        return "peer stream started";
    case QUIC_CONNECTION_EVENT_STREAMS_AVAILABLE:
        return "streams available";
    case QUIC_CONNECTION_EVENT_PEER_NEEDS_STREAMS:
        return "peer needs streams";
    case QUIC_CONNECTION_EVENT_IDEAL_PROCESSOR_CHANGED:
        return "ideal processor changed";
    case QUIC_CONNECTION_EVENT_DATAGRAM_STATE_CHANGED:
        return "datagram state changed";
    case QUIC_CONNECTION_EVENT_DATAGRAM_RECEIVED:
        return "datagram received";
    case QUIC_CONNECTION_EVENT_DATAGRAM_SEND_STATE_CHANGED:
        return "datagram send state changed";
    case QUIC_CONNECTION_EVENT_RESUMED:
        return "resumed";
    case QUIC_CONNECTION_EVENT_RESUMPTION_TICKET_RECEIVED:
        return "resumption ticket received";
    case QUIC_CONNECTION_EVENT_PEER_CERTIFICATE_RECEIVED:
        return "peer certificate received";
    case QUIC_CONNECTION_EVENT_RELIABLE_RESET_NEGOTIATED:
        return "reliable reset negotiated";
    case QUIC_CONNECTION_EVENT_ONE_WAY_DELAY_NEGOTIATED:
        return "one way delay negotiated";
    }
    return "unknown";
}

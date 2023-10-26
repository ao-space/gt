#include "listener.h"
#include "connection.hpp"
#include "quic.hpp"
#include <iostream>

class Listener {
  public:
    Listener(void *context) : context(context) {}

    ~Listener() {
        if (listener != nullptr) {
            MsQuic->ListenerClose(listener);
            MsQuic->SetCallbackHandler(listener, nullptr, nullptr);
            listener = nullptr;
        }
        if (configuration != nullptr) {
            MsQuic->ConfigurationClose(configuration);
            configuration = nullptr;
        }
    }

    bool Start(char *addr, uint64_t idleTimeoutMs, char *keyFile, char *certFile, char *password) {
        QUIC_SETTINGS settings = {};
        settings.IdleTimeoutMs = idleTimeoutMs;
        settings.IsSet.IdleTimeoutMs = true;
        settings.ServerResumptionLevel = QUIC_SERVER_RESUME_AND_ZERORTT;
        settings.IsSet.ServerResumptionLevel = true;
        settings.PeerBidiStreamCount = 1024;
        settings.IsSet.PeerBidiStreamCount = true;
        QUIC_STATUS status = MsQuic->ConfigurationOpen(Registration, &ALPN, 1, &settings,
                                                       sizeof(settings), NULL, &configuration);
        if (QUIC_FAILED(status)) {
            return false;
        }

        QUIC_CREDENTIAL_CONFIG CredConfig = {};
        CredConfig.Flags = QUIC_CREDENTIAL_FLAG_NONE;
        if (strlen(password) == 0) {
            CredConfig.Type = QUIC_CREDENTIAL_TYPE_CERTIFICATE_FILE;
            QUIC_CERTIFICATE_FILE CertificateFile = {};
            CertificateFile.PrivateKeyFile = keyFile;
            CertificateFile.CertificateFile = certFile;
            CredConfig.CertificateFile = &CertificateFile;
            std::cout << "key and cert : " << keyFile << " | " << certFile << std::endl;
        } else {
            CredConfig.Type = QUIC_CREDENTIAL_TYPE_CERTIFICATE_FILE_PROTECTED;
            QUIC_CERTIFICATE_FILE_PROTECTED certFileProtected = {};
            certFileProtected.PrivateKeyFile = keyFile;
            certFileProtected.CertificateFile = certFile;
            certFileProtected.PrivateKeyPassword = password;
            CredConfig.CertificateFileProtected = &certFileProtected;
        }

        status = MsQuic->ConfigurationLoadCredential(configuration, &CredConfig);
        if (QUIC_FAILED(status)) {
            return false;
        }

        auto cb = [](HQUIC listener, void *context, QUIC_LISTENER_EVENT *event) -> QUIC_STATUS {
            return ((Listener *)context)->callback(event);
        };
        status = MsQuic->ListenerOpen(Registration, cb, this, &listener);
        if (QUIC_FAILED(status)) {
            return false;
        }

        QUIC_ADDR quicAddr;
        auto ok = QuicAddrFromString(addr, 0, &quicAddr);
        if (!ok) {
            return false;
        }
        status = MsQuic->ListenerStart(listener, &ALPN, 1, &quicAddr);
        if (QUIC_FAILED(status)) {
            return false;
        }

        return true;
    }

    char *GetAddr() {
        QUIC_ADDR addr;
        uint32_t addrLen = sizeof(addr);
        QUIC_STATUS status =
            MsQuic->GetParam(listener, QUIC_PARAM_LISTENER_LOCAL_ADDRESS, &addrLen, &addr);
        if (QUIC_FAILED(status)) {
            return nullptr;
        }
        auto ok = QuicAddrToString(&addr, &addrStr);
        if (!ok) {
            return nullptr;
        }

        return addrStr.Address;
    }

  private:
    HQUIC configuration = {};
    HQUIC listener = {};
    void *context = {};
    QUIC_ADDR_STR addrStr = {};

    QUIC_STATUS callback(QUIC_LISTENER_EVENT *event) {
        QUIC_STATUS status;
        Connection *conn;

        switch (event->Type) {
        case QUIC_LISTENER_EVENT_NEW_CONNECTION:
            status =
                MsQuic->ConnectionSetConfiguration(event->NEW_CONNECTION.Connection, configuration);
            if (QUIC_FAILED(status)) {
                MsQuic->ConnectionClose(event->NEW_CONNECTION.Connection);
                break;
            }
            conn = new Connection(event->NEW_CONNECTION.Connection);
            OnNewConnection(this, conn, context);
            break;
        case QUIC_LISTENER_EVENT_STOP_COMPLETE:
            break;
        default:
            break;
        }
        return QUIC_STATUS_SUCCESS;
    }
};

void *NewListener(char *addr, uint64_t idleTimeoutMs, char *keyFile, char *certFile, char *password,
                  void *context) {
    auto listener = new Listener(context);
    auto ok = listener->Start(addr, idleTimeoutMs, keyFile, certFile, password);
    if (!ok) {
        delete listener;
        listener = nullptr;
    }
    return listener;
}

void DeleteListener(void *listener) { delete (Listener *)listener; }

char *GetListenerAddr(void *listener) { return ((Listener *)listener)->GetAddr(); }

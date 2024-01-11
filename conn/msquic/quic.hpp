#ifndef QUIC_HPP
#define QUIC_HPP

#include <msquic.h>

extern const QUIC_API_TABLE *MsQuic;
extern HQUIC Registration;
extern const QUIC_BUFFER ALPN;

const char *StringStatus(QUIC_STATUS status);
const char *StringEvent(QUIC_CONNECTION_EVENT_TYPE type);
const char *StringEvent(QUIC_STREAM_EVENT_TYPE type);

#endif /* QUIC_HPP */
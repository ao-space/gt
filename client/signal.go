package client

import (
	connection "github.com/isrc-cas/gt/conn"
)

func handleError(tunnel *conn) (err error) {
	var peekBytes []byte
	peekBytes, err = tunnel.Reader.Peek(2)
	if err != nil {
		return
	}
	code := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
	_, err = tunnel.Reader.Discard(2)
	if err != nil {
		return
	}
	switch connection.Error(code) {
	case connection.ErrInvalidIDAndSecret:
		tunnel.Logger.Error().Str("err", "invalid id and secret").Msg("read error signal")
	case connection.ErrFailedToOpenTCPPort:
		var peekBytes []byte
		peekBytes, err = tunnel.Reader.Peek(2)
		if err != nil {
			return
		}
		serviceIndex := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
		_, err = tunnel.Reader.Discard(2)
		if err != nil {
			return
		}
		var local string
		if s := tunnel.client.services.Load(); s != nil && serviceIndex < uint16(len(*s)) {
			local = (*s)[serviceIndex].LocalURL.String()
		}
		tunnel.Logger.Error().
			Str("local", local).
			Str("err", "failed to open tcp port").
			Msg("read error signal")
	case connection.ErrReachedMaxConnections:
		tunnel.Logger.Error().Str("err", "reached the max connections").Msg("read error signal")
	case connection.ErrHostNumberLimited:
		tunnel.Logger.Error().Str("err", "the number of host prefixes exceeded the upper limit").Msg("read error signal")
	case connection.ErrHostConflict:
		tunnel.Logger.Error().Str("err", "host conflict").Msg("read error signal")
	case connection.ErrHostRegexMismatch:
		tunnel.Logger.Error().Str("err", "host regex mismatch").Msg("read error signal")
	case connection.ErrDifferentConfigClientConnected:
		tunnel.Logger.Error().Str("err", "another client that with different config already connected").Msg("read error signal")
	case connection.ErrReachedMaxOptions:
		tunnel.Logger.Error().Str("err", "the number of options exceeded the upper limit").Msg("read error signal")
	case connection.ErrTCPNumberLimited:
		tunnel.Logger.Error().Str("err", "the number of tcp ports exceeded the upper limit").Msg("read error signal")
	default:
		tunnel.Logger.Error().Str("err", "unknown error").Msg("read error signal")
	}
	return
}

func handleInfo(tunnel *conn) (err error) {
	var peekBytes []byte
	peekBytes, err = tunnel.Reader.Peek(2)
	if err != nil {
		return
	}
	code := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
	_, err = tunnel.Reader.Discard(2)
	if err != nil {
		return
	}
	switch connection.Info(code) {
	case connection.InfoTCPPortOpened:
		peekBytes, err = tunnel.Reader.Peek(2)
		if err != nil {
			return
		}
		serviceIndex := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
		_, err = tunnel.Reader.Discard(2)
		if err != nil {
			return
		}
		peekBytes, err = tunnel.Reader.Peek(2)
		if err != nil {
			return
		}
		tcpPort := uint16(peekBytes[1]) | uint16(peekBytes[0])<<8
		_, err = tunnel.Reader.Discard(2)
		if err != nil {
			return
		}
		var local string
		if s := tunnel.client.services.Load(); s != nil && serviceIndex < uint16(len(*s)) {
			local = (*s)[serviceIndex].LocalURL.String()
		}
		tunnel.Logger.Info().Uint16("serviceIndex", serviceIndex).
			Str("local", local).
			Uint16("tcp port", tcpPort).
			Msg("tcp port opened")
	default:
		tunnel.Logger.Info().Msg("read unknown info signal")
	}
	return
}

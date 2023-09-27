package conn

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/quic-go/quic-go"
	"math/big"
	"net"
)

type QuicConnection struct {
	quic.Connection
	quic.Stream
}

type QuicListener struct {
	quic.Listener
}

var _ net.Conn = &QuicConnection{}
var _ net.Listener = &QuicListener{}

func QuicDial(addr string, config *tls.Config) (net.Conn, error) {
	config.NextProtos = []string{"gt-quic"}
	conn, err := quic.DialAddr(context.Background(), addr, config, &quic.Config{})
	if err != nil {
		panic(err)
	}
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		panic(err)
	}
	nc := &QuicConnection{
		Connection: conn,
		Stream:     stream,
	}
	return nc, err
}

func QuicListen(addr string, config *tls.Config) (net.Listener, error) {
	config.NextProtos = []string{"gt-quic"}
	listener, err := quic.ListenAddr(addr, config, &quic.Config{})
	if err != nil {
		panic(err)
	}
	ln := &QuicListener{
		Listener: *listener,
	}
	return ln, err
}

func (ln *QuicListener) Accept() (net.Conn, error) {
	conn, _ := ln.Listener.Accept(context.Background())
	stream, err := conn.AcceptStream(context.Background())
	nc := &QuicConnection{
		Connection: conn,
		Stream:     stream,
	}
	return nc, err
}

func GenerateTLSConfig() *tls.Config {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &ecdsaKey.PublicKey, ecdsaKey)
	if err != nil {
		panic(err)
	}
	keyBytes, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: keyBytes})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"gt-quic"},
	}
}

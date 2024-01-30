package conn

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/isrc-cas/gt/predef"
	"github.com/quic-go/quic-go"
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
	conn, err := quic.DialAddr(context.Background(), addr, config, &quic.Config{EnableDatagrams: true})
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
	listener, err := quic.ListenAddr(addr, config, &quic.Config{EnableDatagrams: true})
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

const ProbeTimes = 10

func GetQuicProbesResults(addr string, timeout time.Duration) (avgRtt float64, pktLoss float64, err error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	conn, err := QuicDial(addr, tlsConfig)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()
	sendBuffer := []byte{predef.MagicNumber, 0x02, 0x00}
	_, err = conn.Write(sendBuffer)
	if err != nil {
		return
	}

	var totalSuccessNum int64
	var totalDelay int64
	var buf []byte
	for i := 0; i < ProbeTimes; i++ {
		bs := [9]byte{}
		bs[0] = byte(i)
		now := time.Now().UnixMilli()
		binary.BigEndian.PutUint64(bs[1:], uint64(now))
		err = conn.(*QuicConnection).SendDatagram(bs[:])
		if err != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		buf, err = conn.(*QuicConnection).ReceiveDatagram(ctx)
		cancel()
		if err != nil {
			return
		}
		if len(buf) >= 9 {
			no := buf[0]
			u := binary.BigEndian.Uint64(buf[1:])
			interval := time.Now().Sub(time.UnixMilli(int64(u))).Milliseconds()
			totalSuccessNum += 1
			totalDelay += interval
			if no == ProbeTimes-1 {
				break
			}
		}
	}
	avgRtt = float64(totalDelay) / (float64(ProbeTimes))
	pktLoss = 1 - float64(totalSuccessNum)/float64(ProbeTimes)
	_, err = conn.Write([]byte{0xFF})
	return
}

package conn

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	quicbbr "github.com/DrakenLibra/gt-bbr"
	"github.com/isrc-cas/gt/predef"
	"github.com/quic-go/quic-go"
	"math/big"
	"net"
	"sync/atomic"
	"time"
)

type QuicConnection struct {
	quic.Connection
	quic.Stream
}

type QuicBbrConnection struct {
	quicbbr.Session
	quicbbr.Stream
}

type QuicListener struct {
	quic.Listener
}

type QuicBbrListener struct {
	quicbbr.Listener
}

var _ net.Conn = &QuicConnection{}
var _ net.Listener = &QuicListener{}
var _ net.Conn = &QuicBbrConnection{}
var _ net.Listener = &QuicBbrListener{}

func (c *QuicBbrConnection) Close() error {
	err := c.Stream.Close()
	err = c.Session.Close()
	return err
}

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

func QuicBbrDial(addr string, config *tls.Config) (net.Conn, error) {
	config.NextProtos = []string{"gt-quic"}
	conn, err := quicbbr.DialAddr(addr, config, &quicbbr.Config{})
	if err != nil {
		panic(err)
	}
	stream, err := conn.OpenStreamSync()
	if err != nil {
		panic(err)
	}
	nc := &QuicBbrConnection{
		Session: conn,
		Stream:  stream,
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

func QuicBbrListen(addr string, config *tls.Config) (net.Listener, error) {
	config.NextProtos = []string{"gt-quic"}
	listener, err := quicbbr.ListenAddr(addr, config, &quicbbr.Config{})
	if err != nil {
		panic(err)
	}
	ln := &QuicBbrListener{
		Listener: listener,
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

func (ln *QuicBbrListener) Accept() (net.Conn, error) {
	conn, _ := ln.Listener.Accept()
	stream, err := conn.AcceptStream()
	nc := &QuicBbrConnection{
		Session: conn,
		Stream:  stream,
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

func GetQuicProbesResults(addr string) (avgRtt float64, pktLoss float64, err error) {
	totalNum := 100
	var totalSuccessNum int64 = 0
	var totalDelay int64 = 0
	var buf []byte
	probeCloseError := &quic.ApplicationError{
		Remote:       false,
		ErrorCode:    0x42,
		ErrorMessage: "close QUIC probe connection",
	}
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true

	conn, err := QuicDial(addr, tlsConfig)
	if err != nil {
		return
	}
	sendBuffer := []byte{predef.MagicNumber, 0x02}
	_, err = conn.Write(sendBuffer)
	if err != nil {
		return
	}

	for i := 0; i < totalNum; i++ {
		go func() {
			err = conn.(*QuicConnection).SendMessage([]byte(time.Now().Format("2006-01-02 15:04:05.000000000")))
			if err != nil {
				return
			}
		}()
	}

	for {
		timer := time.AfterFunc(3*time.Second, func() {
			err = conn.(*QuicConnection).CloseWithError(0x42, "close QUIC probe connection")
			if err != nil {
				return
			}
		})
		buf, err = conn.(*QuicConnection).ReceiveMessage()
		if err != nil {
			// QUIC的stream关闭时会返回io.EOF，但是QUIC的不可靠数据包Datagram是在connection层面进行发送的
			// 因此需要通过quic.ApplicationError判断QUIC connection是否由于应用程序主动关闭
			if err.Error() == probeCloseError.Error() {
				err = nil
				break
			} else {
				return
			}
		}
		if buf != nil {
			sendTine, _ := time.ParseInLocation("2006-01-02 15:04:05.000000000", string(buf), time.Local)
			interval := time.Now().Sub(sendTine).Microseconds()
			atomic.AddInt64(&totalSuccessNum, 1)
			atomic.AddInt64(&totalDelay, interval)
		}
		timer.Stop()
	}

	avgRtt = float64(atomic.LoadInt64(&totalDelay)) / (float64(1000 * totalNum))
	pktLoss = 1 - float64(atomic.LoadInt64(&totalSuccessNum))/float64(totalNum)
	return
}

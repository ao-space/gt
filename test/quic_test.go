package test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestQuic(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
		if err != nil {
			panic(err)
		}
		if request.FormValue("hello") != "world" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = writer.Write([]byte("ok"))
		if err != nil {
			panic(err)
		}
	})
	hs := &http.Server{Handler: mux}
	l, err := net.Listen("tcp", "127.0.0.1:12080")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := hs.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		err := hs.Serve(l)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:8080",
		"-quicAddr", "127.0.0.1:10080",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-timeout", "10s",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", "http://" + l.Addr().String(),
		"-remote", fmt.Sprintf("quic://%v", s.GetQuicListenerAddrPort()),
		"-remoteTimeout", "5s",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.OnTunnelClose.Store(func() {
		panic("tunnel should not be closed")
	})

	conn, err := net.Dial("tcp", s.GetListenerAddrPort().String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write([]byte("GET "))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(12 * time.Second)

	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	resp, err := httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/test?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid status code")
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(all) != "ok" {
		t.Fatal("invalid resp")
	}
	c.OnTunnelClose.Store(func() {})
	t.Logf("%s", all)
	s.Shutdown()
}

func TestQuicBbr(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(writer http.ResponseWriter, request *http.Request) {
		err := request.ParseForm()
		if err != nil {
			panic(err)
		}
		if request.FormValue("hello") != "world" {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = writer.Write([]byte("ok"))
		if err != nil {
			panic(err)
		}
	})
	hs := &http.Server{Handler: mux}
	l, err := net.Listen("tcp", "127.0.0.1:12080")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := hs.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		err := hs.Serve(l)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:8080",
		"-quicAddr", "127.0.0.1:10080",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-timeout", "10s",
		"-bbr",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", "http://" + l.Addr().String(),
		"-remote", fmt.Sprintf("quic://%v", s.GetQuicListenerAddrPort()),
		"-remoteTimeout", "5s",
		"-bbr",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	c.OnTunnelClose.Store(func() {
		panic("tunnel should not be closed")
	})

	conn, err := net.Dial("tcp", s.GetListenerAddrPort().String())
	if err != nil {
		t.Fatal(err)
	}
	_, err = conn.Write([]byte("GET "))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(12 * time.Second)

	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	resp, err := httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/test?hello=world")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid status code")
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(all) != "ok" {
		t.Fatal("invalid resp")
	}
	c.OnTunnelClose.Store(func() {})
	t.Logf("%s", all)
	s.Shutdown()
}

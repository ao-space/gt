// Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/buger/jsonparser"
	"github.com/gorilla/websocket"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/server"
)

// 多线程安全
type safeBuffer struct {
	buf     *bytes.Buffer
	rwMutex *sync.RWMutex
}

func (sb *safeBuffer) Write(p []byte) (int, error) {
	sb.rwMutex.Lock()
	defer sb.rwMutex.Unlock()

	return sb.buf.Write(p)
}

func newStringWriter() (io.Writer, func() string) {
	buf := new(bytes.Buffer)
	rwMutex := &sync.RWMutex{}

	return &safeBuffer{buf: buf, rwMutex: rwMutex}, func() string {
		rwMutex.RLock()
		defer rwMutex.RUnlock()

		return buf.String()
	}
}

func setupServer(args []string, out io.Writer) (s *server.Server, err error) {
	s, err = server.New(args, out)
	if err != nil {
		return
	}
	err = s.Start()
	return
}

func setupClient(args []string, out io.Writer) (c *client.Client, err error) {
	c, err = client.New(args, out)
	if err != nil {
		return
	}
	err = c.Start()
	if err != nil {
		return
	}
	err = c.WaitUntilReady(30 * time.Second)
	if err != nil {
		return
	}
	time.Sleep(500 * time.Millisecond) // 等待 0.5 秒让 tunnel started 进入日志
	return
}

type clientOption struct {
	args []string
	out  io.Writer
}

func setupClients(cOptions ...clientOption) (cSlice []*client.Client, err error) {
	var wg sync.WaitGroup
	var errPointer atomic.Pointer[error]
	cSlice = make([]*client.Client, len(cOptions))
	for i := range cOptions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, err := setupClient(cOptions[i].args, cOptions[i].out)
			cSlice[i] = c
			if err != nil {
				errPointer.Store(&err)
			}
		}(i)
	}
	wg.Wait()
	if errPointer.Load() != nil {
		err = *errPointer.Load()
	}
	return
}

func setupHTTPClient(addr string, tlsConfig *tls.Config) *http.Client {
	dialFn := func(ctx context.Context, network string, address string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext(ctx, network, addr)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:       tlsConfig,
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialFn,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       5 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	return httpClient
}

func TestServerAndClient(t *testing.T) {
	t.Parallel()
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
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
		"-local", "http://www.baidu.com",
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	resp, err := httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com")
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
	t.Logf("%s", all)
}

func TestClientAndServerWithLocalServer(t *testing.T) {
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

	l, err := net.Listen("tcp", "127.0.0.1:0")
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
		"-addr", "127.0.0.1:0",
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
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
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
	t.Logf("%s", all)
	s.Shutdown()
}

func TestClientAndServerWithLocalWebsocket(t *testing.T) {
	t.Parallel()
	upgrader := websocket.Upgrader{} // use default options

	echo := func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}
		defer func(c *websocket.Conn) {
			err := c.Close()
			if err != nil {
				panic(err)
			}
		}(c)
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNoStatusReceived) {
					return
				}
				panic(err)
			}
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				panic(err)
			}
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/test", echo)
	hs := &http.Server{Handler: mux}

	l, err := net.Listen("tcp", "127.0.0.1:0")
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
		"-addr", "127.0.0.1:0",
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
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	dialFn := func(ctx context.Context, network string, address string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext(ctx, network, s.GetListenerAddrPort().String())
	}
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		NetDialContext:   dialFn,
	}
	ws, _, err := dialer.Dial("ws://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/test", nil)
	if err != nil {
		t.Fatal("dial:", err)
	}

	done := make(chan struct{})
	msg := make(chan string, 1)

	go func() {
		defer close(done)
		for i := 0; i < 3; i++ {
			_, message, err := ws.ReadMessage()
			if err != nil {
				panic(err)
			}
			m := <-msg
			if m != string(message) {
				panic("not equal")
			}
			log.Printf("client recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

FOR:
	for {
		select {
		case <-done:
			err := ws.WriteControl(websocket.CloseMessage, nil, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			err = ws.Close()
			if err != nil {
				t.Fatal(err)
			}
			break FOR
		case tick := <-ticker.C:
			ts := tick.String()
			err := ws.WriteMessage(websocket.TextMessage, []byte(ts))
			if err != nil {
				t.Fatal(err)
			}
			msg <- ts
		}
	}

	s.Shutdown()
}

func TestPing(t *testing.T) {
	t.Parallel()
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
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
		"-local", "http://www.baidu.com",
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	time.Sleep(20 * time.Second)
	if s.GetTunneling() != 1 {
		t.Fatal("zero tunneling?!")
	}
	s.Shutdown()
}

func TestAPIStatus(t *testing.T) {
	t.Parallel()
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-apiAddr", "127.0.0.1:0",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	httpClient := setupHTTPClient(s.GetAPIListenerAddrPort().String(), nil)
	resp, err := httpClient.Get("http://status.example.com/status") // 只要路径是 /status 就行
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid status code")
	}
	resp1, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", resp1)

	time.Sleep(time.Second)
	resp, err = httpClient.Get("http://status.example.com/status")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid status code")
	}
	resp2, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", resp2)

	if !bytes.Equal(resp1, resp2) {
		t.Fatalf("resp1(%s) != resp2(%s)", resp1, resp2)
	}
	s.Shutdown()
}

func TestAuthAPI(t *testing.T) {
	t.Parallel()

	// 模拟 AuthAPI
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		all, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		id, _ := jsonparser.GetString(all, "networkClientId")
		secret, _ := jsonparser.GetString(all, "networkSecretKey")
		if id != "05797ac9-86ae-40b0-b767-7a41e03a5486" || secret != "eec1eabf-2c59-4e19-bf10-34707c17ed89" {
			panic("invalid id or secret")
		}
		_, err = rw.Write([]byte("{\"result\":true}"))
		if err != nil {
			panic(err)
		}
	})
	httpServer := http.Server{
		Handler: mux,
	}
	httpLisener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := httpServer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		err := httpServer.Serve(httpLisener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-authAPI", "http://" + httpLisener.Addr().String() + "/",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", "http://www.baidu.com",
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// 通过 http 测试
	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	resp, err := httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com")
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
	t.Logf("%s", all)
}

func TestRemoteAPI(t *testing.T) {
	t.Parallel()

	// 启动服务端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// 模拟 RemoteAPI
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("Request-Id")
		if requestID == "" {
			panic("invalid Request-Id")
		}

		id := r.URL.Query().Get("network_client_id")
		if id != "05797ac9-86ae-40b0-b767-7a41e03a5486" {
			panic("invalid id")
		}

		_, err := rw.Write([]byte("{\"serverAddress\":\"tcp://" + s.GetListenerAddrPort().String() + "\"}"))
		if err != nil {
			panic(err)
		}
	})
	httpServer := http.Server{
		Handler: mux,
	}
	httpLisener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := httpServer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		err := httpServer.Serve(httpLisener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	// 启动客户端
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", "http://www.baidu.com",
		"-remoteAPI", "http://" + httpLisener.Addr().String() + "/",
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// 通过 http 测试
	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	resp, err := httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com")
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
	t.Logf("%s", all)
}

func TestAutoSecret(t *testing.T) {
	t.Parallel()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-local", "http://www.baidu.com/",
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// 通过 http 测试
	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	resp, err := httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com")
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
	t.Logf("%s", all)
}

func TestSNI(t *testing.T) {
	t.Parallel()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "",
		"-sniAddr", "127.0.0.1:0",
		"-id", "www",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-timeout", "10s",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "www",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", "https://www.baidu.com",
		"-remote", s.GetSNIListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// 通过 https 测试
	httpClient := setupHTTPClient(s.GetSNIListenerAddrPort().String(), &tls.Config{})
	resp, err := httpClient.Get("https://www.baidu.com")
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
	t.Logf("%s", all)
}

func TestClientAndServerWithHTTPMUXHeader(t *testing.T) {
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

	l, err := net.Listen("tcp", "127.0.0.1:0")
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
		"-addr", "127.0.0.1:0",
		"-httpMUXHeader", "EID",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", fmt.Sprintf("http://%s", l.Addr().String()),
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	req, err := http.NewRequest("GET", "http://abc.example.com/test?hello=world", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header["EID"] = []string{"05797ac9-86ae-40b0-b767-7a41e03a5486"}
	resp, err := httpClient.Do(req)
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
	t.Logf("%s", all)
	s.Shutdown()
}

func TestTCP(t *testing.T) {
	t.Parallel()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-tcpRange", "1024-65535",
		"-tcpNumber", "1",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	client1LogWriter, client1Log := newStringWriter()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", "tcp://www.baidu.com:80",
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTCPRandom",
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, client1LogWriter)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	time.Sleep(100 * time.Millisecond) // 等待服务端完成 TCP 端口分配

	// 从客户端的日志中获取 tcp 端口
	match := regexp.MustCompile(`tcp port=(\d+)`).FindStringSubmatch(client1Log())
	if len(match) != 2 {
		t.Fatal("failed to get tcp port from client log")
	}
	tcpPort := match[1]

	// 通过 tcp 测试
	resp, err := http.Get("http://localhost:" + tcpPort + "/")
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
	if len(all) > 100 {
		all = all[:100]
	}
	t.Logf("%s", all)
}

func TestSpeedLimit(t *testing.T) {
	t.Parallel()

	// 启动 http 服务
	mux := http.NewServeMux()
	waitUpload := make(chan struct{}, 1)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			_, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			waitUpload <- struct{}{}
		case "GET":
			_, err := w.Write(make([]byte, 4096))
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	httpServer := http.Server{
		Handler: mux,
	}
	httpLisener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := httpServer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		err := httpServer.Serve(httpLisener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-speed", "1024",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-secret", "eec1eabf-2c59-4e19-bf10-34707c17ed89",
		"-local", "http://" + httpLisener.Addr().String() + "/",
		"-remote", s.GetListenerAddrPort().String(),
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)

	// 上传 4096 字节的内容所需时间应该在 4 到 5 秒
	buffer := bytes.NewBuffer(make([]byte, 4096))
	startTime := time.Now()
	resp, err := httpClient.Post("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/", "application/octet-stream", buffer)
	if err != nil {
		t.Fatal(err)
	}
	<-waitUpload
	intervals := time.Since(startTime)
	if intervals < 4*time.Second || intervals > 5*time.Second {
		t.Fatalf("intervals: %v, intervals < 4*time.Second || intervals > 5*time.Second", intervals)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 下载 4096 字节的内容所需时间应该在 4 到 5 秒
	startTime = time.Now()
	resp, err = httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	intervals = time.Since(startTime)
	if intervals < 4*time.Second || intervals > 5*time.Second {
		t.Fatalf("intervals: %v, intervals < 4*time.Second || intervals > 5*time.Second", intervals)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 等待 3 秒后再次测试
	time.Sleep(3 * time.Second)

	// 上传 4096 字节的内容所需时间应该在 4 到 5 秒
	buffer = bytes.NewBuffer(make([]byte, 4096))
	startTime = time.Now()
	resp, err = httpClient.Post("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/", "application/octet-stream", buffer)
	if err != nil {
		t.Fatal(err)
	}
	<-waitUpload
	intervals = time.Since(startTime)
	if intervals < 4*time.Second || intervals > 5*time.Second {
		t.Fatalf("intervals: %v, intervals < 4*time.Second || intervals > 5*time.Second", intervals)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// 下载 4096 字节的内容所需时间应该在 4 到 5 秒
	startTime = time.Now()
	resp, err = httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	intervals = time.Since(startTime)
	if intervals < 4*time.Second || intervals > 5*time.Second {
		t.Fatalf("intervals: %v, intervals < 4*time.Second || intervals > 5*time.Second", intervals)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidIDOrSecret(t *testing.T) {
	t.Parallel()

	// 启动服务端、客户端
	client1LogWriter, client1Log := newStringWriter()
	client2LogWriter, client2Log := newStringWriter()
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-id", "id1",
		"-secret", "secret1",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cSlice, err := setupClients(clientOption{
		args: []string{ // ID 不存在
			"client", "-id=id2", "-secret=secret2", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost",
		},
		out: client1LogWriter,
	}, clientOption{
		args: []string{ // secret 错误
			"client", "-id=id1", "-secret=secret2", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost",
		},
		out: client2LogWriter,
	})
	if err == nil {
		t.Fatal("expect err not nil")
	}
	defer func() {
		for _, c := range cSlice {
			c.Close()
		}
	}()

	// client1 失败
	if !strings.Contains(client1Log(), "invalid id and secret") {
		t.Fatal("client1 not failed")
	}

	// client2 失败
	if !strings.Contains(client2Log(), "invalid id and secret") {
		t.Fatal("client2 not failed")
	}
}

func TestConnectionsLimit(t *testing.T) {
	t.Parallel()

	// 启动 http 服务
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		if err != nil {
			t.Fatal(err)
		}
	})
	httpServer := http.Server{
		Handler: mux,
	}
	httpLisener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := httpServer.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	go func() {
		err := httpServer.Serve(httpLisener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// 启动服务端、客户端
	client1LogWriter, client1Log := newStringWriter()
	client2LogWriter, client2Log := newStringWriter()
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-connections", "3",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cSlice, err := setupClients(clientOption{
		args: []string{ // 成功
			"client",
			"-id", "id1",
			"-secret", "secret1",
			"-remote", s.GetListenerAddrPort().String(),
			"-local", "http://" + httpLisener.Addr().String(),
			"-remoteConnections", "3",
			"-remoteIdleConnections", "3",
			"-remoteTimeout", "5s",
			"-reconnectDelay", "24h", // 不重连
			"-useLocalAsHTTPHost",
		},
		out: client1LogWriter,
	}, clientOption{
		args: []string{ // 前 3 个 tunnel 成功，后 2 个失败
			"client",
			"-id", "id2",
			"-secret", "secret2",
			"-remote", s.GetListenerAddrPort().String(),
			"-local", "http://" + httpLisener.Addr().String(),
			"-remoteConnections", "5",
			"-remoteIdleConnections", "5",
			"-remoteTimeout", "5s",
			"-reconnectDelay", "24h", // 不重连
			"-useLocalAsHTTPHost",
		},
		out: client2LogWriter,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, c := range cSlice {
			c.Close()
		}
	}()

	// 10 秒内分别向 2 个客户端均匀地发送 1000 个 HTTP 请求，这一步是为了检测通信过程中的 data race
	var resps sync.Map
	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	for i := 0; i < 1000; i++ {
		go func() {
			resp, err := httpClient.Get("http://id1.example.com/")
			if err != nil && !errors.Is(err, io.EOF) {
				panic(err)
			}
			if resp != nil {
				resps.Store(resp, nil)
			}
		}()
		go func() {
			resp, err := httpClient.Get("http://id2.example.com/")
			if err != nil && !errors.Is(err, io.EOF) {
				panic(err)
			}
			if resp != nil {
				resps.Store(resp, nil)
			}
		}()
		time.Sleep(10 * time.Second / 1000)
	}
	resps.Range(func(key, value interface{}) bool {
		err := key.(*http.Response).Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		resps.Delete(key)
		return true
	})

	// client1 成功
	if strings.Count(client1Log(), "tunnel started") != 3 {
		t.Fatal("client1 not successful")
	}

	// client2 失败
	if strings.Count(client2Log(), "reached the max connections") != 2 {
		t.Fatal("client2 not failed")
	}
}

func TestReconnectLimit(t *testing.T) {
	t.Parallel()

	// 启动服务端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-id", "id1",
		"-secret", "secret1",
		"-reconnectDuration", "6s",
		"-reconnectTimes", "0",
		"-logLevel", "debug",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// client1 失败
	client1LogWriter, client1Log := newStringWriter()
	c1, err := setupClient([]string{
		"client",
		"-id", "id1",
		"-secret", "secret2",
		"-remote", s.GetListenerAddrPort().String(),
		"-local", "http://www.baidu.com/",
		"-remoteTimeout", "5s",
		"-reconnectDelay", "1s",
		"-useLocalAsHTTPHost",
	},
		client1LogWriter)
	if err == nil {
		t.Fatal("expect err not nil")
	}

	// 客户端在 30 秒的时间内不断重试（大约 30 次），但只有 5 次有错误信号的返回，其他的请求都被忽略了
	if strings.Count(client1Log(), "invalid id and secret") != 5 {
		t.Log("client1Log", client1Log())
		t.Fatal("client1 not failed")
	}

	c1.Close()

	// client2 一开始仍然处于被限制的状态，连接失败，但 reconnectDuration 过后成功
	client2LogWriter, client2Log := newStringWriter()
	c2, err := setupClient([]string{
		"client",
		"-id", "id1",
		"-secret", "secret1",
		"-remote", s.GetListenerAddrPort().String(),
		"-local", "http://www.baidu.com/",
		"-remoteTimeout", "5s",
		"-reconnectDelay", "1s",
		"-useLocalAsHTTPHost",
	}, client2LogWriter)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Count(client2Log(), "tunnel started") != 1 {
		t.Log("client2Log", client2Log())
		t.Fatal("client2 not successful")
	}

	c2.Close()
}

func TestHostNumber(t *testing.T) {
	t.Parallel()

	// 生成配置文件
	err := os.WriteFile("test_host_number_server.yaml", []byte(`
users:
  id1:
    secret: secret1
  id2:
    secret: secret2
  id3:
    secret: secret3
    host:
      number: 2
  id4:
    secret: secret4
    host:
      number: 2
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove("test_host_number_server.yaml")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// 启动服务端、客户端
	client1LogWriter, client1Log := newStringWriter()
	client2LogWriter, client2Log := newStringWriter()
	client3LogWriter, client3Log := newStringWriter()
	client4LogWriter, client4Log := newStringWriter()
	s, err := setupServer([]string{
		"server",
		"-config", "test_host_number_server.yaml",
		"-addr", "127.0.0.1:0",
		"-hostNumber", "1",
		"-hostWithID",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cSlice, err := setupClients(clientOption{
		args: []string{
			"client", "-id=id1", "-secret=secret1", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=1",
		},
		out: client1LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id2", "-secret=secret2", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=1",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=2",
		},
		out: client2LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id3", "-secret=secret3", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=1",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=2",
		},
		out: client3LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id4", "-secret=secret4", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=1",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=2",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=3",
		},
		out: client4LogWriter,
	})
	defer func() {
		for _, c := range cSlice {
			c.Close()
		}
	}()

	// client1 成功
	if !strings.Contains(client1Log(), "tunnel started") {
		t.Fatal("client1 not successful")
	}

	// client2 失败
	if !strings.Contains(client2Log(), "the number of host prefixes exceeded the upper limit") {
		t.Fatal("client2 not failed")
	}

	// client3 成功
	if !strings.Contains(client3Log(), "tunnel started") {
		t.Fatal("client3 not successful")
	}

	// client4 失败
	if !strings.Contains(client4Log(), "the number of host prefixes exceeded the upper limit") {
		t.Fatal("client4 not failed")
	}
}

func TestHostRegex(t *testing.T) {
	t.Parallel()

	// 生成配置文件
	err := os.WriteFile("test_host_regex_server.yaml", []byte(`
users:
  id1:
    secret: secret1
  id2:
    secret: secret2
  id3:
    secret: secret3
    host:
      regex:
        - ^[a-z]+$
  id4:
    secret: secret4
    host:
      regex:
        - ^[a-z]+$
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove("test_host_regex_server.yaml")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// 启动服务端、客户端
	client1LogWriter, client1Log := newStringWriter()
	client2LogWriter, client2Log := newStringWriter()
	client3LogWriter, client3Log := newStringWriter()
	client4LogWriter, client4Log := newStringWriter()
	s, err := setupServer([]string{
		"server",
		"-config", "test_host_regex_server.yaml",
		"-addr", "127.0.0.1:0",
		"-hostRegex", "^[0-9]+$",
		"-hostNumber", "2",
		"-hostWithID",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cSlice, err := setupClients(clientOption{
		args: []string{
			"client", "-id=id1", "-secret=secret1", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=1",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=2",
		},
		out: client1LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id2", "-secret=secret2", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=a",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=b",
		},
		out: client2LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id3", "-secret=secret3", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=a",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=b",
		},
		out: client3LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id4", "-secret=secret4", "-remote", s.GetListenerAddrPort().String(),
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=1",
			"-local=http://www.baidu.com/", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-hostPrefix=2",
		},
		out: client4LogWriter,
	})
	if err == nil {
		t.Fatal("expect err not nil")
	}
	defer func() {
		for _, c := range cSlice {
			c.Close()
		}
	}()

	// client1 成功
	if !strings.Contains(client1Log(), "tunnel started") {
		t.Fatal("client1 not successful")
	}

	// client2 失败
	if !strings.Contains(client2Log(), "host regex mismatch") {
		t.Fatal("client2 not failed")
	}

	// client3 成功
	if !strings.Contains(client3Log(), "tunnel started") {
		t.Fatal("client3 not successful")
	}

	// client4 失败
	if !strings.Contains(client4Log(), "host regex mismatch") {
		t.Fatal("client4 not failed")
	}
}

func TestTCPNumberAndTCPRange(t *testing.T) {
	t.Parallel()

	// 生成配置文件
	err := os.WriteFile("test_tcp_number_server.yaml", []byte(`
users:
 id1:
   secret: secret1
 id2:
   secret: secret2
 id3:
   secret: secret3
 id4:
   secret: secret4
   tcp:
     - range: 41000-42000
     - range: 42001-43000
 id5:
   secret: secret5
   tcp:
     - range: 51001-51999
 id6:
   secret: secret6
   tcp:
     - range: 50000-51000
     - range: 52000-60000
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove("test_tcp_number_server.yaml")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// 启动服务端、客户端
	client1LogWriter, client1Log := newStringWriter()
	client2LogWriter, client2Log := newStringWriter()
	client3LogWriter, client3Log := newStringWriter()
	client4LogWriter, client4Log := newStringWriter()
	client5LogWriter, client5Log := newStringWriter()
	client6LogWriter, client6Log := newStringWriter()
	s, err := setupServer([]string{
		"server",
		"-config", "test_tcp_number_server.yaml",
		"-addr", "127.0.0.1:0",
		"-tcpNumber", "2",
		"-tcpRange", "10000-20000",
		"-tcpRange", "20000-30000",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	cSlice, err := setupClients(clientOption{
		args: []string{
			"client", "-id=id1", "-secret=secret1", "-remote", s.GetListenerAddrPort().String(),
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=10000",
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=20000",
		},
		out: client1LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id2", "-secret=secret2", "-remote", s.GetListenerAddrPort().String(),
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=10001",
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=10002",
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=30000",
		},
		out: client2LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id3", "-secret=secret3", "-remote", s.GetListenerAddrPort().String(),
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=9999",
		},
		out: client3LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id4", "-secret=secret4", "-remote", s.GetListenerAddrPort().String(),
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=42000",
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=43000",
		},
		out: client4LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id5", "-secret=secret5", "-remote", s.GetListenerAddrPort().String(),
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=60001",
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=60002",
		},
		out: client5LogWriter,
	}, clientOption{
		args: []string{
			"client", "-id=id6", "-secret=secret6", "-remote", s.GetListenerAddrPort().String(),
			"-local=tcp://www.baidu.com:80", "-remoteTimeout=5s", "-useLocalAsHTTPHost", "-remoteTCPPort=20001",
		},
		out: client6LogWriter,
	})
	if err == nil {
		t.Fatal("expect err not nil")
	}
	defer func() {
		for _, c := range cSlice {
			c.Close()
		}
	}()

	// client1 成功
	if !strings.Contains(client1Log(), "tunnel started") {
		t.Fatal("client1 not successful")
	}

	// client2 失败
	if !strings.Contains(client2Log(), "the number of tcp ports exceeded the upper limit") {
		t.Fatal("client2 not failed")
	}

	// client3 失败
	if !strings.Contains(client3Log(), "failed to open tcp port") {
		t.Fatal("client3 not failed")
	}

	// client4 成功
	if !strings.Contains(client4Log(), "tunnel started") {
		t.Fatal("client4 not successful")
	}

	// client5 失败
	if !strings.Contains(client5Log(), "failed to open tcp port") {
		t.Fatal("client5 not failed")
	}

	// client6 失败
	if !strings.Contains(client6Log(), "failed to open tcp port") {
		t.Fatal("client6 not failed")
	}
}

func TestHostPrefixConflict(t *testing.T) {
	t.Parallel()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	client1LogWriter, client1Log := newStringWriter()
	client2LogWriter, client2Log := newStringWriter()
	cSlice, err := setupClients(clientOption{
		args: []string{
			"client",
			"-id", "id1",
			"-secret", "secret1",
			"-remote", s.GetListenerAddrPort().String(),
			"-local", "http://www.baidu.com/",
			"-remoteTimeout", "5s",
			"-useLocalAsHTTPHost",
			"-hostPrefix", "a",
			"-reconnectDelay", "24h", // 不重连
		},
		out: client1LogWriter,
	}, clientOption{
		args: []string{
			"client",
			"-id", "id2",
			"-secret", "secret2",
			"-remote", s.GetListenerAddrPort().String(),
			"-local", "http://www.baidu.com/",
			"-remoteTimeout", "5s",
			"-useLocalAsHTTPHost",
			"-hostPrefix", "a",
			"-reconnectDelay", "24h", // 不重连
		},
		out: client2LogWriter,
	})
	defer func() {
		for _, c := range cSlice {
			c.Close()
		}
	}()
	t.Log(err)

	// client1 或者 client2 有 host conflict
	if strings.Count(client1Log(), "host conflict") != 1 && strings.Count(client2Log(), "host conflict") != 1 {
		t.Fatal("client1 or client2 not host conflict")
	}
}

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
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestFailToDialLocalServer(t *testing.T) {
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
	c.OnTunnelClose.Store(func() {
		panic("tunnel should not be closed")
	})
	httpClient := setupHTTPClient(s.GetListenerAddrPort().String(), nil)
	_, err = httpClient.Get("http://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com/test?hello=world")
	if err == nil {
		t.Fatal("should failed to connect")
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

func TestInCompleteHTTPReqToServer(t *testing.T) {
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
	//err = conn.Close()
	//if err != nil {
	//	t.Fatal(err)
	//}
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

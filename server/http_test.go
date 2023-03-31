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

package server

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/isrc-cas/gt/bufio"
	"github.com/isrc-cas/gt/util"
)

const headerTargetPrefix = "Target-ID:"

func TestPeekHostReader(t *testing.T) {
	text := "GET / HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"User-Agent: curl/7.64.1\r\n" +
		"Accept: */*"
	host, err := peekHost(bufio.NewReader(strings.NewReader(text)))
	if err != nil {
		t.Fatal(err)
	}
	if len(host) < 1 || string(host) != "localhost" {
		t.Fatal()
	}
	t.Logf("host: %s", host)
}

func TestPeekTargetReader(t *testing.T) {
	text := "GET / HTTP/1.1\r\n" +
		"Host: localhost\r\n" +
		"Target-ID: target.localhost\r\n" +
		"User-Agent: curl/7.64.1\r\n" +
		"Accept: */*"
	target, err := peekHeader(bufio.NewReader(strings.NewReader(text)), headerTargetPrefix)
	if err != nil {
		t.Fatal(err)
	}
	if len(target) < 1 || string(target) != "target.localhost" {
		t.Fatal()
	}
	t.Logf("target: %s", target)
}

func TestPeekHostReaderError(t *testing.T) {
	text := "GET / HTTP/1.1\r\n" +
		"Host: \r\n" +
		"User-Agent: curl/7.64.1\r\n" +
		"Accept: */*"
	host, err := peekHost(bufio.NewReader(strings.NewReader(text)))
	if err == nil {
		t.Fatal()
	}
	t.Logf("host: %s", host)
}

func TestPeekHostReaderNoHost(t *testing.T) {
	text := "GET / HTTP/1.1\r\n" +
		"User-Agent: curl/7.64.1\r\n" +
		"Accept: */*"
	value, err := peekHost(bufio.NewReader(strings.NewReader(text)))
	if err == nil {
		t.Fatal()
	}
	t.Logf("value: %s, err: %s", value, err)
	value, err = peekHeader(bufio.NewReader(strings.NewReader(text)), headerTargetPrefix)
	if err == nil {
		t.Fatal()
	}
	t.Logf("value: %s, err: %s", value, err)
}

func TestPeekHostReaderInvalidHeaders(t *testing.T) {
	text := util.RandomString(17 * 1024)
	host, err := peekHost(bufio.NewReaderSize(strings.NewReader(text), 20*1024))
	if err == nil || !errors.Is(err, ErrInvalidHTTPProtocol) {
		t.Fatal()
	}
	t.Logf("host: %s, err: %s", host, err)
	text = "GET / HTTP/1.1\r\n\r\n" + text
	host, err = peekHost(bufio.NewReaderSize(strings.NewReader(text), 20*1024))
	if err == nil || !errors.Is(err, io.EOF) {
		t.Fatal()
	}
	t.Logf("host: %s, err: %s", host, err)
}

func TestParseTokenFromHost(t *testing.T) {
	id, err := parseIDFromHost([]byte("id"))
	if err == nil {
		t.Fatal("invalid id should returns error")
	}
	t.Log(id, err)
	id, err = parseIDFromHost([]byte("id.com"))
	if err == nil {
		t.Fatal("invalid id should returns error")
	}
	t.Log(id, err)
	id, err = parseIDFromHost([]byte("abc.id.com"))
	if err != nil {
		t.Fatal("invalid id should not returns error", err)
	}
	if !bytes.Equal(id, []byte("abc")) {
		t.Fatal("only 'abc' should be returned")
	}
	t.Logf("%s", id)
}

func BenchmarkParseTokenFromHost(b *testing.B) {
	host := []byte("abc.id.com")
	var id []byte
	var err error
	for i := 0; i < b.N; i++ {
		id, err = parseIDFromHost(host)
		if err != nil {
			b.Fatal("invalid id should not returns error", err)
		}
	}
	_ = id
}

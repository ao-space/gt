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
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"
)

func TestServices(t *testing.T) {
	t.Parallel()

	// 生成配置文件
	err := os.WriteFile("test_services_server.yaml", []byte(`
users:
  id3:
    secret: secret3
    host:
      withID: true
      number: 4
    tcp:
      - range: 1024-65535
        number: 4
  id4:
    secret: secret4
    host:
      withID: true
      number: 4
    tcp:
      - range: 1024-65535
        number: 4
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile("test_services_client1.yaml", []byte(`
services:
- local: http://www.baidu.com
  useLocalAsHTTPHost: true
  hostPrefix: id1-3
- local: http://m.baidu.com
  useLocalAsHTTPHost: true
  hostPrefix: id1-4
- local: tcp://www.baidu.com:80
  useLocalAsHTTPHost: true
  remoteTCPRandom: true
- local: tcp://m.baidu.com:80
  useLocalAsHTTPHost: true
  remoteTCPRandom: true
`), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile("test_services_client2.yaml", []byte(`
services:
- local: http://www.baidu.com
  useLocalAsHTTPHost: true
  hostPrefix: 3
- local: http://m.baidu.com
  useLocalAsHTTPHost: true
  hostPrefix: 4
- local: tcp://www.baidu.com:80
  useLocalAsHTTPHost: true
  remoteTCPRandom: true
- local: tcp://m.baidu.com:80
  useLocalAsHTTPHost: true
  remoteTCPRandom: true
  `), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove("test_services_server.yaml")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove("test_services_client1.yaml")
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove("test_services_client2.yaml")
		if err != nil {
			t.Fatal(err)
		}
	}()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-config", "test_services_server.yaml",
		"-addr", "127.0.0.1:0",
		"-id", "id1",
		"-secret", "secret1",
		"-id", "id2",
		"-secret", "secret2",
		"-hostNumber", "4",
		"-tcpRange", "1024-65535",
		"-tcpNumber", "4",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	client1LogWriter, client1Log := newStringWriter()
	client2LogWriter, client2Log := newStringWriter()
	cSlice, err := setupClients(clientOption{
		args: []string{
			"client", "-config=test_services_client1.yaml", "-remote=" + s.GetListenerAddrPort().String(), "-remoteTimeout=5s", "-id=id1", "-secret=secret1",
			"-local=http://www.baidu.com", "-useLocalAsHTTPHost", // hostPrefix=id1
			"-local=http://m.baidu.com", "-useLocalAsHTTPHost", "-hostPrefix=id1-2",
			"-local=tcp://www.baidu.com:80", "-useLocalAsHTTPHost", "-remoteTCPRandom",
			"-local=tcp://m.baidu.com:80", "-useLocalAsHTTPHost", "-remoteTCPRandom",
		},
		out: client1LogWriter,
	}, clientOption{
		args: []string{
			"client", "-config=test_services_client2.yaml", "-remote=" + s.GetListenerAddrPort().String(), "-remoteTimeout=5s", "-id=id3", "-secret=secret3",
			"-local=http://www.baidu.com", "-useLocalAsHTTPHost", // hostPrefix=id3
			"-local=http://m.baidu.com", "-useLocalAsHTTPHost", "-hostPrefix=2",
			"-local=tcp://www.baidu.com:80", "-useLocalAsHTTPHost", "-remoteTCPRandom",
			"-local=tcp://m.baidu.com:80", "-useLocalAsHTTPHost", "-remoteTCPRandom",
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
	time.Sleep(100 * time.Millisecond) // 等待服务器完成 TCP 端口分配

	// 从客户端日志中获取 tcp 端口
	var client1TCPPort1, client1TCPPort2, client1TCPPort3, client1TCPPort4 string
	match := regexp.MustCompile(`tcp port (\d+)`).FindAllStringSubmatch(client1Log(), -1)
	if len(match) != 4 {
		t.Fatal("invalid client1 log")
	}
	client1TCPPort1 = match[0][1]
	client1TCPPort2 = match[1][1]
	client1TCPPort3 = match[2][1]
	client1TCPPort4 = match[3][1]
	var client2TCPPort1, client2TCPPort2, client2TCPPort3, client2TCPPort4 string
	match = regexp.MustCompile(`tcp port (\d+)`).FindAllStringSubmatch(client2Log(), -1)
	if len(match) != 4 {
		t.Fatal("invalid client2 log")
	}
	client2TCPPort1 = match[0][1]
	client2TCPPort2 = match[1][1]
	client2TCPPort3 = match[2][1]
	client2TCPPort4 = match[3][1]

	// 对所有服务进行并发测试
	tests := []struct {
		serverAddr string
		url        string
	}{
		{s.GetListenerAddrPort().String(), "http://id1.example.com"},
		{s.GetListenerAddrPort().String(), "http://id1-2.example.com"},
		{s.GetListenerAddrPort().String(), "http://id1-3.example.com"},
		{s.GetListenerAddrPort().String(), "http://id1-4.example.com"},
		{"", "http://127.0.0.1:" + client1TCPPort1},
		{"", "http://127.0.0.1:" + client1TCPPort2},
		{"", "http://127.0.0.1:" + client1TCPPort3},
		{"", "http://127.0.0.1:" + client1TCPPort4},
		{s.GetListenerAddrPort().String(), "http://id3.example.com"},
		{s.GetListenerAddrPort().String(), "http://id3-2.example.com"},
		{s.GetListenerAddrPort().String(), "http://id3-3.example.com"},
		{s.GetListenerAddrPort().String(), "http://id3-4.example.com"},
		{"", "http://127.0.0.1:" + client2TCPPort1},
		{"", "http://127.0.0.1:" + client2TCPPort2},
		{"", "http://127.0.0.1:" + client2TCPPort3},
		{"", "http://127.0.0.1:" + client2TCPPort4},
	}
	for _, tt := range tests {
		var resp *http.Response
		var err error
		if tt.serverAddr != "" {
			httpClient := setupHTTPClient(tt.serverAddr, nil)
			resp, err = httpClient.Get(tt.url)
		} else {
			resp, err = http.Get(tt.url)
		}
		if err != nil {
			t.Fatal(err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatal("invalid status code")
			return
		}
		all, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
			return
		}
		if len(all) > 100 {
			all = all[:100]
		}
		fmt.Printf("%s\n", all)
	}
}

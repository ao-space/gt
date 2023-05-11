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

package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/util"
	"github.com/rs/zerolog"
)

// Server provides internal api service.
type Server struct {
	http.Server
	logger         zerolog.Logger
	checkTunnelMtx sync.Mutex
	RemoteAddr     string
	RemoteSchema   string

	// status response cache
	statusRespCache     []byte
	statusRespCacheTime time.Time

	id         atomic.Value
	secret     atomic.Value
	idConflict func(id string) bool
	muxHeader  string
}

// ID 返回 api server 生成的 id
func (s *Server) ID() string {
	idValue := s.id.Load()
	if idValue == nil {
		return ""
	}
	return idValue.(string)
}

// Secret 返回 api server 生成的 secret
func (s *Server) Secret() string {
	secretValue := s.secret.Load()
	if secretValue == nil {
		return ""
	}
	return secretValue.(string)
}

// NewServer returns an api server instance.
func NewServer(addr string, logger zerolog.Logger, idConflict func(id string) bool, header string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: mux,
		},
		logger:     logger,
		idConflict: idConflict,
		muxHeader:  header,
	}
	mux.HandleFunc("/status", s.status)
	mux.HandleFunc("/statusResp", s.statusResp)
	return s
}

func (s *Server) status(writer http.ResponseWriter, _ *http.Request) {
	err := s.check(writer)
	if err != nil {
		s.logger.Warn().Err(err).Msg("failed to check status")
		writer.WriteHeader(http.StatusServiceUnavailable)
		r := `{"status": "failed", "version":"` + predef.Version + `"}`
		_, err = writer.Write([]byte(r))
		if err != nil {
			s.logger.Warn().Err(err).Msg("failed to resp failed status")
		}
	}
}

func (s *Server) statusResp(writer http.ResponseWriter, _ *http.Request) {
	r := `{"status": "ok", "version":"` + predef.Version + `"}`
	_, err := writer.Write([]byte(r))
	if err != nil {
		s.logger.Warn().Err(err).Msg("failed to responses to statusResp request")
	}
}

func (s *Server) randomIDSecret() error {
	retries := 10
	for i := 0; i < retries; i++ {
		id := util.RandomString(predef.DefaultIDSize)
		if s.idConflict(id) {
			continue
		}
		s.id.Store(id)
		s.secret.Store(util.RandomString(predef.DefaultSecretSize))
		return nil
	}
	return fmt.Errorf("random id and secret still conflict after %v retries", retries)
}

// Auth 验证是不是 api server 生成的 id 和 secret
func (s *Server) Auth(id string, secret string) (ok bool) {
	ok = id == s.ID() && secret == s.Secret()
	return
}

func (s *Server) check(writer http.ResponseWriter) (err error) {
	s.checkTunnelMtx.Lock()
	defer s.checkTunnelMtx.Unlock()

	err = s.randomIDSecret()
	if err != nil {
		return
	}
	defer func() {
		s.id.Store("")
		s.secret.Store("")
	}()

	if len(s.statusRespCache) > 0 &&
		!s.statusRespCacheTime.IsZero() &&
		time.Since(s.statusRespCacheTime) <= 15*time.Second {
		_, err = writer.Write(s.statusRespCache)
		return
	}

	id := s.ID()
	cArgs := []string{
		"client",
		"-id", id,
		"-secret", s.Secret(),
		"-local", "http://" + s.Addr,
		"-remote", s.RemoteSchema + s.RemoteAddr,
		"-logLevel", "info",
		"-remoteCertInsecure",
	}
	c, err := client.New(cArgs, nil)
	if err != nil {
		return
	}
	err = c.Start()
	if err != nil {
		return
	}
	defer c.Close()
	dialFn := func(ctx context.Context, network string, address string) (net.Conn, error) {
		return (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext(ctx, network, s.RemoteAddr)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialFn,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	err = c.WaitUntilReady(30 * time.Second)
	if err != nil {
		return
	}
	var url string
	switch s.RemoteSchema {
	case "tcp://":
		url = fmt.Sprintf("http://%v.example.com/statusResp", id)
	case "tls://":
		url = fmt.Sprintf("https://%v.example.com/statusResp", id)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	if len(s.muxHeader) > 0 && s.muxHeader != "Host" {
		req.Header[s.muxHeader] = []string{id}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return
	}
	err = resp.Body.Close()
	if err != nil {
		return
	}
	s.statusRespCache = bs
	s.statusRespCacheTime = time.Now()
	_, err = writer.Write(bs)
	return
}

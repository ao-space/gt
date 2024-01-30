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
	"errors"
	"net"
	"net/http"

	"github.com/rs/zerolog"
)

// Server provides internal api service.
type Server struct {
	http.Server
	logger   zerolog.Logger
	Listener *VirtualListener
}

type ctxKey string

var rawConn ctxKey = "rawConn"

// NewServer returns an api server instance.
func NewServer(logger zerolog.Logger) *Server {
	mux := http.NewServeMux()
	s := &Server{
		Server: http.Server{
			Handler: mux,
			ConnContext: func(ctx context.Context, c net.Conn) context.Context {
				return context.WithValue(ctx, rawConn, c)
			},
		},
		Listener: NewVirtualListener(),
		logger:   logger,
	}
	mux.HandleFunc("/", s.exchangeWebRTCInfo)
	return s
}

func (s *Server) Start() {
	go func() {
		err := s.Serve(s.Listener)
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("failed to serve")
		}
		s.logger.Info().Msg("serve done")
	}()
}

func (s *Server) exchangeWebRTCInfo(writer http.ResponseWriter, request *http.Request) {
	defer func() {
		if request.Body != nil {
			request.Body.Close()
		}
	}()
	value := request.Context().Value(rawConn)
	if value == nil {
		return
	}
	rawConn, ok := value.(*Conn)
	if !ok {
		return
	}
	op := request.Header.Get("WebRTC-OP")
	switch op {
	case "get-offer":
		s.getOffer(rawConn, writer, request)
	case "resp-answer":
		s.receiveAnswer(rawConn, writer, request)
	default:
		s.receiveOffer(rawConn, writer, request)
	}
}

func (s *Server) receiveOffer(conn *Conn, writer http.ResponseWriter, request *http.Request) {
	conn.ProcessOffer(request, writer)
}

func (s *Server) getOffer(conn *Conn, writer http.ResponseWriter, request *http.Request) {
	conn.GetOffer(request, writer)
}

func (s *Server) receiveAnswer(conn *Conn, writer http.ResponseWriter, request *http.Request) {
	conn.ProcessAnswer(request, writer)
}

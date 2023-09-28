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

package client

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/predef"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

// Config is a client config.
type Config struct {
	Version  string // 目前未使用
	Services services
	Options
}

// Options is the config options for a client.
type Options struct {
	Config                string        `arg:"config" yaml:"-" usage:"The config file path to load"`
	ID                    string        `yaml:"id" usage:"The unique id used to connect to server. Now it's the prefix of the domain."`
	Secret                string        `yaml:"secret" usage:"The secret used to verify the id"`
	ReconnectDelay        time.Duration `yaml:"reconnectDelay" usage:"The delay before reconnect. Supports values like '30s', '5m'"`
	Remote                string        `yaml:"remote" usage:"The remote server url. Supports tcp:// and tls://, default tcp://"`
	RemoteSTUN            string        `yaml:"remoteSTUN" usage:"The remote STUN server address"`
	RemoteAPI             string        `yaml:"remoteAPI" usage:"The API to get remote server url"`
	RemoteCert            string        `yaml:"remoteCert" usage:"The path to remote cert"`
	RemoteCertInsecure    bool          `yaml:"remoteCertInsecure" usage:"Accept self-signed SSL certs from remote"`
	RemoteConnections     uint          `yaml:"remoteConnections" usage:"The max number of server connections in the pool. Valid value is 1 to 10"`
	RemoteIdleConnections uint          `yaml:"remoteIdleConnections" usage:"The number of idle server connections kept in the pool"`
	RemoteTimeout         time.Duration `yaml:"remoteTimeout" usage:"The timeout of remote connections. Supports values like '30s', '5m'"`

	HostPrefix         config.PositionSlice[string]        `yaml:"-" arg:"hostPrefix" usage:"The server will recognize this host prefix and forward data to local"`
	RemoteTCPPort      config.PositionSlice[uint16]        `yaml:"-" arg:"remoteTCPPort" usage:"The TCP port that the remote server will open"`
	RemoteTCPRandom    config.PositionSlice[bool]          `yaml:"-" arg:"remoteTCPRandom" usage:"Whether to choose a random tcp port by the remote server"`
	Local              config.PositionSlice[string]        `yaml:"-" arg:"local" usage:"The local service url"`
	LocalTimeout       config.PositionSlice[time.Duration] `yaml:"-" arg:"localTimeout" usage:"The timeout of local connections. Supports values like '30s', '5m'"`
	UseLocalAsHTTPHost config.PositionSlice[bool]          `yaml:"-" arg:"useLocalAsHTTPHost" usage:"Use the local address as host"`

	SentryDSN         string               `yaml:"sentryDSN" usage:"Sentry DSN to use"`
	SentryLevel       config.Slice[string] `yaml:"sentryLevel" usage:"Sentry levels: trace, debug, info, warn, error, fatal, panic (default [\"error\", \"fatal\", \"panic\"])"`
	SentrySampleRate  float64              `yaml:"sentrySampleRate" usage:"Sentry sample rate for event submission: [0.0 - 1.0]"`
	SentryRelease     string               `yaml:"sentryRelease" usage:"Sentry release to be sent with events"`
	SentryEnvironment string               `yaml:"sentryEnvironment" usage:"Sentry environment to be sent with events"`
	SentryServerName  string               `yaml:"sentryServerName" usage:"Sentry server name to be reported"`
	SentryDebug       bool                 `yaml:"sentryDebug" usage:"Sentry debug mode, the debug information is printed to help you understand what sentry is doing"`

	WebRTCConnectionIdleTimeout time.Duration `yaml:"webrtcConnectionIdleTimeout" usage:"The timeout of WebRTC connection. Supports values like '30s', '5m'"`
	WebRTCLogLevel              string        `yaml:"webrtcLogLevel" usage:"WebRTC log level: verbose, info, warning, error"`
	WebRTCMinPort               uint16        `yaml:"webrtcMinPort" usage:"The min port of WebRTC peer connection"`
	WebRTCMaxPort               uint16        `yaml:"webrtcMaxPort" usage:"The max port of WebRTC peer connection"`

	TCPForwardAddr        string `yaml:"tcpForwardAddr" usage:"The address of TCP forward"`
	TCPForwardHostPrefix  string `yaml:"tcpForwardHostPrefix" usage:"The host prefix of TCP forward"`
	TCPForwardConnections uint   `yaml:"tcpForwardConnections" usage:"The max number of TCP forward peer connections in the pool. Valid value is 1 to 10"`

	LogFile         string `yaml:"logFile" usage:"Path to save the log file"`
	LogFileMaxSize  int64  `yaml:"logFileMaxSize" usage:"Max size of the log files"`
	LogFileMaxCount uint   `yaml:"logFileMaxCount" usage:"Max count of the log files"`
	LogLevel        string `yaml:"logLevel" usage:"Log level: trace, debug, info, warn, error, fatal, panic, disable"`
	Version         bool   `arg:"version" yaml:"-" usage:"Show the version of this program"`

	Signal string `arg:"s" yaml:"-" usage:"Send signal to client processes. Supports values: reload, restart, stop, kill"`

	OpenBBR bool `yaml:"bbr" usage:"Use bbr as congestion control algorithm when GT use QUIC connection. Default algorithm is Cubic."`
}

func defaultConfig() Config {
	return Config{
		Options: Options{
			ReconnectDelay:        5 * time.Second,
			RemoteTimeout:         45 * time.Second,
			RemoteConnections:     3,
			RemoteIdleConnections: 1,

			SentrySampleRate: 1.0,
			SentryRelease:    predef.Version,

			WebRTCConnectionIdleTimeout: 5 * time.Minute,
			WebRTCLogLevel:              "warning",

			TCPForwardConnections: 3,

			LogFileMaxCount: 7,
			LogFileMaxSize:  512 * 1024 * 1024,
			LogLevel:        zerolog.InfoLevel.String(),

			OpenBBR: false,
		},
	}
}

type clientURL struct {
	*url.URL
}

func (c *clientURL) UnmarshalYAML(value *yaml.Node) (err error) {
	c.URL, err = url.Parse(value.Value)
	return
}

type service struct {
	HostPrefix         string        `yaml:"hostPrefix"`
	RemoteTCPPort      uint16        `yaml:"remoteTCPPort"`
	RemoteTCPRandom    *bool         `yaml:"remoteTCPRandom"`
	LocalURL           clientURL     `yaml:"local"`
	LocalTimeout       time.Duration `yaml:"localTimeout"`
	UseLocalAsHTTPHost bool          `yaml:"useLocalAsHTTPHost"`
}

func (s *service) String() string {
	sb := &strings.Builder{}
	sb.WriteString("service {")
	sb.WriteString("hostPrefix: ")
	sb.WriteString(s.HostPrefix)
	sb.WriteString(", local: ")
	sb.WriteString(s.LocalURL.String())
	sb.WriteString(", remoteTCPPort: ")
	sb.WriteString(strconv.Itoa(int(s.RemoteTCPPort)))
	if s.RemoteTCPRandom != nil {
		sb.WriteString(", remoteTCPRandom: ")
		sb.WriteString(fmt.Sprintf("%t", *s.RemoteTCPRandom))
	}
	sb.WriteString("}")
	return sb.String()
}

type services []service

func (ss services) String() string {
	sb := &strings.Builder{}
	sb.WriteByte('[')
	for i, s := range ss {
		sb.WriteString(s.String())
		if i != len(ss)-1 {
			sb.WriteByte(',')
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

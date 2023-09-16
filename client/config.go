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
	"encoding/json"
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
	Config                string   `arg:"config" yaml:"-" json:",omitempty" usage:"The config file path to load"`
	ID                    string   `yaml:"id" usage:"The unique id used to connect to server. Now it's the prefix of the domain."`
	Secret                string   `yaml:"secret" usage:"The secret used to verify the id"`
	ReconnectDelay        Duration `yaml:"reconnectDelay" json:",omitempty" usage:"The delay before reconnect. Supports values like '30s', '5m'"`
	Remote                string   `yaml:"remote" json:",omitempty" usage:"The remote server url. Supports tcp:// and tls://, default tcp://"`
	RemoteSTUN            string   `yaml:"remoteSTUN" json:",omitempty" usage:"The remote STUN server address"`
	RemoteAPI             string   `yaml:"remoteAPI" json:",omitempty" usage:"The API to get remote server url"`
	RemoteCert            string   `yaml:"remoteCert" json:",omitempty" usage:"The path to remote cert"`
	RemoteCertInsecure    bool     `yaml:"remoteCertInsecure" json:",omitempty" usage:"Accept self-signed SSL certs from remote"`
	RemoteConnections     uint     `yaml:"remoteConnections" json:",omitempty" usage:"The max number of server connections in the pool. Valid value is 1 to 10"`
	RemoteIdleConnections uint     `yaml:"remoteIdleConnections" json:",omitempty" usage:"The number of idle server connections kept in the pool"`
	RemoteTimeout         Duration `yaml:"remoteTimeout" json:",omitempty" usage:"The timeout of remote connections. Supports values like '30s', '5m'"`

	HostPrefix         config.PositionSlice[string]   `yaml:"-" arg:"hostPrefix" usage:"The server will recognize this host prefix and forward data to local"`
	RemoteTCPPort      config.PositionSlice[uint16]   `yaml:"-" json:"-" arg:"remoteTCPPort" usage:"The TCP port that the remote server will open"`
	RemoteTCPRandom    config.PositionSlice[bool]     `yaml:"-" json:"-" arg:"remoteTCPRandom" usage:"Whether to choose a random tcp port by the remote server"`
	Local              config.PositionSlice[string]   `yaml:"-" json:"-" arg:"local" usage:"The local service url"`
	LocalTimeout       config.PositionSlice[Duration] `yaml:"-" json:"-" arg:"localTimeout" usage:"The timeout of local connections. Supports values like '30s', '5m'"`
	UseLocalAsHTTPHost config.PositionSlice[bool]     `yaml:"-" json:"-" arg:"useLocalAsHTTPHost" usage:"Use the local address as host"`

	SentryDSN         string               `yaml:"sentryDSN" json:",omitempty" usage:"Sentry DSN to use"`
	SentryLevel       config.Slice[string] `yaml:"sentryLevel" json:",omitempty" usage:"Sentry levels: trace, debug, info, warn, error, fatal, panic (default [\"error\", \"fatal\", \"panic\"])"`
	SentrySampleRate  float64              `yaml:"sentrySampleRate" json:",omitempty" usage:"Sentry sample rate for event submission: [0.0 - 1.0]"`
	SentryRelease     string               `yaml:"sentryRelease" json:",omitempty" usage:"Sentry release to be sent with events"`
	SentryEnvironment string               `yaml:"sentryEnvironment" json:",omitempty" usage:"Sentry environment to be sent with events"`
	SentryServerName  string               `yaml:"sentryServerName" json:",omitempty" usage:"Sentry server name to be reported"`
	SentryDebug       bool                 `yaml:"sentryDebug" json:",omitempty" usage:"Sentry debug mode, the debug information is printed to help you understand what sentry is doing"`

	WebRTCConnectionIdleTimeout Duration `yaml:"webrtcConnectionIdleTimeout" usage:"The timeout of WebRTC connection. Supports values like '30s', '5m'"`
	WebRTCLogLevel              string   `yaml:"webrtcLogLevel" json:",omitempty" usage:"WebRTC log level: verbose, info, warning, error"`
	WebRTCMinPort               uint16   `yaml:"webrtcMinPort" json:",omitempty" usage:"The min port of WebRTC peer connection"`
	WebRTCMaxPort               uint16   `yaml:"webrtcMaxPort" json:",omitempty" usage:"The max port of WebRTC peer connection"`

	TCPForwardAddr        string `yaml:"tcpForwardAddr" json:",omitempty" usage:"The address of TCP forward"`
	TCPForwardHostPrefix  string `yaml:"tcpForwardHostPrefix" json:",omitempty" usage:"The host prefix of TCP forward"`
	TCPForwardConnections uint   `yaml:"tcpForwardConnections" json:",omitempty" usage:"The max number of TCP forward peer connections in the pool. Valid value is 1 to 10"`

	LogFile         string `yaml:"logFile" json:",omitempty" usage:"Path to save the log file"`
	LogFileMaxSize  int64  `yaml:"logFileMaxSize" json:",omitempty" usage:"Max size of the log files"`
	LogFileMaxCount uint   `yaml:"logFileMaxCount" json:",omitempty" usage:"Max count of the log files"`
	LogLevel        string `yaml:"logLevel" json:",omitempty" usage:"Log level: trace, debug, info, warn, error, fatal, panic, disable"`
	Version         bool   `arg:"version" yaml:"-" json:"-" usage:"Show the version of this program"`

	EnableWebServer bool   `arg:"web"  yaml:"web" json:"-" usage:"Enable web server"`
	WebAddr         string `arg:"webAddr"  yaml:"webAddr" json:"-" usage:"Web server address"`
	WebPort         uint16 `arg:"webPort" yaml:"webPort" json:"-" usage:"Web server port"`
	EnablePprof     bool   `arg:"pprof"  yaml:"pprof" json:"-" usage:"Enable pprof in web server"`
	SigningKey      string `arg:"signingKey" yaml:"signingKey" json:"-" usage:"JWT signing key for web server"`
	Admin           string `arg:"admin" yaml:"admin" json:"-" usage:"Admin username use for login in web server"`
	Password        string `arg:"password" yaml:"password" json:"-" usage:"Admin password use for login in web server"`

	Signal string `arg:"s" yaml:"-" json:"-" usage:"Send signal to client processes. Supports values: reload, restart, stop, kill"`
}

func defaultConfig() Config {
	return Config{
		Options: Options{
			ReconnectDelay:        Duration(5 * time.Second),
			RemoteTimeout:         Duration(45 * time.Second),
			RemoteConnections:     3,
			RemoteIdleConnections: 1,

			SentrySampleRate: 1.0,
			SentryRelease:    predef.Version,

			WebRTCConnectionIdleTimeout: Duration(5 * time.Minute),
			WebRTCLogLevel:              "warning",

			TCPForwardConnections: 3,

			LogFileMaxCount: 7,
			LogFileMaxSize:  512 * 1024 * 1024,
			LogLevel:        zerolog.InfoLevel.String(),

			EnableWebServer: false,
			WebAddr:         "127.0.0.1",
			WebPort:         8080,
			EnablePprof:     false,
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
func (c *clientURL) MarshalYAML() (interface{}, error) {
	return c.URL.String(), nil
}

func (c *clientURL) UnmarshalJSON(data []byte) error {
	var urlStr string
	if err := json.Unmarshal(data, &urlStr); err != nil {
		return err
	}
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	c.URL = parsedURL
	return nil
}

func (c *clientURL) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.URL.String())
}

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) (err error) {
	duration, err := time.ParseDuration(value.Value)
	if err != nil {
		return
	}
	*d = Duration(duration)
	return
}

func (d *Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(*d).String(), nil
}

func (d *Duration) UnmarshalJSON(data []byte) (err error) {
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:

		duration, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(duration)
	default:
		return fmt.Errorf("invalid duration: %v", v)
	}
	return nil
}

func (d *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(*d).String())
}

type service struct {
	HostPrefix         string    `yaml:"hostPrefix"`
	RemoteTCPPort      uint16    `yaml:"remoteTCPPort"`
	RemoteTCPRandom    *bool     `yaml:"remoteTCPRandom"`
	LocalURL           clientURL `yaml:"local"`
	LocalTimeout       Duration  `yaml:"localTimeout"`
	UseLocalAsHTTPHost bool      `yaml:"useLocalAsHTTPHost"`
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

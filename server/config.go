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
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/isrc-cas/gt/config"
	connection "github.com/isrc-cas/gt/conn"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/server/sync"
	"github.com/isrc-cas/gt/util"
	"github.com/rs/zerolog"
)

// Config is a server config.
type Config struct {
	Version string          `yaml:"-" json:"-"` // 目前未使用
	Users   map[string]user `yaml:"users,omitempty"`
	TCPs    []tcp           `yaml:"tcp,omitempty" json:",omitempty"`
	Host    host            `yaml:",omitempty" json:",omitempty"`
	Options
}

// Options is the config Options for a server.
type Options struct {
	Config        string `arg:"config" yaml:"-" json:"-" usage:"The config file path to load"`
	Addr          string `yaml:"addr,omitempty" usage:"The address to listen on. Supports values like: '80', ':80' or '0.0.0.0:80'"`
	TLSAddr       string `yaml:"tlsAddr,omitempty" json:",omitempty" usage:"The address for tls to listen on. Supports values like: '443', ':443' or '0.0.0.0:443'"`
	TLSMinVersion string `yaml:"tlsVersion,omitempty" json:",omitempty" usage:"The tls min version. Supports values: tls1.1, tls1.2, tls1.3"`
	CertFile      string `yaml:"certFile,omitempty" json:",omitempty" usage:"The path to cert file"`
	KeyFile       string `yaml:"keyFile,omitempty" json:",omitempty" usage:"The path to key file"`

	IDs               config.Slice[string] `arg:"id" yaml:"-" json:"-" usage:"The user id"`
	Secrets           config.Slice[string] `arg:"secret" yaml:"-" json:"-" usage:"The secret for user id"`
	Users             string               `yaml:"users,omitempty" json:"UserPath,omitempty" usage:"The users yaml file to load"`
	AuthAPI           string               `yaml:"authAPI,omitempty" json:",omitempty" usage:"The API to authenticate user with id and secret"`
	AllowAnyClient    bool                 `yaml:"allowAnyClient,omitempty" json:",omitempty" usage:"Allow any client to connect to the server"`
	TCPRanges         config.Slice[string] `arg:"tcpRange" yaml:"-" json:"-" usage:"The tcp port range, like 1024-65535"`
	TCPNumbers        config.Slice[string] `arg:"tcpNumber" yaml:"-" json:"-" usage:"The number of tcp ports allowed to be opened for each id"`
	Speed             uint32               `yaml:"speed,omitempty" json:",omitempty" usage:"The max number of bytes the client can transfer per second"`
	Connections       uint32               `yaml:"connections,omitempty" json:",omitempty" usage:"The max number of tunnel connections for a client"`
	ReconnectTimes    uint32               `yaml:"reconnectTimes,omitempty" json:",omitempty" usage:"The max number of times the client fails to reconnect"`
	ReconnectDuration config.Duration      `yaml:"reconnectDuration,omitempty" json:",omitempty" json:",omitempty" usage:"The time that the client cannot connect after the number of failed reconnections reaches the max number"`
	HostNumber        uint32               `arg:"hostNumber" yaml:"-" json:"-" usage:"The number of host-based services that the user can start"`
	HostRegex         config.Slice[string] `arg:"hostRegex" yaml:"-" json:"-" usage:"The host prefix started by user must conform to one of these rules"`
	HostWithID        bool                 `arg:"hostWithID" yaml:"-" json:"-" usage:"The prefix of host will become the form of id-host"`

	HTTPMUXHeader       string `yaml:"httpMUXHeader,omitempty" json:",omitempty" usage:"The http multiplexing header to be used"`
	MaxHandShakeOptions uint16 `yaml:"maxHandShakeOptions,omitempty" json:",omitempty" usage:"The max number of hand shake options"`

	Timeout                        config.Duration `yaml:"timeout,omitempty" json:",omitempty" usage:"The timeout of connections. Supports values like '30s', '5m'"`
	TimeoutOnUnidirectionalTraffic bool            `yaml:"timeoutOnUnidirectionalTraffic,omitempty" json:",omitempty" usage:"Timeout will happens when traffic is unidirectional"`

	// internal api service
	APIAddr          string `yaml:"apiAddr,omitempty" json:",omitempty" usage:"The address to listen on for internal api service. Supports values like: '8080', ':8080' or '0.0.0.0:8080'"`
	APICertFile      string `yaml:"apiCertFile,omitempty" json:",omitempty" usage:"The path to cert file"`
	APIKeyFile       string `yaml:"apiKeyFile,omitempty" json:",omitempty" usage:"The path to key file"`
	APITLSMinVersion string `yaml:"apiTLSVersion,omitempty" json:",omitempty" usage:"The tls min version. Supports values: tls1.1, tls1.2, tls1.3"`

	STUNAddr     string `yaml:"stunAddr,omitempty" json:",omitempty" usage:"The address to listen on for STUN service. Supports values like: '3478', ':3478' or '0.0.0.0:3478'"`
	STUNLogLevel string `yaml:"stunLogLevel,omitempty" json:",omitempty" usage:"Log level: trace, debug, info, warn, error, disable"`

	SNIAddr string `yaml:"sniAddr,omitempty" json:",omitempty" usage:"The address to listen on for raw tls proxy. Host comes from Server Name Indication. Supports values like: '443', ':443' or '0.0.0.0:443'"`

	SentryDSN         string               `yaml:"sentryDSN,omitempty" json:",omitempty" usage:"Sentry DSN to use"`
	SentryLevel       config.Slice[string] `yaml:"sentryLevel,omitempty" json:",omitempty" usage:"Sentry levels: trace, debug, info, warn, error, fatal, panic (default [\"error\", \"fatal\", \"panic\"])"`
	SentrySampleRate  float64              `yaml:"sentrySampleRate,omitempty" json:",omitempty" usage:"Sentry sample rate for event submission: [0.0 - 1.0]"`
	SentryRelease     string               `yaml:"sentryRelease,omitempty" json:",omitempty" usage:"Sentry release to be sent with events"`
	SentryEnvironment string               `yaml:"sentryEnvironment,omitempty" json:",omitempty" usage:"Sentry environment to be sent with events"`
	SentryServerName  string               `yaml:"sentryServerName,omitempty" json:",omitempty" usage:"Sentry server name to be reported"`
	SentryDebug       bool                 `yaml:"sentryDebug,omitempty" json:",omitempty" usage:"Sentry debug mode, the debug information is printed to help you understand what sentry is doing"`

	LogFile         string `yaml:"logFile,omitempty" json:",omitempty" usage:"Path to save the log file"`
	LogFileMaxSize  int64  `yaml:"logFileMaxSize,omitempty" json:",omitempty" usage:"Max size of the log files"`
	LogFileMaxCount uint   `yaml:"logFileMaxCount,omitempty" json:",omitempty" usage:"Max count of the log files"`
	LogLevel        string `yaml:"logLevel,omitempty" json:",omitempty" usage:"Log level: trace, debug, info, warn, error, fatal, panic, disable"`
	Version         bool   `arg:"version" yaml:"-" json:"-" usage:"Show the version of this program"`

	EnableWebServer bool   `arg:"web"  yaml:"web,omitempty" json:"-" usage:"Enable web server"`
	WebAddr         string `arg:"webAddr"  yaml:"webAddr,omitempty" json:"-" usage:"Web server address"`
	WebPort         uint16 `arg:"webPort" yaml:"webPort,omitempty" json:"-" usage:"Web server port"`
	EnablePprof     bool   `arg:"pprof"  yaml:"pprof,omitempty" json:"-" usage:"Enable pprof in web server"`
	SigningKey      string `arg:"signingKey" yaml:"signingKey,omitempty" json:"-" usage:"JWT signing key for web server"`
	Admin           string `arg:"admin" yaml:"admin,omitempty" json:"-" usage:"Admin username use for login in web server"`
	Password        string `arg:"password" yaml:"password,omitempty" json:"-" usage:"Admin password use for login in web server"`
}

func defaultConfig() Config {
	return Config{
		Options: Options{
			Addr:             "80",
			Timeout:          config.Duration{Duration: 90 * time.Second},
			TLSMinVersion:    "tls1.2",
			APITLSMinVersion: "tls1.2",
			LogFileMaxCount:  7,
			LogFileMaxSize:   512 * 1024 * 1024,
			LogLevel:         zerolog.InfoLevel.String(),
			STUNLogLevel:     "warn",

			SentrySampleRate: 1.0,
			SentryRelease:    predef.Version,

			HTTPMUXHeader: "Host",

			Connections:       10,
			ReconnectTimes:    3,
			ReconnectDuration: config.Duration{Duration: 5 * time.Minute},

			HostNumber: 0,

			MaxHandShakeOptions: 30,
		},
	}
}

// tcp 管理
type tcp struct {
	Range  string `json:",omitempty"`
	Number uint16 `json:",omitempty"`

	PortRange util.PortRange `yaml:"-" json:"-"`

	usedPort atomic.Int32
}

func (t *tcp) openTCPPort(tcpPort uint16) (listener net.Listener, err error) {
	if t.usedPort.Load() >= int32(t.Number) {
		return nil, connection.ErrFailedToOpenTCPPort
	}

	listener, err = net.Listen("tcp", ":"+strconv.Itoa(int(tcpPort)))
	if err == nil {
		t.usedPort.Add(1)
	}
	return
}

// 目前用不到这个函数，只是用 openTCPPort 来限制客户端开启的 tcp 端口数量
// func (t *tcp) closeTCPPort(listener net.Listener) (err error) {
// 	err = listener.Close()
// 	t.usedPort--
// 	return
// }

func (t *tcp) parseRange() (err error) {
	t.PortRange, err = util.NewPortRangeFromString(t.Range)
	return
}

// user 用户权限细节
type user struct {
	Secret      string
	TCPs        []tcp  `yaml:"tcp" json:",omitempty"`
	Speed       uint32 `yaml:"speed" json:",omitempty"`
	Connections uint32 `yaml:"connections" json:",omitempty"`
	Host        host   `json:",omitempty"`
	temp        bool
}

// users 客户端的权限管理
type users struct {
	sync.Map
}

// 合并 users 配置文件和命令行的 users
func (u *users) mergeUsers(users map[string]user, ids, secrets []string) error {
	for id, ud := range users {
		u.Store(id, ud)
	}

	if len(ids) != len(secrets) {
		return errors.New("the number of id does not match the number of secret")
	}
	for i := 0; i < len(ids); i++ {
		u.Store(ids[i], user{
			Secret: secrets[i],
		})
	}

	return u.verify()
}

func (u *users) verify() (err error) {
	u.Range(func(idValue, userValue interface{}) bool {
		id := idValue.(string)
		user := userValue.(user)
		if len(id) < predef.MinIDSize || len(id) > predef.MaxIDSize {
			err = fmt.Errorf("invalid id length: '%s'", id)
		}

		if len(user.Secret) < predef.MinSecretSize || len(user.Secret) > predef.MaxSecretSize {
			err = fmt.Errorf("invalid secret length: '%s'", user.Secret)
		}
		return true
	})
	return
}

func (u *users) empty() (empty bool) {
	empty = true
	u.Range(func(key, value interface{}) bool {
		empty = false
		return false
	})
	return
}

func (u *users) auth(id string, secret string) (result user, err error) {
	value, ok := u.Load(id)
	if !ok {
		err = ErrInvalidUser
		return
	}
	result, ok = value.(user)
	if !ok {
		err = ErrInvalidUser
		return
	}
	if result.Secret != secret {
		err = ErrInvalidUser
	}
	return
}

func (u *users) isIDConflict(id string) bool {
	_, ok := u.Load(id)
	return ok
}

// host 管理
type host struct {
	Number   *uint32               `json:",omitempty"`
	RegexStr *config.Slice[string] `yaml:"regex" json:",omitempty"`
	Regex    *[]*regexp.Regexp     `yaml:"-" json:"-"`
	WithID   *bool                 `yaml:"withID" json:",omitempty"`
	Prefixes map[string]struct{}   `yaml:"-" json:"-"`
}

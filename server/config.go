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
	Version string          // 目前未使用
	Users   map[string]user `yaml:"users"`
	TCPs    []tcp           `yaml:"tcp"`
	Host    host
	Options
}

// Options is the config Options for a server.
type Options struct {
	Config        string `arg:"config" yaml:"-" usage:"The config file path to load"`
	Addr          string `yaml:"addr" usage:"The address to listen on. Supports values like: '80', ':80' or '0.0.0.0:80'"`
	TLSAddr       string `yaml:"tlsAddr" usage:"The address for tls to listen on. Supports values like: '443', ':443' or '0.0.0.0:443'"`
	TLSMinVersion string `yaml:"tlsVersion" usage:"The tls min version. Supports values: tls1.1, tls1.2, tls1.3"`
	CertFile      string `yaml:"certFile" usage:"The path to cert file"`
	KeyFile       string `yaml:"keyFile" usage:"The path to key file"`

	IDs               config.Slice[string] `arg:"id" yaml:"-" usage:"The user id"`
	Secrets           config.Slice[string] `arg:"secret" yaml:"-" usage:"The secret for user id"`
	Users             string               `yaml:"users" usage:"The users yaml file to load"`
	AuthAPI           string               `yaml:"authAPI" usage:"The API to authenticate user with id and secret"`
	AllowAnyClient    bool                 `yaml:"allowAnyClient" usage:"Allow any client to connect to the server"`
	TCPRanges         config.Slice[string] `yaml:"tcpRange" usage:"The tcp port range, like 1024-65535"`
	TCPNumbers        config.Slice[string] `yaml:"tcpNumber" usage:"The number of tcp ports allowed to be opened for each id"`
	Speed             uint32               `yaml:"speed" usage:"The max number of bytes the client can transfer per second"`
	Connections       uint32               `yaml:"connections" usage:"The max number of tunnel connections for a client"`
	ReconnectTimes    uint32               `yaml:"reconnectTimes" usage:"The max number of times the client fails to reconnect"`
	ReconnectDuration time.Duration        `yaml:"reconnectDuration" usage:"The time that the client cannot connect after the number of failed reconnections reaches the max number"`
	HostNumber        uint32               `yaml:"hostNumber" usage:"The number of host-based services that the user can start"`
	HostRegex         config.Slice[string] `yaml:"hostRegex" usage:"The host prefix started by user must conform to one of these rules"`
	HostWithID        bool                 `yaml:"hostWithID" usage:"The prefix of host will become the form of id-host"`

	HTTPMUXHeader string `yaml:"httpMUXHeader" usage:"The http multiplexing header to be used"`

	Timeout                        time.Duration `yaml:"timeout" usage:"The timeout of connections. Supports values like '30s', '5m'"`
	TimeoutOnUnidirectionalTraffic bool          `yaml:"timeoutOnUnidirectionalTraffic" usage:"Timeout will happens when traffic is unidirectional"`

	// internal api service
	APIAddr          string `yaml:"apiAddr" usage:"The address to listen on for internal api service. Supports values like: '8080', ':8080' or '0.0.0.0:8080'"`
	APICertFile      string `yaml:"apiCertFile" usage:"The path to cert file"`
	APIKeyFile       string `yaml:"apiKeyFile" usage:"The path to key file"`
	APITLSMinVersion string `yaml:"apiTLSVersion" usage:"The tls min version. Supports values: tls1.1, tls1.2, tls1.3"`

	STUNAddr string `yaml:"stunAddr" usage:"The address to listen on for STUN service. Supports values like: '3478', ':3478' or '0.0.0.0:3478'"`

	SNIAddr string `yaml:"sniAddr" usage:"The address to listen on for raw tls proxy. Host comes from Server Name Indication. Supports values like: '443', ':443' or '0.0.0.0:443'"`

	SentryDSN         string               `yaml:"sentryDSN" usage:"Sentry DSN to use"`
	SentryLevel       config.Slice[string] `yaml:"sentryLevel" usage:"Sentry levels: trace, debug, info, warn, error, fatal, panic (default [\"error\", \"fatal\", \"panic\"])"`
	SentrySampleRate  float64              `yaml:"sentrySampleRate" usage:"Sentry sample rate for event submission: [0.0 - 1.0]"`
	SentryRelease     string               `yaml:"sentryRelease" usage:"Sentry release to be sent with events"`
	SentryEnvironment string               `yaml:"sentryEnvironment" usage:"Sentry environment to be sent with events"`
	SentryServerName  string               `yaml:"sentryServerName" usage:"Sentry server name to be reported"`
	SentryDebug       bool                 `yaml:"sentryDebug" usage:"Sentry debug mode, the debug information is printed to help you understand what sentry is doing"`

	LogFile         string `yaml:"logFile" usage:"Path to save the log file"`
	LogFileMaxSize  int64  `yaml:"logFileMaxSize" usage:"Max size of the log files"`
	LogFileMaxCount uint   `yaml:"logFileMaxCount" usage:"Max count of the log files"`
	LogLevel        string `yaml:"logLevel" usage:"Log level: trace, debug, info, warn, error, fatal, panic, disable"`
	Version         bool   `arg:"version" yaml:"-" usage:"Show the version of this program"`
}

func defaultConfig() Config {
	return Config{
		Options: Options{
			Addr:             "80",
			Timeout:          90 * time.Second,
			TLSMinVersion:    "tls1.2",
			APITLSMinVersion: "tls1.2",
			LogFileMaxCount:  7,
			LogFileMaxSize:   512 * 1024 * 1024,
			LogLevel:         zerolog.InfoLevel.String(),

			SentrySampleRate: 1.0,
			SentryRelease:    predef.Version,

			HTTPMUXHeader: "Host",

			Connections:       10,
			ReconnectTimes:    3,
			ReconnectDuration: 5 * time.Minute,
			HostNumber:        1,
		},
	}
}

// tcp 管理
type tcp struct {
	Range  string
	Number uint16

	PortRange *util.PortRange `yaml:"-"`

	usedPort uint32
}

func (t *tcp) openTCPPort(tcpPort uint16) (listener net.Listener, err error) {
	if t.usedPort >= uint32(t.Number) {
		return nil, connection.ErrFailedToOpenTCPPort
	}

	listener, err = net.Listen("tcp", ":"+strconv.Itoa(int(tcpPort)))
	if err == nil {
		t.usedPort++ // 只在单线程中调用，不需要加锁
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
	TCPs        []tcp  `yaml:"tcp"`
	Speed       uint32 `yaml:"speed"`
	Connections uint32 `yaml:"connections"`
	Host        host
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

func (u *users) auth(id string, secret string) (ok bool) {
	value, ok := u.Load(id)
	if !ok {
		return
	}
	ud, ok := value.(user)
	ok = ok && ud.Secret == secret
	return
}

func (u *users) isIDConflict(id string) bool {
	_, ok := u.Load(id)
	return ok
}

// host 管理
type host struct {
	Number   *uint32
	RegexStr *config.Slice[string] `yaml:"regex"`
	Regex    *[]*regexp.Regexp     `yaml:"-"`
	WithID   *bool                 `yaml:"withID"`

	usedHost uint32
}

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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/buger/jsonparser"
	"github.com/isrc-cas/gt/client/api"
	"github.com/isrc-cas/gt/client/webrtc"
	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/util"
)

// New parses the command line args and creates a Client. out 用于测试
func New(args []string, out io.Writer) (c *Client, err error) {
	conf := defaultConfig()
	err = config.ParseFlags(args, &conf, &conf.Options)
	if err != nil {
		return
	}
	if conf.Options.Version {
		_, _ = fmt.Println(predef.Version)
		os.Exit(0)
	}

	l, err := logger.Init(logger.Options{
		FilePath:          conf.LogFile,
		Out:               out,
		RotationCount:     conf.LogFileMaxCount,
		RotationSize:      conf.LogFileMaxSize,
		Level:             conf.LogLevel,
		SentryDSN:         conf.SentryDSN,
		SentryLevels:      conf.SentryLevel,
		SentrySampleRate:  conf.SentrySampleRate,
		SentryRelease:     conf.SentryRelease,
		SentryEnvironment: conf.SentryEnvironment,
		SentryServerName:  conf.SentryServerName,
		SentryDebug:       conf.SentryDebug,
	})
	if err != nil {
		return
	}

	c = &Client{
		config:  conf,
		Logger:  l,
		tunnels: make(map[*conn]struct{}),
		peers:   make(map[uint32]*peerTask),
	}
	c.tunnelsCond = sync.NewCond(c.tunnelsRWMtx.RLocker())
	c.apiServer = api.NewServer(l.With().Str("scope", "api").Logger())
	c.apiServer.ReadTimeout = 30 * time.Second
	return
}

type dialer struct {
	host      string
	stun      string
	tlsConfig *tls.Config
	dialFn    func() (conn net.Conn, err error)
}

func (d *dialer) init(c *Client, remote string, stun string) (err error) {
	var u *url.URL
	u, err = url.Parse(remote)
	if err != nil {
		err = fmt.Errorf("remote url (-remote option) '%s' is invalid, cause %s", remote, err.Error())
		return
	}
	d.stun = stun
	c.Logger.Info().Str("remote", remote).Str("stun", stun).Msg("remote url")
	switch u.Scheme {
	case "tls":
		if len(u.Port()) < 1 {
			u.Host = net.JoinHostPort(u.Host, "443")
		}
		tlsConfig := &tls.Config{}
		if len(c.config.RemoteCert) > 0 {
			var cf []byte
			cf, err = os.ReadFile(c.config.RemoteCert)
			if err != nil {
				err = fmt.Errorf("failed to read remote cert file (-remoteCert option) '%s', cause %s", c.config.RemoteCert, err.Error())
				return
			}
			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM(cf)
			if !ok {
				err = fmt.Errorf("failed to parse remote cert file (-remoteCert option) '%s'", c.config.RemoteCert)
				return
			}
			tlsConfig.RootCAs = roots
		}
		if c.config.RemoteCertInsecure {
			tlsConfig.InsecureSkipVerify = true
		}
		d.host = u.Host
		d.tlsConfig = tlsConfig
		d.dialFn = d.tlsDial
	case "tcp":
		if len(u.Port()) < 1 {
			u.Host = net.JoinHostPort(u.Host, "80")
		}
		d.host = u.Host
		d.dialFn = d.dial
	default:
		err = fmt.Errorf("remote url (-remote option) '%s' is invalid", remote)
	}
	return
}

func (d *dialer) initWithRemote(c *Client) (err error) {
	return d.init(c, c.config.Remote, c.config.RemoteSTUN)
}

func (d *dialer) initWithRemoteAPI(c *Client) (err error) {
	req, err := http.NewRequest("GET", c.config.RemoteAPI, nil)
	if err != nil {
		return
	}
	query := req.URL.Query()
	query.Add("network_client_id", c.config.ID)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Request-Id", strconv.FormatInt(time.Now().Unix(), 10))
	client := http.Client{
		Timeout: c.config.RemoteTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	r, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("invalid http status code %d, body: %s", resp.StatusCode, string(r))
		return
	}
	addr, err := jsonparser.GetString(r, "serverAddress")
	if err != nil {
		return
	}
	stunAddr, err := jsonparser.GetString(r, "stunAddress")
	if err != nil {
		if err != jsonparser.KeyPathNotFoundError {
			return
		}
	}
	err = d.init(c, addr, stunAddr)
	return
}

func (d *dialer) dial() (conn net.Conn, err error) {
	return net.Dial("tcp", d.host)
}

func (d *dialer) tlsDial() (conn net.Conn, err error) {
	return tls.Dial("tcp", d.host, d.tlsConfig)
}

// Start runs the client agent.
func (c *Client) Start() (err error) {
	c.Logger.Info().Interface("config", c.config).Msg(predef.Version)

	var level webrtc.LoggingSeverity
	switch c.config.WebRTCLogLevel {
	case "verbose":
		level = webrtc.LoggingSeverityVerbose
	case "info":
		level = webrtc.LoggingSeverityInfo
	case "warning":
		level = webrtc.LoggingSeverityWarning
	case "error":
		level = webrtc.LoggingSeverityError
	default:
		level = webrtc.LoggingSeverityNone
	}
	webrtc.SetLog(level, func(severity webrtc.LoggingSeverity, message, tag string) {
		switch severity {
		case webrtc.LoggingSeverityVerbose:
			c.Logger.Debug().Str("tag", tag).Msg("google-webrtc: " + message)
		case webrtc.LoggingSeverityInfo:
			c.Logger.Info().Str("tag", tag).Msg("google-webrtc: " + message)
		case webrtc.LoggingSeverityWarning:
			c.Logger.Warn().Str("tag", tag).Msg("google-webrtc: " + message)
		case webrtc.LoggingSeverityError:
			c.Logger.Error().Str("tag", tag).Msg("google-webrtc: " + message)
		}
	})

	if len(c.config.ID) < predef.MinIDSize || len(c.config.ID) > predef.MaxIDSize {
		err = fmt.Errorf("agent id (-id option) '%s' is invalid", c.config.ID)
		return
	}
	if c.config.Secret == "" {
		c.config.Secret = util.RandomString(predef.DefaultSecretSize)
	} else if len(c.config.Secret) < predef.MinSecretSize || len(c.config.Secret) > predef.MaxSecretSize {
		err = fmt.Errorf("agent secret (-secret option) '%s' is invalid", c.config.Secret)
		return
	}

	err = c.parseServices()
	if err != nil {
		return
	}

	var dialer dialer
	if len(c.config.Remote) > 0 {
		if !strings.Contains(c.config.Remote, "://") {
			c.config.Remote = "tcp://" + c.config.Remote
		}
		err = dialer.initWithRemote(c)
		if err != nil {
			return
		}
	}
	if len(c.config.RemoteAPI) > 0 {
		if !strings.HasPrefix(c.config.RemoteAPI, "http://") &&
			!strings.HasPrefix(c.config.RemoteAPI, "https://") {
			err = fmt.Errorf("remote api url (-remoteAPI option) '%s' must begin with http:// or https://", c.config.RemoteAPI)
			return
		}
		for len(dialer.host) == 0 {
			if atomic.LoadUint32(&c.closing) == 1 {
				err = errors.New("client is closing")
				return
			}
			err = dialer.initWithRemoteAPI(c)
			if err == nil {
				break
			}
			c.Logger.Error().Err(err).Msg("failed to query server address")
			time.Sleep(c.config.ReconnectDelay)
		}
	}
	if len(dialer.host) == 0 {
		err = errors.New("option -remote or -remoteAPI must be specified")
		return
	}

	if c.config.RemoteConnections < 1 {
		c.config.RemoteConnections = 1
	} else if c.config.RemoteConnections > 10 {
		c.config.RemoteConnections = 10
	}
	if c.config.RemoteIdleConnections < 1 {
		c.config.RemoteIdleConnections = 1
	} else if c.config.RemoteIdleConnections > c.config.RemoteConnections {
		c.config.RemoteIdleConnections = c.config.RemoteConnections
	}
	c.idleManager = newIdleManager(c.config.RemoteIdleConnections)

	for i := uint(1); i <= c.config.RemoteConnections; i++ {
		go c.connectLoop(dialer, i)
	}
	c.apiServer.Start()

	// tcpforward
	if c.config.TCPForwardConnections < 1 {
		c.config.TCPForwardConnections = 1
	} else if c.config.TCPForwardConnections > 10 {
		c.config.TCPForwardConnections = 10
	}
	if c.config.TCPForwardHostPrefix != "" {
		c.tcpForwardListener, err = net.Listen("tcp", c.config.TCPForwardAddr)
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to listen TCP forward")
			return
		}
		c.Logger.Info().Str("addr", c.tcpForwardListener.Addr().String()).Msg("Listening TCP forward")
		go c.tcpForwardStart(dialer)
	}

	return
}

// Close stops the client agent.
func (c *Client) Close() {
	if !atomic.CompareAndSwapUint32(&c.closing, 0, 1) {
		return
	}
	defer c.Logger.Close()
	c.tunnelsRWMtx.Lock()
	for t := range c.tunnels {
		t.SendForceCloseSignal()
		t.Close()
	}
	c.tunnelsRWMtx.Unlock()
	c.peersRWMtx.Lock()
	for _, p := range c.peers {
		p.Close()
	}
	c.peersRWMtx.Unlock()
	c.idleManager.Close()
	c.Logger.Info().Err(c.apiServer.Close()).Msg("api server close")
	if c.tcpForwardListener != nil {
		_ = c.tcpForwardListener.Close()
	}
}

func (c *Client) initConn(d dialer) (result *conn, err error) {
	c.initConnMtx.Lock()
	defer c.initConnMtx.Unlock()

	conn, err := d.dialFn()
	if err != nil {
		return
	}
	result = newConn(conn, c)
	result.stuns = append(result.stuns, d.stun)
	err = result.init()
	if err != nil {
		result.Close()
	}
	return
}

func (c *Client) connect(d dialer, connID uint) (closing bool) {
	defer func() {
		if !predef.Debug {
			if e := recover(); e != nil {
				c.Logger.Error().Msgf("recovered panic: %#v\n%s", e, debug.Stack())
			}
		}
	}()

	exit := c.idleManager.InitIdle(connID)
	if !exit {
		c.Logger.Info().Uint("connID", connID).Msg("trying to connect to remote")
		conn, err := c.initConn(d)
		if err == nil {
			conn.Logger = c.Logger.With().Uint("connID", connID).Logger()
			conn.readLoop(connID)
		} else {
			c.Logger.Error().Err(err).Uint("connID", connID).Msg("failed to connect to remote")
		}
	} else {
		c.Logger.Info().Uint("connID", connID).Msg("wait to connect to remote")
	}

	if atomic.LoadUint32(&c.closing) == 1 {
		return true
	}
	time.Sleep(c.config.ReconnectDelay)
	c.idleManager.SetWait(connID)
	c.idleManager.WaitIdle(connID)

	for len(c.config.RemoteAPI) > 0 {
		if atomic.LoadUint32(&c.closing) == 1 {
			return true
		}
		err := d.initWithRemoteAPI(c)
		if err == nil {
			break
		}
		c.Logger.Error().Err(err).Msg("failed to query server address")
		time.Sleep(c.config.ReconnectDelay)
	}
	return
}

func (c *Client) connectLoop(d dialer, connID uint) {
	for atomic.LoadUint32(&c.closing) == 0 {
		if c.connect(d, connID) {
			break
		}
	}
	c.Logger.Info().Msg("connect loop exited")
}

func (c *Client) addTunnel(conn *conn) {
	c.tunnelsRWMtx.Lock()
	c.tunnels[conn] = struct{}{}
	c.tunnelsRWMtx.Unlock()
	c.tunnelsCond.Broadcast()
}

func (c *Client) removeTunnel(conn *conn) {
	c.tunnelsRWMtx.Lock()
	delete(c.tunnels, conn)
	c.tunnelsRWMtx.Unlock()
}

var errTimeout = errors.New("timeout")

// WaitUntilReady waits until the client connected to server
func (c *Client) WaitUntilReady(timeout time.Duration) (err error) {
	c.tunnelsRWMtx.RLock()
	defer c.tunnelsRWMtx.RUnlock()
	for len(c.tunnels) < 1 {
		var e atomic.Value
		func() {
			timer := time.AfterFunc(timeout, func() {
				e.Store(errTimeout)
				c.tunnelsCond.Broadcast()
			})
			defer timer.Stop()
			c.tunnelsCond.Wait()
		}()
		v := e.Load()
		if v == nil {
			return
		}
		err = v.(error)
		if err != nil {
			return
		}
	}
	return
}

func (c *Client) parseServices() (err error) {
	// 将命令行和配置文件的数据填充到 c.services
	configServicesLen := len(c.config.Local) // 当长度为 1 的时候不需要位置信息
	configServices := make([]service, configServicesLen)
	for i := 0; i < configServicesLen; i++ {
		configServices[i].LocalURL.URL, err = url.Parse(c.config.Local[i].Value)
		if err != nil {
			err = fmt.Errorf("local url (-local option) '%s' is invalid, cause %s", c.config.Local[i].Value, err.Error())
			return
		}

		for _, x := range c.config.RemoteTCPPort {
			if configServicesLen == 1 ||
				(x.Position > c.config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < c.config.Local[i+1].Position)) {
				configServices[i].RemoteTCPPort = x.Value
			}
		}
		for _, x := range c.config.RemoteTCPRandom {
			if configServicesLen == 1 ||
				(x.Position > c.config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < c.config.Local[i+1].Position)) {
				configServices[i].RemoteTCPRandom = &x.Value
			}
		}
		for _, x := range c.config.LocalTimeout {
			if configServicesLen == 1 ||
				(x.Position > c.config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < c.config.Local[i+1].Position)) {
				configServices[i].LocalTimeout = x.Value
			}
		}
		for _, x := range c.config.UseLocalAsHTTPHost {
			if configServicesLen == 1 ||
				(x.Position > c.config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < c.config.Local[i+1].Position)) {
				configServices[i].UseLocalAsHTTPHost = x.Value
			}
		}
		for _, x := range c.config.HostPrefix {
			if configServicesLen == 1 ||
				(x.Position > c.config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < c.config.Local[i+1].Position)) {
				configServices[i].HostPrefix = x.Value
			}
		}
	}
	c.services = append(configServices, c.config.Services...)

	for i := 0; i < len(c.services); i++ {
		if c.services[i].LocalURL.URL == nil {
			err = errors.New("local url (-local option) cannot be empty")
			return
		}

		// 设置默认值
		if c.services[i].LocalTimeout == 0 {
			c.services[i].LocalTimeout = 120 * time.Second
		}
		if c.services[i].RemoteTCPRandom == nil {
			c.services[i].RemoteTCPRandom = new(bool)
			*c.services[i].RemoteTCPRandom = c.services[i].LocalURL.Scheme == "tcp" && c.services[i].RemoteTCPPort == 0
		}
		if (c.services[i].LocalURL.Scheme == "http" || c.services[i].LocalURL.Scheme == "https") &&
			c.services[i].HostPrefix == "" {
			c.services[i].HostPrefix = c.config.ID
		}

		// 处理 LocalURL
		switch c.services[i].LocalURL.Scheme {
		case "http":
			if !strings.Contains(c.services[i].LocalURL.Host, ":") {
				c.services[i].LocalURL.Host += ":80"
			}
		case "https":
			if !strings.Contains(c.services[i].LocalURL.Host, ":") {
				c.services[i].LocalURL.Host += ":443"
			}
		case "tcp":
			if c.services[i].LocalURL.Port() == "" {
				err = errors.New("-local option should contain port when local url (-local option) begin with tcp://")
				return
			}
			if c.services[i].RemoteTCPPort == 0 && !*c.services[i].RemoteTCPRandom {
				err = errors.New("-remoteTCPPort or -remoteTCPRandom option should be set when local url (-local option) begin with tcp://")
				return
			}
		default:
			err = fmt.Errorf("local url (-local option) '%s' must begin with http://, https:// or tcp://", c.config.Local[i].Value)
			return
		}

		// 判断 HostPrefix 的合法性
		if len(c.services[i].HostPrefix) > 0 &&
			(len(c.services[i].HostPrefix) < predef.MinHostPrefixSize || len(c.services[i].HostPrefix) > predef.MaxHostPrefixSize) {
			err = fmt.Errorf("host prefix (-hostPrefix option) '%s' is invalid", c.services[i].HostPrefix)
			return
		}
	}

	// HostPrefix 不能重复
	for i := 0; i < len(c.services); i++ {
		for j := i + 1; j < len(c.services); j++ {
			if len(c.services[i].HostPrefix) > 0 &&
				c.services[i].HostPrefix == c.services[j].HostPrefix {
				err = fmt.Errorf("duplicated host-prefix: %v", c.services[i].HostPrefix)
				return
			}
		}
	}

	// services 不能为空
	if len(c.services) == 0 {
		err = errors.New("no service is configured")
		return
	}

	return
}

// GetTCPForwardListenerAddrPort 获取 tcp forward listener 地址，返回值可能为空
func (c *Client) GetTCPForwardListenerAddrPort() (addrPort netip.AddrPort) {
	if c.tcpForwardListener == nil {
		return
	}
	addrPort = c.tcpForwardListener.Addr().(*net.TCPAddr).AddrPort()
	return
}

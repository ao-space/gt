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
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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
	"syscall"
	"time"

	"github.com/buger/jsonparser"
	"github.com/davecgh/go-spew/spew"
	"github.com/isrc-cas/gt/client/api"
	"github.com/isrc-cas/gt/client/webrtc"
	"github.com/isrc-cas/gt/config"
	connection "github.com/isrc-cas/gt/conn"
	"github.com/isrc-cas/gt/conn/msquic"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/util"
	"github.com/shirou/gopsutil/v3/process"
)

// New parses the command line args and creates a Client. out 用于测试
func New(args []string, out io.Writer) (c *Client, err error) {
	conf := getDefaultConfig(args)
	err = config.ParseFlags(args, &conf, &conf.Options)
	if err != nil {
		return
	}
	if conf.Options.Version {
		_, _ = fmt.Println(predef.Version)
		os.Exit(0)
	}
	if len(conf.Options.Signal) > 0 {
		err = processSignal(conf.Options.Signal)
		if err != nil {
			return
		}
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
		Logger:  l,
		tunnels: make(map[*conn]struct{}),
		peers:   make(map[uint32]PeerTask),
	}
	c.config.Store(&conf)
	c.tunnelsCond = sync.NewCond(c.tunnelsRWMtx.RLocker())
	c.apiServer = api.NewServer(l.With().Str("scope", "api").Logger())
	c.apiServer.ReadTimeout = 30 * time.Second
	c.webrtcThreadPool = webrtc.NewThreadPool(3)
	return
}
func getDefaultConfig(args []string) (conf Config) {
	if predef.IsNoArgs() {
		conf = defaultConfigWithNoArgs()
	} else {
		conf = defaultConfig()
		if util.Contains(args, "-webAddr") {
			conf.Config = predef.GetDefaultClientConfigPath()
		}
	}
	return
}

func processSignal(signal string) (err error) {
	switch signal {
	case "reload":
		err := sig(syscall.SIGHUP)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	case "restart":
		err := sig(syscall.SIGQUIT)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	case "stop":
		err := sig(syscall.SIGTERM)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	case "kill":
		err := sig(syscall.SIGKILL)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	default:
		err = fmt.Errorf("unknown value of '-s': %q", signal)
	}
	return
}

func sig(sig syscall.Signal) (err error) {
	processes, err := process.Processes()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	tid := os.Getpid()
	p, err := process.NewProcess(int32(tid))
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	e, err := p.Exe()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, proc := range processes {
		pid := int(proc.Pid)
		if pid == tid {
			continue
		}
		var exe string
		exe, err = proc.Exe()
		if err != nil {
			if os.IsNotExist(err) || os.IsPermission(err) {
				continue
			}
			_, _ = fmt.Fprintln(os.Stderr, err)
			return
		}
		if strings.HasPrefix(exe, e) {
			p, err := os.FindProcess(pid)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return err
			}
			err = p.Signal(sig)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				return err
			}
			fmt.Printf("sent signal to process %d.\n", pid)
		}
	}
	return
}

type dialer struct {
	tcp        string
	tls        string
	quic       string
	bbr        bool
	preferQuic bool
	stuns      []string
	tlsConfig  *tls.Config
}

func (d *dialer) isReady() (ready bool) {
	if len(d.tls) > 0 {
		return true
	}
	if len(d.quic) > 0 {
		return true
	}
	if len(d.tcp) > 0 {
		return true
	}

	return false
}

func (d *dialer) init(c *Client, remotes []string, stuns []string) (err error) {
	d.stuns = stuns
	for _, remote := range remotes {
		var u *url.URL
		u, err = url.Parse(remote)
		if err != nil {
			err = fmt.Errorf("remote url (-remote option) '%s' is invalid, cause %s", remote, err.Error())
			return
		}
		c.Logger.Info().Str("remote", remote).Strs("stuns", stuns).Msg("remote url")
		switch u.Scheme {
		case "tls":
			if len(u.Port()) < 1 {
				u.Host = net.JoinHostPort(u.Host, "443")
			}
			tlsConfig := &tls.Config{}
			if len(c.Config().RemoteCert) > 0 {
				var cf []byte
				cf, err = os.ReadFile(c.Config().RemoteCert)
				if err != nil {
					err = fmt.Errorf("failed to read remote cert file (-remoteCert option) '%s', cause %s", c.Config().RemoteCert, err.Error())
					return
				}
				roots := x509.NewCertPool()
				ok := roots.AppendCertsFromPEM(cf)
				if !ok {
					err = fmt.Errorf("failed to parse remote cert file (-remoteCert option) '%s'", c.Config().RemoteCert)
					return
				}
				tlsConfig.RootCAs = roots
			}
			if c.Config().RemoteCertInsecure {
				tlsConfig.InsecureSkipVerify = true
			}
			d.tls = u.Host
			d.tlsConfig = tlsConfig
		case "tcp":
			if len(u.Port()) < 1 {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			d.tcp = u.Host
		case "quic":
			if len(u.Port()) < 1 {
				u.Host = net.JoinHostPort(u.Host, "443")
			}
			tlsConfig := &tls.Config{}
			if len(c.Config().RemoteCert) > 0 {
				var cf []byte
				cf, err = os.ReadFile(c.Config().RemoteCert)
				if err != nil {
					err = fmt.Errorf("failed to read remote cert file (-remoteCert option) '%s', cause %s", c.Config().RemoteCert, err.Error())
					return
				}
				roots := x509.NewCertPool()
				ok := roots.AppendCertsFromPEM(cf)
				if !ok {
					err = fmt.Errorf("failed to parse remote cert file (-remoteCert option) '%s'", c.Config().RemoteCert)
					return
				}
				tlsConfig.RootCAs = roots
			}
			if c.Config().RemoteCertInsecure {
				tlsConfig.InsecureSkipVerify = true
			}
			d.quic = u.Host
			d.tlsConfig = tlsConfig
		default:
			err = fmt.Errorf("remote url (-remote option) '%s' is invalid", remote)
		}
	}
	return
}

func (d *dialer) initWithRemoteAPI(c *Client) (err error) {
	req, err := http.NewRequest("GET", c.Config().RemoteAPI, nil)
	if err != nil {
		return
	}
	query := req.URL.Query()
	query.Add("network_client_id", c.Config().ID)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Request-Id", strconv.FormatInt(time.Now().Unix(), 10))
	client := http.Client{
		Timeout: c.Config().RemoteTimeout.Duration,
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
		if !errors.Is(err, jsonparser.KeyPathNotFoundError) {
			return
		}
	}
	err = d.init(c, []string{addr}, []string{stunAddr})
	return
}

func (d *dialer) tcpDial() (conn net.Conn, err error) {
	return net.Dial("tcp", d.tcp)
}

func (d *dialer) tlsDial() (conn net.Conn, err error) {
	return tls.Dial("tcp", d.tls, d.tlsConfig)
}

func (d *dialer) quicDial() (conn net.Conn, err error) {
	return connection.QuicDial(d.quic, d.tlsConfig)
}

func (d *dialer) msquicDial() (conn net.Conn, err error) {
	return msquic.MsquicDial(d.quic, d.tlsConfig)
}

func (d *dialer) dial() (conn net.Conn, err error) {
	if !d.preferQuic && len(d.tls) > 0 {
		return d.tlsDial()
	}
	if len(d.quic) > 0 {
		if d.bbr {
			return d.msquicDial()
		}
		return d.quicDial()
	}
	if len(d.tcp) > 0 {
		return d.tcpDial()
	}

	return nil, errors.New("no dialer available")
}

func (c *Client) processRemotes() (result []dialer, err error) {
	m := make(map[string][]string)
	for index, remote := range c.Config().Remote {
		if !strings.Contains(remote, "://") {
			remote = "tcp://" + remote
			c.Config().Remote[index] = remote
		}
		var u *url.URL
		u, err = url.Parse(remote)
		if err != nil {
			return
		}
		hostname := u.Hostname()
		value, ok := m[hostname]
		if ok {
			value = append(value, remote)
			m[hostname] = value
		} else {
			m[hostname] = []string{remote}
		}
	}
	result = make([]dialer, 0, len(m))
	for _, remotes := range m {
		var d dialer
		err = d.init(c, remotes, c.Config().RemoteSTUN)
		if err != nil {
			return
		}
		result = append(result, d)
	}
	return
}

// Start runs the client agent.
func (c *Client) Start() (err error) {
	c.Logger.Info().Msg(predef.Version)

	var level webrtc.LoggingSeverity
	switch c.Config().WebRTCLogLevel {
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

	if len(c.Config().ID) < predef.MinIDSize || len(c.Config().ID) > predef.MaxIDSize {
		err = fmt.Errorf("agent id (-id option) '%s' is invalid", c.Config().ID)
		return
	}
	if c.Config().Secret == "" {
		c.Config().Secret = util.RandomString(predef.DefaultSecretSize)
	} else if len(c.Config().Secret) < predef.MinSecretSize || len(c.Config().Secret) > predef.MaxSecretSize {
		err = fmt.Errorf("agent secret (-secret option) '%s' is invalid", c.Config().Secret)
		return
	}

	err = c.parseServices()
	if err != nil {
		return
	}

	var ds []dialer
	if len(c.Config().Remote) > 0 {
		ds, err = c.processRemotes()
		if err != nil {
			return
		}
	}
	if len(c.Config().RemoteAPI) > 0 {
		if !strings.HasPrefix(c.Config().RemoteAPI, "http://") &&
			!strings.HasPrefix(c.Config().RemoteAPI, "https://") {
			err = fmt.Errorf("remote api url (-remoteAPI option) '%s' must begin with http:// or https://", c.Config().RemoteAPI)
			return
		}
		var dialer dialer
		for len(ds) == 0 {
			if atomic.LoadUint32(&c.closing) == 1 {
				err = errors.New("client is closing")
				return
			}
			err = dialer.initWithRemoteAPI(c)
			if err == nil {
				ds = append(ds, dialer)
				break
			}
			c.Logger.Error().Err(err).Msg("failed to query server address")
			time.Sleep(c.Config().ReconnectDelay.Duration)
		}
	}
	if len(ds) == 0 {
		err = errors.New("option -remote or -remoteAPI must be specified")
		return
	}

	if c.Config().RemoteConnections < 1 {
		c.Config().RemoteConnections = 1
	} else if c.Config().RemoteConnections > 10 {
		c.Config().RemoteConnections = 10
	}
	if c.Config().RemoteIdleConnections < 1 {
		c.Config().RemoteIdleConnections = 1
	} else if c.Config().RemoteIdleConnections > c.Config().RemoteConnections {
		c.Config().RemoteIdleConnections = c.Config().RemoteConnections
	}
	if c.Config().WebRTCRemoteConnections < 1 {
		c.Config().WebRTCRemoteConnections = 1
	} else if !predef.Debug && c.Config().WebRTCRemoteConnections > 50 {
		c.Config().WebRTCRemoteConnections = 50
	}
	c.idleManager = newIdleManager(c.Config().RemoteIdleConnections)

	conf4Log := *c.Config()
	conf4Log.Secret = "******"
	conf4Log.Password = "******"
	conf4Log.SigningKey = "******"
	c.Logger.Info().Msg(spew.Sdump(conf4Log))
	connID := uint(0)
	for _, dialer := range ds {
		for i := uint(1); i <= c.Config().RemoteConnections; i++ {
			connID += 1
			go c.connectLoop(dialer, connID)
			c.waitTunnelsShutdown.Add(1)
		}
	}
	c.apiServer.Start()

	// tcpforward
	if c.Config().TCPForwardConnections < 1 {
		c.Config().TCPForwardConnections = 1
	}
	if c.Config().TCPForwardHostPrefix != "" {
		c.tcpForwardListener, err = net.Listen("tcp", c.Config().TCPForwardAddr)
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to listen TCP forward")
			return
		}
		c.Logger.Info().Str("addr", c.tcpForwardListener.Addr().String()).Msg("Listening TCP forward")
		for _, dialer := range ds {
			go c.tcpForwardStart(dialer)
		}
	}

	return
}

func (c *Client) Config() *Config {
	return c.config.Load()
}

func (c *Client) GetConnectionPoolStatus() (status map[uint]Status) {
	if c.idleManager == nil {
		return
	}
	return c.idleManager.GetConnectionStatus()
}

func (c *Client) GetConnectionPoolNetInfo() (pools []PoolInfo) {
	c.tunnelsRWMtx.RLock()
	defer c.tunnelsRWMtx.RUnlock()
	for conn := range c.tunnels {
		pools = append(pools, PoolInfo{
			LocalAddr:  conn.LocalAddr(),
			RemoteAddr: conn.RemoteAddr(),
		})
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
	if c.idleManager != nil {
		c.idleManager.Close()
	}
	c.Logger.Info().Err(c.apiServer.Close()).Msg("api server close")
	if c.tcpForwardListener != nil {
		_ = c.tcpForwardListener.Close()
	}
	//c.webrtcThreadPool.Close()
}

// Shutdown stops the client gracefully.
func (c *Client) Shutdown() {
	defer c.Logger.Close()
	c.ShutdownWithoutClosingLogger()
}

func (c *Client) ShutdownWithoutClosingLogger() {
	if !atomic.CompareAndSwapUint32(&c.closing, 0, 1) {
		return
	}

	c.tunnelsRWMtx.Lock()
	for t := range c.tunnels {
		t.SendCloseSignal()
	}
	c.tunnelsRWMtx.Unlock()
	c.peersRWMtx.Lock()
	for _, p := range c.peers {
		p.Close()
	}
	c.peersRWMtx.Unlock()

	if c.idleManager != nil {
		c.idleManager.Close()
	}
	c.waitTunnelsShutdown.Wait()

	c.Logger.Info().Err(c.apiServer.Close()).Msg("api server close")
	if c.tcpForwardListener != nil {
		_ = c.tcpForwardListener.Close()
	}
	//c.webrtcThreadPool.Close()
}

func (c *Client) initConn(d dialer, connID uint) (result *conn, err error) {
	c.initConnMtx.Lock()
	defer c.initConnMtx.Unlock()

	conn, err := d.dial()
	if err != nil {
		return
	}
	result = newConn(conn, c)
	result.stuns = append(result.stuns, d.stuns...)
	result.Logger = c.Logger.With().Uint("connID", connID).Logger()
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

	c.idleManager.initMtx.Lock()
	exit := c.idleManager.Init(connID)
	if !exit {
		c.Logger.Info().Uint("connID", connID).Msg("trying to connect to remote")
		conn, err := c.initConn(d, connID)
		if err == nil {
			c.idleManager.SetIdle(connID)
			c.idleManager.initMtx.Unlock()
			conn.readLoop(connID)
		} else {
			c.idleManager.initMtx.Unlock()
			c.Logger.Error().Err(err).Uint("connID", connID).Msg("failed to connect to remote")
		}
	} else {
		c.idleManager.initMtx.Unlock()
		c.Logger.Info().Uint("connID", connID).Msg("wait to connect to remote")
	}
	c.idleManager.SetWait(connID)
	if atomic.LoadUint32(&c.closing) == 1 {
		return true
	}
	time.Sleep(c.Config().ReconnectDelay.Duration)
	c.idleManager.WaitIdle(connID)

	for len(c.Config().RemoteAPI) > 0 {
		if atomic.LoadUint32(&c.closing) == 1 {
			return true
		}
		err := d.initWithRemoteAPI(c)
		if err == nil {
			break
		}
		c.Logger.Error().Uint("connID", connID).Err(err).Msg("failed to query server address")
		time.Sleep(c.Config().ReconnectDelay.Duration)
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
	c.waitTunnelsShutdown.Done()
}

func (c *Client) addTunnel(conn *conn) {
	c.tunnelsRWMtx.Lock()
	c.tunnels[conn] = struct{}{}
	conn.services.Store(c.services.Load())
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
	var e atomic.Value
	timer := time.AfterFunc(timeout, func() {
		e.Store(errTimeout)
		c.tunnelsCond.Broadcast()
	})
	defer timer.Stop()
	for len(c.tunnels) < 1 {
		c.tunnelsCond.Wait()
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

var errNoService = errors.New("no service is configured")

func (c *Client) parseServices() (err error) {
	services, err := parseServices(c.Config())
	if err != nil {
		return
	}
	if len(services) == 0 {
		err = errNoService
		return
	}
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", services)))
	cs := [32]byte{}
	h.Sum(cs[:0])
	c.configChecksum.Store(&cs)
	c.services.Store(&services)
	c.Logger.Info().Hex("checksum", cs[:]).Str("services", services.String()).Msg("parse services")
	return
}

func parseServices(config *Config) (result services, err error) {
	// 将命令行和配置文件的数据填充到 c.services
	configServicesLen := len(config.Local) // 当长度为 1 的时候不需要位置信息
	configServices := make(services, configServicesLen)
	for i := 0; i < configServicesLen; i++ {
		configServices[i].LocalURL.URL, err = url.Parse(config.Local[i].Value)
		if err != nil {
			err = fmt.Errorf("local url (-local option) '%s' is invalid, cause %s", config.Local[i].Value, err.Error())
			return
		}

		for _, x := range config.RemoteTCPPort {
			if configServicesLen == 1 ||
				(x.Position > config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < config.Local[i+1].Position)) {
				configServices[i].RemoteTCPPort = x.Value
			}
		}
		for _, x := range config.RemoteTCPRandom {
			if configServicesLen == 1 ||
				(x.Position > config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < config.Local[i+1].Position)) {
				configServices[i].RemoteTCPRandom = &x.Value
			}
		}
		for _, x := range config.LocalTimeout {
			if configServicesLen == 1 ||
				(x.Position > config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < config.Local[i+1].Position)) {
				configServices[i].LocalTimeout.Duration = x.Value
			}
		}
		for _, x := range config.UseLocalAsHTTPHost {
			if configServicesLen == 1 ||
				(x.Position > config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < config.Local[i+1].Position)) {
				configServices[i].UseLocalAsHTTPHost = x.Value
			}
		}
		for _, x := range config.HostPrefix {
			if configServicesLen == 1 ||
				(x.Position > config.Local[i].Position &&
					(i == configServicesLen-1 || x.Position < config.Local[i+1].Position)) {
				configServices[i].HostPrefix = x.Value
			}
		}
	}
	result = append(configServices, config.Services...)

	usedIDASHostPrefix := false
	for i := 0; i < len(result); i++ {
		if result[i].LocalURL.URL == nil {
			err = errors.New("local url (-local option) cannot be empty")
		}

		// 设置默认值
		if result[i].LocalTimeout.Duration == 0 {
			result[i].LocalTimeout.Duration = 120 * time.Second
		}
		if result[i].RemoteTCPRandom == nil {
			result[i].RemoteTCPRandom = new(bool)
			*result[i].RemoteTCPRandom = result[i].LocalURL.Scheme == "tcp" && result[i].RemoteTCPPort == 0
		}
		if (result[i].LocalURL.Scheme == "http" || result[i].LocalURL.Scheme == "https") &&
			result[i].HostPrefix == "" {
			if !usedIDASHostPrefix {
				result[i].HostPrefix = config.ID
				usedIDASHostPrefix = true
			} else {
				err = errors.New("multi-services needs multiple hostPrefix")
				return
			}
		}

		// 处理 LocalURL
		switch result[i].LocalURL.Scheme {
		case "http":
			if !strings.Contains(result[i].LocalURL.Host, ":") {
				result[i].LocalURL.Host += ":80"
			}
		case "https":
			if !strings.Contains(result[i].LocalURL.Host, ":") {
				result[i].LocalURL.Host += ":443"
			}
		case "tcp":
			if result[i].LocalURL.Port() == "" {
				err = errors.New("-local option should contain port when local url (-local option) begin with tcp://")
				return
			}
			if result[i].RemoteTCPPort == 0 && !*result[i].RemoteTCPRandom {
				err = errors.New("-remoteTCPPort or -remoteTCPRandom option should be set when local url (-local option) begin with tcp://")
				return
			}
		default:
			err = fmt.Errorf("local url (-local option) '%s' must begin with http://, https:// or tcp://", result[i].LocalURL.String())
			return
		}

		// 判断 HostPrefix 的合法性
		if len(result[i].HostPrefix) > 0 &&
			(len(result[i].HostPrefix) < predef.MinHostPrefixSize || len(result[i].HostPrefix) > predef.MaxHostPrefixSize) {
			err = fmt.Errorf("host prefix (-hostPrefix option) '%s' is invalid", result[i].HostPrefix)
			return
		}
	}

	// HostPrefix 不能重复
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if len(result[i].HostPrefix) > 0 &&
				result[i].HostPrefix == result[j].HostPrefix {
				err = fmt.Errorf("duplicated host-prefix: %v", result[i].HostPrefix)
				return
			}
		}
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

// ReloadServices reload services from config file
func (c *Client) ReloadServices() (err error) {
	if !c.reloading.CompareAndSwap(false, true) {
		return errors.New("already reloading services")
	}
	defer func() {
		c.reloading.Store(false)
	}()

	conf := getDefaultConfig(os.Args)
	// ignore the webSetting
	conf.WebAddr = c.Config().WebAddr
	conf.WebKeyFile = c.Config().WebKeyFile
	conf.WebCertFile = c.Config().WebCertFile
	conf.EnablePprof = c.Config().EnablePprof
	conf.SigningKey = c.Config().SigningKey
	conf.Admin = c.Config().Admin
	conf.Password = c.Config().Password

	err = config.ParseFlags(os.Args, &conf, &conf.Options)
	if err != nil {
		return
	}

	ncb, err := json.Marshal(conf.Options)
	if err != nil {
		return
	}
	ocb, err := json.Marshal(c.Config().Options)
	if err != nil {
		return
	}
	same := bytes.Equal(ncb, ocb)
	c.Logger.Info().
		Str("newOptions", string(ncb)).
		Str("oldOptions", string(ocb)).
		Bool("isSame", same).
		Msg("the options section of configs")
	if !same {
		return errors.New("the options section of config file changed")
	}

	services, err := parseServices(&conf)
	if err != nil {
		return
	}
	if len(services) == 0 {
		err = errNoService
		return
	}
	checksum := [32]byte{}
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", services)))
	h.Sum(checksum[:0])
	c.Logger.Info().Hex("checksum", checksum[:]).Str("services", services.String()).Msg("parse services")

	if checksum == *c.configChecksum.Load() {
		return errors.New("config did not change")
	}

	buf := pool.BytesPool.Get().([]byte)
	defer pool.BytesPool.Put(buf)
	i := copy(buf, connection.ServicesBytes)
	n := gen(conf, services, buf[i:])

	conf4Log := conf
	conf4Log.Secret = "******"
	c.Logger.Info().Str("config", "reloading").Msg(spew.Sdump(conf4Log))

	c.initConnMtx.Lock()
	defer c.initConnMtx.Unlock()
	c.config.Store(&conf)
	c.services.Store(&services)
	c.configChecksum.Store(&checksum)

	c.tunnelsRWMtx.RLock()
	defer c.tunnelsRWMtx.RUnlock()
	for t := range c.tunnels {
		_, err = t.Write(buf[:n+i])
		if err != nil {
			return
		}
		t.Logger.Info().Msg("sent reload info")
		c.reloadWaitGroup.Add(1)
	}
	timer := time.AfterFunc(15*time.Second, func() {
		c.Logger.Warn().Msg("reload timeout")
		os.Exit(1)
	})
	c.reloadWaitGroup.Wait()
	timer.Stop()
	return
}

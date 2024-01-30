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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/libp2p/go-reuseport"
	"io"
	"math"
	"net"
	"net/http"
	"net/netip"
	"os"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	gosync "sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/buger/jsonparser"
	"github.com/davecgh/go-spew/spew"
	"github.com/isrc-cas/gt/config"
	connection "github.com/isrc-cas/gt/conn"
	"github.com/isrc-cas/gt/conn/msquic"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/server/api"
	"github.com/isrc-cas/gt/server/sync"
	"github.com/isrc-cas/gt/util"
	"github.com/pion/logging"
	"github.com/pion/turn/v3"
	"github.com/shirou/gopsutil/v3/process"
)

// Server is a network agent server.
type Server struct {
	config       Config
	users        users
	portsManager portsManager
	Logger       logger.Logger
	id2Client    sync.Map
	closing      uint32
	tlsListener  net.Listener
	listener     net.Listener
	sniListener  net.Listener
	quicListener net.Listener
	accepted     uint64
	served       uint64
	failed       uint64
	tunneling    uint64
	apiServer    *api.Server
	apiListener  net.Listener
	authUser     func(id string, secret string) (user, error)
	removeClient func(id string)
	stunServer   *turn.Server
	turnListener net.PacketConn

	// 重连限制
	reconnect        map[string]uint32
	reconnectRWMutex gosync.RWMutex

	hostPrefix2Client    sync.Map // key: hostPrefix(string) value: *client
	tlsHostPrefix2Client sync.Map // key: hostPrefix(string) value: *client
}

// New parses the command line args and creates a Server. out 用于测试
func New(args []string, out io.Writer) (s *Server, err error) {
	var conf Config
	if util.IsNoArgs() {
		conf = defaultConfigWithNoArgs()
	} else {
		conf = DefaultConfig()
		if util.Contains(args, "-webAddr") {
			conf.Config = util.GetDefaultServerConfigPath()
		}
	}
	err = config.ParseFlags(args, &conf, &conf.Options)
	if err != nil {
		return
	}
	conf.ConfigType = "server"
	err = MergeConfig(&conf)
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

	s = &Server{
		config:    conf,
		Logger:    l,
		reconnect: make(map[string]uint32),
	}
	return
}

func processSignal(signal string) (err error) {
	switch signal {
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
func (s *Server) tlsListen() (err error) {
	var tlsConfig *tls.Config
	tlsConfig, err = newTLSConfig(s.config.CertFile, s.config.KeyFile, s.config.TLSMinVersion)
	if err != nil {
		return
	}
	listener, err := reuseport.Listen("tcp", s.config.TLSAddr)
	if err != nil {
		err = fmt.Errorf("can not listen on addr '%s', cause %s, please check option 'tlsAddr'", s.config.TLSAddr, err.Error())
		return
	}
	s.tlsListener = tls.NewListener(listener, tlsConfig)
	s.Logger.Info().Str("addr", s.tlsListener.Addr().String()).Msg("Listening TLS")
	go s.acceptLoop(s.tlsListener, func(c *conn) {
		c.handle(c.handleHTTP)
	})
	return
}

func (s *Server) listen() (err error) {
	s.listener, err = reuseport.Listen("tcp", s.config.Addr)
	if err != nil {
		err = fmt.Errorf("can not listen on addr '%s', cause %s, please check option 'addr'", s.config.Addr, err.Error())
		return
	}
	s.Logger.Info().Str("addr", s.listener.Addr().String()).Msg("Listening")
	go s.acceptLoop(s.listener, func(c *conn) {
		c.handle(c.handleHTTP)
	})
	return
}

func (s *Server) quicListen(openBBR bool) (err error) {
	var tlsConfig *tls.Config
	if len(s.config.CertFile) > 0 && len(s.config.KeyFile) > 0 {
		tlsConfig, err = newTLSConfig(s.config.CertFile, s.config.KeyFile, s.config.TLSMinVersion)
	} else {
		tlsConfig = connection.GenerateTLSConfig()
	}
	if err != nil {
		return
	}
	if openBBR {
		//s.quicListener, err = connection.QuicBbrListen(s.config.QuicAddr, tlsConfig)
		//s.quicListener, err = quic.NewListenr(s.config.QuicAddr, 10_000, s.config.KeyFile, s.config.CertFile, "")
		s.quicListener, err = msquic.MsquicListen(s.config.QuicAddr, s.config.KeyFile, s.config.CertFile)
	} else {
		s.quicListener, err = connection.QuicListen(s.config.QuicAddr, tlsConfig)
	}
	if err != nil {
		err = fmt.Errorf("can not listen on addr '%s', cause %s, please check option 'addr'", s.config.QuicAddr, err.Error())
		return
	}
	s.Logger.Info().Str("QuicAddr", s.quicListener.Addr().String()).Msg("Listening")
	go s.acceptLoop(s.quicListener, func(c *conn) {
		c.handle(c.handleHTTP)
	})
	return
}

func (s *Server) sniListen() (err error) {
	s.sniListener, err = reuseport.Listen("tcp", s.config.SNIAddr)
	if err != nil {
		err = fmt.Errorf("can not listen on addr '%s', cause %s, please check option 'sniAddr'", s.config.SNIAddr, err.Error())
		return
	}
	s.Logger.Info().Str("sniAddr", s.sniListener.Addr().String()).Msg("Listening SNI")
	go s.acceptLoop(s.sniListener, func(c *conn) {
		c.handle(c.handleSNI)
	})
	return
}

func (s *Server) acceptLoop(l net.Listener, handle func(*conn)) {
	var err error
	defer func() {
		if !predef.Debug {
			if e := recover(); e != nil {
				s.Logger.Error().Msgf("recovered panic: %#v\n%s", e, debug.Stack())
			}
		}
		if errors.Is(err, net.ErrClosed) {
			err = nil
		}
		s.Logger.Info().Str("addr", l.Addr().String()).Err(err).Msg("acceptLoop ended")
	}()
	s.Logger.Info().Str("addr", l.Addr().String()).Msg("acceptLoop started")
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		if atomic.LoadUint32(&s.closing) > 0 {
			return
		}
		var conn net.Conn
		conn, err = l.Accept()
		if err != nil {
			if atomic.LoadUint32(&s.closing) > 0 {
				return
			}
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				s.Logger.Error().Err(err).Dur("delay", tempDelay).Msg("Server accept error")
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		atomic.AddUint64(&s.accepted, 1)
		c := newConn(conn, s)
		go handle(c)
	}
}

// Start runs the server.
func (s *Server) Start() (err error) {
	s.Logger.Info().Msg(predef.Version)
	err = s.users.mergeUsers(s.config.Users, nil, nil)
	if err != nil {
		return
	}
	users := make(map[string]*user)
	err = config.Yaml2Interface(s.config.Options.Users, users)
	if err != nil {
		return
	}
	err = s.users.mergeUsers(users, s.config.IDs, s.config.Secrets)
	if err != nil {
		return
	}
	err = s.parseTCPs()
	if err != nil {
		return
	}
	err = s.parseHost()
	if err != nil {
		return
	}

	if len(s.config.HTTPMUXHeader) <= 0 {
		err = fmt.Errorf("HTTP multiplexing header (-httpMUXHeader option) '%s' is invalid", s.config.HTTPMUXHeader)
		return
	}

	if len(s.config.AuthAPI) > 0 {
		s.authUser = nil
		s.removeClient = s.removeClientOnly
	} else if s.users.empty() {
		s.Logger.Warn().Msg("working on -allowAnyClient mode, because no user is configured")
		s.authUser = s.authUserOrCreateUser
		s.removeClient = s.removeClientAndUser
	} else if !s.config.AllowAnyClient {
		s.authUser = s.authUserWithConfig
		s.removeClient = s.removeClientOnly
	} else {
		s.authUser = s.authUserOrCreateUser
		s.removeClient = s.removeClientAndTempUser
	}
	if len(s.config.APIAddr) > 0 {
		if strings.IndexByte(s.config.APIAddr, ':') == -1 {
			s.config.APIAddr = ":" + s.config.APIAddr
		}
		apiServer := api.NewServer(
			s.config.APIAddr,
			s.Logger.With().Str("scope", "api").Logger(),
			s.users.isIDConflict,
			s.config.HTTPMUXHeader,
		)
		s.apiServer = apiServer
	}

	var listening bool
	if len(s.config.TLSAddr) > 0 && len(s.config.CertFile) > 0 && len(s.config.KeyFile) > 0 {
		if strings.IndexByte(s.config.TLSAddr, ':') == -1 {
			s.config.TLSAddr = ":" + s.config.TLSAddr
		}
		err = s.tlsListen()
		if err != nil {
			return
		}
		listening = true
	}
	if len(s.config.QuicAddr) > 0 {
		//if strings.IndexByte(s.config.QuicAddr, ':') == -1 {
		//	s.config.QuicAddr = ":" + s.config.QuicAddr
		//}
		// TODO Go 的冒号加空格会同时监听 IPv4 和 IPv6，这里行为与其他的不一致
		colonIndex := strings.IndexByte(s.config.QuicAddr, ':')
		if colonIndex == -1 {
			s.config.QuicAddr = "0.0.0.0:" + s.config.QuicAddr
		}
		if colonIndex == 0 {
			s.config.QuicAddr = "0.0.0.0" + s.config.QuicAddr
		}
		err = s.quicListen(s.config.OpenBBR)
		if err != nil {
			return
		}
		listening = true
	}
	if len(s.config.Addr) > 0 {
		if strings.IndexByte(s.config.Addr, ':') == -1 {
			s.config.Addr = ":" + s.config.Addr
		}
		err = s.listen()
		if err != nil {
			return
		}
		listening = true
	}
	if len(s.config.SNIAddr) > 0 {
		if strings.IndexByte(s.config.SNIAddr, ':') == -1 {
			s.config.SNIAddr = ":" + s.config.SNIAddr
		}
		err = s.sniListen()
		if err != nil {
			return
		}
		listening = true
	}
	if !listening {
		err = errors.New("no services is providing, please check the config")
		return
	}

	if len(s.config.STUNAddr) > 0 {
		err = s.startSTUNServer()
		if err != nil {
			return
		}
	}

	if len(s.config.APIAddr) > 0 {
		err = s.startAPIServer()
		if err != nil {
			return
		}
	}

	conf4log := *s.Config()
	conf4log.Password = "******"
	conf4log.SigningKey = "******"

	s.Logger.Info().Msg(spew.Sdump(conf4log))
	return
}

func (s *Server) startSTUNServer() (err error) {
	if strings.IndexByte(s.config.STUNAddr, ':') == -1 {
		s.config.STUNAddr = ":" + s.config.STUNAddr
	}
	s.turnListener, err = reuseport.ListenPacket("udp", s.config.STUNAddr)
	if err != nil {
		return
	}
	stunLogger := s.Logger.With().Str("scope", "stun").Logger()
	stunLogger.Info().Str("addr", s.config.STUNAddr).Msg("Listening")
	factory := logging.NewDefaultLoggerFactory()
	factory.Writer = stunLogger
	var lv logging.LogLevel
	switch strings.ToUpper(s.config.STUNLogLevel) {
	default:
		fallthrough
	case "DISABLE":
		lv = logging.LogLevelDisabled
	case "ERROR":
		lv = logging.LogLevelError
	case "WARN":
		lv = logging.LogLevelWarn
	case "INFO":
		lv = logging.LogLevelInfo
	case "DEBUG":
		lv = logging.LogLevelDebug
	case "TRACE":
		lv = logging.LogLevelTrace
	}
	factory.DefaultLogLevel = lv
	server, err := turn.NewServer(turn.ServerConfig{
		Realm:         "ao.space",
		LoggerFactory: factory,
		AuthHandler: func(username, realm string, srcAddr net.Addr) (key []byte, ok bool) {
			return
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: s.turnListener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorNone{
					Address: "0.0.0.0",
				},
			},
		},
	})

	s.Logger.Info().Str("addr", s.turnListener.LocalAddr().String()).Msg("Listening STUN")
	s.stunServer = server
	return
}

func newTLSConfig(cert, key, tlsMinVersion string) (tlsConfig *tls.Config, err error) {
	crt, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		err = fmt.Errorf("invalid cert and key, cause %s", err.Error())
		return
	}
	tlsConfig = &tls.Config{}
	tlsConfig.Certificates = []tls.Certificate{crt}
	switch strings.ToLower(tlsMinVersion) {
	case "tls1.1":
		tlsConfig.MinVersion = tls.VersionTLS11
	default:
		fallthrough
	case "tls1.2":
		tlsConfig.MinVersion = tls.VersionTLS12
		tlsConfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		}
	case "tls1.3":
		tlsConfig.MinVersion = tls.VersionTLS13
	}
	return
}

func (s *Server) startAPIServer() (err error) {
	if s.tlsListener != nil {
		s.apiServer.RemoteSchema = "tls://"
		s.apiServer.RemoteAddr = s.tlsListener.Addr().String()
	} else if s.listener != nil {
		s.apiServer.RemoteSchema = "tcp://"
		s.apiServer.RemoteAddr = s.listener.Addr().String()
	}
	if len(s.config.APICertFile) > 0 && len(s.config.APIKeyFile) > 0 {
		var tlsConfig *tls.Config
		tlsConfig, err = newTLSConfig(s.config.APICertFile, s.config.APIKeyFile, s.config.APITLSMinVersion)
		if err != nil {
			return
		}
		ln, err := reuseport.Listen("tcp", s.config.APIAddr)
		if err != nil {
			return fmt.Errorf("can not listen on addr '%s', cause %s, please check option 'tlsAddr'", s.config.APIAddr, err.Error())
		}
		s.apiListener = tls.NewListener(ln, tlsConfig)
	} else {
		s.apiListener, err = reuseport.Listen("tcp", s.config.APIAddr)
		if err != nil {
			return fmt.Errorf("can not listen on addr '%s', cause %s, please check option 'apiAddr'", s.config.APIAddr, err.Error())
		}
	}
	s.Logger.Info().Str("addr", s.apiListener.Addr().String()).Msg("Listening API")
	s.apiServer.Addr = s.apiListener.Addr().String()
	go func() {
		err := s.apiServer.Serve(s.apiListener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		s.Logger.Info().Err(err).Msg("api server closed")
	}()
	return nil
}

type authParam struct {
	NetworkClientId  string   `json:"networkClientId"`
	NetworkSecretKey string   `json:"networkSecretKey"`
	AppletTokens     []string `json:"appletTokens"`
}

func (s *Server) authWithAPI(id string, secret string, prefixes []string) (hostPrefixes map[string]struct{}, ok bool, err error) {
	var bs bytes.Buffer
	encoder := json.NewEncoder(&bs)
	p := &authParam{
		NetworkClientId:  id,
		NetworkSecretKey: secret,
		AppletTokens:     prefixes,
	}
	err = encoder.Encode(p)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", s.config.AuthAPI, &bs)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Request-Id", strconv.FormatInt(time.Now().Unix(), 10))
	client := http.Client{
		Timeout: s.config.Timeout.Duration,
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
	ok, err = jsonparser.GetBoolean(r, "result")
	if ok {
		hostPrefixes = make(map[string]struct{})
		hostPrefixes[id] = struct{}{}
		_, err = jsonparser.ArrayEach(r, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			hostPrefixes[string(value)] = struct{}{}
		}, "appletTokens")
		if errors.Is(err, jsonparser.KeyPathNotFoundError) {
			err = nil
		}
	}
	return
}

// Close stops the server.
func (s *Server) Close() {
	if !atomic.CompareAndSwapUint32(&s.closing, 0, 1) {
		return
	}
	defer s.Logger.Close()
	event := s.Logger.Info()
	if s.apiServer != nil {
		event.AnErr("api", s.apiServer.Close())
	}
	if s.stunServer != nil {
		event.AnErr("turn", s.stunServer.Close())
	}
	if s.listener != nil {
		event.AnErr("listener", s.listener.Close())
	}
	if s.tlsListener != nil {
		event.AnErr("tlsListener", s.tlsListener.Close())
	}
	if s.sniListener != nil {
		event.AnErr("sniListener", s.sniListener.Close())
	}
	s.id2Client.Range(func(key, value interface{}) bool {
		if c, ok := value.(*client); ok && c != nil {
			c.close()
		}
		return true
	})
	event.Msg("server stopped")
}

// IsClosing tells is the server stopping.
func (s *Server) IsClosing() (closing bool) {
	return atomic.LoadUint32(&s.closing) > 0
}

// Shutdown stops the server gracefully.
func (s *Server) Shutdown() {
	defer s.Logger.Close()
	s.ShutdownWithoutClosingLogger()
}

func (s *Server) ShutdownWithoutClosingLogger() {
	if !atomic.CompareAndSwapUint32(&s.closing, 0, 1) {
		return
	}
	event := s.Logger.Info()
	if s.apiServer != nil {
		event.AnErr("api", s.apiServer.Close())
	}
	if s.stunServer != nil {
		event.AnErr("turn", s.stunServer.Close())
	}
	if s.listener != nil {
		event.AnErr("listener", s.listener.Close())
	}
	if s.tlsListener != nil {
		event.AnErr("tlsListener", s.tlsListener.Close())
	}
	if s.sniListener != nil {
		event.AnErr("sniListener", s.sniListener.Close())
	}
	for {
		accepted := s.GetAccepted()
		served := s.GetServed()
		failed := s.GetFailed()
		tunneling := s.GetTunneling()
		if accepted == served+failed+tunneling {
			break
		}

		i := 0
		s.id2Client.Range(func(key, value interface{}) bool {
			i++
			if c, ok := value.(*client); ok && c != nil {
				c.shutdown()
			}
			return true
		})
		if i == 0 {
			break
		}

		s.Logger.Info().
			Uint64("accepted", accepted).
			Uint64("served", served).
			Uint64("failed", failed).
			Uint64("tunneling", tunneling).
			Msg("server shutting down")
		time.Sleep(3 * time.Second)
	}
	s.id2Client.Range(func(key, value interface{}) bool {
		if c, ok := value.(*client); ok && c != nil {
			c.close()
		}
		return true
	})
	event.Msg("server stopped")
}

func (s *Server) getOrCreateClient(id string, fn func() interface{}) (result *client, exists bool) {
	value, exists := s.id2Client.LoadOrCreate(id, fn)
	result = value.(*client)
	return
}

func (s *Server) getHostPrefix(hostPrefix string) (c clientWithServiceIndex, ok bool) {
	value, ok := s.hostPrefix2Client.Load(hostPrefix)
	if ok {
		c = value.(clientWithServiceIndex)
	}
	return
}

func (s *Server) getOrCreateHostPrefix(hostPrefix string, tls bool, fn func() interface{}) (c clientWithServiceIndex, ok bool) {
	if !tls {
		actual, loaded := s.hostPrefix2Client.LoadOrCreate(hostPrefix, fn)
		return actual.(clientWithServiceIndex), loaded
	} else {
		actual, loaded := s.tlsHostPrefix2Client.LoadOrCreate(hostPrefix, fn)
		return actual.(clientWithServiceIndex), loaded
	}
}

func (s *Server) storeHostPrefix(hostPrefix string, tls bool, c clientWithServiceIndex) {
	if !tls {
		s.hostPrefix2Client.Store(hostPrefix, c)
	} else {
		s.tlsHostPrefix2Client.Store(hostPrefix, c)
	}
}

func (s *Server) removeHostPrefix(hostPrefix string, tls bool) {
	if !tls {
		s.hostPrefix2Client.Delete(hostPrefix)
	} else {
		s.tlsHostPrefix2Client.Delete(hostPrefix)
	}
}

func (s *Server) getTLSHostPrefix(hostPrefix string) (c clientWithServiceIndex, ok bool) {
	value, ok := s.tlsHostPrefix2Client.Load(hostPrefix)
	if ok {
		c = value.(clientWithServiceIndex)
	}
	return
}

// GetAccepted returns value of accepted
func (s *Server) GetAccepted() uint64 {
	return atomic.LoadUint64(&s.accepted)
}

// GetServed returns value of served
func (s *Server) GetServed() uint64 {
	return atomic.LoadUint64(&s.served)
}

// GetFailed returns value of served
func (s *Server) GetFailed() uint64 {
	return atomic.LoadUint64(&s.failed)
}

// GetTunneling returns value of tunneling
func (s *Server) GetTunneling() uint64 {
	return atomic.LoadUint64(&s.tunneling)
}

// ErrInvalidUser is returned if id and secret are invalid
var ErrInvalidUser = errors.New("invalid user")

func (s *Server) authUserWithConfig(id string, secret string) (u user, err error) {
	if len(id) < 1 || len(secret) < 1 {
		err = ErrInvalidUser
		return
	}
	u, err = s.users.auth(id, secret)
	if err != nil {
		if s.apiServer != nil && s.apiServer.Auth(id, secret) {
			u = s.newTempUserForAPIServer()
			err = nil
			return
		}
	}
	return
}

func (s *Server) newTempUserForAPIServer() user {
	return user{
		TCPNumber:    &s.config.TCPNumber,
		Speed:        s.config.Speed,
		Connections:  s.config.Connections,
		Host:         s.config.Host,
		portsManager: &s.portsManager,
	}
}

func (s *Server) authUserWithAPI(id string, secret string, prefixes []string) (u user, err error) {
	if len(id) < 1 || len(secret) < 1 {
		err = ErrInvalidUser
		return
	}
	hostPrefixes, ok, err := s.authWithAPI(id, secret, prefixes)
	if err != nil {
		return
	}
	if !ok {
		if s.apiServer != nil && s.apiServer.Auth(id, secret) {
			u = s.newTempUserForAPIServer()
			err = nil
			return
		}
		err = ErrInvalidUser
		return
	}
	u = user{
		TCPNumber:    &s.config.TCPNumber,
		Speed:        s.config.Speed,
		Connections:  s.config.Connections,
		Host:         s.config.Host,
		portsManager: &s.portsManager,
	}
	u.Host.Prefixes = hostPrefixes
	return
}

func (s *Server) authUserOrCreateUser(id, secret string) (u user, err error) {
	if s.apiServer != nil && s.apiServer.Auth(id, secret) {
		u = s.newTempUserForAPIServer()
		err = nil
		return
	}

	value, _ := s.users.LoadOrCreate(id, func() interface{} {
		return user{
			Secret:       secret,
			TCPNumber:    &s.config.TCPNumber,
			Speed:        s.config.Speed,
			Connections:  s.config.Connections,
			Host:         s.config.Host,
			temp:         true,
			portsManager: &s.portsManager,
		}
	})
	var ok bool
	u, ok = value.(user)
	if !ok {
		err = ErrInvalidUser
		return
	}
	if u.Secret != secret {
		err = ErrInvalidUser
	}
	return
}

func (s *Server) removeClientOnly(id string) {
	s.id2Client.Delete(id)
}

func (s *Server) removeClientAndUser(id string) {
	s.id2Client.Delete(id)
	s.users.Delete(id)
}

func (s *Server) removeClientAndTempUser(id string) {
	s.id2Client.Delete(id)

	value, loaded := s.users.Load(id)
	if loaded && value.(user).temp {
		s.users.Delete(id)
	}
}

// tcp 相关配置，命令行的优先级高于配置文件
func (s *Server) parseTCPs() (err error) {
	ports := make(map[uint16]struct{}, 65536)
	all := make(map[uint16]struct{}, 65536)
	for _, tcp := range s.config.TCPs {
		pr, err := util.NewPortRangeFromString(tcp.Range)
		if err != nil {
			return err
		}
		for i := pr.Min; i <= pr.Max; i++ {
			ports[i] = struct{}{}
			all[i] = struct{}{}
			if i == math.MaxUint16 {
				break
			}
		}
	}

	for _, tcp := range s.config.TCPRanges {
		pr, err := util.NewPortRangeFromString(tcp)
		if err != nil {
			return err
		}
		for i := pr.Min; i <= pr.Max; i++ {
			ports[i] = struct{}{}
			all[i] = struct{}{}
			if i == math.MaxUint16 {
				break
			}
		}
	}

	s.portsManager.ports = ports

	// 处理用户 tcp
	s.users.Range(func(key, value interface{}) bool {
		u := value.(user)
		if u.TCPNumber == nil {
			u.TCPNumber = &s.config.TCPNumber
		}
		if len(u.TCPs) == 0 { // 如果用户没有设置则使用全局的
			u.portsManager = &s.portsManager
		} else {
			ports := make(map[uint16]struct{})
			for _, tcp := range u.TCPs {
				var pr util.PortRange
				pr, err = util.NewPortRangeFromString(tcp.Range)
				if err != nil {
					return false
				}
				for i := pr.Min; i <= pr.Max; i++ {
					if _, ok := all[i]; ok {
						err = fmt.Errorf("tcp port %d is used by global", i)
						return false
					}
					ports[i] = struct{}{}
					all[i] = struct{}{}
					if i == math.MaxUint16 {
						break
					}
				}
			}
			u.portsManager = &portsManager{ports: ports}
		}
		s.users.Store(key, u)
		return true
	})
	return
}

// host 相关配置，命令行的优先级高于配置文件
func (s *Server) parseHost() (err error) {
	// 合并 host regex
	hostRegexMap := make(map[string]struct{})
	if s.config.Host.RegexStr == nil {
		s.config.Host.RegexStr = &s.config.HostRegex
	}
	for _, regex := range *s.config.Host.RegexStr {
		hostRegexMap[regex] = struct{}{}
	}
	for _, regex := range s.config.HostRegex {
		hostRegexMap[regex] = struct{}{}
	}
	*s.config.Host.RegexStr = config.Slice[string]{}
	for hostRegex := range hostRegexMap {
		*s.config.Host.RegexStr = append(*s.config.Host.RegexStr, hostRegex)
	}

	// 处理全局 host
	if s.config.Host.Number == nil {
		s.config.Host.Number = &s.config.HostNumber
	}
	s.config.Host.Regex = new([]*regexp.Regexp)
	for str := range hostRegexMap {
		regex, err := regexp.Compile(str)
		if err != nil {
			return err
		}
		*s.config.Host.Regex = append(*s.config.Host.Regex, regex)
	}
	if s.config.Host.WithID == nil {
		s.config.Host.WithID = &s.config.HostWithID
	}

	// 提前将用户的参数设置为用户设置的值或全局的值，避免在热点代码中重复判断
	s.users.Range(func(key, value interface{}) bool {
		u := value.(user)

		// speed
		if u.Speed <= 0 {
			u.Speed = s.config.Speed
		}

		// connections
		if u.Connections <= 0 {
			u.Connections = s.config.Connections
		}

		// host
		if u.Host.Number == nil {
			u.Host.Number = s.config.Host.Number
		}
		if u.Host.RegexStr == nil {
			u.Host.RegexStr = s.config.Host.RegexStr
		}
		u.Host.Regex = new([]*regexp.Regexp)
		for _, str := range *u.Host.RegexStr {
			var regex *regexp.Regexp
			regex, err = regexp.Compile(str)
			if err != nil {
				return false
			}
			*u.Host.Regex = append(*u.Host.Regex, regex)
		}
		if u.Host.WithID == nil {
			u.Host.WithID = s.config.Host.WithID
		}

		s.users.Store(key, u)
		return true
	})
	if err != nil {
		return
	}

	return
}

// GetListenerAddrPort 获取 listener 地址，返回值可能为空
func (s *Server) GetListenerAddrPort() (addrPort netip.AddrPort) {
	if s.listener == nil {
		return
	}
	addrPort = s.listener.Addr().(*net.TCPAddr).AddrPort()
	return
}

// GetSNIListenerAddrPort 获取 sni listener 地址，返回值可能为空
func (s *Server) GetSNIListenerAddrPort() (addrPort netip.AddrPort) {
	if s.sniListener == nil {
		return
	}
	addrPort = s.sniListener.Addr().(*net.TCPAddr).AddrPort()
	return
}

// GetTLSListenerAddrPort 获取 tls listener 地址，返回值可能为空
func (s *Server) GetTLSListenerAddrPort() (addrPort netip.AddrPort) {
	if s.tlsListener == nil {
		return
	}
	addrPort = s.tlsListener.Addr().(*net.TCPAddr).AddrPort()
	return
}

// GetQuicListenerAddrPort 获取 QUIC listener 地址，返回值可能为空
func (s *Server) GetQuicListenerAddrPort() (addrPort string) {
	if s.tlsListener == nil {
		return
	}
	addrPort = s.quicListener.Addr().String()
	return
}

// GetAPIListenerAddrPort 获取 api listener 地址，返回值可能为空
func (s *Server) GetAPIListenerAddrPort() (addrPort netip.AddrPort) {
	if s.apiListener == nil {
		return
	}
	addrPort = s.apiListener.Addr().(*net.TCPAddr).AddrPort()
	return
}

// GetSTUNListenerAddrPort 获取 turn listener 地址，返回值可能为空
func (s *Server) GetSTUNListenerAddrPort() (addrPort netip.AddrPort) {
	if s.turnListener == nil {
		return
	}
	addrPort = s.turnListener.LocalAddr().(*net.UDPAddr).AddrPort()
	return
}

func (s *Server) Config() *Config {
	return &s.config
}

func (s *Server) GetConnectionInfo() (info []ConnectionInfo) {
	s.id2Client.Range(func(key, value interface{}) bool {
		client := value.(*client)
		clientInfo := client.GetConnectionInfo()
		info = append(info, clientInfo...)
		return true
	})
	return
}

func MergeConfig(cfg *Config) (err error) {
	defaultConfig := DefaultConfig()
	reflectedSavedConfig := reflect.ValueOf(&cfg.Options).Elem()
	reflectedDefaultConfig := reflect.ValueOf(&defaultConfig.Options).Elem()

	for i := 0; i < reflectedSavedConfig.NumField(); i++ {
		field := reflectedSavedConfig.Field(i)
		if field.IsZero() && field.Kind() != reflect.Slice && reflectedDefaultConfig.Field(i).IsZero() != true {
			field.Set(reflectedDefaultConfig.Field(i))
		}
	}
	return
}

package request

import (
	"encoding/json"
	//cliConfig "github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/config"
	//srvConfig "github.com/isrc-cas/gt/server"
	"net/url"
	"time"
)

//type Client cliConfig.Config
//type Server srvConfig.Config

//Client Settings

// Config is a client config.
type ClientConfig struct {
	Version  string        `json:"version,omitempty"` // 目前未使用
	Services []service     `json:"services,omitempty"`
	Options  ClientOptions `json:"options"`
}

// Options is the config options for a client.
type ClientOptions struct {
	Config                string        `arg:"config" yaml:"-" usage:"The config file path to load" json:"config,omitempty"`
	ID                    string        `yaml:"id" usage:"The unique id used to connect to server. Now it's the prefix of the domain." json:"ID,omitempty"`
	Secret                string        `yaml:"secret" usage:"The secret used to verify the id" json:"secret,omitempty"`
	ReconnectDelay        time.Duration `yaml:"reconnectDelay" usage:"The delay before reconnect. Supports values like '30s', '5m'" json:"reconnectDelay,omitempty"`
	Remote                string        `yaml:"remote" usage:"The remote server url. Supports tcp:// and tls://, default tcp://" json:"remote,omitempty"`
	RemoteSTUN            string        `yaml:"remoteSTUN" usage:"The remote STUN server address" json:"remoteSTUN,omitempty"`
	RemoteAPI             string        `yaml:"remoteAPI" usage:"The API to get remote server url" json:"remoteAPI,omitempty"`
	RemoteCert            string        `yaml:"remoteCert" usage:"The path to remote cert" json:"remoteCert,omitempty"`
	RemoteCertInsecure    bool          `yaml:"remoteCertInsecure"  usage:"Accept self-signed SSL certs from remote" json:"remoteCertInsecure,omitempty"`
	RemoteConnections     uint          `yaml:"remoteConnections" usage:"The max number of server connections in the pool. Valid value is 1 to 10" json:"remoteConnections,omitempty"`
	RemoteIdleConnections uint          `yaml:"remoteIdleConnections" usage:"The number of idle server connections kept in the pool" json:"remoteIdleConnections,omitempty"`
	RemoteTimeout         time.Duration `yaml:"remoteTimeout" usage:"The timeout of remote connections. Supports values like '30s', '5m'" json:"remoteTimeout,omitempty"`

	HostPrefix         config.PositionSlice[string]        `yaml:"hostPrefix" usage:"The server will recognize this host prefix and forward data to local" json:"hostPrefix,omitempty"`
	RemoteTCPPort      config.PositionSlice[uint16]        `yaml:"-" arg:"remoteTCPPort" usage:"The TCP port that the remote server will open" json:"remoteTCPPort,omitempty"`
	RemoteTCPRandom    config.PositionSlice[bool]          `yaml:"-" arg:"remoteTCPRandom" usage:"Whether to choose a random tcp port by the remote server" json:"remoteTCPRandom,omitempty"`
	Local              config.PositionSlice[string]        `yaml:"-" arg:"local" usage:"The local service url" json:"local,omitempty"`
	LocalTimeout       config.PositionSlice[time.Duration] `yaml:"-" arg:"localTimeout" usage:"The timeout of local connections. Supports values like '30s', '5m'" json:"localTimeout,omitempty"`
	UseLocalAsHTTPHost config.PositionSlice[bool]          `yaml:"-" arg:"useLocalAsHTTPHost" usage:"Use the local address as host" json:"useLocalAsHTTPHost,omitempty"`

	SentryDSN         string               `yaml:"sentryDSN" usage:"Sentry DSN to use" json:"sentryDSN,omitempty"`
	SentryLevel       config.Slice[string] `yaml:"sentryLevel" usage:"Sentry levels: trace, debug, info, warn, error, fatal, panic (default [\"error\", \"fatal\", \"panic\"])" json:"sentryLevel,omitempty"`
	SentrySampleRate  float64              `yaml:"sentrySampleRate" usage:"Sentry sample rate for event submission: [0.0 - 1.0]" json:"sentrySampleRate,omitempty"`
	SentryRelease     string               `yaml:"sentryRelease" usage:"Sentry release to be sent with events" json:"sentryRelease,omitempty"`
	SentryEnvironment string               `yaml:"sentryEnvironment" usage:"Sentry environment to be sent with events" json:"sentryEnvironment,omitempty"`
	SentryServerName  string               `yaml:"sentryServerName" usage:"Sentry server name to be reported" json:"sentryServerName,omitempty"`
	SentryDebug       bool                 `yaml:"sentryDebug" usage:"Sentry debug mode, the debug information is printed to help you understand what sentry is doing" json:"sentryDebug,omitempty"`

	WebRTCConnectionIdleTimeout time.Duration `yaml:"webrtcConnectionIdleTimeout" usage:"The timeout of WebRTC connection. Supports values like '30s', '5m'" json:"webRTCConnectionIdleTimeout,omitempty"`
	WebRTCLogLevel              string        `yaml:"webrtcLogLevel" usage:"WebRTC log level: verbose, info, warning, error" json:"webRTCLogLevel,omitempty"`
	WebRTCMinPort               uint16        `yaml:"webrtcMinPort" usage:"The min port of WebRTC peer connection" json:"webRTCMinPort,omitempty"`
	WebRTCMaxPort               uint16        `yaml:"webrtcMaxPort" usage:"The max port of WebRTC peer connection" json:"webRTCMaxPort,omitempty"`

	TCPForwardAddr        string `yaml:"tcpForwardAddr" usage:"The address of TCP forward" json:"TCPForwardAddr,omitempty"`
	TCPForwardHostPrefix  string `yaml:"tcpForwardHostPrefix" usage:"The host prefix of TCP forward" json:"TCPForwardHostPrefix,omitempty"`
	TCPForwardConnections uint   `yaml:"tcpForwardConnections" usage:"The max number of TCP forward peer connections in the pool. Valid value is 1 to 10" json:"TCPForwardConnections,omitempty"`

	LogFile         string `yaml:"logFile" usage:"Path to save the log file" json:"logFile,omitempty"`
	LogFileMaxSize  int64  `yaml:"logFileMaxSize" usage:"Max size of the log files" json:"logFileMaxSize,omitempty"`
	LogFileMaxCount uint   `yaml:"logFileMaxCount" usage:"Max count of the log files" json:"logFileMaxCount,omitempty"`
	LogLevel        string `yaml:"logLevel" usage:"Log level: trace, debug, info, warn, error, fatal, panic, disable" json:"logLevel,omitempty"`
	Version         bool   `arg:"version" yaml:"-" usage:"Show the version of this program" json:"version,omitempty"`
}

type clientURL struct {
	*url.URL
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *clientURL) UnmarshalJSON(b []byte) error {
	var rawURL string
	err := json.Unmarshal(b, &rawURL)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	c.URL = parsedURL
	return nil
}

type service struct {
	HostPrefix         string        `yaml:"hostPrefix" json:"hostPrefix,omitempty"`
	RemoteTCPPort      uint16        `yaml:"remoteTCPPort" json:"remoteTCPPort,omitempty"`
	RemoteTCPRandom    *bool         `yaml:"remoteTCPRandom" json:"remoteTCPRandom,omitempty"`
	LocalURL           clientURL     `yaml:"local" json:"localURL"`
	LocalTimeout       time.Duration `yaml:"localTimeout" json:"localTimeout,omitempty"`
	UseLocalAsHTTPHost bool          `yaml:"useLocalAsHTTPHost" json:"useLocalAsHTTPHost,omitempty"`
}

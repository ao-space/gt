package main

import "C"
import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/lib/client"
	"github.com/isrc-cas/gt/lib/server"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/util"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {}

type op struct {
	OP OPValue `json:"op,omitempty"`
}

type OPValue string

const (
	ready                OPValue = "ready"
	gracefulShutdown     OPValue = "gracefulShutdown"
	gracefulShutdownDone OPValue = "gracefulShutdownDone"
	shutdown             OPValue = "shutdown"
	shutdownDone         OPValue = "shutdownDone"
)

func writeJson(json []byte) (err error) {
	l := [4]byte{}
	binary.BigEndian.PutUint32(l[:], uint32(len(json)))
	_, err = os.Stdout.Write(l[:])
	if err != nil {
		return
	}
	_, err = os.Stdout.Write(json)
	return
}

func readJson() (json []byte, err error) {
	l := [4]byte{}
	_, err = os.Stdin.Read(l[:])
	if err != nil {
		return
	}
	jl := binary.BigEndian.Uint32(l[:])
	if jl > 8*1024 {
		err = errors.New("json too large")
		return
	}
	json = make([]byte, jl)
	_, err = io.ReadFull(os.Stdin, json)
	return
}

func handleStdIO(logger logger.Logger, ch chan os.Signal) {
	go func() {
		var err error
		defer logger.Info().Err(err).Msg("handleStdIO done")
		for {
			var bs []byte
			bs, err = readJson()
			if err != nil {
				return
			}
			var op op
			err = json.Unmarshal(bs, &op)
			if err != nil {
				return
			}
			switch op.OP {
			case gracefulShutdown:
				ch <- syscall.SIGQUIT
			case shutdown:
				ch <- syscall.SIGTERM
			}
		}
	}()
}

func writeOP(logger logger.Logger, op op) {
	bs, err := json.Marshal(op)
	if err != nil {
		logger.Info().Err(err).Interface("op", op).Msg("failed to marshal op")
		return
	}
	err = writeJson(bs)
	if err != nil {
		logger.Info().Err(err).Interface("op", op).Msg("failed to write op")
	}
}

//export RunServer
func RunServer(args []string) {
	util.SetArgs(args)
	s, err := server.New(args, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}
	defer s.Close()

	webServer, err := libserver.StartWebServer(s)
	if err != nil {
		s.Logger.Fatal().Err(err).Msg("failed to start web server")
	}
	defer func() {
		err = libserver.ShutdownWebServer(webServer)
		if err != nil {
			s.Logger.Error().Err(err).Msg("failed to shutdown web server")
		}
	}()

	if len(s.Config().WebAddr) == 0 || libserver.CheckConfigFile(s) {
		err = s.Start()
		if err != nil {
			if len(s.Config().WebAddr) == 0 {
				// web server is not started, exit
				s.Logger.Fatal().Err(err).Msg("failed to start")
			} else {
				// web server is started, continue for web server
				s.Logger.Error().Err(err).Msg("failed to start GT Server, please utilize the web server interface for further GT Server configuration.")
			}
		} else {
			writeOP(s.Logger, op{OP: ready})
		}
	}

	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	handleStdIO(s.Logger, osSig)

	for sig := range osSig {
		s.Logger.Info().Str("signal", sig.String()).Msg("received os signal")
		switch sig {
		case syscall.SIGINT:
			return
		default:
			s.Logger.Info().Msg("wait 3m to stop immediately")
			time.AfterFunc(3*time.Minute, func() {
				os.Exit(1)
			})
			err = libserver.ShutdownWebServer(webServer)
			if err != nil {
				s.Logger.Error().Err(err).Msg("failed to shutdown web server")
			}
			s.Shutdown()
			switch sig {
			case syscall.SIGQUIT:
				writeOP(s.Logger, op{OP: gracefulShutdownDone})
			case syscall.SIGTERM:
				writeOP(s.Logger, op{OP: shutdownDone})
			}
			os.Exit(0)
		}
	}
}

//export RunClient
func RunClient(args []string) {
	util.SetArgs(args)
	c, err := client.New(args, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client")
	}
	defer c.Close()

	webServer, err := libclient.StartWebServer(c)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to start web server")
	}
	defer func() {
		err = libclient.ShutdownWebServer(webServer)
		if err != nil {
			c.Logger.Error().Err(err).Msg("failed to shutdown web server")
		}
	}()

	if len(c.Config().WebAddr) == 0 || libclient.CheckConfigFile(c) {
		err = c.Start()
		if err != nil {
			if len(c.Config().WebAddr) == 0 {
				// web server is not started, exit
				c.Logger.Fatal().Err(err).Msg("failed to start")
			} else {
				// web server is started, continue for web server
				c.Logger.Error().Err(err).Msg("failed to start GT Client, please utilize the web server interface for further GT Client configuration.")
			}
		} else {
			go func() {
				for i := 0; i < 10; i++ {
					if c.WaitUntilReady(30*time.Second) == nil {
						writeOP(c.Logger, op{OP: ready})
						break
					}
				}
			}()
		}
	}

	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	handleStdIO(c.Logger, osSig)

	for sig := range osSig {
		c.Logger.Info().Str("signal", sig.String()).Msg("received os signal")
		switch sig {
		case syscall.SIGHUP:
			// reload the config
			err := c.ReloadServices(args)
			c.Logger.Info().Err(err).Msg("reload services done")
		case syscall.SIGINT:
			return
		default:
			c.Logger.Info().Msg("wait 30s to stop immediately")
			time.AfterFunc(30*time.Second, func() {
				os.Exit(1)
			})
			err = libclient.ShutdownWebServer(webServer)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to shutdown web server")
			}
			c.Shutdown()
			switch sig {
			case syscall.SIGQUIT:
				writeOP(c.Logger, op{OP: gracefulShutdownDone})
			case syscall.SIGTERM:
				writeOP(c.Logger, op{OP: shutdownDone})
			}
			os.Exit(0)
		}
	}
}

/*
 * Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import "C"
import (
	"encoding/json"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/lib/client"
	"github.com/isrc-cas/gt/lib/server"
	"github.com/isrc-cas/gt/logger"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/util"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {}

func handleStdIO(logger logger.Logger, ch chan os.Signal) {
	go func() {
		var err error
		defer logger.Info().Err(err).Msg("handleStdIO done")
		for {
			var bs []byte
			bs, err = util.ReadJson()
			if err != nil {
				return
			}
			var op util.OP
			err = json.Unmarshal(bs, &op)
			if err != nil {
				return
			}
			switch op.OP {
			case util.GracefulShutdown:
				ch <- syscall.SIGQUIT
			case util.Shutdown:
				ch <- syscall.SIGTERM
			}
		}
	}()
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
			err := util.WriteOP(util.OP{OP: util.Ready})
			if err != nil {
				s.Logger.Error().Err(err).Msg("failed to send ready signal to stdio")
			}
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
				err := util.WriteOP(util.OP{OP: util.GracefulShutdownDone})
				if err != nil {
					s.Logger.Error().Err(err).Msg("failed to send graceful shutdown signal to stdio")
				}
			case syscall.SIGTERM:
				err := util.WriteOP(util.OP{OP: util.ShutdownDone})
				if err != nil {
					s.Logger.Error().Err(err).Msg("failed to send shutdown signal to stdio")
				}
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
						err := util.WriteOP(util.OP{OP: util.Ready})
						if err != nil {
							c.Logger.Error().Err(err).Msg("failed to send ready signal to stdio")
						}
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
				err := util.WriteOP(util.OP{OP: util.GracefulShutdownDone})
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to send graceful shutdown signal to stdio")
				}
			case syscall.SIGTERM:
				err := util.WriteOP(util.OP{OP: util.ShutdownDone})
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to send shutdown signal to stdio")
				}
			}
			os.Exit(0)
		}
	}
}

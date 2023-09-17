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

package main

import (
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/server/web"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	s, err := server.New(os.Args, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}
	defer s.Close()
	err = s.Start()
	if err != nil {
		s.Logger.Fatal().Err(err).Msg("failed to start")
	}

	err = startWebServer(s)
	if err != nil {
		s.Logger.Fatal().Err(err).Msg("failed to start web server")
	}

	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	for sig := range osSig {
		s.Logger.Info().Str("signal", sig.String()).Msg("received os signal")
		switch sig {
		case syscall.SIGTERM:
			return
		case syscall.SIGQUIT:
			// restart, start a new process and then shutdown gracefully
			err = shutdownWebServer(s)
			if err != nil {
				s.Logger.Error().Err(err).Msg("failed to shutdown web server")
				continue
			}
			// avoid port conflict
			s.ShutdownWithoutClosingLogger()

			err = runCmd(os.Args)
			if err != nil {
				s.Logger.Error().Err(err).Msg("failed to start new process")
				continue
			}
			s.Logger.Info().Msg("Restart successfully")
			s.Logger.Close()
			os.Exit(0)
		default:
			s.Logger.Info().Msg("wait 3m to stop immediately")
			time.AfterFunc(3*time.Minute, func() {
				os.Exit(1)
			})
			s.Shutdown()
			os.Exit(0)
		}
	}
}

func runCmd(args []string) (err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Process.Release()
	return
}

func startWebServer(s *server.Server) (err error) {
	if s.Config().EnableWebServer {
		err = web.NewWebServer(s)
		if err != nil {
			return
		}
	}
	return
}
func shutdownWebServer(s *server.Server) (err error) {
	if s.Config().EnableWebServer {
		s.Logger.Info().Msg("try to shutdown web server")
		err = web.ShutdownWebServer()
		if err != nil {
			return
		}
		s.Logger.Info().Msg("web server stopped")
	}
	return
}

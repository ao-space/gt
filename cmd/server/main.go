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
	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/server/web"
	"github.com/isrc-cas/gt/util"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
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

	webServer, err := startWebServer(s)
	if err != nil {
		s.Logger.Fatal().Err(err).Msg("failed to start web server")
	}

	if len(s.Config().WebAddr) == 0 || checkConfigFile(s) {
		err = s.Start()
		if err != nil {
			if len(s.Config().WebAddr) == 0 {
				// web server is not started, exit
				s.Logger.Fatal().Err(err).Msg("failed to start")
			} else {
				// web server is started, continue for web server
				s.Logger.Error().Err(err).Msg("failed to start GT Server, please utilize the web server interface for further GT Server configuration.")
			}
		}
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
			err = shutdownWebServer(webServer)
			if err != nil {
				s.Logger.Error().Err(err).Msg("failed to shutdown web server")
				continue
			}
			// avoid port conflict
			s.ShutdownWithoutClosingLogger()

			err = runCmd(os.Args, s)
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

func runCmd(args []string, s *server.Server) (err error) {
	args = checkAndSetLogPath(args, s)
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

func startWebServer(s *server.Server) (*web.Server, error) {
	if len(s.Config().WebAddr) != 0 {
		return web.NewWebServer(s)
	}
	return nil, nil
}
func shutdownWebServer(webServer *web.Server) (err error) {
	if webServer == nil {
		return
	}
	err = webServer.Shutdown()
	return
}

// checkConfigFile checks whether the config file exists to determine whether to start the server
func checkConfigFile(s *server.Server) bool {
	configPath := s.Config().Config
	if len(configPath) == 0 {
		configPath = predef.GetDefaultServerConfigPath()
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}
	return true
}

func checkAndSetLogPath(args []string, s *server.Server) []string {
	if len(s.Config().LogFile) != 0 {
		return args
	}

	if len(args) == 1 || util.Contains(args, "-webAddr") {
		if err := updateConfigLogPath(s); err != nil {
			s.Logger.Error().Err(err).Msg("failed to update config log path")
		}
		return args
	}

	return append(args, "-logFile", predef.GetDefaultServerLogPath())
}

func updateConfigLogPath(s *server.Server) error {
	configPath := s.Config().Config
	if len(configPath) == 0 {
		configPath = predef.GetDefaultServerConfigPath()
	}

	var tmp server.Config
	if err := config.Yaml2Interface(configPath, &tmp); err != nil {
		// ignore error when config file does not exist
		if !os.IsNotExist(err) {
			return err
		}
	}

	if len(tmp.LogFile) != 0 {
		// already set in config file
		return nil
	}

	tmp.LogFile = predef.GetDefaultServerLogPath()
	yamlData, err := yaml.Marshal(&tmp)
	if err != nil {
		return err
	}
	if err = util.WriteYamlToFile(configPath, yamlData); err != nil {
		return err
	}

	return nil
}

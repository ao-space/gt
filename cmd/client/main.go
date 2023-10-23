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
	"github.com/isrc-cas/gt/client/web"
	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/util"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/isrc-cas/gt/client"
	"github.com/rs/zerolog/log"
)

func main() {
	c, err := client.New(os.Args, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client")
	}
	defer c.Close()

	webServer, err := startWebServer(c)
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to start web server")
	}

	if len(c.Config().WebAddr) == 0 || checkConfigFile(c) {
		err = c.Start()
		if err != nil {
			if len(c.Config().WebAddr) == 0 {
				// web server is not started, exit
				c.Logger.Fatal().Err(err).Msg("failed to start")
			} else {
				// web server is started, continue for web server
				c.Logger.Error().Err(err).Msg("failed to start GT Client, please utilize the web server interface for further GT Client configuration.")
			}
		}
	}

	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	for sig := range osSig {
		c.Logger.Info().Str("signal", sig.String()).Msg("received os signal")
		switch sig {
		case syscall.SIGHUP:
			// reload the config
			err := c.ReloadServices()
			c.Logger.Info().Err(err).Msg("reload services done")
		case syscall.SIGTERM:
			return
		case syscall.SIGQUIT:
			// restart, start a new process and then shutdown gracefully
			err = shutdownWebServer(webServer)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to shutdown web server")
				continue
			}
			// avoid port conflict
			c.ShutdownWithoutClosingLogger()

			err = runCmd(os.Args, c)
			if err != nil {
				c.Logger.Error().Err(err).Msg("failed to start new process")
				continue
			}
			// yield control to the new process
			// will use api to wait for connected response of new process before shutdown
			c.Logger.Info().Msg("wait 5s to shutdown gracefully")
			c.Logger.Close()
			os.Exit(0)
		default:
			c.Logger.Info().Msg("wait 30s to stop immediately")
			time.AfterFunc(30*time.Second, func() {
				os.Exit(1)
			})
			c.Shutdown()
			os.Exit(0)
		}
	}
}
func runCmd(args []string, c *client.Client) (err error) {
	args = checkAndSetLogPath(args, c)
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
func startWebServer(c *client.Client) (*web.Server, error) {
	if len(c.Config().WebAddr) != 0 {
		return web.NewWebServer(c)
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
func checkConfigFile(c *client.Client) bool {
	configPath := c.Config().Config
	if len(configPath) == 0 {
		configPath = predef.GetDefaultClientConfigPath()
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}
	return true
}
func checkAndSetLogPath(args []string, c *client.Client) []string {
	if len(c.Config().LogFile) != 0 {
		return args
	}

	if len(args) == 1 || util.Contains(args, "-webAddr") {
		if err := updateConfigLogPath(c); err != nil {
			c.Logger.Error().Err(err).Msg("failed to update log path in config")
		}
		return args
	}

	return append(args, "-logFile", predef.GetDefaultClientLogPath())
}

func updateConfigLogPath(c *client.Client) error {
	configPath := c.Config().Config
	if len(configPath) == 0 {
		configPath = predef.GetDefaultClientConfigPath()
	}

	var tmp client.Config
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

	tmp.LogFile = predef.GetDefaultClientLogPath()
	yamlData, err := yaml.Marshal(&tmp)
	if err != nil {
		return err
	}
	if err = util.WriteYamlToFile(configPath, yamlData); err != nil {
		return err
	}

	return nil
}

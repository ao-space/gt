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
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web"
	"github.com/rs/zerolog/log"
)

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

func main() {
	c, err := client.New(os.Args, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create client")
	}
	defer c.Close()
	err = c.Start()
	if err != nil {
		c.Logger.Fatal().Err(err).Msg("failed to start")
	}

	startWebServer(c)

	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	for {
		select {
		case sig := <-osSig:
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

				if c.Config().EnableWebServer {
					// stop web server
					c.Logger.Info().Msg("try to stop web server")
					err := web.ShutdownWebServer()
					if err != nil {
						c.Logger.Error().Err(err).Msg("failed to stop web server")
						continue
					}
					c.Logger.Info().Msg("web server stopped")
				}

				err := runCmd(os.Args)
				if err != nil {
					c.Logger.Error().Err(err).Msg("failed to start new process")
					continue
				}
				// yield control to the new process
				// will use api to wait for connected response of new process before shutdown
				c.Logger.Info().Msg("wait 5s to shutdown gracefully")
				time.Sleep(5 * time.Second)
				fallthrough
			default:
				c.Logger.Info().Msg("wait 30s to stop immediately")
				time.AfterFunc(30*time.Second, func() {
					os.Exit(1)
				})
				os.Exit(0)
			}
		}
	}
}

func startWebServer(c *client.Client) {
	if c.Config().EnableWebServer {
		web.NewWebServer(c)
	}
}

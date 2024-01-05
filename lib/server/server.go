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

package libserver

import (
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/server/web"
	"github.com/isrc-cas/gt/util"
	"os"
)

func StartWebServer(s *server.Server) (*web.Server, error) {
	if len(s.Config().WebAddr) != 0 {
		return web.NewWebServer(s)
	}
	return nil, nil
}
func ShutdownWebServer(webServer *web.Server) (err error) {
	if webServer == nil {
		return
	}
	err = webServer.Shutdown()
	return
}

// CheckConfigFile checks whether the config file exists to determine whether to start the server
func CheckConfigFile(s *server.Server) bool {
	configPath := s.Config().Config
	if len(configPath) == 0 {
		configPath = util.GetDefaultServerConfigPath()
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}
	return true
}

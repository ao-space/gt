package service

import (
	"errors"
	"fmt"
	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/util"
	"github.com/shirou/gopsutil/v3/net"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func VerifyUser(user request.User, s *server.Server) (err error) {
	if user.Username == s.Config().Admin && user.Password == s.Config().Password {
		return nil
	} else {
		return errors.New("username or password is wrong, please try again")
	}
}

func GetMenu(s *server.Server) (menu []request.Menu) {
	menu = []request.Menu{
		//Home
		{
			Path:      "/home/index",
			Name:      "home",
			Component: "/home/index",
			Meta: request.MetaProps{
				Icon:        "HomeFilled",
				Title:       "Home",
				IsHide:      false,
				IsFull:      false,
				IsAffix:     true,
				IsKeepAlive: false,
			},
		},
		//Connection
		{
			Path:      "/connection",
			Name:      "connection",
			Component: "/connection/index",
			Meta: request.MetaProps{
				Icon:        "Connection",
				Title:       "Connection Status",
				IsHide:      false,
				IsFull:      false,
				IsAffix:     false,
				IsKeepAlive: false,
			},
		},
		//Server Config
		{
			Path:      "/config/server",
			Name:      "server",
			Component: "/config/ServerConfig/index",
			Meta: request.MetaProps{
				Icon:        "Setting",
				Title:       "Server",
				IsHide:      false,
				IsFull:      false,
				IsAffix:     false,
				IsKeepAlive: true,
			},
		},
	}
	//pprof
	if s.Config().EnablePprof {
		pprofLink := fmt.Sprintf("http://%s:%d/debug/pprof/", s.Config().WebAddr, s.Config().WebPort)

		menu = append(menu, request.Menu{
			Path:      "/pprof",
			Name:      "pprof",
			Component: "/pprof/index",
			Meta: request.MetaProps{
				Icon:        "link",
				Title:       "pprof",
				IsLink:      pprofLink,
				IsHide:      false,
				IsFull:      false,
				IsAffix:     false,
				IsKeepAlive: false,
			},
		})
	}
	return
}

// GetConnectionInfo returns the connection info (both in pool and external) of the server
func GetConnectionInfo(s *server.Server) (serverPool []request.SimplifiedConnectionWithID, external []request.SimplifiedConnection, err error) {
	pid := int32(os.Getpid())
	conns, err := net.ConnectionsPid("all", pid)
	if err != nil {
		return
	}
	pools := s.GetConnectionInfo()
	poolsInfo := util.SelectedMatchingConnections(conns, pools)
	externalConnection := util.FilterOutMatchingConnections(conns, util.SwitchToPoolInfo(pools))

	serverPool = util.SimplifyConnectionsWithID(poolsInfo)
	external = util.SimplifyConnections(externalConnection)
	return
}

func GetConfigFromFile(s *server.Server) (cfg server.Config, err error) {
	fullPath := s.Config().Options.Config
	if fullPath == "" {
		err = errors.New("config path is empty")
		return
	}
	err = config.Yaml2Interface(fullPath, &cfg)
	if err != nil {
		return
	}
	return
}

func SaveConfigToFile(cfg *server.Config) (fullPath string, err error) {
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}
	if cfg.Options.Config != "" {
		fullPath = cfg.Options.Config
	} else {
		fullPath = filepath.Join(util.GetAppDir(), "server.yaml")
		cfg.Options.Config = fullPath
	}
	err = util.WriteYamlToFile(fullPath, yamlData)
	if err != nil {
		return
	}
	return
}

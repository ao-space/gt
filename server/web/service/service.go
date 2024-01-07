package service

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/predef"
	"github.com/isrc-cas/gt/server"
	util2 "github.com/isrc-cas/gt/util"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/util"
	"github.com/jinzhu/copier"
	"github.com/shirou/gopsutil/v3/net"
	"gopkg.in/yaml.v3"
	"os"
)

func VerifyUser(user request.User, s *server.Server) (err error) {
	if user.Username == s.Config().Admin && user.Password == s.Config().Password {
		return nil
	} else {
		return errors.New("username or password is wrong, please try again")
	}
}

func ChangeUserInfo(user request.UserInfo, s *server.Server) error {
	cfg, err := InheritConfig(s)
	if err != nil {
		return err
	}
	cfg.Admin = user.Username
	cfg.Password = user.Password
	cfg.EnablePprof = user.EnablePprof

	conf4Log := cfg
	conf4Log.Password = "******"
	conf4Log.SigningKey = "******"

	_, err = SaveConfigToFile(&cfg)
	if err != nil {
		return err
	}
	s.Logger.Info().Str("config", "change user info").Msg(spew.Sdump(conf4Log))
	return nil
}

func GetMenu(s *server.Server, lang string) (menu []request.Menu) {
	if lang == "zh" {
		menu = []request.Menu{
			//Home
			{
				Path:      "/home/index",
				Name:      "home",
				Component: "/home/index",
				Meta: request.MetaProps{
					Icon:        "HomeFilled",
					Title:       "主页",
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
					Title:       "连接状态",
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
					Title:       "服务端",
					IsHide:      false,
					IsFull:      false,
					IsAffix:     false,
					IsKeepAlive: true,
				},
			},
		}
	} else {
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
	}
	//pprof
	if s.Config().EnablePprof {
		enableHTTPS := true
		if len(s.Config().WebCertFile) == 0 && len(s.Config().WebKeyFile) == 0 {
			enableHTTPS = false
		}

		webUrl, err := util.SwitchToValidWebURL(s.Config().WebAddr, enableHTTPS)
		if err != nil {
			s.Logger.Error().Err(err).Msg("switch to valid web url failed while getting pprof link")
			return
		}
		pprofLink := fmt.Sprintf("%s/debug/pprof/", webUrl)

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
	if len(fullPath) == 0 {
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
		fullPath = predef.GetDefaultServerConfigPath()
		cfg.Options.Config = fullPath
	}
	err = util2.WriteYamlToFile(fullPath, yamlData)
	if err != nil {
		return
	}
	return
}

// InheritImmutableConfigFields copy immutable fields from original to new
func InheritImmutableConfigFields(original *server.Config, new *server.Config) (err error) {
	if original == nil {
		err = errors.New("original config is nil")
		return
	}
	new.Config = original.Config
	new.WebAddr = original.WebAddr
	new.WebCertFile = original.WebCertFile
	new.WebKeyFile = original.WebKeyFile
	new.EnablePprof = original.EnablePprof
	new.SigningKey = original.SigningKey
	new.Admin = original.Admin
	new.Password = original.Password
	return
}

func InheritConfig(s *server.Server) (cfg server.Config, err error) {
	// Get From File
	cfg, err = GetConfigFromFile(s)
	if err != nil {
		// Get From Running
		err = copier.Copy(&cfg, s.Config()) // SigningKey is also copied
		if err != nil {
			return
		}
	}
	return
}

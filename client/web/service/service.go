package service

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/predef"
	util2 "github.com/isrc-cas/gt/util"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/util"
	"github.com/jinzhu/copier"
	"github.com/shirou/gopsutil/v3/net"

	"gopkg.in/yaml.v3"
	"os"
)

func VerifyUser(user request.User, c *client.Client) (err error) {
	if user.Username == c.Config().Admin && user.Password == c.Config().Password {
		return nil
	} else {
		return errors.New("username or password is wrong, please try again")
	}
}

// ChangeUserInfo Match the user information in the configuration file
func ChangeUserInfo(user request.UserInfo, c *client.Client) error {
	cfg, err := InheritConfig(c)
	if err != nil {
		return err
	}
	cfg.Admin = user.Username
	cfg.Password = user.Password
	cfg.EnablePprof = user.EnablePprof

	conf4log := cfg
	conf4log.Password = "******"
	conf4log.SigningKey = "******"
	_, err = SaveConfigToFile(&cfg)
	if err != nil {
		return err
	}
	c.Logger.Info().Str("config", "change user info").Msg(spew.Sdump(conf4log))
	return nil
}

func GetMenu(c *client.Client, lang string) (menu []request.Menu) {
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
			//Client Config
			{
				Path:      "/config/client",
				Name:      "client",
				Component: "/config/ClientConfig/index",
				Meta: request.MetaProps{
					Icon:        "Setting",
					Title:       "客户端",
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
			//Client Config
			{
				Path:      "/config/client",
				Name:      "client",
				Component: "/config/ClientConfig/index",
				Meta: request.MetaProps{
					Icon:        "Setting",
					Title:       "Client",
					IsHide:      false,
					IsFull:      false,
					IsAffix:     false,
					IsKeepAlive: true,
				},
			},
		}
	}
	//pprof
	if c.Config().EnablePprof {
		enableHTTPS := true
		if len(c.Config().WebCertFile) == 0 && len(c.Config().WebKeyFile) == 0 {
			enableHTTPS = false
		}
		webUrl, err := util.SwitchToValidWebURL(c.Config().WebAddr, enableHTTPS)
		if err != nil {
			c.Logger.Error().Err(err).Msg("switch to valid web url failed while getting pprof link")
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

func GetConnectionPoolStatus(c *client.Client) map[uint]client.Status {
	return c.GetConnectionPoolStatus()
}

// GetConnectionInfo get all connections of current process
// except connections of pools
func GetConnectionInfo(c *client.Client) (info []request.SimplifiedConnection, err error) {
	pid := int32(os.Getpid())
	conns, err := net.ConnectionsPid("all", pid)
	if err != nil {
		return
	}
	pools := c.GetConnectionPoolNetInfo()

	filter := util.FilterOutMatchingConnections(conns, pools)
	info = util.SimplifyConnections(filter)
	return
}

// GetConfigFromFile need to set configPath before
func GetConfigFromFile(c *client.Client) (cfg client.Config, err error) {
	fullPath := c.Config().Options.Config
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

// SaveConfigToFile if didn't set configPath before,
// it will save config to clientConfig.yaml by default
func SaveConfigToFile(cfg *client.Config) (fullPath string, err error) {

	// switch config type to yaml
	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		return
	}
	if cfg.Config != "" {
		fullPath = cfg.Options.Config
	} else {
		fullPath = predef.GetDefaultClientConfigPath()
		cfg.Options.Config = fullPath
	}
	err = util2.WriteYamlToFile(fullPath, yamlData)
	if err != nil {
		return
	}
	return
}

// InheritImmutableConfigFields copy immutable fields from original to new
func InheritImmutableConfigFields(original *client.Config, new *client.Config) (err error) {
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
func InheritConfig(c *client.Client) (cfg client.Config, err error) {
	// Get From File
	cfg, err = GetConfigFromFile(c)
	if err != nil {
		// Get From Running
		err = copier.Copy(&cfg, c.Config()) // SigningKey is also copied
		if err != nil {
			return
		}
	}
	return
}

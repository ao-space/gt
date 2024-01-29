package service

import (
	"bytes"
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
	"reflect"

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
	SeparateConfig(&cfg)
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
	buff := bytes.Buffer{}
	yamlEncoder := yaml.NewEncoder(&buff)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(cfg)
	if err != nil {
		return
	}
	if cfg.Config != "" {
		fullPath = cfg.Options.Config
	} else {
		fullPath = predef.GetDefaultClientConfigPath()
		cfg.Options.Config = fullPath
	}
	err = util2.WriteYamlToFile(fullPath, buff.Bytes())
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
	new.ConfigType = original.ConfigType
	return
}
func InheritConfig(c *client.Client) (cfg client.Config, err error) {
	// Get From File
	cfg, err = GetMergedConfig(c)
	if err != nil {
		// Get From Running
		err = copier.Copy(&cfg, c.Config()) // SigningKey is also copied
		if err != nil {
			return
		}
	}
	return
}
func GetMergedConfig(c *client.Client) (cfg client.Config, err error) {
	cfg, err = GetConfigFromFile(c)
	if err != nil {
		return
	}
	defaultConfig := client.DefaultConfig()
	reflectedSavedConfig := reflect.ValueOf(&cfg.Options).Elem()
	reflectedDefaultConfig := reflect.ValueOf(&defaultConfig.Options).Elem()

	for i := 0; i < reflectedSavedConfig.NumField(); i++ {
		field := reflectedSavedConfig.Field(i)
		if field.IsZero() && field.Kind() != reflect.Slice && reflectedDefaultConfig.Field(i).IsZero() != true {
			reflectedSavedConfig.Field(i).Set(reflectedDefaultConfig.Field(i))
		}
	}
	return
}

func SeparateConfig(newConfig *client.Config) {
	defaultConfig := client.DefaultConfig()
	reflectedNewConfig := reflect.ValueOf(&newConfig.Options).Elem()
	reflectedOldConfig := reflect.ValueOf(&defaultConfig.Options).Elem()

	for i := 0; i < reflectedNewConfig.NumField(); i++ {
		field := reflectedNewConfig.Field(i)
		if field.Kind() == reflect.Slice {
			continue
		}
		if reflect.DeepEqual(reflectedNewConfig.Field(i).Interface(), reflectedOldConfig.Field(i).Interface()) {
			field.SetZero()
		}
	}
	for index := range newConfig.Services {
		if newConfig.Services[index].RemoteTCPRandom != nil && *newConfig.Services[index].RemoteTCPRandom == false {
			newConfig.Services[index].RemoteTCPRandom = nil
		}
	}
}

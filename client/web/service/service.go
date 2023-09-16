package service

import (
	"errors"
	"fmt"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/config"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/util"
	"github.com/shirou/gopsutil/v3/net"

	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func VerifyUser(user request.User, c *client.Client) (err error) {
	if user.Username == c.Config().Admin && user.Password == c.Config().Password {
		return nil
	} else {
		return errors.New("username or password is wrong, please try again")
	}
}

func GenerateToken(signingKey string, user request.User) (token string, err error) {
	j := util.NewJWT(signingKey)
	claims := j.CreateClaims(user.Username, "gt-client")
	token, err = j.CreateToken(claims)
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetMenu(c *client.Client) (menu []request.Menu) {
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
	//pprof
	if c.Config().EnablePprof {
		pprofLink := fmt.Sprintf("http://%s:%d/debug/pprof", c.Config().WebAddr, c.Config().WebPort)
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

// SaveConfigToFile if didn't set configPath before,
// it will save config to clientConfig.yaml by default
func SaveConfigToFile(cfg *client.Config) (fullPath string, err error) {

	// switch config type to yaml
	yamlData, err := yaml.Marshal(&cfg)
	if err != nil {
		return
	}
	if cfg.Config != "" {
		fullPath = cfg.Config
	} else {
		fullPath = filepath.Join(util.GetAppDir(), "client.yaml")
	}
	err = util.WriteYamlToFile(fullPath, yamlData)
	if err != nil {
		return
	}
	return
}

// SendSignal will create a new process to restart the services
func SendSignal(signal string) (err error) {
	execPath, err := os.Executable()
	if err != nil {
		return
	}
	var cmd *exec.Cmd
	switch signal {
	case "reload":
		cmd = exec.Command(execPath, "-s", "reload")
	case "restart":
		cmd = exec.Command(execPath, "-s", "restart")
	case "stop":
		cmd = exec.Command(execPath, "-s", "stop")
	case "kill":
		cmd = exec.Command(execPath, "-s", "kill")
	default:
		err = errors.New("unknown signal")
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		return
	}
	err = cmd.Process.Release()
	return
}
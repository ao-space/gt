package service

import (
	"errors"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/client/web/model/request"
	"github.com/isrc-cas/gt/client/web/util"
	"github.com/isrc-cas/gt/config"
	psNet "github.com/shirou/gopsutil/v3/net"
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
	claims := j.CreateClaims(user.Username)
	token, err = j.CreateToken(claims)
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetServerInfo() (server *util.Server, err error) {
	var s util.Server
	s.Os = util.InitOS()
	if s.Cpu, err = util.InitCPU(); err != nil {
		return nil, err
	}
	if s.Ram, err = util.InitRAM(); err != nil {
		return nil, err
	}
	if s.Disk, err = util.InitDisk(); err != nil {
		return nil, err
	}
	return &s, nil
}

func GetConnectionPoolStatus(c *client.Client) map[uint]client.Status {
	return c.GetConnectionPoolStatus()
}

// GetConnectionInfo get all connections of current process
// except connections of pools
func GetConnectionInfo(c *client.Client) (info []request.SimplifiedConnection, err error) {
	pid := int32(os.Getpid())
	conns, err := psNet.ConnectionsPid("all", pid)
	if err != nil {
		return
	}
	pools := c.GetConnectionPoolNetInfo()

	filter := util.FilterOutMatchingConnections(conns, pools)
	info = util.SimplifyConnections(filter)
	return
}

// GetConfigFormFile need to set configPath before
func GetConfigFormFile(c *client.Client) (cfg client.Config, err error) {
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
		fullPath = filepath.Join(util.GetAppDir(), "clientConfig.yaml")
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

package util

import (
	"errors"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func OpenBrowser(webUrl string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", webUrl)
	case "darwin":
		cmd = exec.Command("open", webUrl)
	case "linux":
		// Check if running as root
		if os.Geteuid() == 0 {
			originalUser := os.Getenv("SUDO_USER")
			if originalUser == "" {
				originalUser = "nobody" // fallback user if SUDO_USER is not set
			}
			cmd = exec.Command("sudo", "-u", originalUser, "xdg-open", webUrl)
		} else {
			cmd = exec.Command("xdg-open", webUrl)
		}
	default:
		return errors.New("unsupported platform to open browser")
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return nil
}
func SwitchToValidWebURL(addr string, enableHTTPS bool) (string, error) {
	if strings.IndexByte(addr, ':') == -1 {
		addr = ":" + addr
	}
	if strings.HasPrefix(addr, ":") {
		addr = "localhost" + addr
	}
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		if enableHTTPS {
			addr = "https://" + addr
		} else {
			addr = "http://" + addr
		}
	}
	_, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	return addr, nil
}
func CreateUrlWithTempKey(url string, tempKey string) string {
	return url + "/#/verify?key=" + tempKey
}

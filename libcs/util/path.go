package util

import (
	"os"
	"path/filepath"
)

func GetAppDir() string {
	if path, err := os.Getwd(); err == nil {
		return path
	}
	return "."
}

func WriteYamlToFile(fullPath string, data []byte) error {

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func Contains(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

var (
	defaultClientConfigPath string
	defaultClientLogPath    string
	defaultServerConfigPath string
	defaultServerLogPath    string
)

func init() {
	defaultClientConfigPath = filepath.Join(GetAppDir(), "client.yaml")
	defaultClientLogPath = filepath.Join(GetAppDir(), "client.log")
	defaultServerConfigPath = filepath.Join(GetAppDir(), "server.yaml")
	defaultServerLogPath = filepath.Join(GetAppDir(), "server.log")
}

func GetDefaultClientConfigPath() string {
	return defaultClientConfigPath
}
func GetDefaultClientLogPath() string {
	return defaultClientLogPath
}
func GetDefaultServerConfigPath() string {
	return defaultServerConfigPath
}
func GetDefaultServerLogPath() string {
	return defaultServerLogPath
}

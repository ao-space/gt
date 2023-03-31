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

package config

import (
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Version string
	Options
}

type Options struct {
	Config              string        `arg:"config" yaml:"-" usage:"The config file path to load"`
	ID                  string        `yaml:"id" usage:"The unique id used to connect to server"`
	Secret              Slice[string] `arg:"secret" yaml:"-" usage:"The secret for user id"`
	Server              string        `yaml:"server" usage:"The server url"`
	ServerCert          string        `yaml:"serverCert" usage:"The cert path of server"`
	ServerCertInsecure  bool          `yaml:"serverCertInsecure" usage:"Accept self-signed SSL certs from the server"`
	ServerTimeout       time.Duration `yaml:"serverTimeout" usage:"The timeout for server connections"`
	Service             string        `yaml:"service" usage:"The service gateway url"`
	ServiceCert         string        `yaml:"serviceCert" usage:"The cert path of service gateway"`
	ServiceCertInsecure bool          `yaml:"serviceCertInsecure" usage:"Accept self-signed SSL certs from the service gateway"`
	ServiceTimeout      time.Duration `yaml:"serviceTimeout" usage:"The timeout for service gateway connections"`
	LogFile             string        `yaml:"logFile" usage:"Path to save the log file"`
	LogFileMaxSize      int64         `yaml:"logFileMaxSize" usage:"Max size of the log files"`
	LogFileMaxCount     uint          `yaml:"logFileMaxCount" usage:"Max count of the log files"`
	LogLevel            string        `yaml:"logLevel" usage:"Log level: trace, debug, info, warn, error, fatal, panic, disable"`
	Version             bool          `arg:"version" yaml:"-" usage:"Show the version of this program"`
}

func defaultConfig() Config {
	return Config{
		Options: Options{
			ServerTimeout:   120 * time.Second,
			ServiceTimeout:  120 * time.Second,
			LogFileMaxCount: 7,
			LogFileMaxSize:  512 * 1024 * 1024,
			LogLevel:        zerolog.InfoLevel.String(),
		},
	}
}

func TestParseFlags(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    Config
	}{
		{
			"config",
			args{[]string{"net", "-config", "testdata/config.yaml"}},
			false,
			Config{
				Version: "1.0",
				Options: Options{
					Config:              "testdata/config.yaml",
					Server:              "tls://localhost:443",
					ServerCertInsecure:  true,
					ServerTimeout:       60 * time.Second,
					Service:             "https://localhost:8443",
					ServiceCertInsecure: true,
					ServiceTimeout:      180 * time.Second,
					LogFileMaxCount:     7,
					LogFileMaxSize:      512 * 1024 * 1024,
					LogLevel:            zerolog.InfoLevel.String(),
				},
			},
		},
		{
			"overwrite config",
			args{[]string{"net", "-config", "testdata/config.yaml", "-server", "tls://localhost:9443", "-logFileMaxCount", "8", "-secret", "1", "-secret", "2"}},
			false,
			Config{
				Version: "1.0",
				Options: Options{
					Config:              "testdata/config.yaml",
					Server:              "tls://localhost:9443",
					ServerCertInsecure:  true,
					ServerTimeout:       60 * time.Second,
					Service:             "https://localhost:8443",
					ServiceCertInsecure: true,
					ServiceTimeout:      180 * time.Second,
					LogFileMaxCount:     8,
					LogFileMaxSize:      512 * 1024 * 1024,
					LogLevel:            zerolog.InfoLevel.String(),
					Secret:              Slice[string]{"1", "2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := defaultConfig()
			if err := ParseFlags(tt.args.args, &config, &config.Options); (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(&tt.want, &config) {
				t.Errorf("ParseFlags() got = \n%#v\n, want \n%#v", config, tt.want)
			}
		})
	}
}

type positionSlice struct {
	BoolPositionSlice     PositionSlice[bool]          `yaml:"boolPositionSlice"`
	DurationPositionSlice PositionSlice[time.Duration] `yaml:"durationPositionSlice"`
	StringPositionSlice   PositionSlice[string]        `yaml:"stringPositionSlice"`
	Uint16PositionSlice   PositionSlice[uint16]        `yaml:"uint16PositionSlice"`
}

func TestPositionFromCommandLine(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedResult positionSlice
	}{
		{
			name: "normal",
			args: []string{
				"net",
				"-boolPositionSlice=true",  // bool 不能使用空格，flag 不支持这种方式
				"-boolPositionSlice=false", // bool 不能使用空格，flag 不支持这种方式
				"-durationPositionSlice", "1s",
				"-durationPositionSlice", "2s",
				"-stringPositionSlice", "abc",
				"-stringPositionSlice", "def",
				"-uint16PositionSlice", "1",
				"-uint16PositionSlice", "2",
			},
			expectedResult: positionSlice{
				BoolPositionSlice: PositionSlice[bool]{
					{
						Value:    true,
						Position: 0,
					},
					{
						Value:    false,
						Position: 1,
					},
				},
				DurationPositionSlice: PositionSlice[time.Duration]{
					{
						Value:    1 * time.Second,
						Position: 2,
					},
					{
						Value:    2 * time.Second,
						Position: 3,
					},
				},
				StringPositionSlice: PositionSlice[string]{
					{
						Value:    "abc",
						Position: 4,
					},
					{
						Value:    "def",
						Position: 5,
					},
				},
				Uint16PositionSlice: PositionSlice[uint16]{
					{
						Value:    1,
						Position: 6,
					},
					{
						Value:    2,
						Position: 7,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := positionSlice{}
			err := ParseFlags(tt.args, &result, &result)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&result, &tt.expectedResult) {
				t.Fatalf("not equal \nresult: %#v\nexpected result: %#v\n", result, tt.expectedResult)
			}
		})
	}
}

type pointer struct {
	Config              string         `yaml:"config"`
	BoolPointer1        *bool          `yaml:"boolPointer1"`        // 命令行
	BoolPointer2        *bool          `yaml:"boolPointer2"`        // 配置文件
	BoolPointer3        *bool          `yaml:"boolPointer3"`        // 空指针
	StringSlicePointer1 *Slice[string] `yaml:"stringSlicePointer1"` // 命令行
	StringSlicePointer2 *Slice[string] `yaml:"stringSlicePointer2"` // 配置文件
	StringSlicePointer3 *Slice[string] `yaml:"stringSlicePointer3"` // 空指针
	Uint32Pointer1      *uint32        `yaml:"uint32Pointer1"`      // 命令行
	Uint32Pointer2      *uint32        `yaml:"uint32Pointer2"`      // 配置文件
	Uint32Pointer3      *uint32        `yaml:"uint32Pointer3"`      // 空指针
}

func TestPointer(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedResult pointer
	}{
		{
			name: "normal",
			args: []string{
				"net",
				"-config", "testdata/pointer.yaml",
				"-boolPointer1=true", // bool 不能使用空格，flag 不支持这种方式
				"-stringSlicePointer1", "^a$",
				"-stringSlicePointer1", "^b$",
				"-uint32Pointer1", "1",
			},
			expectedResult: pointer{
				Config:              "testdata/pointer.yaml",
				BoolPointer1:        &[]bool{true}[0],
				BoolPointer2:        &[]bool{false}[0],
				StringSlicePointer1: &Slice[string]{"^a$", "^b$"},
				StringSlicePointer2: &Slice[string]{"^c$", "^d$"},
				Uint32Pointer1:      &[]uint32{1}[0],
				Uint32Pointer2:      &[]uint32{2}[0],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pointer{}
			err := ParseFlags(tt.args, &result, &result)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&result, &tt.expectedResult) {
				t.Fatalf("not equal \nresult: %#v\nexpected result: %#v\n", result, tt.expectedResult)
			}
		})
	}
}

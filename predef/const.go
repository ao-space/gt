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

package predef

import (
	"github.com/isrc-cas/gt/util"
	"os"
	"path/filepath"
	"time"
)

const (
	// MinIDSize is the minimum size of ID
	MinIDSize = 1
	// MaxIDSize is the maximum size of ID
	MaxIDSize = 200
	// DefaultIDSize ID 的默认长度
	DefaultIDSize = 64
	// MinSecretSize 表示 secret 长度的最小值
	MinSecretSize = MinIDSize
	// MaxSecretSize 表示 secret 长度的最大值
	MaxSecretSize = MaxIDSize
	// DefaultSecretSize secret 的默认长度
	DefaultSecretSize = DefaultIDSize
	// DefaultSigningKeySize signing key 的默认长度
	DefaultSigningKeySize = 32
	// DefaultAdminSize admin 的默认长度
	DefaultAdminSize = 8
	// DefaultPasswordSize password 的默认长度
	DefaultPasswordSize = 8
	// DefaultTokenDuration token 的默认有效期
	DefaultTokenDuration = 30 * time.Minute
	// MinHostPrefixSize 表示 host 前缀长度的最小值
	MinHostPrefixSize = MinIDSize
	// MaxHostPrefixSize 表示 host 前缀长度的最大值
	MaxHostPrefixSize = MaxIDSize
	// MaxHTTPHeaderSize max ending of host in http headers
	MaxHTTPHeaderSize = 2 * 1024
)

// OP is the type of operations
type OP = uint16

const (
	// Data is a data operation
	Data OP = iota
	// Close is a close operation
	Close
	// ServicesData is a multiple service data
	ServicesData
)

// 通信协议的 option
// 扩展规则
//
//	前两位 + 1 为 option 的长度，单位是字节
//	比如 0100 0001 0000 0000，开头的 01 表示该 option 长度为 2 字节，后 14 位为 option 的内容
var (
	IDAsHostPrefix      = []byte{0}
	OpenTCPPort         = []byte{1}
	OptionAndNextOption = []byte{2}
	OpenHost            = []byte{3}
	IDAsTLSHostPrefix   = []byte{4}
	OpenTLSHost         = []byte{5}
)

// MagicNumber 常量数字，见 https://en.wikipedia.org/wiki/Magic_number_(programming)
const MagicNumber byte = 0xF0

var (
	defaultClientConfigPath string
	defaultClientLogPath    string
	defaultServerConfigPath string
	defaultServerLogPath    string
)

// IsNoArgs 表示是否没有参数, 用于判断是否需要提供更多的默认配置
var isNoArgs = false

func init() {
	defaultClientConfigPath = filepath.Join(util.GetAppDir(), "client.yaml")
	defaultClientLogPath = filepath.Join(util.GetAppDir(), "client.log")
	defaultServerConfigPath = filepath.Join(util.GetAppDir(), "server.yaml")
	defaultServerLogPath = filepath.Join(util.GetAppDir(), "server.log")
	if len(os.Args) <= 1 {
		isNoArgs = true
	}
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
func IsNoArgs() bool {
	return isNoArgs
}

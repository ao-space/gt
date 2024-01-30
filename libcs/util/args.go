package util

import (
	"os"
	"sync/atomic"
)

var (
	NoArgs     atomic.Bool
	globalArgs atomic.Pointer[[]string]
)

func init() {
	if len(os.Args) <= 1 {
		NoArgs.Store(true)
	}
}

// IsNoArgs 表示是否没有参数, 用于判断是否需要提供更多的默认配置
func IsNoArgs() bool {
	return NoArgs.Load()
}

func Args() []string {
	load := globalArgs.Load()
	if load != nil {
		return *load
	}
	return os.Args
}

func SetArgs(args []string) {
	if len(args) <= 1 {
		NoArgs.Store(true)
	}
	globalArgs.Store(&args)
}

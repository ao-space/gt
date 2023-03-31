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
	"fmt"
	"reflect"
	"time"
)

// BasicType time.Duration 包含在 ~int64
type BasicType interface {
	~string |
		~bool |
		~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uintptr | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Slice 切片类型的命令行参数
type Slice[T BasicType] []T

// String flag.Value 接口
func (s *Slice[T]) String() string {
	return fmt.Sprintf("%v", *s)
}

// Set flag.Value 接口
func (s *Slice[T]) Set(value string) error {
	switch any(*s).(type) {
	case Slice[time.Duration]:
		v, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		reflect.ValueOf(s).Elem().Set(
			reflect.Append(
				reflect.ValueOf(s).Elem(),
				reflect.ValueOf(v),
			),
		) // TODO 目前泛型还不支持 switch，所以只能用 reflect
	default:
		var empty T
		*s = append(*s, empty)
		fmt.Sscanf(value, "%v", &(*s)[len(*s)-1])
	}
	return nil
}

// Get flag.Getter 接口
func (s *Slice[T]) Get() interface{} {
	return *s
}

// IsBoolFlag flag.boolFlag 接口
func (s *Slice[T]) IsBoolFlag() bool {
	_, ok := any(s).(*Slice[bool])
	return ok
}

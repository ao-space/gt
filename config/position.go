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
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

// Position 带位置信息的命令行参数
type Position[T BasicType] struct {
	Value    T
	Position uint32
}

// String flag.Value 接口
func (b *Position[T]) String() string {
	return fmt.Sprintf("Value: %v Position: %v", b.Value, b.Position)
}

// Set flag.Value 接口
func (b *Position[T]) Set(value string) error {
	switch any(b.Value).(type) {
	case time.Duration:
		v, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		reflect.ValueOf(&b.Value).Elem().Set(reflect.ValueOf(v)) // TODO 目前泛型还不支持 switch，所以只能用 reflect
	default:
		fmt.Sscanf(value, "%v", &b.Value)
	}
	b.Position = atomic.LoadUint32(&position)
	atomic.AddUint32(&position, 1)
	if atomic.LoadUint32(&position) == 0 {
		return errors.New("position out of uint32 range")
	}
	return nil
}

// Get flag.Getter 接口
func (b *Position[T]) Get() interface{} {
	return *b
}

// IsBoolFlag flag.boolFlag 接口
func (b *Position[T]) IsBoolFlag() bool {
	_, ok := any(b.Value).(*bool)
	return ok
}

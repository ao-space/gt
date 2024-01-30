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
)

// PositionSlice 带位置信息的切片类型的命令行参数
type PositionSlice[T BasicType] []Position[T]

// String flag.Value 接口
func (p *PositionSlice[T]) String() string {
	return fmt.Sprintf("%v", *p)
}

// Set flag.Value 接口
func (p *PositionSlice[T]) Set(value string) error {
	var empty Position[T]
	err := empty.Set(value)
	if err != nil {
		return err
	}
	*p = append(*p, empty)
	return nil
}

// Get flag.Getter 接口
func (p *PositionSlice[T]) Get() interface{} {
	return *p
}

// IsBoolFlag flag.boolFlag 接口
func (p *PositionSlice[T]) IsBoolFlag() bool {
	_, ok := any(p).(*PositionSlice[bool])
	return ok
}

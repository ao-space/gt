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

package util

import (
	"errors"
	"strconv"
	"strings"
)

// PortRange 端口范围
type PortRange struct {
	Min uint16
	Max uint16
}

// NewPortRangeFromString 新建 PortRange，需要处理 3 种情况
// 22-80 -> 22-80
// 80 -> 80-80
// 0 -> 1-65535
func NewPortRangeFromString(portRangeStr string) (_ *PortRange, err error) {
	var min uint64
	var max uint64

	i := strings.IndexByte(portRangeStr, '-')
	if i == -1 {
		port, err := strconv.ParseUint(portRangeStr, 10, 16)
		if err != nil {
			return nil, err
		}
		if port == 0 {
			min = 1
			max = 65535
		} else {
			min = port
			max = port
		}
		return NewPortRangeFromNumber(uint16(min), uint16(max))
	}

	min, err = strconv.ParseUint(portRangeStr[:i], 10, 16)
	if err != nil {
		return nil, err
	}
	max, err = strconv.ParseUint(portRangeStr[i+1:], 10, 16)
	if err != nil {
		return nil, err
	}
	return NewPortRangeFromNumber(uint16(min), uint16(max))
}

// NewPortRangeFromNumber 新建 PortRange
func NewPortRangeFromNumber(min, max uint16) (*PortRange, error) {
	pr := &PortRange{}

	if min == 0 {
		min = 1
	}

	pr.Min = min
	pr.Max = max
	if pr.Min > pr.Max {
		return nil, errors.New("the minimum value is greater than the maximum value")
	}
	return pr, nil
}

func (pr *PortRange) String() string {
	return "{Min: " + strconv.Itoa(int(pr.Min)) + ", Max: " + strconv.Itoa(int(pr.Max)) + "}"
}

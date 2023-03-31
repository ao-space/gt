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
	"sync/atomic"
	"testing"
	"time"
)

func TestPosition(t *testing.T) {
	// 重置 position，不能使用 t.Parallel()
	atomic.StoreUint32(&position, 0)

	// time.Duration 这是一个特殊的类型
	var durationPosition Position[time.Duration]
	err := durationPosition.Set("1s")
	if err != nil {
		t.Fatal(err)
	}
	if durationPosition.String() != "Value: 1s Position: 0" {
		t.Fatal("durationPosition.String() != \"Value: 1s Position: 0\"")
	}

	// string
	var stringPosition Position[string]
	err = stringPosition.Set("hello")
	if err != nil {
		t.Fatal(err)
	}
	if stringPosition.String() != "Value: hello Position: 1" {
		t.Fatal("stringPosition.String() != \"Value: hello Position: 1\"")
	}

	// bool
	var boolPosition Position[bool]
	err = boolPosition.Set("true")
	if err != nil {
		t.Fatal(err)
	}
	if boolPosition.String() != "Value: true Position: 2" {
		t.Fatal("boolPosition.String() != \"Value: true Position: 2\"")
	}

	// float32
	var float32Position Position[float32]
	err = float32Position.Set("1.1")
	if err != nil {
		t.Fatal(err)
	}
	if float32Position.String() != "Value: 1.1 Position: 3" {
		t.Fatal("float32Position.String() != \"Value: 1.1 Position: 3\"")
	}

	// float64
	var float64Position Position[float64]
	err = float64Position.Set("1.1")
	if err != nil {
		t.Fatal(err)
	}
	if float64Position.String() != "Value: 1.1 Position: 4" {
		t.Fatal("float64Position.String() != \"Value: 1.1 Position: 4\"")
	}

	// int
	var intPosition Position[int]
	err = intPosition.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if intPosition.String() != "Value: 1 Position: 5" {
		t.Fatal("intPosition.String() != \"Value: 1 Position: 5\"")
	}

	// int8
	var int8Position Position[int8]
	err = int8Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if int8Position.String() != "Value: 1 Position: 6" {
		t.Fatal("int8Position.String() != \"Value: 1 Position: 6\"")
	}

	// int16
	var int16Position Position[int16]
	err = int16Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if int16Position.String() != "Value: 1 Position: 7" {
		t.Fatal("int16Position.String() != \"Value: 1 Position: 7\"")
	}

	// int32
	var int32Position Position[int32]
	err = int32Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if int32Position.String() != "Value: 1 Position: 8" {
		t.Fatal("int32Position.String() != \"Value: 1 Position: 8\"")
	}

	// int64
	var int64Position Position[int64]
	err = int64Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if int64Position.String() != "Value: 1 Position: 9" {
		t.Fatal("int64Position.String() != \"Value: 1 Position: 9\"")
	}

	// uintptr
	var uintptrPosition Position[uintptr]
	err = uintptrPosition.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if uintptrPosition.String() != "Value: 1 Position: 10" {
		t.Fatal("uintptrPosition.String() != \"Value: 1 Position: 10\"")
	}

	// uint
	var uintPosition Position[uint]
	err = uintPosition.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if uintPosition.String() != "Value: 1 Position: 11" {
		t.Fatal("uintPosition.String() != \"Value: 1 Position: 11\"")
	}

	// uint8
	var uint8Position Position[uint8]
	err = uint8Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if uint8Position.String() != "Value: 1 Position: 12" {
		t.Fatal("uint8Position.String() != \"Value: 1 Position: 12\"")
	}

	// uint16
	var uint16Position Position[uint16]
	err = uint16Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if uint16Position.String() != "Value: 1 Position: 13" {
		t.Fatal("uint16Position.String() != \"Value: 1 Position: 13\"")
	}

	// uint32
	var uint32Position Position[uint32]
	err = uint32Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if uint32Position.String() != "Value: 1 Position: 14" {
		t.Fatal("uint32Position.String() != \"Value: 1 Position: 14\"")
	}

	// uint64
	var uint64Position Position[uint64]
	err = uint64Position.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	if uint64Position.String() != "Value: 1 Position: 15" {
		t.Fatal("uint64Position.String() != \"Value: 1 Position: 15\"")
	}
}

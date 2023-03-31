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

func TestPositionSlice(t *testing.T) {
	// 重置 position，不能使用 t.Parallel()
	atomic.StoreUint32(&position, 0)

	// time.Duration 这是一个特殊的类型
	var durationSlice PositionSlice[time.Duration]
	err := durationSlice.Set("1s")
	if err != nil {
		t.Fatal(err)
	}
	err = durationSlice.Set("2s")
	if err != nil {
		t.Fatal(err)
	}
	if durationSlice.String() != "[{1s 0} {2s 1}]" {
		t.Fatal("durationSlice.String() != \"[{1s 0} {2s 1}]\"")
	}

	// string
	var stringSlice PositionSlice[string]
	err = stringSlice.Set("hello")
	if err != nil {
		t.Fatal(err)
	}
	err = stringSlice.Set("world")
	if err != nil {
		t.Fatal(err)
	}
	if stringSlice.String() != "[{hello 2} {world 3}]" {
		t.Fatal("stringSlice.String() != \"[{hello 2} {world 3}]\"")
	}

	// bool
	var boolSlice PositionSlice[bool]
	err = boolSlice.Set("true")
	if err != nil {
		t.Fatal(err)
	}
	err = boolSlice.Set("false")
	if err != nil {
		t.Fatal(err)
	}
	if boolSlice.String() != "[{true 4} {false 5}]" {
		t.Fatal("boolSlice.String() != \"[{true 4} {false 5}]\"")
	}

	// float32
	var float32Slice PositionSlice[float32]
	err = float32Slice.Set("1.1")
	if err != nil {
		t.Fatal(err)
	}
	err = float32Slice.Set("2.2")
	if err != nil {
		t.Fatal(err)
	}
	if float32Slice.String() != "[{1.1 6} {2.2 7}]" {
		t.Fatal("float32Slice.String() != \"[{1.1 6} {2.2 7}]\"")
	}

	// float64
	var float64Slice PositionSlice[float64]
	err = float64Slice.Set("1.1")
	if err != nil {
		t.Fatal(err)
	}
	err = float64Slice.Set("2.2")
	if err != nil {
		t.Fatal(err)
	}
	if float64Slice.String() != "[{1.1 8} {2.2 9}]" {
		t.Fatal("float64Slice.String() != \"[{1.1 8} {2.2 9}]\"")
	}

	// int
	var intSlice PositionSlice[int]
	err = intSlice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = intSlice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if intSlice.String() != "[{1 10} {2 11}]" {
		t.Fatal("intSlice.String() != \"[{1 10} {2 11}]\"")
	}

	// int8
	var int8Slice PositionSlice[int8]
	err = int8Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int8Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int8Slice.String() != "[{1 12} {2 13}]" {
		t.Fatal("int8Slice.String() != \"[{1 12} {2 13}]\"")
	}

	// int16
	var int16Slice PositionSlice[int16]
	err = int16Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int16Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int16Slice.String() != "[{1 14} {2 15}]" {
		t.Fatal("int16Slice.String() != \"[{1 14} {2 15}]\"")
	}

	// int32
	var int32Slice PositionSlice[int32]
	err = int32Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int32Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int32Slice.String() != "[{1 16} {2 17}]" {
		t.Fatal("int32Slice.String() != \"[{1 16} {2 17}]\"")
	}

	// int64
	var int64Slice PositionSlice[int64]
	err = int64Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int64Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int64Slice.String() != "[{1 18} {2 19}]" {
		t.Fatal("int64Slice.String() != \"[{1 18} {2 19}]\"")
	}

	// uintptr
	var uintptrSlice PositionSlice[uintptr]
	err = uintptrSlice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uintptrSlice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uintptrSlice.String() != "[{1 20} {2 21}]" {
		t.Fatal("uintptrSlice.String() != \"[{1 20} {2 21}]\"")
	}

	// uint
	var uintSlice PositionSlice[uint]
	err = uintSlice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uintSlice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uintSlice.String() != "[{1 22} {2 23}]" {
		t.Fatal("uintSlice.String() != \"[{1 22} {2 23}]\"")
	}

	// uint8
	var uint8Slice PositionSlice[uint8]
	err = uint8Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint8Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint8Slice.String() != "[{1 24} {2 25}]" {
		t.Fatal("uint8Slice.String() != \"[{1 24} {2 25}]\"")
	}

	// uint16
	var uint16Slice PositionSlice[uint16]
	err = uint16Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint16Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint16Slice.String() != "[{1 26} {2 27}]" {
		t.Fatal("uint16Slice.String() != \"[{1 26} {2 27}]\"")
	}

	// uint32
	var uint32Slice PositionSlice[uint32]
	err = uint32Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint32Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint32Slice.String() != "[{1 28} {2 29}]" {
		t.Fatal("uint32Slice.String() != \"[{1 28} {2 29}]\"")
	}

	// uint64
	var uint64Slice PositionSlice[uint64]
	err = uint64Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint64Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint64Slice.String() != "[{1 30} {2 31}]" {
		t.Fatal("uint64Slice.String() != \"[{1 30} {2 31}]\"")
	}
}

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
	"testing"
	"time"
)

func TestSlice(t *testing.T) {
	t.Parallel()

	// time.Duration 这是一个特殊的类型
	var durationSlice Slice[time.Duration]
	err := durationSlice.Set("1s")
	if err != nil {
		t.Fatal(err)
	}
	err = durationSlice.Set("2s")
	if err != nil {
		t.Fatal(err)
	}
	if durationSlice.String() != "[1s 2s]" {
		t.Fatal("durationSlice.String() != \"[1s 2s]\"")
	}

	// string
	var stringSlice Slice[string]
	err = stringSlice.Set("hello")
	if err != nil {
		t.Fatal(err)
	}
	err = stringSlice.Set("world")
	if err != nil {
		t.Fatal(err)
	}
	if stringSlice.String() != "[hello world]" {
		t.Fatal("stringSlice.String() != \"[hello world]\"")
	}

	// bool
	var boolSlice Slice[bool]
	err = boolSlice.Set("true")
	if err != nil {
		t.Fatal(err)
	}
	err = boolSlice.Set("false")
	if err != nil {
		t.Fatal(err)
	}
	if boolSlice.String() != "[true false]" {
		t.Fatal("boolSlice.String() != \"[true false]\"")
	}

	// float32
	var float32Slice Slice[float32]
	err = float32Slice.Set("1.1")
	if err != nil {
		t.Fatal(err)
	}
	err = float32Slice.Set("2.2")
	if err != nil {
		t.Fatal(err)
	}
	if float32Slice.String() != "[1.1 2.2]" {
		t.Fatal("float32Slice.String() != \"[1.1 2.2]\"")
	}

	// float64
	var float64Slice Slice[float64]
	err = float64Slice.Set("1.1")
	if err != nil {
		t.Fatal(err)
	}
	err = float64Slice.Set("2.2")
	if err != nil {
		t.Fatal(err)
	}
	if float64Slice.String() != "[1.1 2.2]" {
		t.Fatal("float64Slice.String() != \"[1.1 2.2]\"")
	}

	// int
	var intSlice Slice[int]
	err = intSlice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = intSlice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if intSlice.String() != "[1 2]" {
		t.Fatal("intSlice.String() != \"[1 2]\"")
	}

	// int8
	var int8Slice Slice[int8]
	err = int8Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int8Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int8Slice.String() != "[1 2]" {
		t.Fatal("int8Slice.String() != \"[1 2]\"")
	}

	// int16
	var int16Slice Slice[int16]
	err = int16Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int16Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int16Slice.String() != "[1 2]" {
		t.Fatal("int16Slice.String() != \"[1 2]\"")
	}

	// int32
	var int32Slice Slice[int32]
	err = int32Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int32Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int32Slice.String() != "[1 2]" {
		t.Fatal("int32Slice.String() != \"[1 2]\"")
	}

	// int64
	var int64Slice Slice[int64]
	err = int64Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = int64Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if int64Slice.String() != "[1 2]" {
		t.Fatal("int64Slice.String() != \"[1 2]\"")
	}

	// uintptr
	var uintptrSlice Slice[uintptr]
	err = uintptrSlice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uintptrSlice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uintptrSlice.String() != "[1 2]" {
		t.Fatal("uintptrSlice.String() != \"[1 2]\"")
	}

	// uint
	var uintSlice Slice[uint]
	err = uintSlice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uintSlice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uintSlice.String() != "[1 2]" {
		t.Fatal("uintSlice.String() != \"[1 2]\"")
	}

	// uint8
	var uint8Slice Slice[uint8]
	err = uint8Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint8Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint8Slice.String() != "[1 2]" {
		t.Fatal("uint8Slice.String() != \"[1 2]\"")
	}

	// uint16
	var uint16Slice Slice[uint16]
	err = uint16Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint16Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint16Slice.String() != "[1 2]" {
		t.Fatal("uint16Slice.String() != \"[1 2]\"")
	}

	// uint32
	var uint32Slice Slice[uint32]
	err = uint32Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint32Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint32Slice.String() != "[1 2]" {
		t.Fatal("uint32Slice.String() != \"[1 2]\"")
	}

	// uint64
	var uint64Slice Slice[uint64]
	err = uint64Slice.Set("1")
	if err != nil {
		t.Fatal(err)
	}
	err = uint64Slice.Set("2")
	if err != nil {
		t.Fatal(err)
	}
	if uint64Slice.String() != "[1 2]" {
		t.Fatal("uint64Slice.String() != \"[1 2]\"")
	}
}

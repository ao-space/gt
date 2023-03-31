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

package util_test

import (
	"testing"
	"time"

	"github.com/isrc-cas/gt/util"
)

func TestBlockValue(t *testing.T) {
	bv := util.NewBlockValue[int]()

	startTime := time.Now()
	go func() {
		time.Sleep(1 * time.Second)
		bv.Set(&[]int{1}[0])
	}()

	// 第一次 Get 需要等待至少 1 秒
	v := bv.Get()
	if time.Since(startTime) < 1*time.Second {
		t.Fatal("BlockValue not blocking")
	}
	if *v != 1 {
		t.Fatal("v != 1")
	}

	// 第二次 Get 的时间不应该超过 100 毫秒
	startTime = time.Now()
	v = bv.Get()
	if time.Since(startTime) > 100*time.Millisecond {
		t.Fatal("BlockValue blocking")
	}
	if *v != 1 {
		t.Fatal("v != 1")
	}
}

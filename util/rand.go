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
	"math/rand"
	"sync"
	"time"
)

var (
	r    *rand.Rand
	lock sync.Mutex
)

func init() {
	r = rand.New(rand.NewSource(time.Now().Unix()))
}

// RandomString 随机字符串
func RandomString(n int) string {
	letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]byte, n)
	lock.Lock()
	defer lock.Unlock()
	for i := range s {
		s[i] = letters[r.Intn(len(letters))]
	}
	return string(s)
}

func RandomPort() int {
	lock.Lock()
	defer lock.Unlock()
	return r.Intn(65535-1024) + 1024
}

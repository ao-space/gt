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
	"sync/atomic"
)

// BlockValue 当 Set 方法没有调用时 Get 方法会一直阻塞
type BlockValue[T any] struct {
	v           atomic.Pointer[T]
	waitSet     chan struct{}
	initialized atomic.Bool
}

func NewBlockValue[T any]() BlockValue[T] {
	return BlockValue[T]{
		waitSet: make(chan struct{}, 1),
	}
}

func (bv *BlockValue[T]) Get() *T {
	if !bv.initialized.Load() {
		<-bv.waitSet
	}
	return bv.v.Load()
}

func (bv *BlockValue[T]) Set(v *T) {
	bv.v.Store(v)
	if bv.initialized.Load() {
		return
	}
	bv.initialized.Store(true)
	close(bv.waitSet)
}

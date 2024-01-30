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

package webrtc

/*
#include "threadpool.h"
*/
import "C"
import "unsafe"

type ThreadPool struct {
	p unsafe.Pointer
}

func NewThreadPool(threadNum uint32) *ThreadPool {
	return &ThreadPool{p: C.NewThreadPool(C.uint32_t(threadNum))}
}

func (tp *ThreadPool) GetThread() unsafe.Pointer {
	return C.GetThreadPoolThread(tp.p)
}

func (tp *ThreadPool) GetSocketThread() unsafe.Pointer {
	return C.GetThreadPoolSocketThread(tp.p)
}

func (tp *ThreadPool) Close() {
	C.DeleteThreadPool(tp.p)
}

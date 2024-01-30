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

package pool

import (
	"io"
	"sync"

	"github.com/isrc-cas/gt/bufio"
)

const (
	// MaxBufferSize max tunnel message size
	MaxBufferSize = 4 * 1024
)

// BytesPool is a pool of []byte that cap and len are MaxBufferSize
var BytesPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, MaxBufferSize)
	},
}

var readersPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReaderWithBuf(BytesPool.Get().([]byte))
	},
}

// GetReader returns a *bufio.Reader in the pool
func GetReader(reader io.Reader) *bufio.Reader {
	r := readersPool.Get().(*bufio.Reader)
	r.Reset(reader)
	return r
}

// PutReader puts the *bufio.Reader in the pool
func PutReader(r *bufio.Reader) {
	readersPool.Put(r)
}

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

package rotatelogs

import (
	"fmt"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
)

func TestGenFilename(t *testing.T) {
	// Mock time
	ts := []time.Time{
		time.Time{},
		(time.Time{}).Add(24 * time.Hour),
	}

	for _, xt := range ts {
		rl, err := New(
			"/path/to/%Y/%m/%d",
			WithClock(clockwork.NewFakeClockAt(xt)),
		)
		if !assert.NoError(t, err, "New should succeed") {
			return
		}

		defer rl.Close()

		fn := rl.genFilename()
		expected := fmt.Sprintf("/path/to/%04d/%02d/%02d",
			xt.Year(),
			xt.Month(),
			xt.Day(),
		)

		if !assert.Equal(t, expected, fn) {
			return
		}
	}
}

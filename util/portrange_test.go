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
	"reflect"
	"testing"
)

func TestPortRange(t *testing.T) {
	tests := []struct {
		args           string
		expectedResult *PortRange
	}{
		{
			args: "22-80",
			expectedResult: &PortRange{
				Min: 22,
				Max: 80,
			},
		},
		{
			args: "80",
			expectedResult: &PortRange{
				Min: 80,
				Max: 80,
			},
		},
		{
			args: "0",
			expectedResult: &PortRange{
				Min: 1,
				Max: 65535,
			},
		},
	}
	for _, test := range tests {
		result, err := NewPortRangeFromString(test.args)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result, test.expectedResult) {
			t.Fatalf("unexpected result\n%#v\n%#v\n", result, test.expectedResult)
		}
	}
}

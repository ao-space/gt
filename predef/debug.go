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

//go:build !release
// +build !release

package predef

import (
	"log"
	"net/http"
	// used for prof
	_ "net/http/pprof"
	"os"
	"strings"
)

// Debug enables the logs of read and write operations
var Debug = true

func init() {
	env, ok := os.LookupEnv("DEBUG_REQ")
	if ok {
		if strings.ToLower(env) == "true" {
			Debug = true
		} else {
			Debug = false
		}
	}
	prof, ok := os.LookupEnv("DEBUG_PROF")
	if ok {
		go func() {
			log.Println(http.ListenAndServe(prof, nil))
		}()
	}
}

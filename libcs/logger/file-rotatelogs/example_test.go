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

package rotatelogs_test

import (
	"fmt"
	rotatelogs "github.com/isrc-cas/gt/logger/file-rotatelogs"
	"io/ioutil"
	"os"
)

func ExampleForceNewFile() {
	logDir, err := ioutil.TempDir("", "rotatelogs_test")
	if err != nil {
		fmt.Println("could not create log directory ", err)
		return
	}
	logPath := fmt.Sprintf("%s/test.log", logDir)

	for i := 0; i < 2; i++ {
		writer, err := rotatelogs.New(logPath,
			rotatelogs.ForceNewFile(),
		)
		if err != nil {
			fmt.Println("Could not open log file ", err)
			return
		}

		n, err := writer.Write([]byte("test"))
		if err != nil || n != 4 {
			fmt.Println("Write failed ", err, " number written ", n)
			return
		}
		err = writer.Close()
		if err != nil {
			fmt.Println("Close failed ", err)
			return
		}
	}

	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		fmt.Println("ReadDir failed ", err)
		return
	}
	for _, file := range files {
		fmt.Println(file.Name(), file.Size())
	}

	err = os.RemoveAll(logDir)
	if err != nil {
		fmt.Println("RemoveAll failed ", err)
		return
	}
	// OUTPUT:
	// test.log 4
	// test.log.1 4
}

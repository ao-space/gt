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
#include <stdlib.h>
#include "logging.h"
*/
import "C"
import "sync"

type LoggingSeverity int

const (
	LoggingSeverityVerbose LoggingSeverity = iota
	LoggingSeverityInfo
	LoggingSeverityWarning
	LoggingSeverityError
	LoggingSeverityNone
)

func (l *LoggingSeverity) String() string {
	switch *l {
	case LoggingSeverityVerbose:
		return "verbose"
	case LoggingSeverityInfo:
		return "info"
	case LoggingSeverityWarning:
		return "warning"
	case LoggingSeverityError:
		return "error"
	case LoggingSeverityNone:
		return "none"
	}
	panic("unreachable")
}

var (
	onLogMessageGlobalRWMutex sync.RWMutex
	onLogMessageGlobal        func(severity LoggingSeverity, message string, tag string)
)

//export onLogMessage
func onLogMessage(severity C.int, messageC *C.char, tagC *C.char) {
	onLogMessageGlobalRWMutex.RLock()
	defer onLogMessageGlobalRWMutex.RUnlock()
	if onLogMessageGlobal == nil {
		return
	}

	message := C.GoString(messageC)
	tag := C.GoString(tagC)
	onLogMessageGlobal(LoggingSeverity(severity), message, tag)
}

// SetLog set logging severity and onLogMessage callback
func SetLog(
	severity LoggingSeverity,
	f func(severity LoggingSeverity, message string, tag string),
) {
	onLogMessageGlobalRWMutex.Lock()
	onLogMessageGlobal = f
	onLogMessageGlobalRWMutex.Unlock()
	C.SetLog(C.int(severity))
}

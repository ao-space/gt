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
	"os"
	"sync"
	"time"

	strftime "github.com/lestrrat-go/strftime"
)

type Handler interface {
	Handle(Event)
}

type HandlerFunc func(Event)

type Event interface {
	Type() EventType
}

type EventType int

const (
	InvalidEventType EventType = iota
	FileRotatedEventType
)

type FileRotatedEvent struct {
	prev    string // previous filename
	current string // current, new filename
}

// RotateLogs represents a log file that gets
// automatically rotated as you write to it.
type RotateLogs struct {
	clock         Clock
	curFn         string
	curBaseFn     string
	globPattern   string
	generation    int
	linkName      string
	maxAge        time.Duration
	mutex         sync.RWMutex
	eventHandler  Handler
	outFh         *os.File
	pattern       *strftime.Strftime
	rotationTime  time.Duration
	rotationSize  int64
	rotationCount uint
	forceNewFile  bool
}

// Clock is the interface used by the RotateLogs
// object to determine the current time
type Clock interface {
	Now() time.Time
}
type clockFn func() time.Time

// UTC is an object satisfying the Clock interface, which
// returns the current time in UTC
var UTC = clockFn(func() time.Time { return time.Now().UTC() })

// Local is an object satisfying the Clock interface, which
// returns the current time in the local timezone
var Local = clockFn(time.Now)

// Option is used to pass optional arguments to
// the RotateLogs constructor
type Option interface {
	Name() string
	Value() interface{}
}

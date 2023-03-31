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

package client

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const (
	running status = iota
	idle
	wait
)

type status int

type idleManager struct {
	status     map[uint]status
	statusMtx  sync.RWMutex
	statusCond *sync.Cond
	min        uint
	close      atomic.Bool
}

func (m *idleManager) String() string {
	m.statusMtx.RLock()
	defer m.statusMtx.RUnlock()
	return fmt.Sprintf("%#v", m.status)
}

func newIdleManager(min uint) *idleManager {
	m := &idleManager{
		status: make(map[uint]status),
		min:    min,
	}
	m.statusCond = sync.NewCond(&m.statusMtx)
	return m
}

func (m *idleManager) InitIdle(id uint) (exit bool) {
	m.statusMtx.Lock()
	defer m.statusMtx.Unlock()

	if v, ok := m.status[id]; ok && v == idle {
		return false
	}
	m.status[id] = idle
	var n uint
	for _, s := range m.status {
		switch s {
		case running:
			n++
		case idle:
			n++
		}
	}
	if n <= m.min {
		return false
	}

	m.status[id] = wait
	return true
}

func (m *idleManager) ChangeToWait(id uint) (ok bool) {
	m.statusMtx.Lock()
	defer m.statusMtx.Unlock()

	var n uint
	for _, s := range m.status {
		switch s {
		case idle:
			n++
		}
	}
	if n <= m.min {
		return false
	}

	m.status[id] = wait
	return true
}

func (m *idleManager) IdleCount() (n int) {
	m.statusMtx.RLock()
	defer m.statusMtx.RUnlock()
	n = 0
	for _, s := range m.status {
		switch s {
		case idle:
			n++
		}
	}
	return
}

func (m *idleManager) WaitCount() (n int) {
	m.statusMtx.RLock()
	defer m.statusMtx.RUnlock()
	n = 0
	for _, s := range m.status {
		switch s {
		case wait:
			n++
		}
	}
	return
}

func (m *idleManager) RunningCount() (n int) {
	m.statusMtx.RLock()
	defer m.statusMtx.RUnlock()
	n = 0
	for _, s := range m.status {
		switch s {
		case running:
			n++
		}
	}
	return
}

func (m *idleManager) SetRunningWithTaskCount(id uint, taskCount uint32) {
	m.statusMtx.RLock()
	s := m.status[id]
	m.statusMtx.RUnlock()
	if s == running {
		if taskCount >= 3 {
			m.statusCond.Signal()
		}
		return
	}

	m.statusMtx.Lock()
	if m.status[id] != running {
		m.status[id] = running
	}
	m.statusMtx.Unlock()
}

func (m *idleManager) SetRunning(id uint) {
	m.statusMtx.RLock()
	s := m.status[id]
	m.statusMtx.RUnlock()
	if s == running {
		return
	}

	m.statusMtx.Lock()
	if m.status[id] != running {
		m.status[id] = running
		defer m.statusCond.Signal()
	}
	m.statusMtx.Unlock()
}

func (m *idleManager) SetIdle(id uint) {
	m.statusMtx.Lock()
	m.status[id] = idle
	m.statusMtx.Unlock()
}

func (m *idleManager) SetWait(id uint) {
	m.statusMtx.Lock()
	m.status[id] = wait
	m.statusMtx.Unlock()
}

func (m *idleManager) WaitIdle(id uint) {
	m.statusMtx.Lock()
	defer m.statusMtx.Unlock()

	for !m.close.Load() {
		wait := false
		for _, s := range m.status {
			if s == idle {
				wait = true
				break
			}
		}
		if wait {
			m.statusCond.Wait()
			continue
		}
		m.status[id] = idle
		return
	}
}

func (m *idleManager) Close() {
	m.close.Store(true)
	m.statusCond.Broadcast()
}

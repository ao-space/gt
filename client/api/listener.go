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

package api

import (
	"errors"
	"fmt"
	"net"
)

type VirtualListener struct {
	acceptCh chan net.Conn
}

func (v *VirtualListener) AcceptCh() chan<- net.Conn {
	return v.acceptCh
}

func NewVirtualListener() *VirtualListener {
	return &VirtualListener{acceptCh: make(chan net.Conn, 1)}
}

func (v *VirtualListener) Accept() (c net.Conn, err error) {
	c, ok := <-v.acceptCh
	if !ok {
		err = errors.New("listener closed")
	}
	return
}

func (v *VirtualListener) Close() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	close(v.acceptCh)
	return err
}

func (v *VirtualListener) Addr() net.Addr {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.2:1")
	return addr
}

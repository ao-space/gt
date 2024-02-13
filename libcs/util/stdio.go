/*
 * Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
)

type OP struct {
	OP OPValue `json:"op,omitempty"`
}

type OPValue string

const (
	Ready                OPValue = "ready"
	GracefulShutdown     OPValue = "gracefulShutdown"
	GracefulShutdownDone OPValue = "gracefulShutdownDone"
	Shutdown             OPValue = "shutdown"
	ShutdownDone         OPValue = "shutdownDone"
	Reconnect            OPValue = "reconnect"
)

var writeMtx sync.Mutex

func WriteJson(json []byte) (err error) {
	writeMtx.Lock()
	defer writeMtx.Unlock()

	l := [4]byte{}
	binary.BigEndian.PutUint32(l[:], uint32(len(json)))
	_, err = os.Stdout.Write(l[:])
	if err != nil {
		return
	}
	_, err = os.Stdout.Write(json)
	return
}

func ReadJson() (json []byte, err error) {
	l := [4]byte{}
	_, err = os.Stdin.Read(l[:])
	if err != nil {
		return
	}
	jl := binary.BigEndian.Uint32(l[:])
	if jl > 8*1024 {
		err = errors.New("json too large")
		return
	}
	json = make([]byte, jl)
	_, err = io.ReadFull(os.Stdin, json)
	return
}

func WriteOP(op OP) (err error) {
	bs, err := json.Marshal(op)
	if err != nil {
		return
	}
	err = WriteJson(bs)
	return
}

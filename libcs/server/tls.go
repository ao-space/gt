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

package server

import (
	"errors"

	"github.com/isrc-cas/gt/bufio"
)

func peekTLSHost(reader *bufio.Reader) ([]byte, error) {
	_, err := reader.Peek(1) // 保证 Client Hello 已经被缓冲，否则 bufLen 可能为 0
	if err != nil {
		return nil, err
	}
	bufLen := reader.Buffered()
	buf, err := reader.Peek(bufLen)
	if err != nil {
		return nil, err
	}
	bufIndex := 0

	// 判断 Record Layer 是否是 Handshake
	if bufIndex+1 > bufLen {
		return nil, errors.New("failed to read Record Layer Type")
	}
	recordLayerType := buf[bufIndex]
	bufIndex++
	if recordLayerType != 22 {
		return nil, errors.New("the Record Layer type is not Handshake")
	}
	bufIndex += 2 + 2 // Record Layer Version, recordLayerLen

	// 判断 Handshake Type 是否是 Client Hello
	if bufIndex+1 > bufLen {
		return nil, errors.New("failed to read Handshake Type")
	}
	handshakeType := buf[bufIndex]
	bufIndex++
	if handshakeType != 1 {
		return nil, errors.New("the Handshake Type is not Client Hello")
	}
	bufIndex += 3 + 2 + 32 // Handshake Length, Handshake Version, Handshake Random
	if bufIndex+1 > bufLen {
		return nil, errors.New("failed to read Session ID Length")
	}
	sessionIDLen := buf[bufIndex]
	bufIndex += 1 + int(sessionIDLen)
	if bufIndex+2 > bufLen {
		return nil, errors.New("failed to read Cipher Suites Length")
	}
	cipherSuitesLen := uint16(buf[bufIndex+1]) | uint16(buf[bufIndex])<<8
	bufIndex += 2 + int(cipherSuitesLen)
	if bufIndex+1 > bufLen {
		return nil, errors.New("failed to read Compression Methods Length")
	}
	compressionMethodsLen := buf[bufIndex]
	bufIndex += 1 + int(compressionMethodsLen)
	if bufIndex+2 > bufLen {
		return nil, errors.New("failed to read Extensions Length")
	}
	extensionsLen := int(buf[bufIndex+1]) | int(buf[bufIndex])<<8
	bufIndex += 2

	// 遍历 Extensions
	for extensionsLen > 0 {
		if bufIndex+2 > bufLen {
			return nil, errors.New("failed to read Extension Type")
		}
		extensionType := uint16(buf[bufIndex+1]) | uint16(buf[bufIndex])<<8
		bufIndex += 2
		extensionsLen -= 2
		if bufIndex+2 > bufLen {
			return nil, errors.New("failed to read Extension Length")
		}
		extensionLen := uint16(buf[bufIndex+1]) | uint16(buf[bufIndex])<<8
		bufIndex += 2
		extensionsLen -= 2
		// 判断 Extension Type 是否是 Server Name Indication
		if extensionType != 0 {
			bufIndex += int(extensionLen)
			extensionsLen -= int(extensionLen)
			continue
		}
		bufIndex += 2 // Sever Name List Length
		// extensionsLen -= 2
		if bufIndex+1 > bufLen {
			return nil, errors.New("failed to read Server Name Type")
		}
		serverNameType := buf[bufIndex]
		bufIndex++
		// extensionsLen--
		// 判断 Server Name Type 是否是 host_name
		if serverNameType != 0 {
			return nil, errors.New("the Server Name Type is not host_name")
		}
		if bufIndex+2 > bufLen {
			return nil, errors.New("failed to read Server Name Length")
		}
		serverNameLen := uint16(buf[bufIndex+1]) | uint16(buf[bufIndex])<<8
		bufIndex += 2
		// extensionsLen -= 2
		if bufIndex+int(serverNameLen) > bufLen {
			return nil, errors.New("failed to read Server Name")
		}
		serverName := buf[bufIndex : bufIndex+int(serverNameLen)]
		return serverName, nil
	}
	return nil, errors.New("failed to read Server Name Indication")
}

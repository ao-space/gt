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

package client

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/isrc-cas/gt/client/api"
	"github.com/isrc-cas/gt/client/webrtc"
	"github.com/isrc-cas/gt/logger"
)

// Client is a network agent client.
type Client struct {
	config              atomic.Pointer[Config]
	Logger              logger.Logger
	initConnMtx         sync.Mutex
	closing             uint32
	tunnels             map[*conn]struct{}
	tunnelsRWMtx        sync.RWMutex
	peers               map[uint32]*peerTask
	peersRWMtx          sync.RWMutex
	tunnelsCond         *sync.Cond
	idleManager         *idleManager
	apiServer           *api.Server
	services            atomic.Pointer[services]
	tcpForwardListener  net.Listener
	webrtcThreadPool    *webrtc.ThreadPool
	waitTunnelsShutdown sync.WaitGroup
	configChecksum      atomic.Pointer[[32]byte]
	reloadWaitGroup     sync.WaitGroup
	reloading           atomic.Bool

	// test purpose only
	OnTunnelClose atomic.Value

	// indicate which remote is chosen to establish tunnel
	chosenRemoteLabel int
}

func (c *conn) onTunnelClose() {
	cb := c.client.OnTunnelClose.Load()
	if cb != nil {
		if cb, ok := cb.(func()); ok {
			cb()
		}
	}
}

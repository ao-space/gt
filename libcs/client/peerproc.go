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
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/isrc-cas/gt/bufio"
	"github.com/isrc-cas/gt/client/api"
	"github.com/isrc-cas/gt/client/std"
	"github.com/isrc-cas/gt/pool"
	"github.com/isrc-cas/gt/predef"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type peerProcessTask struct {
	Logger   zerolog.Logger
	id       uint32
	closing  uint32
	initDone uint32
	state    uint8
	dataLen  [2]byte
	n        int
	data     []byte
	tunnel   *conn
	timer    *time.Timer
	apiConn  *api.Conn
	cmd      *exec.Cmd
	stdout   io.ReadCloser
	stdin    io.WriteCloser
	stderr   io.ReadCloser
}

type op struct {
	Config      *opConfig      `json:"config,omitempty"`
	OfferSDP    string         `json:"offerSDP,omitempty"`
	AnswerSDP   string         `json:"answerSDP,omitempty"`
	Candidate   string         `json:"candidate,omitempty"`
	GetOfferSDP *opGetOfferSDP `json:"getOfferSDP,omitempty"`
}

type opConfig struct {
	Stuns      []string          `json:"stuns,omitempty"`
	HTTPRoutes map[string]string `json:"httpRoutes,omitempty"`
	TCPRoutes  map[string]string `json:"tcpRoutes,omitempty"`
	PortMin    uint16            `json:"portMin,omitempty"`
	PortMax    uint16            `json:"portMax,omitempty"`
	Timeout    uint16            `json:"timeout,omitempty"`
}

type opGetOfferSDP struct {
	ChannelName string `json:"channelName"`
}

func (pt *peerProcessTask) init() (err error) {
	cmd := exec.Command(os.Args[0], "sub-p2p")
	pt.stdin, err = cmd.StdinPipe()
	if err != nil {
		return
	}
	pt.stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	pt.stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		var err error
		defer func() {
			pt.Logger.Info().Err(err).AnErr("wait", cmd.Wait()).Msg("peer process wait done")
			pt.closeWithLock()
		}()
		reader := bufio.NewReader(pt.stderr)
		for {
			var line string
			line, err = reader.ReadString('\n')
			if err != nil {
				return
			}
			pt.Logger.Info().Msg(line)
			if strings.HasPrefix(line, "p2p done") {
				break
			}
		}
	}()
	if pt.timer != nil {
		pt.timer.Stop()
	}
	config := pt.tunnel.client.Config()
	httpRouters := make(map[string]string)
	tcpRouters := make(map[string]string)
	for i, s := range config.Services {
		if len(s.HostPrefix) > 0 {
			if i == 0 {
				httpRouters["@"] = s.LocalURL.String()
			}
			httpRouters[s.HostPrefix] = s.LocalURL.String()
		} else {
			tcpRouters[strconv.FormatUint(uint64(atomic.LoadUint32(&s.remoteTCPPort)), 10)] = s.LocalURL.String()
		}
	}
	js, err := json.Marshal(&op{
		Config: &opConfig{
			Stuns:      pt.tunnel.stuns,
			HTTPRoutes: httpRouters,
			TCPRoutes:  tcpRouters,
			PortMin:    config.WebRTCMinPort,
			PortMax:    config.WebRTCMaxPort,
			Timeout:    uint16(config.WebRTCConnectionIdleTimeout.Duration.Seconds()),
		},
	})
	if err != nil {
		return
	}
	err = pt.writeJson(js)
	pt.cmd = cmd
	return
}

func (pt *peerProcessTask) writeJson(json []byte) (err error) {
	l := [4]byte{}
	binary.BigEndian.PutUint32(l[:], uint32(len(json)))
	_, err = pt.stdin.Write(l[:])
	if err != nil {
		return
	}
	_, err = pt.stdin.Write(json)
	return
}

func (pt *peerProcessTask) readJson() (json []byte, err error) {
	l := [4]byte{}
	_, err = pt.stdout.Read(l[:])
	if err != nil {
		return
	}
	jl := binary.BigEndian.Uint32(l[:])
	if jl > 8*1024 {
		err = errors.New("json too large")
		return
	}
	json = make([]byte, jl)
	_, err = io.ReadFull(pt.stdout, json)
	return
}

func (pt *peerProcessTask) Close() {
	if !atomic.CompareAndSwapUint32(&pt.closing, 0, 1) {
		return
	}
	defer pool.BytesPool.Put(pt.data)
	client := pt.tunnel.client
	delete(client.peers, pt.id)
	if pt.timer != nil {
		pt.timer.Stop()
	}
	var cmdErr error
	if pt.cmd != nil {
		cmdErr = pt.cmd.Process.Kill()
	}
	pt.Logger.Info().AnErr("cmd", cmdErr).Msg("peer task closed")
}

func (pt *peerProcessTask) CloseWithLock() {
	if !atomic.CompareAndSwapUint32(&pt.closing, 0, 1) {
		return
	}
	defer pool.BytesPool.Put(pt.data)
	client := pt.tunnel.client
	client.peersRWMtx.Lock()
	delete(client.peers, pt.id)
	client.peersRWMtx.Unlock()
	if pt.timer != nil {
		pt.timer.Stop()
	}
	var cmdErr error
	if pt.cmd != nil {
		cmdErr = pt.cmd.Process.Kill()
	}
	pt.Logger.Info().AnErr("cmd", cmdErr).Msg("peer task closed")
}

func (pt *peerProcessTask) closeWithLock() {
	if !atomic.CompareAndSwapUint32(&pt.closing, 0, 1) {
		return
	}
	defer pool.BytesPool.Put(pt.data)
	client := pt.tunnel.client
	client.peersRWMtx.Lock()
	delete(client.peers, pt.id)
	client.peersRWMtx.Unlock()
	if pt.timer != nil {
		pt.timer.Stop()
	}
	pt.Logger.Info().Msg("peer task closed")
}

func (pt *peerProcessTask) process(r io.Reader, writer http.ResponseWriter, initFn func() error) {
	var err error
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		if !predef.Debug {
			if e := recover(); e != nil {
				pt.Logger.Info().Interface("panic", e).Msg("process panic")
			}
		}
		pt.Logger.Info().Err(err).Msg("process done")
	}()
	reader := std.NewChunkedReader(r)
	for {
		switch pt.state {
		case dataLength:
			var n int
			n, err = reader.Read(pt.dataLen[pt.n:2])
			if n > 0 {
				pt.n += n
			}
			if pt.n != 2 {
				if err != nil {
					if err == io.EOF || err == io.ErrUnexpectedEOF {
						err = nil
						return
					}
					pt.Logger.Error().Err(err).Hex("data", pt.dataLen[:pt.n]).Int("read n", n).Msg("failed to read data")
					return
				}
				continue
			}
			en := uint(pt.dataLen[0])<<8 | uint(pt.dataLen[1])
			if en > pool.MaxBufferSize {
				err = errors.New("dataLen is too long")
				return
			}
			pt.n = 0
			pt.state = dataBody
			pt.Logger.Debug().Int("read n", n).Uint("len", en).Msg("read data length")
			fallthrough
		case dataBody:
			var n int
			en := uint16(pt.dataLen[0])<<8 | uint16(pt.dataLen[1])
			n, err = reader.Read(pt.data[pt.n:en])
			if n > 0 {
				pt.n += n
			}
			if pt.n != int(en) {
				if err != nil {
					pt.Logger.Error().Err(err).Hex("data", pt.data[:pt.n]).Int("read n", n).Uint16("len", en).Msg("failed to read data")
					return
				}
				continue
			}
			pt.state = processData
			fallthrough
		case processData:
			pt.Logger.Debug().Str("data", string(pt.data[:pt.n])).Msg("read json")
			if atomic.CompareAndSwapUint32(&pt.initDone, 0, 1) {
				err = initFn()
				if err != nil {
					pt.Logger.Error().Err(err).Msg("initFn failed")
					return
				}
			} else {
				js, err := json.Marshal(&op{
					Candidate: string(pt.data[:pt.n]),
				})
				if err != nil {
					pt.Logger.Error().Err(err).Msg("failed to Marshal candidate")
					return
				}
				err = pt.writeJson(js)
				if err != nil {
					pt.Logger.Error().Err(err).Msg("failed to add ice candidate")
					return
				}
				pt.Logger.Debug().Msg("AddICECandidate")
			}

			pt.n = 0
			pt.state = dataLength
			if err != nil {
				if err == io.EOF {
					err = nil
					return
				}
				pt.Logger.Error().Err(err).Hex("data", pt.data[:pt.n]).Msg("err after processData")
				return
			}
			continue
		}
	}
}

func (pt *peerProcessTask) response(writer http.ResponseWriter, sdp []byte) {
	var err error
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		pt.Logger.Info().Err(err).Msg("response done")
	}()

	pt.Logger.Debug().Bytes("sdp", sdp).Msg("local sdp")
	writer.Header().Add("Transfer-Encoding", "chunked")
	n := uint16(len(sdp))
	l := copy(pt.data, []byte{byte(n >> 8), byte(n)})
	l += copy(pt.data[l:], sdp)
	_, err = writer.Write(pt.data[:l])
	if err != nil {
		return
	}

	for {
		candidate, err := pt.readJson()
		if err != nil {
			return
		}
		pt.Logger.Debug().Str("candidate", string(candidate)).Msg("local candidate")
		op := op{}
		err = json.Unmarshal(candidate, &op)
		if err != nil {
			return
		}
		if len(op.Candidate) == 0 {
			break
		}
		n := uint16(len(op.Candidate))
		l := copy(pt.data, []byte{byte(n >> 8), byte(n)})
		l += copy(pt.data[l:], op.Candidate)
		_, err = writer.Write(pt.data[:l])
		if err != nil {
			return
		}
	}
}

func (pt *peerProcessTask) processOffer(r *http.Request, writer http.ResponseWriter) {
	pt.process(r.Body, writer, func() (err error) {
		err = pt.init()
		if err != nil {
			return
		}

		js, err := json.Marshal(&op{
			OfferSDP: string(pt.data[:pt.n]),
		})
		if err != nil {
			return
		}
		err = pt.writeJson(js)
		if err != nil {
			return
		}

		js, err = pt.readJson()
		if err != nil {
			return
		}
		op := op{}
		err = json.Unmarshal(js, &op)
		if err != nil {
			return
		}
		pt.response(writer, []byte(op.AnswerSDP))
		return
	})
}

func (pt *peerProcessTask) getOffer(_ *http.Request, writer http.ResponseWriter) {
	var err error
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		pt.Logger.Info().Err(err).Msg("getOffer done")
	}()
	err = pt.init()
	if err != nil {
		return
	}

	js, err := json.Marshal(&op{
		GetOfferSDP: &opGetOfferSDP{ChannelName: "default"},
	})
	if err != nil {
		return
	}
	err = pt.writeJson(js)
	if err != nil {
		return
	}

	js, err = pt.readJson()
	if err != nil {
		return
	}
	op := op{}
	err = json.Unmarshal(js, &op)
	if err != nil {
		return
	}
	pt.response(writer, []byte(op.OfferSDP))

	_, err = writer.Write([]byte(fmt.Sprintf(`{"id":%d}`, pt.id)))
}

func (pt *peerProcessTask) processAnswer(r *http.Request, writer http.ResponseWriter) {
	c := pt.tunnel
	var err error
	var shouldClose bool
	defer func() {
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
		}
		pt.Logger.Info().Err(err).Msg("processAnswer done")
		if shouldClose {
			pt.CloseWithLock()
		}
	}()
	task := pt
	idValue := r.Header.Get("WebRTC-OP-ID")
	id, err := strconv.ParseUint(idValue, 10, 32)
	if err != nil {
		return
	}
	if uint32(id) != pt.id {
		shouldClose = true
		client := c.client
		client.peersRWMtx.RLock()
		pt, ok := client.peers[uint32(id)]
		client.peersRWMtx.RUnlock()
		if !ok {
			err = errors.New("invalid task id")
			return
		}
		task = pt.(*peerProcessTask)
	}
	task.process(r.Body, writer, func() (err error) {
		js, err := json.Marshal(&op{
			AnswerSDP: string(task.data[:task.n]),
		})
		if err != nil {
			return
		}
		err = task.writeJson(js)
		return
	})
}

func (pt *peerProcessTask) APIWriter() *std.PipeWriter {
	return pt.apiConn.PipeWriter
}

func (pt *peerProcessTask) APIConn() *api.Conn {
	return pt.apiConn
}

type PeerTask interface {
	Close()
	CloseWithLock()
	APIWriter() *std.PipeWriter
	APIConn() *api.Conn
}

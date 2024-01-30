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

package test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	validityPeriod = 365 * 24 * time.Hour // 证书有效期
	ecdsaCurve     = elliptic.P256()      // ECDSA 256 位的强度约等于 RSA 3072 位
)

// generateTLSKeyAndCert host 的格式为 "id1.example.com,id2.example.com"
func generateTLSKeyAndCert(hosts, keyPath, certPath string) (err error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(validityPeriod)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}
	x590Cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	hostSlice := strings.Split(hosts, ",")
	for _, host := range hostSlice {
		ip := net.ParseIP(host)
		if ip != nil {
			x590Cert.IPAddresses = append(x590Cert.IPAddresses, ip)
		} else {
			x590Cert.DNSNames = append(x590Cert.DNSNames, host)
		}
	}

	// key
	ecdsaKey, err := ecdsa.GenerateKey(ecdsaCurve, rand.Reader)
	if err != nil {
		return
	}
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return
	}
	keyBytes, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		return
	}
	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if err != nil {
		return
	}
	err = keyFile.Close()
	if err != nil {
		return
	}

	// crt
	certBytes, err := x509.CreateCertificate(rand.Reader, x590Cert, x590Cert, &ecdsaKey.PublicKey, ecdsaKey)
	if err != nil {
		return
	}
	certFile, err := os.Create(certPath)
	if err != nil {
		return
	}
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return
	}
	err = certFile.Close()
	return
}

func TestTLS(t *testing.T) {
	// 生成 TLS 证书
	const (
		keyFile  = "tls.key"
		certFile = "tls.crt"
	)
	err := generateTLSKeyAndCert("*.example.com,localhost", keyFile, certFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove(keyFile)
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove(certFile)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// 启动服务端、客户端
	s, err := setupServer([]string{
		"server",
		"-addr", "127.0.0.1:0",
		"-tlsAddr", "127.0.0.1:0",
		"-keyFile", keyFile,
		"-certFile", certFile,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	c, err := setupClient([]string{
		"client",
		"-id", "05797ac9-86ae-40b0-b767-7a41e03a5486",
		"-local", "http://www.baidu.com/",
		"-remote", fmt.Sprintf("tls://localhost:%v", s.GetTLSListenerAddrPort().Port()), // 这里不能使用 127.0.0.1
		"-remoteCert", certFile,
		"-remoteTimeout", "5s",
		"-useLocalAsHTTPHost",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// 通过 https 测试
	rootCAs := x509.NewCertPool()
	certBytes, err := os.ReadFile(certFile)
	if err != nil {
		t.Fatal(err)
	}
	ok := rootCAs.AppendCertsFromPEM(certBytes)
	if !ok {
		t.Fatal("failed to add cert from pem")
	}
	httpClient := setupHTTPClient(s.GetTLSListenerAddrPort().String(), &tls.Config{
		RootCAs: rootCAs,
	})
	resp, err := httpClient.Get("https://05797ac9-86ae-40b0-b767-7a41e03a5486.example.com")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("invalid status code")
	}
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) > 100 {
		all = append([]byte(nil), all[:100]...)
	}
	t.Logf("%s", all)
}

package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckConfig(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name: "all options set",
			args: []string{
				"client",
				"-web",
				"-webAddr", "localhost",
				"-webPort", "8080",
				"-admin", "admin",
				"-password", "password",
				"-signingKey", "signingKey",
				"-pprof", "true",
			},
			expectErr: false,
		},
		{
			name: "missing webAddr",
			args: []string{
				"client",
				"-web",
				"-webPort", "8080",
				"-admin", "admin",
				"-password", "password",
				"-signingKey", "signingKey",
				"-pprof", "true",
			},
			expectErr: true,
		},
		{
			name: "missing webPort",
			args: []string{
				"client",
				"-web",
				"-webAddr", "localhost",
				"-admin", "admin",
				"-password", "password",
				"-signingKey", "signingKey",
				"-pprof", "true",
			},
			expectErr: true,
		},
		{
			name: "missing admin",
			args: []string{
				"client",
				"-web",
				"-webAddr", "localhost",
				"-webPort", "8080",
				"-password", "password",
				"-signingKey", "signingKey",
				"-pprof", "true",
			},
			expectErr: true,
		},
		{
			name: "missing password",
			args: []string{
				"client",
				"-web",
				"-webAddr", "localhost",
				"-webPort", "8080",
				"-admin", "admin",
				"-signingKey", "signingKey",
				"-pprof", "true",
			},
			expectErr: true,
		},
		{
			name: "missing signingKey",
			args: []string{
				"client",
				"-web",
				"-webAddr", "localhost",
				"-webPort", "8080",
				"-admin", "admin",
				"-password", "password",
				"-pprof", "true",
			},
			expectErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := client.New(tt.args, nil)
			if err != nil {
				t.Errorf("Error while initializing the client: %v", err)
			}
			if err := checkConfig(c); (err != nil) != tt.expectErr {
				t.Errorf("CheckConfig() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestSetRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	clientWithPprof := initializeClient(t, generateClientArgs(true))
	clientWithoutPprof := initializeClient(t, generateClientArgs(false))

	r := gin.New()
	setRoutes(clientWithPprof, r)

	loginPayload := map[string]interface{}{
		"username": "admin",
		"password": "password",
	}
	token, err := performLogin(r, loginPayload)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	tests := []struct {
		name    string
		method  string
		path    string
		client  *client.Client
		status  int
		headers map[string]string
	}{
		{"Check Login Route", "POST", "/api/login", clientWithPprof, http.StatusOK, nil},
		{"Check Server Info Route", "GET", "/api/server/info", clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Running Config Route", "GET", "/api/config/running", clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		{"Check Config From File Route", "GET", "/api/config/file", clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		{"Check Save Config Route", "POST", "/api/config/save", clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Server Info Route", "GET", "/api/server/info", clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		//Can't test these routes because they will kill the test server
		//{"Check Server Restart Route", "PUT", "/api/server/restart",clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		//{"Check Server Shutdown Route", "PUT", "/api/server/shutdown",clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		//{"Check Server Kill Route", "PUT", "/api/server/kill",clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Connections Route", "GET", "/api/connection/list", clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Permissions Route", "GET", "/api/permission/menu", clientWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Pprof Route", "GET", "/debug/pprof/", clientWithPprof, http.StatusOK, nil},
		{"Check Pprof Route", "GET", "/debug/pprof/", clientWithoutPprof, http.StatusNotFound, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			setRoutes(tt.client, r)

			w := performRequestWithHeaders(r, tt.method, tt.path, nil, tt.headers)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}
func generateClientArgs(withPprof bool) []string {
	args := []string{
		"client",
		"-web",
		"-webAddr", "localhost",
		"-webPort", "8080",
		"-admin", "admin",
		"-password", "password",
		"-signingKey", "signingKey",
	}
	if withPprof {
		args = append(args, "-pprof", "true")
	}
	return args
}
func initializeClient(t *testing.T, args []string) *client.Client {
	c, err := client.New(args, nil)
	if err != nil {
		t.Fatalf("Error while initializing the client: %v", err)
	}
	return c
}

func performLogin(r *gin.Engine, loginPayload map[string]interface{}) (string, error) {
	payloadBytes, err := json.Marshal(loginPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal login payload: %v", err)
	}
	payloadReader := bytes.NewReader(payloadBytes)

	w := performRequestWithHeaders(r, "POST", "/api/login", payloadReader, nil)
	if w.Code != http.StatusOK {
		return "", fmt.Errorf("failed to login during test setup with status: %d, response: %s", w.Code, w.Body.String())
	}

	response := map[string]interface{}{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal login response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{}) // 获取"data"字段并断言为map[string]interface{}
	if !ok {
		return "", fmt.Errorf("data not found or is not an object in login response")
	}

	token, exists := data["token"] // 从"data"中获取token
	if !exists {
		return "", fmt.Errorf("token not found in login response")
	}

	return token.(string), nil
}

func performRequestWithHeaders(r *gin.Engine, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/server"
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
				"server",
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
				"server",
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
				"server",
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
				"server",
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
				"server",
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
			name: "webPort less than or equal to 0",
			args: []string{
				"server",
				"-web",
				"-webAddr", "localhost",
				"-webPort", "0",
				"-admin", "admin",
				"-password", "password",
				"-signingKey", "signingKey",
				"-pprof", "true",
			},
			expectErr: true,
		},
		{
			name: "signingKey is empty string but other options set",
			args: []string{
				"server",
				"-web",
				"-webAddr", "localhost",
				"-webPort", "8080",
				"-admin", "admin",
				"-password", "password",
				"-signingKey", "",
				"-pprof", "true",
			},
			expectErr: false, // Because it'll get a random value if not set
		},
		{
			name: "all options not set",
			args: []string{
				"server",
				"-web",
				"-pprof", "true",
			},
			expectErr: true, // Multiple errors can occur. We assume at least one error will be thrown.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := server.New(tt.args, nil)
			if err != nil {
				t.Fatalf("Error while initializing the server: %v", err)
			}

			err = checkConfig(s)
			if (err != nil) != tt.expectErr {
				t.Errorf("checkConfig() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestSetRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	serverWithPprof := initializeServer(t, generateServerArgs(true))
	serverWithoutPprof := initializeServer(t, generateServerArgs(false))

	r := gin.New()
	setRoutes(serverWithPprof, r)

	loginPayload := map[string]interface{}{
		"username": "admin4test",
		"password": "password4test",
	}
	token, err := performLogin(r, loginPayload)
	if err != nil {
		t.Fatalf("Failed to perform login: %v", err)
	}

	tests := []struct {
		name    string
		method  string
		path    string
		server  *server.Server
		status  int
		headers map[string]string
	}{
		{"Check Login Route", "POST", "/api/login", serverWithPprof, http.StatusOK, nil},
		{"Check Server Info Route", "GET", "/api/server/info", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Running Config Route", "GET", "/api/config/running", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		{"Check Config From File Route", "GET", "/api/config/file", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		{"Check Save Config Route", "POST", "/api/config/save", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Server Info Route", "GET", "/api/server/info", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		//Can't test these routes because they will kill the test server
		//{"Check Server Restart Route", "PUT", "/api/server/restart", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		//{"Check Server Shutdown Route", "PUT", "/api/server/shutdown", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		//{"Check Server Kill Route", "PUT", "/api/server/kill", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Connections Route", "GET", "/api/connection/list", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Permissions Route", "GET", "/api/permission/menu", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Pprof Route", "GET", "/debug/pprof/", serverWithPprof, http.StatusOK, nil},

		{"Check Pprof Route", "GET", "/debug/pprof/", serverWithoutPprof, http.StatusNotFound, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			setRoutes(tt.server, r)

			w := performRequestWithHeaders(r, tt.method, tt.path, nil, tt.headers)
			assert.Equal(t, tt.status, w.Code)
		})
	}
}

func generateServerArgs(withPprof bool) []string {
	args := []string{
		"server",
		"-web",
		"-webAddr", "localhost",
		"-webPort", "8080",
		"-admin", "admin4test",
		"-password", "password4test",
		"-signingKey", "signingKey4test",
	}

	if withPprof {
		args = append(args, "-pprof")
	}

	return args
}

func initializeServer(t *testing.T, args []string) *server.Server {
	s, err := server.New(args, nil)
	if err != nil {
		t.Fatalf("Error while initializing the server: %v", err)
	}
	return s
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

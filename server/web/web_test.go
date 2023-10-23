package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/server"
	wServer "github.com/isrc-cas/gt/web/server"
	"io"
	"io/fs"
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
				"-webAddr", "localhost:8000",
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
				"-webAddr", "localhost:8000",
				"-admin", "admin",
				"-password", "password",
				"-signingKey", "",
				"-pprof", "true",
			},
			expectErr: false, // Because it'll get a random value if not set
		},
		// Can't test this because os.Args is not equal to tt.args
		//{
		//	name: "all options not set",
		//	args: []string{
		//		"server",
		//	},
		//	expectErr: false, // server New will create a default webAddr
		//},
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

type RouteTestCase struct {
	name           string
	method         string
	path           string
	server         *server.Server
	headers        map[string]string
	expectRedirect bool
}

func executeTestCase(t *testing.T, r *gin.Engine, testCase RouteTestCase, noRouteContent []byte) {
	w := performRequestWithHeaders(r, testCase.method, testCase.path, nil, testCase.headers)

	if bytes.Equal(w.Body.Bytes(), noRouteContent) {
		if testCase.expectRedirect {
			t.Logf("Request was caught by NoRoute as expected.")
		} else {
			t.Errorf("Request was unexpectedly caught by NoRoute.")
		}
	} else {
		if testCase.expectRedirect {
			t.Errorf("Request was not caught by NoRoute but was expected to be.")
		} else {
			t.Logf("Request was not caught by NoRoute as expected.")
		}
	}
}

func TestSetRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	data, err := fs.ReadFile(FS, "dist/index.html")
	if err != nil {
		t.Fatalf("Failed to read dist/index.html: %v", err)
	}
	noRouteContent := data

	serverWithPprof := initializeServer(t, generateServerArgs(true))
	serverWithoutPprof := initializeServer(t, generateServerArgs(false))

	r := gin.New()
	tm := wServer.NewTokenManager(wServer.DefaultTokenManagerConfig())
	err = setRoutes(serverWithPprof, tm, r)
	if err != nil {
		t.Fatalf("Failed to set routes: %v", err)
	}
	loginPayload := map[string]interface{}{
		"username": "admin4test",
		"password": "password4test",
	}
	token, err := performLogin(r, loginPayload)
	if err != nil {
		t.Fatalf("Failed to perform login: %v", err)
	}
	testCases := []RouteTestCase{
		{"Check Health Route", "GET", "/api/health", serverWithPprof, nil, false},
		{"Check Login Route", "POST", "/api/login", serverWithPprof, nil, false},
		{"Check Server Info Route", "GET", "/api/server/info", serverWithPprof, map[string]string{"x-token": token}, false},

		{"Check UserInfo Route", "GET", "/api/user/info", serverWithPprof, map[string]string{"x-token": token}, false},
		{"Check UserChangeInfo Route", "POST", "/api/user/change", serverWithPprof, map[string]string{"x-token": token}, false},

		{"Check Running Config Route", "GET", "/api/config/running", serverWithPprof, map[string]string{"x-token": token}, false},
		{"Check Config From File Route", "GET", "/api/config/file", serverWithPprof, map[string]string{"x-token": token}, false},
		{"Check Save Config Route", "POST", "/api/config/save", serverWithPprof, map[string]string{"x-token": token}, false},

		{"Check Server Info Route", "GET", "/api/server/info", serverWithPprof, map[string]string{"x-token": token}, false},
		//Can't test these routes because they will kill the test server
		//{"Check Server Restart Route", "PUT", "/api/server/restart", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		//{"Check Server Shutdown Route", "PUT", "/api/server/shutdown", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},
		//{"Check Server Kill Route", "PUT", "/api/server/kill", serverWithPprof, http.StatusOK, map[string]string{"x-token": token}},

		{"Check Connections Route", "GET", "/api/connection/list", serverWithPprof, map[string]string{"x-token": token}, false},
		{"Check Permissions Route", "GET", "/api/permission/menu", serverWithPprof, map[string]string{"x-token": token}, false},

		{"Check Pprof Route with pprof permission", "GET", "/debug/pprof/", serverWithPprof, nil, false},
		{"Check Pprof Route without pprof permission", "GET", "/debug/pprof/", serverWithoutPprof, nil, true},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r := gin.New()
			err := setRoutes(testCase.server, tm, r)
			if err != nil {
				t.Errorf("setRoutes() error = %v", err)
			}
			executeTestCase(t, r, testCase, noRouteContent)
		})
	}
}

func generateServerArgs(withPprof bool) []string {
	args := []string{
		"server",
		"-webAddr", "localhost:8080",
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

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data not found or is not an object in login response")
	}

	token, exists := data["token"]
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

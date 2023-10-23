package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/web/server"
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
				"client",
				"-webAddr", "localhost:7000",
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
				"-admin", "admin",
				"-password", "password",
				"-signingKey", "signingKey",
				"-pprof", "true",
			},
			expectErr: true,
		},
		{
			name: "missing signingKey",
			args: []string{
				"client",
				"-webAddr", "localhost:7000",
				"-admin", "admin",
				"-password", "password",
				"-pprof", "true",
			},
			expectErr: false,
		},
		// Can't be tested because of os.Args is not equal to tt.args
		//{
		//	name: "all options not set",
		//	args: []string{
		//		"client",
		//	},
		//	expectErr: false, //client New will set default values
		//},
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

type RouteTestCase struct {
	name           string
	method         string
	path           string
	server         *client.Client
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
	clientWithPprof := initializeClient(t, generateClientArgs(true))
	clientWithoutPprof := initializeClient(t, generateClientArgs(false))

	r := gin.New()
	tm := server.NewTokenManager(server.DefaultTokenManagerConfig())
	err = setRoutes(clientWithPprof, tm, r)
	if err != nil {
		t.Fatalf("Failed to set routes: %v", err)
	}

	loginPayload := map[string]interface{}{
		"username": "admin",
		"password": "password",
	}
	token, err := performLogin(r, loginPayload)
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}
	testCases := []RouteTestCase{
		{"Check Health Route", "GET", "/api/health", clientWithPprof, nil, false},
		{"Check Login Route", "POST", "/api/login", clientWithPprof, nil, false},
		{"Check Server Info Route", "GET", "/api/server/info", clientWithPprof, map[string]string{"x-token": token}, false},

		{"Check User Info Route", "GET", "/api/user/info", clientWithPprof, map[string]string{"x-token": token}, false},
		{"Check UserChangeInfo Route", "POST", "/api/user/change", clientWithoutPprof, map[string]string{"x-token": token}, false},

		{"Check Running Config Route", "GET", "/api/config/running", clientWithPprof, map[string]string{"x-token": token}, false},
		{"Check Config From File Route", "GET", "/api/config/file", clientWithPprof, map[string]string{"x-token": token}, false},
		{"Check Save Config Route", "POST", "/api/config/save", clientWithPprof, map[string]string{"x-token": token}, false},

		{"Check Server Info Route", "GET", "/api/server/info", clientWithPprof, map[string]string{"x-token": token}, false},

		{"Check Connection Info Route", "GET", "/api/connection/list", clientWithPprof, map[string]string{"x-token": token}, false},
		{"Check Permissions Route", "GET", "/api/permission/menu", clientWithPprof, map[string]string{"x-token": token}, false},

		{"Check Pprof Route with pprof enabled", "GET", "/debug/pprof/", clientWithPprof, nil, false},
		{"Check Pprof Route with pprof disabled", "GET", "/debug/pprof/", clientWithoutPprof, nil, true},
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
func generateClientArgs(withPprof bool) []string {
	args := []string{
		"client",
		"-webAddr", "localhost:7000",
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

package api

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/web/server/model/request"
	"github.com/isrc-cas/gt/web/server/model/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	args := []string{
		"server4test",
		"-admin", "admin4test",
		"-password", "password4test",
	}
	server4test, err := server.New(args, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tests := []struct {
		name         string
		input        string
		server       *server.Server
		expectedCode float64
		checkToken   bool
	}{
		{
			name:         "JSON binding failure",
			input:        `{"invalid": "json"`,
			server:       server4test,
			expectedCode: response.ERROR,
		},
		{
			name: "User verification failure",
			input: func() string {
				data, _ := json.Marshal(request.User{
					Username: "wrongUsername",
					Password: "wrongPassword",
				})
				return string(data)
			}(),
			server:       server4test,
			expectedCode: response.ERROR,
		},
		{
			name: "User verification success",
			input: func() string {
				data, _ := json.Marshal(request.User{
					Username: "admin4test",
					Password: "password4test",
				})
				return string(data)
			}(),
			server:       server4test,
			expectedCode: response.SUCCESS,
			checkToken:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.Default()
			server4test = tt.server
			r.POST("/login", Login(server4test))

			req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.input))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			r.ServeHTTP(resp, req)

			var responseBody map[string]interface{}
			err := json.Unmarshal(resp.Body.Bytes(), &responseBody)
			if err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			// Assert the status code
			assert.Equal(t, http.StatusOK, resp.Code)

			// Asserting the code in the response body
			assert.Equal(t, tt.expectedCode, responseBody["code"])

			// Check if token exists in successful response
			if tt.checkToken {
				data, exists := responseBody["data"].(map[string]interface{})
				if assert.True(t, exists, "Data should be a map") {
					_, tokenExists := data["token"]
					assert.True(t, tokenExists, "Token should exist in data")
				}
			}
		})
	}
}

func TestGetMenu(t *testing.T) {
	argsWithPprof := []string{
		"server4test",
		"-pprof",
	}
	serverWithPprof, err := server.New(argsWithPprof, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	argsWithoutPprof := []string{
		"server4test",
	}
	serverWithoutPprof, err := server.New(argsWithoutPprof, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	gin.SetMode(gin.TestMode)

	type Response struct {
		Code int           `json:"code"`
		Data []interface{} `json:"data"`
		Msg  string        `json:"msg"`
	}

	// Test with pprof
	r1 := gin.New()
	r1.GET("/menu", GetMenu(serverWithPprof))
	req1, _ := http.NewRequest(http.MethodGet, "/menu", nil)
	resp1 := httptest.NewRecorder()
	r1.ServeHTTP(resp1, req1)
	menuWithPprof := resp1.Body.String()

	var responseWithPprof Response
	if err := json.Unmarshal([]byte(menuWithPprof), &responseWithPprof); err != nil {
		t.Fatalf("failed to parse response with pprof: %v", err)
	}

	// Test without pprof
	r2 := gin.New()
	r2.GET("/menu", GetMenu(serverWithoutPprof))
	req2, _ := http.NewRequest(http.MethodGet, "/menu", nil)
	resp2 := httptest.NewRecorder()
	r2.ServeHTTP(resp2, req2)
	menuWithoutPprof := resp2.Body.String()

	var responseWithoutPprof Response
	if err := json.Unmarshal([]byte(menuWithoutPprof), &responseWithoutPprof); err != nil {
		t.Fatalf("failed to parse response without pprof: %v", err)
	}

	if len(responseWithPprof.Data)-len(responseWithoutPprof.Data) != 1 {
		t.Errorf("menuWithPprof items: %d, menuWithoutPprof items: %d", len(responseWithPprof.Data), len(responseWithoutPprof.Data))
	}
}

func TestGetServerInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/server/info", GetServerInfo)

	req, _ := http.NewRequest(http.MethodGet, "/server/info", nil)
	resp := httptest.NewRecorder()

	r.ServeHTTP(resp, req)

	//check HTTPStatusCode
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.Code)
	}

	var jsonResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	data, ok := jsonResponse["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'data' field in response")
	}

	_, ok = data["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'serverInfo' field in 'data'")
	}
}

func TestGetRunningConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	args := []string{
		"server4test",
	}
	server4test, err := server.New(args, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	r := gin.New()
	r.GET("/config/running", GetRunningConfig(server4test))

	testConfigEndpoint(t, r, "/config/running", "Expected 'config' field from running config")
}

func TestGetConfigFromFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Server without config file argument
	argsWithoutConfig := []string{
		"server4test",
	}

	serverWithoutConfig, err := server.New(argsWithoutConfig, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	r1 := gin.New()
	r1.GET("/config/file", GetConfigFromFile(serverWithoutConfig))
	testConfigEndpoint(t, r1, "/config/file", "Expected 'config' field from running config")

	// Server with config file argument
	argsWithConfig := []string{
		"server4test",
		"-config", "../../testdata/config.yaml",
	}

	serverWithConfig, err := server.New(argsWithConfig, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	r2 := gin.New()
	r2.GET("/config/file", GetConfigFromFile(serverWithConfig))
	testConfigEndpoint(t, r2, "/config/file", "Expected 'config' field from file")
}

func testConfigEndpoint(t *testing.T, router *gin.Engine, url, errMsg string) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Check HTTPStatusCode
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.Code)
	}

	var jsonResponse map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &jsonResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	data, ok := jsonResponse["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected 'data' field in response")
	}

	_, ok = data["config"].(map[string]interface{})
	if !ok {
		t.Fatalf(errMsg)
	}
}

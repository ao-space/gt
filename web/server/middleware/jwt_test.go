package middleware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/isrc-cas/gt/web/server/model/response"
	"github.com/isrc-cas/gt/web/server/util"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	signingKey := "test_key"
	j := util.NewJWT(signingKey)
	r.Use(JWTAuthMiddleware(signingKey))

	r.GET("/testEndpoint", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code": response.SUCCESS,
			"msg":  "OK",
		})
	})
	r.GET("/verifyClaims", func(c *gin.Context) {
		claimsValue, exists := c.Get("claims")
		if !exists {
			t.Fatal("Claims not set in the context")
		}
		claims, ok := claimsValue.(*util.CustomClaims)
		if !ok {
			t.Fatal("Claims type is not *CustomClaims")
		}
		if claims.Username != "testuser" {
			t.Fatalf("Expected username to be testuser, got: %s", claims.Username)
		}
		c.JSON(200, gin.H{
			"code": response.SUCCESS,
			"msg":  "OK",
		})
	})

	checkResponseCode := func(w *httptest.ResponseRecorder, expectedCode int) {
		body, _ := ioutil.ReadAll(w.Body)
		var resp map[string]interface{}
		err := json.Unmarshal(body, &resp)
		if err != nil {
			t.Fatal("Failed to parse response body")
		}
		if int(resp["code"].(float64)) != expectedCode {
			t.Fatalf("Expected code %d in the response body, got: %f", expectedCode, resp["code"].(float64))
		}
	}

	// Create a valid token
	claims := j.CreateClaims("testuser", "testIssuer")
	validToken, err := j.CreateToken(claims)
	if err != nil {
		t.Fatal("Failed to create a test token")
	}

	// Scenario 1: No token
	req, _ := http.NewRequest("GET", "/testEndpoint", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	checkResponseCode(w, response.OVERDUE)

	// Scenario 2: Valid token
	req, _ = http.NewRequest("GET", "/testEndpoint", nil)
	req.Header.Set("x-token", validToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	checkResponseCode(w, response.SUCCESS)

	// Scenario 3: Expired token
	expiredClaims := util.CustomClaims{
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-util.TokenExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * util.TokenExpireDuration)),
			Issuer:    "testIssuer",
			Subject:   "user token",
		},
	}
	expiredToken, err := j.CreateToken(expiredClaims)
	if err != nil {
		t.Fatal("Failed to create an expired token")
	}
	req, _ = http.NewRequest("GET", "/testEndpoint", nil)
	req.Header.Set("x-token", expiredToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	checkResponseCode(w, response.OVERDUE)

	// Scenario 4: Token with unknown signing key
	invalidKey := "invalid_key"
	jInvalid := util.NewJWT(invalidKey)
	invalidToken, err := jInvalid.CreateToken(claims)
	if err != nil {
		t.Fatal("Failed to create a token with invalid key")
	}
	req, _ = http.NewRequest("GET", "/testEndpoint", nil)
	req.Header.Set("x-token", invalidToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	checkResponseCode(w, response.OVERDUE)

	// Scenario 5: Token with invalid format
	malformedToken := "malformedToken"
	req, _ = http.NewRequest("GET", "/testEndpoint", nil)
	req.Header.Set("x-token", malformedToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	checkResponseCode(w, response.OVERDUE)

	// Verify that the claims are set in the context after `c.Set("claims", claims)`
	req, _ = http.NewRequest("GET", "/verifyClaims", nil)
	req.Header.Set("x-token", validToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	checkResponseCode(w, response.SUCCESS)

}

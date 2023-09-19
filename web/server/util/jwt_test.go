package util

import (
	"math/rand"
	"testing"
)

func TestJWT(t *testing.T) {
	signingKey := "validSigningKey"
	wrongSigningKey := "wrongSigningKey"
	username := "admin"
	issuer := "testIssuer"
	j := NewJWT(signingKey)

	tests := []struct {
		name       string
		jwt        *JWT
		tokenFunc  func() (string, error)
		expectErr  bool
		expectUser string
	}{
		{
			name: "valid token",
			jwt:  j,
			tokenFunc: func() (string, error) {
				return j.CreateToken(j.CreateClaims(username, issuer))
			},
			expectErr:  false,
			expectUser: username,
		},
		{
			name: "wrong signing key",
			jwt:  NewJWT(wrongSigningKey),
			tokenFunc: func() (string, error) {
				return j.CreateToken(j.CreateClaims(username, issuer))
			},
			expectErr: true,
		},
		{
			name: "modified token",
			jwt:  j,
			tokenFunc: func() (string, error) {
				token, _ := j.CreateToken(j.CreateClaims(username, issuer))

				// Get a random index in the token to modify
				randomIdx := rand.Intn(len(token))

				// Get a random character different from the character at randomIdx
				originalChar := token[randomIdx]
				var randomChar rune
				for {
					randomChar = rune(rand.Intn(26) + 'a')
					if byte(randomChar) != originalChar {
						break
					}
				}

				// Replace the character at randomIdx with randomChar
				modifiedToken := token[:randomIdx] + string(randomChar) + token[randomIdx+1:]
				return modifiedToken, nil
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, _ := tt.tokenFunc()
			parsedClaims, err := tt.jwt.ParseToken(token)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, (err != nil))
			}
			if !tt.expectErr && parsedClaims.Username != tt.expectUser {
				t.Errorf("expected username %s, got %s", tt.expectUser, parsedClaims.Username)
			}
		})
	}
}

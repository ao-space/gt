package service

import (
	"github.com/isrc-cas/gt/server"
	"github.com/isrc-cas/gt/web/server/model/request"
	"testing"
)

func TestVerifyUser(t *testing.T) {
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
		name        string
		inputUser   request.User
		server      *server.Server
		expectedErr bool
	}{
		{
			name: "Matching username and password",
			inputUser: request.User{
				Username: "admin4test",
				Password: "password4test",
			},
			server:      server4test,
			expectedErr: false,
		},
		{
			name: "Mismatched username",
			inputUser: request.User{
				Username: "wrongAdmin",
				Password: "password4test",
			},
			server:      server4test,
			expectedErr: true,
		},
		{
			name: "Mismatched password",
			inputUser: request.User{
				Username: "admin4test",
				Password: "wrongPassword",
			},
			server:      server4test,
			expectedErr: true,
		},
		{
			name: "Mismatched username and password",
			inputUser: request.User{
				Username: "wrongAdmin",
				Password: "wrongPassword",
			},
			server:      server4test,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyUser(tt.inputUser, tt.server)
			gotError := (err != nil)
			if gotError != tt.expectedErr {
				t.Errorf("expected error %v, got error %v", tt.expectedErr, gotError)
			}
		})
	}
}

func TestGetMenu(t *testing.T) {
	serverWithPprof, err := server.New([]string{"server4test", "-pprof"}, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	serverWithoutPprof, err := server.New([]string{"server4test"}, nil)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	tests := []struct {
		name        string
		server      *server.Server
		expectedLen int
	}{
		{
			name:        "Without pprof",
			server:      serverWithoutPprof,
			expectedLen: 3,
		},
		{
			name:        "With pprof",
			server:      serverWithPprof,
			expectedLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := GetMenu(tt.server)
			if len(menu) != tt.expectedLen {
				t.Errorf("expected menu length %d, got %d", tt.expectedLen, len(menu))
			}
		})
	}
}

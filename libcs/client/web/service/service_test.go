package service

import (
	"github.com/isrc-cas/gt/client"
	"github.com/isrc-cas/gt/web/server/model/request"
	"testing"
)

func TestVerifyUser(t *testing.T) {
	args := []string{
		"client4test",
		"-admin", "admin4test",
		"-password", "password4test",
	}
	client4test, err := client.New(args, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	tests := []struct {
		name        string
		inputUser   request.User
		client      *client.Client
		expectedErr bool
	}{

		{
			name: "Matching username and password",
			inputUser: request.User{
				Username: "admin4test",
				Password: "password4test",
			},
			client:      client4test,
			expectedErr: false,
		},
		{
			name: "Mismatched username",
			inputUser: request.User{
				Username: "wrongAdmin",
				Password: "password4test",
			},
			client:      client4test,
			expectedErr: true,
		},
		{
			name: "Mismatched password",
			inputUser: request.User{
				Username: "admin4test",
				Password: "wrongPassword",
			},
			client:      client4test,
			expectedErr: true,
		},
		{
			name: "Mismatched username and password",
			inputUser: request.User{
				Username: "wrongAdmin",
				Password: "wrongPassword",
			},
			client:      client4test,
			expectedErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := VerifyUser(test.inputUser, test.client)
			gotErr := err != nil
			if gotErr != test.expectedErr {
				t.Errorf("VerifyUser() error = %v, expectedErr %v", err, test.expectedErr)
			}
		})
	}
}

func TestGetMenu(t *testing.T) {
	clientWithPprof, err := client.New([]string{"client4test", "-pprof"}, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	clientWithoutPprof, err := client.New([]string{"client4test"}, nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	tests := []struct {
		name        string
		client      *client.Client
		expectedLen int
	}{
		{
			name:        "Without pprof",
			client:      clientWithoutPprof,
			expectedLen: 3,
		},
		{
			name:        "With pprof",
			client:      clientWithPprof,
			expectedLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			menu := GetMenu(tt.client)
			if len(menu) != tt.expectedLen {
				t.Errorf("expected menu length %d, got %d", tt.expectedLen, len(menu))
			}
		})
	}
}

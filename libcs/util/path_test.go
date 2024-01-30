package util

import (
	"os"
	"testing"
)

func TestWriteYamlToFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "writeyamltest")
	if err != nil {
		t.Fatalf("Error creating temp directory: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		inputData   []byte
		filePath    string
		expectError bool
		expected    []byte
	}{
		{
			name:      "normal write",
			inputData: []byte("example: data"),
			filePath:  tmpDir + "/test.yaml",
			expected:  []byte("example: data"),
		},
		{
			name:      "overwrite existing file",
			inputData: []byte("new: data"),
			filePath:  tmpDir + "/test2.yaml",
			expected:  []byte("new: data"),
		},
		{
			name:      "write empty data",
			inputData: []byte(""),
			filePath:  tmpDir + "/empty.yaml",
			expected:  []byte(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.name == "overwrite existing file" {
				err := os.WriteFile(tt.filePath, []byte("initial: data"), 0644)
				if err != nil {
					t.Fatalf("Failed to create file: %s", err)
				}
			}

			err := WriteYamlToFile(tt.filePath, tt.inputData)

			// check if we expect an error
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
					return
				}
			} else if err != nil {
				t.Errorf("Didn't expect an error but got: %s", err)
				return
			}

			// check the content of the file
			content, readErr := os.ReadFile(tt.filePath)
			if readErr != nil {
				t.Fatalf("Failed to read file: %s", readErr)
			}

			if string(content) != string(tt.expected) {
				t.Errorf("Expected content %s, got %s", string(tt.expected), string(content))
			}
		})
	}
}

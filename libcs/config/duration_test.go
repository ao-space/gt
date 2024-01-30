package config

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"testing"
	"time"
)

func TestDurationString(t *testing.T) {
	testCases := []struct {
		duration Duration
		expected string
	}{
		{duration: Duration{0}, expected: "0s"},
		{duration: Duration{time.Nanosecond}, expected: "1ns"},
		{duration: Duration{time.Microsecond}, expected: "1µs"},
		{duration: Duration{time.Millisecond}, expected: "1ms"},
		{duration: Duration{time.Second}, expected: "1s"},
		{duration: Duration{time.Minute}, expected: "1m0s"},
		{duration: Duration{time.Hour}, expected: "1h0m0s"},

		{duration: Duration{time.Second * 90}, expected: "1m30s"},
		{duration: Duration{time.Second * 120}, expected: "2m0s"},
		{duration: Duration{time.Hour + 30*time.Minute + 30*time.Second}, expected: "1h30m30s"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			actual := tc.duration.String()
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}

func TestDurationSet(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{input: "", expected: 0, hasError: true},
		{input: "123", expected: 0, hasError: true},
		{input: "1x", expected: 0, hasError: true},
		{input: "hello", expected: 0, hasError: true},

		{input: "1ns", expected: time.Nanosecond, hasError: false},
		{input: "2µs", expected: time.Microsecond * 2, hasError: false},
		{input: "3ms", expected: time.Millisecond * 3, hasError: false},
		{input: "4s", expected: time.Second * 4, hasError: false},
		{input: "5m", expected: time.Minute * 5, hasError: false},
		{input: "6h", expected: time.Hour * 6, hasError: false},
		{input: "7h30m", expected: time.Hour*7 + time.Minute*30, hasError: false},
		{input: "8m9s10h", expected: time.Hour*10 + time.Minute*8 + time.Second*9, hasError: false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			d := &Duration{}
			err := d.Set(tc.input)
			if tc.hasError && err == nil {
				t.Fatalf("expected an error, but got none for input %s", tc.input)
			}
			if !tc.hasError && err != nil {
				t.Fatalf("didn't expect an error, but got: %v for input %s", err, tc.input)
			}
			if d.Duration != tc.expected {
				t.Errorf("for input %s, expected duration %v, got %v", tc.input, tc.expected, d.Duration)
			}
		})
	}
}

func TestDurationGet(t *testing.T) {
	testCases := []struct {
		duration Duration
		expected time.Duration
	}{
		{duration: Duration{time.Second}, expected: time.Second},
		{duration: Duration{time.Minute}, expected: time.Minute},
		{duration: Duration{time.Hour}, expected: time.Hour},
		{duration: Duration{time.Millisecond * 500}, expected: time.Millisecond * 500},
		{duration: Duration{time.Second * 90}, expected: time.Second * 90},
	}

	for _, tc := range testCases {
		t.Run(tc.expected.String(), func(t *testing.T) {
			actual := tc.duration.Get()
			if actualDuration, ok := actual.(time.Duration); ok {
				if actualDuration != tc.expected {
					t.Errorf("expected %v, got %v", tc.expected, actualDuration)
				}
			} else {
				t.Errorf("Get() did not return a time.Duration for %v", tc.duration)
			}
		})
	}
}

func TestDurationUnmarshalYAML(t *testing.T) {
	testCases := []struct {
		yamlInput string
		expected  time.Duration
		hasError  bool
	}{
		{yamlInput: "duration: 1x", expected: 0, hasError: true},
		{yamlInput: "duration: hello", expected: 0, hasError: true},
		{yamlInput: "duration: 123", expected: 0, hasError: true},
		{yamlInput: `duration: ""`, expected: 0, hasError: true},
		{yamlInput: "duration: 1d", expected: 0, hasError: true},

		{yamlInput: "duration:  ", expected: 0, hasError: false},
		{yamlInput: "duration: 1s", expected: time.Second, hasError: false},
		{yamlInput: "duration: 1m", expected: time.Minute, hasError: false},
		{yamlInput: "duration: 1h", expected: time.Hour, hasError: false},
		{yamlInput: "duration: 500ms", expected: time.Millisecond * 500, hasError: false},
		{yamlInput: "duration: 1h30m", expected: time.Hour + time.Minute*30, hasError: false},
	}

	for _, tc := range testCases {
		t.Run(tc.yamlInput, func(t *testing.T) {
			var output struct {
				Duration Duration
			}
			err := yaml.Unmarshal([]byte(tc.yamlInput), &output)
			if tc.hasError && err == nil {
				t.Fatalf("expected an error, but got none for input %s", tc.yamlInput)
			}
			if !tc.hasError && err != nil {
				t.Fatalf("didn't expect an error, but got: %v for input %s", err, tc.yamlInput)
			}
			if output.Duration.Duration != tc.expected {
				t.Errorf("for input %s, expected duration %v, got %v", tc.yamlInput, tc.expected, output.Duration.Duration)
			}
		})
	}
}

func TestDurationMarshalYAML(t *testing.T) {
	testCases := []struct {
		duration Duration
		expected string
	}{
		{duration: Duration{time.Nanosecond}, expected: "1ns"},
		{duration: Duration{time.Millisecond * 500}, expected: "500ms"},
		{duration: Duration{time.Second}, expected: "1s"},
		{duration: Duration{time.Minute}, expected: "1m0s"},
		{duration: Duration{time.Hour}, expected: "1h0m0s"},
		{duration: Duration{time.Hour + time.Minute*30}, expected: "1h30m0s"},
		{duration: Duration{time.Hour + time.Minute*30 + time.Second*15}, expected: "1h30m15s"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result, err := yaml.Marshal(tc.duration)
			if err != nil {
				t.Fatalf("didn't expect an error, but got: %v", err)
			}
			if string(result) != tc.expected+"\n" { // yaml.Marshal adds a newline at the end
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestDurationUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		jsonInput string
		expected  time.Duration
		hasError  bool
	}{
		{jsonInput: `"1x"`, expected: 0, hasError: true},
		{jsonInput: `"hello"`, expected: 0, hasError: true},
		{jsonInput: `"123"`, expected: 0, hasError: true},
		{jsonInput: `""`, expected: 0, hasError: true},
		{jsonInput: `"1d"`, expected: 0, hasError: true},

		{jsonInput: `"1532ms"`, expected: time.Millisecond * 1532, hasError: false},
		{jsonInput: `"1s"`, expected: time.Second, hasError: false},
		{jsonInput: `"5m"`, expected: time.Minute * 5, hasError: false},
		{jsonInput: `"90m"`, expected: time.Minute * 90, hasError: false},
		{jsonInput: `"7h"`, expected: time.Hour * 7, hasError: false},
		{jsonInput: `"1h30m"`, expected: time.Hour + time.Minute*30, hasError: false},
	}

	for _, tc := range testCases {
		t.Run(tc.jsonInput, func(t *testing.T) {
			var d Duration
			err := json.Unmarshal([]byte(tc.jsonInput), &d)
			if tc.hasError && err == nil {
				t.Fatalf("expected an error, but got none for input %s", tc.jsonInput)
			}
			if !tc.hasError && err != nil {
				t.Fatalf("didn't expect an error, but got: %v for input %s", err, tc.jsonInput)
			}
			if d.Duration != tc.expected {
				t.Errorf("for input %s, expected duration %v, got %v", tc.jsonInput, tc.expected, d.Duration)
			}
		})
	}
}

func TestDurationMarshalJSON(t *testing.T) {
	testCases := []struct {
		duration Duration
		expected string
	}{
		{duration: Duration{time.Second}, expected: `"1s"`},
		{duration: Duration{time.Minute}, expected: `"1m0s"`},
		{duration: Duration{time.Hour}, expected: `"1h0m0s"`},
		{duration: Duration{time.Hour + time.Minute*30}, expected: `"1h30m0s"`},
		{duration: Duration{time.Hour + time.Minute*30 + time.Second*15}, expected: `"1h30m15s"`},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result, err := json.Marshal(tc.duration)
			if err != nil {
				t.Fatalf("didn't expect an error, but got: %v", err)
			}
			if string(result) != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

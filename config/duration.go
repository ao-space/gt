package config

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"time"
)

type Duration struct {
	time.Duration
}

// String flag.Value Interface
func (d *Duration) String() string {
	return d.Duration.String()
}

// Set flag.Value Interface
func (d *Duration) Set(s string) error {
	var err error
	d.Duration, err = time.ParseDuration(s)
	return err
}

// Get flag.Getter Interface
func (d *Duration) Get() interface{} {
	return d.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) (err error) {
	duration, err := time.ParseDuration(value.Value)
	if err != nil {
		return
	}
	d.Duration = duration
	return
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return d.Duration.String(), nil
}

func (d *Duration) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

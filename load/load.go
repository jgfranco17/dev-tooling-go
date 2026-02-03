package load

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"
)

// FromJSON loads and parses JSON data from the provided reader
// into any arbitrary Go type.
func FromJSON[T any](data io.Reader) (T, error) {
	var v T
	err := json.NewDecoder(data).Decode(&v)
	return v, err
}

// FromYAML loads and parses YAML data from the provided reader
// into any arbitrary Go type.
func FromYAML[T any](data io.Reader) (T, error) {
	var v T
	err := yaml.NewDecoder(data).Decode(&v)
	return v, err
}

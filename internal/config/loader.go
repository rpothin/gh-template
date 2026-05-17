package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadManifest parses a YAML manifest file from disk.
func LoadManifest(path string) (*Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}
	if err := ValidateManifest(&m); err != nil {
		return nil, fmt.Errorf("validating config file %q: %w", path, err)
	}
	return &m, nil
}

// ParseManifestBytes parses YAML bytes into a Manifest with strict field checking and validation.
// Used by both local file loading and remote manifest fetching.
func ParseManifestBytes(data []byte) (*Manifest, error) {
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	dec.KnownFields(true)
	var m Manifest
	if err := dec.Decode(&m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	if err := ValidateManifest(&m); err != nil {
		return nil, fmt.Errorf("validating manifest: %w", err)
	}
	return &m, nil
}

// ManifestToYAML serializes a Manifest to YAML bytes.
func ManifestToYAML(m *Manifest) ([]byte, error) {
	return yaml.Marshal(m)
}

// WriteManifest serializes a Manifest and writes it to a file.
func WriteManifest(path string, m *Manifest) error {
	data, err := ManifestToYAML(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

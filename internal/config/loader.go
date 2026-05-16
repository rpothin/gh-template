package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadManifest parses a YAML manifest file from disk.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
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

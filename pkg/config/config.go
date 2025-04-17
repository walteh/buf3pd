package config

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
	"gopkg.in/yaml.v3"
)

// Buf3pdDep represents a dependency in the buf3pd configuration
type Buf3pdDep struct {
	Type   string   `yaml:"type"`
	Repo   string   `yaml:"repo"`
	Path   string   `yaml:"path"`
	Ref    string   `yaml:"ref"`
	Filter []string `yaml:"filter"`
}

// Config represents the configuration structure in buf.yaml
type Config struct {
	Path string      `yaml:"path"`
	Deps []Buf3pdDep `yaml:"deps"`
}

// Reader provides an interface for reading configuration
type Reader interface {
	ReadConfig(ctx context.Context, path string) (*Config, error)
}

// FileReader implements the Reader interface
type FileReader struct{}

// NewFileReader creates a new FileReader
func NewFileReader() *FileReader {
	return &FileReader{}
}

// ReadConfig reads the buf3pd configuration from a file
func (r *FileReader) ReadConfig(ctx context.Context, path string) (*Config, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Str("path", path).Msg("reading buf3pd config")

	// Parse buf.yaml
	bufYamlContent, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("reading buf.yaml: %w", err)
	}

	// Split by "---" to find the sections
	sections := strings.Split(string(bufYamlContent), "---")
	if len(sections) < 2 {
		return nil, errors.New("buf.yaml does not contain the expected sections separated by '---'")
	}

	content := sections[1]

	var buf3pdConfig struct {
		Buf3pd *Config `yaml:"buf3pd"`
	}
	if err := yaml.Unmarshal([]byte(content), &buf3pdConfig); err != nil {
		return nil, errors.Errorf("unmarshalling buf3pd config: %w", err)
	}

	if buf3pdConfig.Buf3pd == nil || len(buf3pdConfig.Buf3pd.Deps) == 0 {
		return nil, errors.New("buf3pd config does not contain any dependencies")
	}

	return buf3pdConfig.Buf3pd, nil
}

// ValidatePath ensures the output path exists, creating it if necessary
func ValidatePath(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.Errorf("creating output directory: %w", err)
	}
	return nil
}

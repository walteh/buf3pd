package config

import (
	"context"
	"os"
	"path/filepath"
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

// BufModule represents a module in the buf.yaml modules section
type BufModule struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// BufYaml represents the complete buf.yaml file structure
type BufYaml struct {
	Version  string                 `yaml:"version"`
	Modules  []BufModule            `yaml:"modules,omitempty"`
	Breaking map[string]interface{} `yaml:"breaking,omitempty"`
	Deps     []string               `yaml:"deps,omitempty"`
	Lint     map[string]interface{} `yaml:"lint,omitempty"`
	Buf3pd   *Config                `yaml:"buf3pd,omitempty"`
}

// Reader provides an interface for reading configuration
type Reader interface {
	ReadConfig(ctx context.Context, workDir string, configPath string) (*Config, error)
	ReadBufYaml(ctx context.Context, path string) (*BufYaml, error)
	WriteBufYaml(ctx context.Context, path string, bufYaml *BufYaml) error
	EnsureModulesInBufYaml(ctx context.Context, path string, outputPath string, deps []Buf3pdDep) error
}

// FileReader implements the Reader interface
type FileReader struct{}

// NewFileReader creates a new FileReader
func NewFileReader() *FileReader {
	return &FileReader{}
}

// ReadConfig reads the buf3pd configuration, checking for a dedicated buf3pd.yaml file first
func (r *FileReader) ReadConfig(ctx context.Context, workDir string, configPath string) (*Config, error) {
	log := zerolog.Ctx(ctx)

	// First try to read from buf3pd.yaml if it exists
	buf3pdYamlPath := filepath.Join(workDir, "buf.3pd.yaml")
	if _, err := os.Stat(buf3pdYamlPath); err == nil {
		log.Info().Str("path", buf3pdYamlPath).Msg("reading buf3pd.yaml config")

		content, err := os.ReadFile(buf3pdYamlPath)
		if err != nil {
			return nil, errors.Errorf("reading buf3pd.yaml: %w", err)
		}

		var config Config
		if err := yaml.Unmarshal(content, &config); err != nil {
			return nil, errors.Errorf("unmarshalling buf3pd.yaml: %w", err)
		}

		if len(config.Deps) == 0 {
			return nil, errors.New("buf3pd.yaml does not contain any dependencies")
		}

		return &config, nil
	}

	// Fall back to reading from buf.yaml
	log.Info().Str("path", configPath).Msg("reading buf3pd config from buf.yaml")

	// Read the full buf.yaml file
	bufYaml, err := r.ReadBufYaml(ctx, configPath)
	if err != nil {
		return nil, errors.Errorf("reading buf.yaml: %w", err)
	}

	// Check if the Buf3pd section exists
	if bufYaml.Buf3pd == nil || len(bufYaml.Buf3pd.Deps) == 0 {
		return nil, errors.New("buf.yaml does not contain any buf3pd dependencies")
	}

	return bufYaml.Buf3pd, nil
}

// readBufYamlWithMultiDoc reads buf.yaml handling the multi-document format
func readBufYamlWithMultiDoc(content []byte) (*BufYaml, error) {
	// Split by "---" to find the sections
	sections := strings.Split(string(content), "---")

	// If there are multiple sections, we want the second one
	var yamlContent string
	if len(sections) >= 2 {
		yamlContent = sections[1]
	} else {
		yamlContent = sections[0]
	}

	var bufYaml BufYaml
	if err := yaml.Unmarshal([]byte(yamlContent), &bufYaml); err != nil {
		return nil, errors.Errorf("unmarshalling buf.yaml: %w", err)
	}

	return &bufYaml, nil
}

// ReadBufYaml reads the complete buf.yaml file
func (r *FileReader) ReadBufYaml(ctx context.Context, path string) (*BufYaml, error) {
	log := zerolog.Ctx(ctx)
	log.Info().Str("path", path).Msg("reading buf.yaml")

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("reading buf.yaml: %w", err)
	}

	return readBufYamlWithMultiDoc(content)
}

// WriteBufYaml writes the buf.yaml file preserving the multi-document format
func (r *FileReader) WriteBufYaml(ctx context.Context, path string, bufYaml *BufYaml) error {
	log := zerolog.Ctx(ctx)
	log.Info().Str("path", path).Msg("writing buf.yaml")

	// Read the original file to preserve the structure
	content, err := os.ReadFile(path)
	if err != nil {
		return errors.Errorf("reading buf.yaml: %w", err)
	}

	// Split by "---" to find the sections
	sections := strings.Split(string(content), "---")

	// Marshal the updated YAML
	yamlContent, err := yaml.Marshal(bufYaml)
	if err != nil {
		return errors.Errorf("marshalling buf.yaml: %w", err)
	}

	var outputContent string
	if len(sections) < 2 {
		// No sections found, just write the YAML
		outputContent = string(yamlContent)
	} else {
		// Preserve the first section and update the second
		outputContent = sections[0] + "---\n" + string(yamlContent)
	}

	if err := os.WriteFile(path, []byte(outputContent), 0644); err != nil {
		return errors.Errorf("writing buf.yaml: %w", err)
	}

	return nil
}

// EnsureModulesInBufYaml ensures that all dependencies are properly referenced in the buf.yaml modules section
func (r *FileReader) EnsureModulesInBufYaml(ctx context.Context, path string, outputPath string, deps []Buf3pdDep) error {
	log := zerolog.Ctx(ctx)
	log.Info().Str("path", path).Msg("ensuring modules in buf.yaml")

	// Read the current buf.yaml
	bufYaml, err := r.ReadBufYaml(ctx, path)
	if err != nil {
		return errors.Errorf("reading buf.yaml: %w", err)
	}

	// Create a map of existing modules for quick lookup
	moduleMap := make(map[string]bool)
	for _, module := range bufYaml.Modules {
		moduleMap[module.Name] = true
	}

	// Add modules for each dependency if they don't already exist
	updated := false
	for _, dep := range deps {
		moduleName := dep.Repo
		modulePath := filepath.Join(outputPath, filepath.Base(dep.Repo))

		if !moduleMap[moduleName] {
			bufYaml.Modules = append(bufYaml.Modules, BufModule{
				Name: moduleName,
				Path: modulePath,
			})
			moduleMap[moduleName] = true
			updated = true
			log.Info().Str("name", moduleName).Str("path", modulePath).Msg("added module to buf.yaml")
		}
	}

	// Only write the file if we made changes
	if updated {
		if err := r.WriteBufYaml(ctx, path, bufYaml); err != nil {
			return errors.Errorf("writing buf.yaml: %w", err)
		}
	}

	return nil
}

// ValidatePath ensures the output path exists, creating it if necessary
func ValidatePath(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.Errorf("creating output directory: %w", err)
	}
	return nil
}

package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConfig(t *testing.T) {
	// Setup test context
	ctx := context.Background()
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "buf3pd-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test buf.yaml file
	testBufYaml := `version: v1
---
buf3pd:
  path: proto
  deps:
    - type: git
      repo: github.com/example/repo
      path: proto
      ref: main
      filter:
        - "**/*.proto"
`
	bufYamlPath := filepath.Join(tempDir, "buf.yaml")
	err = os.WriteFile(bufYamlPath, []byte(testBufYaml), 0644)
	assert.NoError(t, err)

	// Test the config reader
	reader := NewFileReader()
	config, err := reader.ReadConfig(ctx, tempDir, bufYamlPath)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "proto", config.Path)
	assert.Len(t, config.Deps, 1)
	assert.Equal(t, "git", config.Deps[0].Type)
	assert.Equal(t, "github.com/example/repo", config.Deps[0].Repo)
	assert.Equal(t, "proto", config.Deps[0].Path)
	assert.Equal(t, "main", config.Deps[0].Ref)
	assert.Len(t, config.Deps[0].Filter, 1)
	assert.Equal(t, "**/*.proto", config.Deps[0].Filter[0])
}

func TestReadConfigFromBuf3pdYaml(t *testing.T) {
	// Setup test context
	ctx := context.Background()
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "buf3pd-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test buf3pd.yaml file
	testBuf3pdYaml := `path: proto
deps:
  - type: git
    repo: github.com/example/repo
    path: proto
    ref: main
    filter:
      - "**/*.proto"
`
	buf3pdYamlPath := filepath.Join(tempDir, "buf3pd.yaml")
	err = os.WriteFile(buf3pdYamlPath, []byte(testBuf3pdYaml), 0644)
	assert.NoError(t, err)

	// Test the config reader
	reader := NewFileReader()
	config, err := reader.ReadConfig(ctx, tempDir, filepath.Join(tempDir, "buf.yaml"))
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "proto", config.Path)
	assert.Len(t, config.Deps, 1)
	assert.Equal(t, "git", config.Deps[0].Type)
	assert.Equal(t, "github.com/example/repo", config.Deps[0].Repo)
	assert.Equal(t, "proto", config.Deps[0].Path)
	assert.Equal(t, "main", config.Deps[0].Ref)
	assert.Len(t, config.Deps[0].Filter, 1)
	assert.Equal(t, "**/*.proto", config.Deps[0].Filter[0])
}

func TestReadBufYaml(t *testing.T) {
	// Setup test context
	ctx := context.Background()
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "buf3pd-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test buf.yaml file with modules
	testBufYaml := `version: v1
---
version: v1
modules:
  - name: github.com/example/repo1
    path: gen/buf3pd/repo1
  - name: github.com/example/repo2
    path: gen/buf3pd/repo2
buf3pd:
  path: proto
  deps:
    - type: git
      repo: github.com/example/repo3
      path: proto
      ref: main
      filter:
        - "**/*.proto"
`
	bufYamlPath := filepath.Join(tempDir, "buf.yaml")
	err = os.WriteFile(bufYamlPath, []byte(testBufYaml), 0644)
	assert.NoError(t, err)

	// Test reading the full buf.yaml file
	reader := NewFileReader()
	bufYaml, err := reader.ReadBufYaml(ctx, bufYamlPath)
	assert.NoError(t, err)
	assert.NotNil(t, bufYaml)
	assert.Equal(t, "v1", bufYaml.Version)
	assert.Len(t, bufYaml.Modules, 2)
	assert.Equal(t, "github.com/example/repo1", bufYaml.Modules[0].Name)
	assert.Equal(t, "gen/buf3pd/repo1", bufYaml.Modules[0].Path)
	assert.Equal(t, "github.com/example/repo2", bufYaml.Modules[1].Name)
	assert.Equal(t, "gen/buf3pd/repo2", bufYaml.Modules[1].Path)
}

func TestEnsureModulesInBufYaml(t *testing.T) {
	// Setup test context
	ctx := context.Background()
	logger := zerolog.New(zerolog.NewTestWriter(t)).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "buf3pd-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test buf.yaml file with existing modules
	testBufYaml := `version: v1
---
version: v1
modules:
  - name: github.com/example/repo1
    path: gen/buf3pd/repo1
buf3pd:
  path: proto
  deps:
    - type: git
      repo: github.com/example/repo1
      path: proto
      ref: main
      filter:
        - "**/*.proto"
    - type: git
      repo: github.com/example/repo2
      path: proto
      ref: main
      filter:
        - "**/*.proto"
`
	bufYamlPath := filepath.Join(tempDir, "buf.yaml")
	err = os.WriteFile(bufYamlPath, []byte(testBufYaml), 0644)
	assert.NoError(t, err)

	// Create test dependencies
	deps := []Buf3pdDep{
		{
			Type: "git",
			Repo: "github.com/example/repo1",
			Path: "proto",
			Ref:  "main",
		},
		{
			Type: "git",
			Repo: "github.com/example/repo2",
			Path: "proto",
			Ref:  "main",
		},
		{
			Type: "git",
			Repo: "github.com/example/repo3",
			Path: "proto",
			Ref:  "main",
		},
	}

	// Test ensuring modules in buf.yaml
	reader := NewFileReader()
	err = reader.EnsureModulesInBufYaml(ctx, bufYamlPath, "proto", deps)
	assert.NoError(t, err)

	// Read the updated buf.yaml
	bufYaml, err := reader.ReadBufYaml(ctx, bufYamlPath)
	require.NoError(t, err)
	assert.Len(t, bufYaml.Modules, 3)

	// Verify that github.com/example/repo1 is still there
	assert.Equal(t, "github.com/example/repo1", bufYaml.Modules[0].Name)
	assert.Equal(t, "gen/buf3pd/repo1", bufYaml.Modules[0].Path)

	// Verify that github.com/example/repo2 and github.com/example/repo3 were added
	moduleNames := make(map[string]bool)
	for _, module := range bufYaml.Modules {
		moduleNames[module.Name] = true
	}
	assert.True(t, moduleNames["github.com/example/repo1"])
	assert.True(t, moduleNames["github.com/example/repo2"])
	assert.True(t, moduleNames["github.com/example/repo3"])
}

func TestValidatePath(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "buf3pd-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test creating a path that doesn't exist yet
	testPath := filepath.Join(tempDir, "new-dir", "sub-dir")
	err = ValidatePath(testPath)
	assert.NoError(t, err)

	// Verify the directory was created
	_, err = os.Stat(testPath)
	assert.NoError(t, err)
}

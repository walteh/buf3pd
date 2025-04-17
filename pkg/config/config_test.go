package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
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
	config, err := reader.ReadConfig(ctx, bufYamlPath)
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

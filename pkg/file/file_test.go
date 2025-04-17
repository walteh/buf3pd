package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDigest(t *testing.T) {
	// Create test files
	file1 := &File{
		Path:    "test/file1.proto",
		Content: []byte("content of file 1"),
	}
	file2 := &File{
		Path:    "test/file2.proto",
		Content: []byte("content of file 2"),
	}

	files := []*File{file1, file2}

	// Test digest calculation
	manager := NewManager()
	digest, err := manager.CalculateDigest(files)
	assert.NoError(t, err)
	assert.NotEmpty(t, digest)

	// Test that the same files produce the same digest
	digest2, err := manager.CalculateDigest(files)
	assert.NoError(t, err)
	assert.Equal(t, digest, digest2)

	// Test that different files produce different digests
	files[1].Content = []byte("modified content")
	digest3, err := manager.CalculateDigest(files)
	assert.NoError(t, err)
	assert.NotEqual(t, digest, digest3)
}

func TestWriteFiles(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "buf3pd-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test files
	files := []*File{
		{
			Path:    "dir1/file1.proto",
			Content: []byte("content of file 1"),
		},
		{
			Path:    "dir2/file2.proto",
			Content: []byte("content of file 2"),
		},
	}

	// Test writing files
	manager := NewManager()
	err = manager.WriteFiles(files, tempDir)
	assert.NoError(t, err)

	// Verify files were written correctly
	for _, file := range files {
		fullPath := filepath.Join(tempDir, file.Path)
		assert.FileExists(t, fullPath)
		content, err := os.ReadFile(fullPath)
		assert.NoError(t, err)
		assert.Equal(t, file.Content, content)
	}
}

func TestFindProtoFiles(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "buf3pd-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	dirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir2"),
		filepath.Join(tempDir, "dir2/subdir"),
	}
	for _, dir := range dirs {
		err = os.MkdirAll(dir, 0755)
		assert.NoError(t, err)
	}

	// Create test files
	files := map[string]string{
		filepath.Join(tempDir, "dir1/file1.proto"):        "content of file 1",
		filepath.Join(tempDir, "dir2/file2.proto"):        "content of file 2",
		filepath.Join(tempDir, "dir2/subdir/file3.proto"): "content of file 3",
		filepath.Join(tempDir, "dir2/subdir/file4.txt"):   "not a proto file",
	}

	for path, content := range files {
		err = os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Test finding proto files without filter
	manager := NewManager()
	protoFiles, err := manager.FindProtoFiles(tempDir, []string{})
	assert.NoError(t, err)
	assert.Len(t, protoFiles, 3)

	// Test finding proto files with filter
	protoFiles, err = manager.FindProtoFiles(tempDir, []string{"dir2/**/*.proto"})
	assert.NoError(t, err)
	assert.Len(t, protoFiles, 2)
	foundDir2 := false
	foundSubdir := false
	for _, file := range protoFiles {
		if file == "dir2/file2.proto" {
			foundDir2 = true
		}
		if file == "dir2/subdir/file3.proto" {
			foundSubdir = true
		}
	}
	assert.True(t, foundDir2)
	assert.True(t, foundSubdir)
}

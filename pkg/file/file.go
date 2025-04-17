package file

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gitlab.com/tozd/go/errors"
)

// File represents a file with its path and content
type File struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

// Handler provides an interface for file operations
type Handler interface {
	FindProtoFiles(directory string, filters []string) ([]string, error)
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, content []byte) error
	WriteFiles(files []*File, basePath string) error
	CalculateDigest(files []*File) (string, error)
}

// Manager implements the Handler interface
type Manager struct{}

// NewManager creates a new Manager
func NewManager() *Manager {
	return &Manager{}
}

// FindProtoFiles finds all proto files in a directory matching the filters
func (m *Manager) FindProtoFiles(directory string, filters []string) ([]string, error) {
	files, err := doublestar.Glob(os.DirFS(directory), "**/*.proto")
	if err != nil {
		return nil, errors.Errorf("finding proto files: %w", err)
	}

	if len(filters) > 0 {
		files = slices.DeleteFunc(files, func(file string) bool {
			for _, filter := range filters {
				ok, err := doublestar.PathMatch(filter, file)
				if err != nil || !ok {
					return true
				}
			}
			return false
		})
	}

	if len(files) == 0 {
		return nil, errors.New("no proto files found in: " + directory)
	}

	return files, nil
}

// ReadFile reads a file from disk
func (m *Manager) ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Errorf("reading file: %w", err)
	}
	return content, nil
}

// WriteFile writes a file to disk
func (m *Manager) WriteFile(path string, content []byte) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.Errorf("creating directory: %w", err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return errors.Errorf("writing file: %w", err)
	}

	return nil
}

// WriteFiles writes multiple files to disk with a base path
func (m *Manager) WriteFiles(files []*File, basePath string) error {
	for _, file := range files {
		outfilePath := filepath.Join(basePath, file.Path)
		if err := os.MkdirAll(filepath.Dir(outfilePath), 0755); err != nil {
			return errors.Errorf("creating output directory: %w", err)
		}
		if err := os.WriteFile(outfilePath, file.Content, 0644); err != nil {
			return errors.Errorf("writing file: %w", err)
		}
	}
	return nil
}

const zeroHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

// CalculateDigest calculates a SHA-256 digest for a slice of files
func (m *Manager) CalculateDigest(files []*File) (string, error) {
	hash := sha256.New()

	// ensure the files are sorted
	slices.SortFunc(files, func(a, b *File) int {
		return strings.Compare(a.Path, b.Path)
	})

	for _, file := range files {
		_, err := hash.Write([]byte(file.Path))
		if err != nil {
			return "", errors.Errorf("writing to hash: %w", err)
		}
		_, err = hash.Write(file.Content)
		if err != nil {
			return "", errors.Errorf("writing to hash: %w", err)
		}
	}

	out := hex.EncodeToString(hash.Sum(nil))

	if out == zeroHash {
		return "", errors.New("zero hash")
	}

	return out, nil
}

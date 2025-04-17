package git

import (
	"os"
	"os/exec"
	"strings"

	"gitlab.com/tozd/go/errors"
)

// Handler provides an interface for git operations
type Handler interface {
	Clone(repo string, path string) error
	FetchTags(repoPath string) error
	Checkout(repoPath string, ref string) error
	GetCommitHash(repoPath string) (string, error)
}

// Manager implements the Handler interface
type Manager struct{}

// NewManager creates a new Manager
func NewManager() *Manager {
	return &Manager{}
}

// Clone clones a git repository to a local path
func (m *Manager) Clone(repo string, path string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", "https://"+repo, path)
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("git clone: %w: %s", err, string(output))
	}
	return nil
}

// FetchTags fetches tags from the origin
func (m *Manager) FetchTags(repoPath string) error {
	cmd := exec.Command("git", "fetch", "origin", "--tags")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("git fetch: %w: %s", err, string(output))
	}
	return nil
}

// Checkout checks out a reference (branch, tag, or commit)
func (m *Manager) Checkout(repoPath string, ref string) error {
	cmd := exec.Command("git", "checkout", ref)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Errorf("git checkout: %w: %s", err, string(output))
	}
	return nil
}

// GetCommitHash gets the current commit hash
func (m *Manager) GetCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	commitHash, err := cmd.Output()
	if err != nil {
		return "", errors.Errorf("git rev-parse: %w", err)
	}
	return strings.TrimSpace(string(commitHash)), nil
}

// CreateTempDir creates a temporary directory for git operations
func CreateTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", "buf3pd-git-")
	if err != nil {
		return "", errors.Errorf("creating temp directory: %w", err)
	}
	return tempDir, nil
}

// CleanupTempDir removes a temporary directory
func CleanupTempDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return errors.Errorf("removing temp directory: %w", err)
	}
	return nil
}

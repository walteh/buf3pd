package deps

import (
	"context"
	"path/filepath"

	"github.com/walteh/buf3pd/pkg/config"
	"github.com/walteh/buf3pd/pkg/file"
	"github.com/walteh/buf3pd/pkg/git"
	"gitlab.com/tozd/go/errors"
)

// NewDepFilesFromRemote creates a DepFiles from a remote repository
func NewDepFilesFromRemote(
	ctx context.Context,
	dep config.Buf3pdDep,
	fileHandler file.Handler,
	gitHandler git.Handler,
) (*DepFiles, error) {
	tempDir, err := git.CreateTempDir()
	if err != nil {
		return nil, errors.Errorf("creating temp directory: %w", err)
	}
	defer git.CleanupTempDir(tempDir)

	// Clone the repository
	if err := gitHandler.Clone(dep.Repo, tempDir); err != nil {
		return nil, errors.Errorf("cloning repository: %w", err)
	}

	// Fetch tags
	if err := gitHandler.FetchTags(tempDir); err != nil {
		return nil, errors.Errorf("fetching tags: %w", err)
	}

	// Checkout the specified reference
	if err := gitHandler.Checkout(tempDir, dep.Ref); err != nil {
		return nil, errors.Errorf("checking out reference: %w", err)
	}

	// Get commit hash
	commit, err := gitHandler.GetCommitHash(tempDir)
	if err != nil {
		return nil, errors.Errorf("getting commit hash: %w", err)
	}

	depFiles := &DepFiles{
		CommitMetadata: commit,
		DepInfo:        dep,
		Files:          []*file.File{},
	}

	if err := depFiles.AddAllNestedProtoFiles(ctx, fileHandler, filepath.Join(tempDir, dep.Path), dep.Filter...); err != nil {
		return nil, errors.Errorf("adding proto files: %w", err)
	}

	if len(depFiles.Files) == 0 {
		return nil, errors.New("no proto files found")
	}

	return depFiles, nil
}

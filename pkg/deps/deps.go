package deps

import (
	"context"
	"path/filepath"
	"slices"
	"strings"

	"github.com/rs/zerolog"
	"github.com/walteh/buf3pd/pkg/config"
	"github.com/walteh/buf3pd/pkg/file"
	"github.com/walteh/buf3pd/pkg/lock"
	"gitlab.com/tozd/go/errors"
)

// DepFiles represents a set of dependency files
type DepFiles struct {
	DepInfo        config.Buf3pdDep `yaml:"dep"`
	Files          []*file.File     `yaml:"files"`
	CommitMetadata string
}

// SortedFiles returns the files sorted by path
func (d *DepFiles) SortedFiles() []*file.File {
	// We could sort them here if needed
	return d.Files
}

// LockEntry creates a lock entry for this dependency
func (d *DepFiles) LockEntry(fileHandler file.Handler) (*lock.Dep, error) {
	digest, err := fileHandler.CalculateDigest(d.Files)
	if err != nil {
		return nil, errors.Errorf("calculating digest: %w", err)
	}

	return &lock.Dep{
		Metadata: lock.LockDepMetadata{
			Type:   d.DepInfo.Type,
			Commit: d.CommitMetadata,
		},
		Repo:   d.DepInfo.Repo,
		Path:   d.DepInfo.Path,
		Ref:    d.DepInfo.Ref,
		Digest: digest,
	}, nil
}

// WriteToDir writes all files to a directory
func (d *DepFiles) WriteToDir(fileHandler file.Handler, relPath string) error {
	return fileHandler.WriteFiles(d.Files, relPath)
}

// AddFile adds a file to the dependency files
func (d *DepFiles) AddFile(fileHandler file.Handler, path string, filePath string) error {
	content, err := fileHandler.ReadFile(filepath.Join(path, filePath))
	if err != nil {
		return errors.Errorf("reading file: %w", err)
	}

	d.Files = append(d.Files, &file.File{
		Path:    filePath,
		Content: content,
	})

	return nil
}

// AddAllNestedProtoFiles adds all proto files matching filters to the dependency files
func (d *DepFiles) AddAllNestedProtoFiles(ctx context.Context, fileHandler file.Handler, path string, filters ...string) error {
	files, err := fileHandler.FindProtoFiles(path, filters)
	if err != nil {
		return errors.Errorf("finding proto files: %w", err)
	}

	for _, filePath := range files {
		zerolog.Ctx(ctx).Info().Str("file", filePath).Str("path", path).Msg("adding file")

		if err := d.AddFile(fileHandler, path, filePath); err != nil {
			return errors.Errorf("adding file: %w", err)
		}
	}

	slices.SortFunc(d.Files, func(a, b *file.File) int {
		return strings.Compare(a.Path, b.Path)
	})

	return nil
}

// Manager provides an interface for managing dependencies
type Manager interface {
	ProcessDependencies(ctx context.Context, config *config.Config, lockFile *lock.File, outputPath string) error
	CheckLocalDependency(ctx context.Context, cfg *config.Config, dep config.Buf3pdDep) (*DepFiles, bool, error)
	FetchRemoteDependency(ctx context.Context, dep config.Buf3pdDep) (*DepFiles, error)
}

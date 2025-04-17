package deps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/walteh/buf3pd/pkg/config"
	"github.com/walteh/buf3pd/pkg/file"
	"gitlab.com/tozd/go/errors"
)

// NewDepFilesFromLocal creates a DepFiles from a local directory
func NewDepFilesFromLocal(
	ctx context.Context,
	cfg *config.Config,
	dep config.Buf3pdDep,
	fileHandler file.Handler,
) (*DepFiles, bool, error) {

	pth := filepath.Join(cfg.Path, filepath.Base(dep.Repo))

	zerolog.Ctx(ctx).Info().Str("path", pth).Msg("processing local dependency")

	// Check if directory exists
	if _, err := os.Stat(pth); os.IsNotExist(err) {
		zerolog.Ctx(ctx).Warn().Str("path", pth).Msg("directory does not exist, skipping local dependency")
		return nil, false, nil
	}

	// Find all proto files in the directory
	protoFiles, err := fileHandler.FindProtoFiles(pth, dep.Filter)
	if err != nil {
		return nil, false, errors.Errorf("finding proto files: %w", err)
	}

	if len(protoFiles) == 0 {
		zerolog.Ctx(ctx).Warn().Str("path", pth).Msg("no proto files found, skipping local dependency")
		return nil, false, nil
	}

	depFiles := &DepFiles{
		DepInfo: dep,
		Files:   []*file.File{},
	}

	for _, filePath := range protoFiles {
		if err := depFiles.AddFile(fileHandler, pth, filePath); err != nil {
			return nil, false, errors.Errorf("adding file: %w", err)
		}
	}

	return depFiles, true, nil
}

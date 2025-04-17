package deps

import (
	"context"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/walteh/buf3pd/pkg/config"
	"github.com/walteh/buf3pd/pkg/file"
	"github.com/walteh/buf3pd/pkg/git"
	"github.com/walteh/buf3pd/pkg/lock"
	"gitlab.com/tozd/go/errors"
)

// DependencyManager implements the Manager interface
type DependencyManager struct {
	fileHandler file.Handler
	gitHandler  git.Handler
	lockManager lock.Manager
}

// NewDependencyManager creates a new DependencyManager
func NewDependencyManager(fileHandler file.Handler, gitHandler git.Handler, lockManager lock.Manager) *DependencyManager {
	return &DependencyManager{
		fileHandler: fileHandler,
		gitHandler:  gitHandler,
		lockManager: lockManager,
	}
}

// ProcessDependencies processes all dependencies in the configuration
func (m *DependencyManager) ProcessDependencies(
	ctx context.Context,
	config *config.Config,
	lockFile *lock.File,
	outputPath string,
) error {
	log := zerolog.Ctx(ctx)
	depFilesToUpdate := []*DepFiles{}

	for _, dep := range config.Deps {
		if dep.Type != "git" {
			log.Warn().Str("type", dep.Type).Msg("unsupported dependency type, skipping")
			continue
		}

		storedLockDep := m.lockManager.EntryFor(lockFile, dep)

		var ok bool
		var tryLoc *DepFiles
		var err error

		// Check if dependency is already processed locally
		tryLoc, ok, err = m.CheckLocalDependency(ctx, config, dep)
		if err != nil {
			return errors.Errorf("checking local dependency: %w", err)
		}

		var depFiles *DepFiles
		var lockDep *lock.Dep
		var skipRemote = false

		if ok {
			// Local dependency found
			realLockDep, err := tryLoc.LockEntry(m.fileHandler)
			if err != nil {
				return errors.Errorf("creating lock entry: %w", err)
			}

			if storedLockDep != nil && storedLockDep.Compare(realLockDep) {
				log.Info().Str("repo", dep.Repo).Str("path", dep.Path).Str("ref", dep.Ref).Msg("dependency already processed")
				skipRemote = true
				realLockDep.Metadata = storedLockDep.Metadata
			} else {
				log.Warn().Any("storedLockDep", storedLockDep).Any("realLockDep", realLockDep).Msg("dependency already processed, but with different commit")
			}

			log.Info().Str("repo", dep.Repo).Str("path", dep.Path).Str("ref", dep.Ref).Msg("using local dependency")
			depFiles = tryLoc
			lockDep = realLockDep
		}

		if !skipRemote {
			// No local dependency, fetch from remote
			log.Info().Str("repo", dep.Repo).Str("path", dep.Path).Str("ref", dep.Ref).Msg("processing git dependency from remote")

			remoteDepFiles, err := m.FetchRemoteDependency(ctx, dep)
			if err != nil {
				return errors.Errorf("fetching remote dependency: %w", err)
			}

			remoteLockDep, err := remoteDepFiles.LockEntry(m.fileHandler)
			if err != nil {
				return errors.Errorf("creating lock entry: %w", err)
			}

			depFiles = remoteDepFiles
			lockDep = remoteLockDep
		}

		// Update lock file
		if storedLockDep != nil {
			*storedLockDep = *lockDep
		} else {
			lockFile.Deps = append(lockFile.Deps, lockDep)
		}

		depFilesToUpdate = append(depFilesToUpdate, depFiles)

		log.Info().Str("repo", dep.Repo).Str("prefix", lockDep.Prefix).Msg("successfully processed dependency")
	}

	// Write updated dependencies to output directory
	for _, depFiles := range depFilesToUpdate {
		err := depFiles.WriteToDir(m.fileHandler, filepath.Join(outputPath, filepath.Base(depFiles.DepInfo.Repo)))
		if err != nil {
			return errors.Errorf("writing dependency files: %w", err)
		}
	}

	return nil
}

// CheckLocalDependency checks if a dependency exists locally
func (m *DependencyManager) CheckLocalDependency(
	ctx context.Context,
	cfg *config.Config,
	dep config.Buf3pdDep,
) (*DepFiles, bool, error) {
	return NewDepFilesFromLocal(ctx, cfg, dep, m.fileHandler)
}

// FetchRemoteDependency fetches a dependency from a remote repository
func (m *DependencyManager) FetchRemoteDependency(
	ctx context.Context,
	dep config.Buf3pdDep,
) (*DepFiles, error) {
	return NewDepFilesFromRemote(ctx, dep, m.fileHandler, m.gitHandler)
}

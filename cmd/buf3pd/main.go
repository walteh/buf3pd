package main

import (
	"context"
	"flag"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/walteh/buf3pd/pkg/config"
	"github.com/walteh/buf3pd/pkg/deps"
	"github.com/walteh/buf3pd/pkg/file"
	"github.com/walteh/buf3pd/pkg/git"
	"github.com/walteh/buf3pd/pkg/lock"
	"gitlab.com/tozd/go/errors"
)

// Version will be set during build
var Version = "dev"

func main() {
	ctx := context.Background()
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	ctx = logger.WithContext(ctx)

	log := zerolog.Ctx(ctx)
	log.Info().Str("version", Version).Msg("starting buf3pd")

	// Parse command line flags
	var (
		bufYamlPath = flag.String("config", "buf.yaml", "Path to buf.yaml file")
		workDir     = flag.String("workdir", ".", "Working directory")
	)
	flag.Parse()

	// Ensure workDir is absolute
	absWorkDir, err := filepath.Abs(*workDir)
	if err != nil {
		log.Fatal().Err(errors.Errorf("resolving absolute path for workdir: %w", err)).Msg("failed to start")
	}

	// Initialize managers
	configReader := config.NewFileReader()
	fileManager := file.NewManager()
	gitManager := git.NewManager()
	lockManager := lock.NewFileManager()
	dependencyManager := deps.NewDependencyManager(fileManager, gitManager, lockManager)

	// Read config
	cfg, err := configReader.ReadConfig(ctx, filepath.Join(absWorkDir, *bufYamlPath))
	if err != nil {
		log.Fatal().Err(errors.Errorf("reading buf3pd config: %w", err)).Msg("failed to read buf3pd config")
	}

	// Read lock file
	lockFile, err := lockManager.ReadLockFile(filepath.Join(absWorkDir, "buf3pd.lock"))
	if err != nil {
		log.Fatal().Err(errors.Errorf("reading lock file: %w", err)).Msg("failed to read lock file")
	}

	// Create the output directory if it doesn't exist
	outputPath := filepath.Join(absWorkDir, cfg.Path)
	if err := config.ValidatePath(outputPath); err != nil {
		log.Fatal().Err(errors.Errorf("validating output path: %w", err)).Msg("failed to validate output path")
	}

	// Process dependencies
	if err := dependencyManager.ProcessDependencies(ctx, cfg, lockFile, outputPath); err != nil {
		log.Fatal().Err(errors.Errorf("processing dependencies: %w", err)).Msg("failed to process dependencies")
	}

	// Write lock file
	if err := lockManager.WriteLockFile(lockFile, filepath.Join(absWorkDir, "buf3pd.lock")); err != nil {
		log.Fatal().Err(errors.Errorf("writing lock file: %w", err)).Msg("failed to write lock file")
	}

	log.Info().Str("path", filepath.Join(absWorkDir, "buf3pd.lock")).Msg("created lock file")
	log.Info().Msg("buf3pd completed successfully")
}

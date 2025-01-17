package faucet

import (
	"context"

	"github.com/CoreumFoundation/coreum-tools/pkg/build"
	"github.com/CoreumFoundation/crust/build/git"
	"github.com/CoreumFoundation/crust/build/golang"
)

const (
	repoURL          = "https://github.com/CoreumFoundation/faucet.git"
	repoPath         = "../faucet"
	dockerBinaryPath = "bin/.cache/docker/faucet/faucet"
	testBinaryPath   = "bin/.cache/integration-tests/faucet"
)

// Build builds faucet in docker.
func Build(ctx context.Context, deps build.DepsFunc) error {
	deps(golang.EnsureGo, ensureRepo)

	return golang.BuildInDocker(ctx, golang.BinaryBuildConfig{
		PackagePath:   "../faucet",
		BinOutputPath: dockerBinaryPath,
	})
}

// BuildIntegrationTests builds faucet integration tests.
func BuildIntegrationTests(ctx context.Context, deps build.DepsFunc) error {
	deps(golang.EnsureGo, ensureRepo)

	return golang.BuildTests(ctx, golang.TestBuildConfig{
		PackagePath:   "../faucet/integration-tests",
		BinOutputPath: testBinaryPath,
		Tags:          []string{"integrationtests"},
	})
}

// Tidy runs `go mod tidy` for faucet repo.
func Tidy(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureRepo)
	return golang.Tidy(ctx, repoPath, deps)
}

// Lint lints faucet repo.
func Lint(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureRepo)
	return golang.Lint(ctx, repoPath, deps)
}

// Test run unit tests in faucet repo.
func Test(ctx context.Context, deps build.DepsFunc) error {
	deps(ensureRepo)
	return golang.Test(ctx, repoPath, deps)
}

func ensureRepo(ctx context.Context, deps build.DepsFunc) error {
	return git.EnsureRepo(ctx, repoURL)
}

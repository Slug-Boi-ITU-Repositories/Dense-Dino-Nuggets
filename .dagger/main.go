// A generated module for Ddn functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

// Credit to the dagger team for the great documentation and function templates

package main

import (
	"context"
	"dagger/ddn/internal/dagger"
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"
)

type Ddn struct{}

// Returns a container that echoes whatever string argument is provided

func (m *Ddn) BuildEnv(src *dagger.Directory) *dagger.Container {
	return dag.Container().
		From("golang:1.25").
		WithDirectory("./src", src).
		WithWorkdir("./src").WithExec([]string{"go", "mod", "tidy"})
}

func (m *Ddn) Build(src *dagger.Directory) *dagger.Directory {

	// define build matrix
	gooses := []string{"linux", "darwin"}
	goarches := []string{"amd64", "arm64"}

	// create empty directory to put build artifacts
	outputs := dag.Directory()

	golang := m.BuildEnv(src)

	for _, goos := range gooses {
		for _, goarch := range goarches {
			// create directory for each OS and architecture
			path := fmt.Sprintf("build/%s-%s/", goos, goarch)

			// build artifact
			build := golang.
				WithEnvVariable("GOOS", goos).
				WithEnvVariable("GOARCH", goarch).
				WithWorkdir("./src").
				WithExec([]string{"go", "build", "-o", path + "/minitwit"})

			// add build to outputs
			outputs = outputs.
				WithDirectory(path, build.Directory(path)).
				WithDirectory(path+"/static", src.Directory("static")).
				WithDirectory(path+"/templates", src.Directory("templates"))
		}
	}

	return outputs
}

func (m *Ddn) Test(ctx context.Context, src *dagger.Directory) (string, error) {
	return m.BuildEnv(src).WithWorkdir("./src").WithExec([]string{"go", "test", "./..."}).Stdout(ctx)
}

func (m *Ddn) Lint(ctx context.Context, src *dagger.Directory) (string, error) {
	return dag.Container().From("golangci/golangci-lint:latest-alpine").
		WithDirectory("./src", src).
		WithWorkdir("./src/src").
		WithExec([]string{"golangci-lint", "run"}).Stdout(ctx)
}

// This includes things like linters and so on when we get that far
// The idea of this function is more so for local testing of the workflow idealy we would create a threaded workflow
// or maybe even multiple workflows that handle each part to make it more easy to see what went wrong
func (m *Ddn) RunAllTests(ctx context.Context, src *dagger.Directory) error {
	// Create error group
	eg, gctx := errgroup.WithContext(ctx)

	// Run linter
	eg.Go(func() error {
		_, err := m.Lint(gctx, src)
		return err
	})

	// Run unit tests
	eg.Go(func() error {
		_, err := m.Test(gctx, src)
		return err
	})

	// Wait for all tests to complete
	// If any test fails, the error will be returned
	return eg.Wait()
}

func (m *Ddn) Publish(ctx context.Context, src *dagger.Directory) (string, error) {
    // First, build the container using your Dockerfile
    container := src.DockerBuild()
    
    // Get Docker Hub credentials from environment variables
    username := os.Getenv("DOCKER_USERNAME")
    password := os.Getenv("DOCKER_PASSWORD")

    if username == "" || password == "" {
        return "", fmt.Errorf("DOCKER_USERNAME and DOCKER_PASSWORD must be set")
    }
    
    // Configure Docker registry authentication
    registry := "docker.io" // Default Docker Hub registry
    auth := dag.SetSecret("DOCKER_PASSWORD", password)
    
    // Tag the container with a version (you might want to make this configurable)
    tag := "latest" // Consider passing this as a parameter or using git commit hash
    imageRef := fmt.Sprintf("%s/%s:%s", username, "minitwitimage", tag)
    
    // Publish the container to Docker Hub
    publishedRef, err := container.
        WithRegistryAuth(registry, username, auth).
        Publish(ctx, imageRef)
    
    if err != nil {
        return "", fmt.Errorf("failed to publish container: %w", err)
    }
    
    return publishedRef, nil
}
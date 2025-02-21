//go:build mage
// +build mage

// Magefile contains the build, dependency management, and utility tasks for the project.
// Run `mage` to list available targets.

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(Deps)
	fmt.Println("Building...")

	if _, err := os.Stat("bin"); os.IsNotExist(err) {
		os.Mkdir("bin", 0755)
	}

	return sh.Run("go", "build", "-o", "bin/todo", "./cmd/todo")
}

// Manage your deps, or running package managers.
func Deps() error {
	fmt.Println("Installing Deps...")
	err := sh.Run("go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("failed to tidy", err)
	}
	return sh.Run("go", "mod", "download")
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	err := os.RemoveAll("bin/")
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error while cleaning: %v\n", err)
	}
}

// Lint runs golangci-lint to analyze the project code.
func Lint() error {
	fmt.Println("Running golangci-lint...")
	return sh.RunV("golangci-lint", "run", "./...")
}

type Install mg.Namespace

// GolangciLint installs golangci-lint. Install snap
func (i Install) Linter() error {
	fmt.Println("Ensuring golangci-lint is installed...")
	// Check if golangci-lint is available
	_, err := sh.Output("which", "golangci-lint")
	if err != nil {
		fmt.Println("golangci-lint is not installed. Installing...")
		return sh.Run("sudo", "snap", "install", "golangci-lint", "--classic")
	}
	return nil
}

// SQLC installs slqc.
func (i Install) SQLC() error {
	fmt.Println("Ensuring sqlc is installed...")
	// Check if sqlc is available
	if err := sh.Run("which", "sqlc"); err != nil {
		fmt.Println("sqlc is not installed. Installing...")
		// Install sqlc using the updated module path
		return sh.Run("go", "install", "github.com/sqlc-dev/sqlc/cmd/sqlc@latest")
	}
	return nil
}

// SQLC namespace for SQL code generation tasks
type SQLC mg.Namespace

// Gen runs SQLC code generation for all queries.
func (SQLC) Gen() error {
	fmt.Println("Running SQLC for all queries...")
	mg.Deps(Install.SQLC)

	// Run sqlc generate for everything
	return sh.RunV("sqlc", "generate", "--file=internal/core/sqlc/sqlc.json")
}

type Docker mg.Namespace

// Build builds the Docker image for the application.
func (Docker) BuildWSL() error {
	version := getVersion()
	imageName := "golang-todo-app"

	fmt.Printf("Building Docker image %s:%s...\n", imageName, version)

	// Build the versioned image
	err := sh.RunV("docker", "build",
		"--build-arg", "GOOS=linux",
		"--build-arg", "GOARCH=amd64",
		"-t", fmt.Sprintf("%s:%s", imageName, version), ".")
	if err != nil {
		return err
	}

	// Tag the image as latest
	fmt.Println("Tagging image with latest...")
	return sh.RunV("docker", "tag",
		fmt.Sprintf("%s:%s", imageName, version),
		fmt.Sprintf("%s:latest", imageName))
}

// BuildPi builds the Docker image for Raspberry Pi (arm64 architecture) with versioning
func (Docker) BuildPi() error {
	version := getVersion()
	imageName := "golang-todo-app-pi"

	fmt.Printf("Building Docker image %s:%s for Raspberry Pi...\n", imageName, version)

	// Build the versioned image
	err := sh.RunV("docker", "build",
		"--build-arg", "GOOS=linux",
		"--build-arg", "GOARCH=arm64",
		"-t", fmt.Sprintf("%s:%s", imageName, version), ".")
	if err != nil {
		return err
	}

	// Tag the image as latest
	fmt.Println("Tagging image with latest...")
	return sh.RunV("docker", "tag",
		fmt.Sprintf("%s:%s", imageName, version),
		fmt.Sprintf("%s:latest", imageName))
}

// Run runs the application container along with the database using Docker Compose.
func (Docker) Run() error {
	fmt.Println("Starting Docker containers with Docker Compose...")
	return sh.RunV("docker-compose", "up", "--build", "-d")
}

// Stop stops and removes the Docker containers.
func (Docker) Stop() error {
	fmt.Println("Stopping and removing Docker containers...")
	return sh.RunV("docker-compose", "down")
}

// Logs displays logs from the application container.
func (Docker) Logs() error {
	fmt.Println("Displaying logs from the application container...")
	return sh.RunV("docker", "logs", "-f", "golang-todo-app")
}

// getVersion gets the version from an environment variable or defaults to "latest"
func getVersion() string {
	version := os.Getenv("VERSION")
	if version == "" {
		version = "latest"
	}
	return version
}

package main

import (
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

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

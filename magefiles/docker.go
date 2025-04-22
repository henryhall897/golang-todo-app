package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

// BuildPi builds, tags, and pushes the Docker image for Raspberry Pi (arm64 architecture) to Docker Hub
func (Docker) BuildPi() error {
	version := getVersion()

	// Retrieve the Docker Hub username from Kubernetes Secret
	cmd := exec.Command("kubectl", "get", "secret", "golang-todo-secret", "-o", "jsonpath={.data.DOCKER_HUB_USERNAME}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to fetch DOCKER_HUB_USERNAME from Kubernetes: %v", err)
	}

	// Decode base64 output
	usernameBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to decode DOCKER_HUB_USERNAME: %v", err)
	}
	dockerHubUsername := string(usernameBytes)

	if dockerHubUsername == "" {
		fmt.Println("ERROR: DOCKER_HUB_USERNAME not found in Kubernetes Secret.")
		return fmt.Errorf("DOCKER_HUB_USERNAME is required")
	}

	imageName := "golang-todo-app-pi"
	fullImageName := fmt.Sprintf("docker.io/%s/%s", dockerHubUsername, imageName)

	// Logging
	fmt.Printf("Building Docker image: %s:%s\n", fullImageName, version)
	fmt.Println("Clearing Docker cache...")

	// Ensure old image layers are fully cleared
	err = sh.RunV("docker", "system", "prune", "-af")
	if err != nil {
		return fmt.Errorf("failed to prune Docker cache: %v", err)
	}

	// Build the versioned image & push directly
	fmt.Println("Building and pushing new image...")
	err = sh.RunV("docker", "buildx", "build",
		"--no-cache",
		"--pull", // Ensure latest base image is used
		"--push", // Push directly to Docker Hub
		"--platform", "linux/arm64",
		"-t", fmt.Sprintf("%s:%s", fullImageName, version), ".")
	if err != nil {
		return fmt.Errorf("failed to build and push image: %v", err)
	}

	fmt.Println("Successfully built and pushed:", fullImageName, version)

	return nil
}

// getVersion gets the version from an environment variable or defaults to "latest"
func getVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "latest"
	}
	return version
}

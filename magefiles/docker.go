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

	// Retrieve the Docker Hub username from the Kubernetes Secret
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

	fmt.Printf("Building Docker image %s:%s for Raspberry Pi...\n", fullImageName, version)

	// Build the versioned image
	err = sh.RunV("docker", "buildx", "build",
		"--build-arg", "GOOS=linux",
		"--build-arg", "GOARCH=arm64",
		"-t", fmt.Sprintf("%s:%s", fullImageName, version), ".")
	if err != nil {
		return fmt.Errorf("failed to build image: %v", err)
	}

	// Tag the image as latest
	fmt.Println("Tagging image with latest...")
	err = sh.RunV("docker", "tag",
		fmt.Sprintf("%s:%s", fullImageName, version),
		fmt.Sprintf("%s:latest", fullImageName))
	if err != nil {
		return fmt.Errorf("failed to tag image: %v", err)
	}

	// Push the versioned and latest tags to Docker Hub
	fmt.Println("Pushing image to Docker Hub...")
	err = sh.RunV("docker", "push", fmt.Sprintf("%s:%s", fullImageName, version))
	if err != nil {
		return fmt.Errorf("failed to push versioned image: %v", err)
	}

	err = sh.RunV("docker", "push", fmt.Sprintf("%s:latest", fullImageName))
	if err != nil {
		return fmt.Errorf("failed to push latest image: %v", err)
	}

	fmt.Println("Image built successfully and pushed to Docker Hub.")
	return nil
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

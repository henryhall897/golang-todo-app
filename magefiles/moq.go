//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/magefile/mage/mg"
)

// Moq namespace for mock generation tasks
type Moq mg.Namespace

// Generate generates mocks for interfaces using moq.
func (Moq) Generate() error {
	fmt.Println("Generating mocks using moq...")

	// Get the root directory of the project
	rootDir, err := getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %w", err)
	}

	// Define the list of interfaces to mock
	mockTargets := []struct {
		Interface string
		Package   string
		Output    string
	}{
		{"UserRepository", "../internal/users", filepath.Join(rootDir, "internal/users/mocks/user_repo_mock.go")},
		{"UserService", "../internal/users", filepath.Join(rootDir, "internal/users/mocks/user_service_mock.go")},
	}

	for _, target := range mockTargets {
		cmd := exec.Command("moq", "-out", target.Output, "-pkg", "mocks", target.Package, target.Interface)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to generate mock for %s: %w", target.Interface, err)
		}
	}

	fmt.Println("Mocks generated successfully.")
	return nil
}

// getProjectRoot finds the root directory of the project.
func getProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Ensure we are in the project root
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent directories")
		}
		dir = parent
	}
}

//go:build mage
// +build mage

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

	return sh.Run("go", "build", "-o", "bin/todo", "./cmd/todo")
}

// Manage your deps, or running package managers.
func Deps() error {
	fmt.Println("Installing Deps...")
	err := sh.Run("go", "mod", "tidy")
	if err != nil {
		fmt.Println("failed to tidy")
	}
	return sh.Run("go", "mod", "download")
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("bin/")
}

// Lint runs golangci-lint to analyze the project code.
func Lint() error {
	fmt.Println("Running golangci-lint...")
	return sh.RunV("golangci-lint", "run", "./...")
}

type Install mg.Namespace

// GolangciLint installs golangci-lint.
func (i Install) Linter() error {
	fmt.Println("Ensuring golangci-lint is installed...")
	// Check if golangci-lint is available
	if err := sh.Run("which", "golangci-lint"); err != nil {
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

type SQLC mg.Namespace

// GenerateSQLC runs sqlc code generation for the users folder.
func (s SQLC) Users() error {
	fmt.Println("Running sqlc for the users folder...")
	// Ensure sqlc is installed
	mg.Deps(Install.SQLC)

	// Run sqlc generate for the users folder
	return sh.RunV("sqlc", "generate", "-f", "internal/users/sqlc.json")
}

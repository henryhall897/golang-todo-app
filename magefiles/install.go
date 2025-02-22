//go:build mage
// +build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

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

//go:build mage

package main

import (
	"fmt"
	"os" // mg contains helpful utility functions, like Deps

	"github.com/magefile/mage/sh"
)

// Deps manages dependencies.
func Deps() error {
	fmt.Println("Installing Deps...")
	err := sh.Run("go", "mod", "tidy")
	if err != nil {
		return fmt.Errorf("failed to tidy: %w", err)
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

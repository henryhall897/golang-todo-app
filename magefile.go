//go:build mage
// +build mage

// Magefile contains the build, dependency management, and utility tasks for the project.
// Run `mage` to list available targets.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
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

type SQLC mg.Namespace

// GenerateSQLC runs sqlc code generation for the users folder.
func (s SQLC) Users() error {
	fmt.Println("Running sqlc for the users folder...")
	// Ensure sqlc is installed
	mg.Deps(Install.SQLC)

	// Run sqlc generate for the users folder
	if _, err := os.Stat("internal/users/sqlc.json"); os.IsNotExist(err) {
		return fmt.Errorf("sqlc.json not found in internal/users: %w", err)
	}
	return sh.RunV("sqlc", "generate", "-f", "internal/users/sqlc.json")

}

// GenerateTodoListSQLC runs sqlc code generation for the todo_list folder.
func (s SQLC) TodoList() error {
	fmt.Println("Running sqlc for the todo_list folder...")
	// Ensure sqlc is installed
	mg.Deps(Install.SQLC)

	// Check if the sqlc.json file exists in the todo_list folder
	if _, err := os.Stat("internal/todo_list/sqlc.json"); os.IsNotExist(err) {
		return fmt.Errorf("sqlc.json not found in internal/todo_list: %w", err)
	}
	// Run sqlc generate for the todo_list folder
	return sh.RunV("sqlc", "generate", "-f", "internal/todo_list/sqlc.json")
}

// GenerateTasksSQLC runs sqlc code generation for the tasks folder.
func (s SQLC) Tasks() error {
	fmt.Println("Running sqlc for the tasks folder...")
	// Ensure sqlc is installed
	mg.Deps(Install.SQLC)

	// Check if the sqlc.json file exists in the tasks folder
	if _, err := os.Stat("internal/tasks/sqlc.json"); os.IsNotExist(err) {
		return fmt.Errorf("sqlc.json not found in internal/tasks: %w", err)
	}
	// Run sqlc generate for the tasks folder
	return sh.RunV("sqlc", "generate", "-f", "internal/tasks/sqlc.json")
}

// LoadEnv loads the .env file.
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, falling back to system environment variables.")
	}
}

type Migrate mg.Namespace

// Up applies all up migrations.
func (m Migrate) Up() error {
	LoadEnv() // Load environment variables
	fmt.Println("Running up migrations...")
	return sh.RunV("migrate", "-path", "migrations", "-database", os.Getenv("DATABASE_URL"), "up")
}

// Down rolls back the last migration.
func (m Migrate) Down() error {
	LoadEnv() // Load environment variables
	fmt.Println("Rolling back the last migration...")
	return sh.RunV("migrate", "-path", "migrations", "-database", os.Getenv("DATABASE_URL"), "down")
}

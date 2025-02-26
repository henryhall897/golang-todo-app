package main

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Gen mg.Namespace

// Generate runs `go generate` for only mocks (moq).
// runs Gen for the application
func (Gen) Generate() error {
	mg.SerialDeps(Gen.Clean)
	if err := sh.Run("sqlc", "generate"); err != nil {
		return err
	}

	return sh.Run("go", "generate", "./...")
}

// CleanGen removes all generated files but keeps `generate.go` files.
func (Gen) Clean() error {
	fmt.Println("Removing generated code in `gen/`")

	if err := sh.Rm("./gen"); err != nil {
		fmt.Printf("Failed to remove ./gen: %v\n", err)
		return err // Return the error to stop execution if needed
	}

	return nil
}

// ensures that no generated files are changed after a fresh generation
func (Gen) Verify() error {
	mg.Deps(Gen.Generate)

	dirtyFiles, err := sh.Output("git", "status", "--porcelain")
	if err != nil {
		return err
	}

	dirtyGenFiles := make([]string, 0)
	for _, dirtyFile := range strings.Split(dirtyFiles, "\n") {
		if strings.Contains(dirtyFile, "gen/") {
			dirtyGenFiles = append(dirtyGenFiles, dirtyFile)
		}
	}

	if len(dirtyGenFiles) > 0 {
		err := sh.RunV("git", "--no-pager", "diff", "gen")
		if err != nil {
			return err
		}

		err = sh.RunV("git", "--no-pager", "diff", "gen")
		if err != nil {
			return err
		}

		return fmt.Errorf("verification failed")
	}

	return nil
}

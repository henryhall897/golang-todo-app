//go:build mage
// +build mage

package main

import (
	"fmt"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// SQLC namespace for SQL code generation tasks
type SQLC mg.Namespace

// Gen runs SQLC code generation for all queries.
func (SQLC) Gen() error {
	fmt.Println("Running SQLC for all queries...")
	mg.Deps(Install.SQLC)

	// Run sqlc generate for everything
	return sh.RunV("sqlc", "generate", "--file=internal/core/sqlc/sqlc.json")
}

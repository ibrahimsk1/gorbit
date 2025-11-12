// Package internal contains contract tests for linting configuration.
// These tests verify that linting tools are properly configured.
//
// Labels: scope:contract loop:g0-work layer:infra
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestLintingConfiguration verifies that linting configuration files exist.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestLintingConfiguration(t *testing.T) {
	repoRoot := findRepoRoot(t)

	t.Run(".golangci.yml exists", func(t *testing.T) {
		golangciPath := filepath.Join(repoRoot, ".golangci.yml")
		if _, err := os.Stat(golangciPath); os.IsNotExist(err) {
			t.Fatalf(".golangci.yml does not exist at %s", golangciPath)
		}
	})

	t.Run("client eslint.config.js exists", func(t *testing.T) {
		eslintPath := filepath.Join(repoRoot, "client", "eslint.config.js")
		if _, err := os.Stat(eslintPath); os.IsNotExist(err) {
			t.Fatalf("client/eslint.config.js does not exist at %s", eslintPath)
		}
	})

	t.Run("client .prettierrc exists", func(t *testing.T) {
		prettierPath := filepath.Join(repoRoot, "client", ".prettierrc")
		if _, err := os.Stat(prettierPath); os.IsNotExist(err) {
			t.Fatalf("client/.prettierrc does not exist at %s", prettierPath)
		}
	})
}

// TestLintingToolsAvailable verifies that linting tools are available.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestLintingToolsAvailable(t *testing.T) {
	// Check golangci-lint
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = filepath.Join(os.Getenv("HOME"), "go")
		}
		golangciPath := filepath.Join(gopath, "bin", "golangci-lint")
		if _, err := os.Stat(golangciPath); os.IsNotExist(err) {
			t.Skip("golangci-lint not available, skipping test")
		}
	}
}

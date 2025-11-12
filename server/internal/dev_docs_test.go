// Package internal contains contract tests for development documentation.
// These tests verify that documentation meets the requirements
// for G0 workspace bootstrap.
//
// Labels: scope:contract loop:g0-work layer:infra
package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDocumentationExists verifies that required documentation files exist.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestDocumentationExists(t *testing.T) {
	repoRoot := findRepoRoot(t)

	t.Run("README.md exists", func(t *testing.T) {
		readmePath := filepath.Join(repoRoot, "README.md")
		if _, err := os.Stat(readmePath); os.IsNotExist(err) {
			t.Fatalf("README.md does not exist at %s", readmePath)
		}
	})

	t.Run("server .env.example exists", func(t *testing.T) {
		envExamplePath := filepath.Join(repoRoot, "server", ".env.example")
		if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
			t.Fatalf("server/.env.example does not exist at %s", envExamplePath)
		}
	})

	t.Run("client .env.example exists", func(t *testing.T) {
		envExamplePath := filepath.Join(repoRoot, "client", ".env.example")
		if _, err := os.Stat(envExamplePath); os.IsNotExist(err) {
			t.Fatalf("client/.env.example does not exist at %s", envExamplePath)
		}
	})
}

// TestReadmeContent verifies that README contains required sections.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestReadmeContent(t *testing.T) {
	repoRoot := findRepoRoot(t)
	readmePath := filepath.Join(repoRoot, "README.md")

	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	readmeContent := strings.ToLower(string(content))
	requiredSections := []string{
		"prerequisites",
		"quick start",
		"development",
		"testing",
	}

	for _, section := range requiredSections {
		if !strings.Contains(readmeContent, section) {
			t.Errorf("README.md missing required section: %s", section)
		}
	}
}


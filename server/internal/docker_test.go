// Package internal contains contract tests for Docker setup.
// These tests verify that Docker configuration meets the requirements
// for G0 workspace bootstrap.
//
// Labels: scope:contract loop:g0-work layer:infra
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestDockerFilesExist verifies that all required Docker files exist.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestDockerFilesExist(t *testing.T) {
	repoRoot := findRepoRoot(t)

	t.Run("server Dockerfile exists", func(t *testing.T) {
		dockerfilePath := filepath.Join(repoRoot, "server", "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			t.Fatalf("server/Dockerfile does not exist at %s", dockerfilePath)
		}
	})

	t.Run("client Dockerfile exists", func(t *testing.T) {
		dockerfilePath := filepath.Join(repoRoot, "client", "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			t.Fatalf("client/Dockerfile does not exist at %s", dockerfilePath)
		}
	})

	t.Run("docker-compose.yml exists", func(t *testing.T) {
		composePath := filepath.Join(repoRoot, "docker-compose.yml")
		if _, err := os.Stat(composePath); os.IsNotExist(err) {
			t.Fatalf("docker-compose.yml does not exist at %s", composePath)
		}
	})

	t.Run("server .dockerignore exists", func(t *testing.T) {
		dockerignorePath := filepath.Join(repoRoot, "server", ".dockerignore")
		if _, err := os.Stat(dockerignorePath); os.IsNotExist(err) {
			t.Fatalf("server/.dockerignore does not exist at %s", dockerignorePath)
		}
	})

	t.Run("client .dockerignore exists", func(t *testing.T) {
		dockerignorePath := filepath.Join(repoRoot, "client", ".dockerignore")
		if _, err := os.Stat(dockerignorePath); os.IsNotExist(err) {
			t.Fatalf("client/.dockerignore does not exist at %s", dockerignorePath)
		}
	})
}

// TestDockerComposeValid verifies that docker-compose.yml is valid.
// This test checks if docker compose can parse the configuration.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestDockerComposeValid(t *testing.T) {
	repoRoot := findRepoRoot(t)

	// Check if docker compose is available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not available, skipping docker-compose validation")
	}

	cmd := exec.Command("docker", "compose", "config", "--quiet")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("docker compose config failed: %v\nOutput: %s", err, string(output))
	}
}

// findRepoRoot finds the repository root directory.
func findRepoRoot(t *testing.T) string {
	serverDir := findServerDir(t)
	return filepath.Dir(serverDir)
}

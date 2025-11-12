// Package internal contains contract tests for test harness configuration.
// These tests verify that test runners and CI are properly configured.
//
// Labels: scope:contract loop:g0-work layer:infra
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestGinkgoConfigured verifies that Ginkgo/Gomega are installed and configured.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestGinkgoConfigured(t *testing.T) {
	serverDir := findServerDir(t)

	// Check if ginkgo binary is available
	if _, err := exec.LookPath("ginkgo"); err != nil {
		// Try to find it in GOPATH/bin
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = filepath.Join(os.Getenv("HOME"), "go")
		}
		ginkgoPath := filepath.Join(gopath, "bin", "ginkgo")
		if _, err := os.Stat(ginkgoPath); os.IsNotExist(err) {
			t.Skip("ginkgo not available, skipping test")
		}
	}

	// Check if suite_test.go exists
	suitePath := filepath.Join(serverDir, "internal", "suite_test.go")
	if _, err := os.Stat(suitePath); os.IsNotExist(err) {
		t.Fatalf("Ginkgo suite_test.go does not exist at %s", suitePath)
	}
}

// TestCIConfiguration verifies that CI configuration exists.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestCIConfiguration(t *testing.T) {
	repoRoot := findRepoRoot(t)

	// Check for GitHub Actions
	ciPath := filepath.Join(repoRoot, ".github", "workflows", "ci.yml")
	if _, err := os.Stat(ciPath); os.IsNotExist(err) {
		t.Fatalf("CI configuration does not exist at %s", ciPath)
	}
}

// TestTestRunnersExecute verifies that test runners can execute.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestTestRunnersExecute(t *testing.T) {
	serverDir := findServerDir(t)

	// Test that go test works
	cmd := exec.Command("go", "test", "./...", "-v")
	cmd.Dir = serverDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed: %v\nOutput: %s", err, string(output))
	}
}

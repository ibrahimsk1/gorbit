// Package internal contains contract tests for the Go workspace setup.
// These tests verify that the workspace structure meets the requirements
// for G0 workspace bootstrap.
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// TestGoWorkspaceStructure verifies that the Go workspace has the required
// directory structure and files.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestGoWorkspaceStructure(t *testing.T) {
	serverDir := findServerDir(t)

	t.Run("go.mod exists", func(t *testing.T) {
		goModPath := filepath.Join(serverDir, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			t.Fatalf("go.mod does not exist at %s", goModPath)
		}
	})

	t.Run("cmd directory exists", func(t *testing.T) {
		cmdDir := filepath.Join(serverDir, "cmd")
		info, err := os.Stat(cmdDir)
		if os.IsNotExist(err) {
			t.Fatalf("cmd directory does not exist at %s", cmdDir)
		}
		if !info.IsDir() {
			t.Fatalf("cmd is not a directory at %s", cmdDir)
		}
	})

	t.Run("internal directory exists", func(t *testing.T) {
		internalDir := filepath.Join(serverDir, "internal")
		info, err := os.Stat(internalDir)
		if os.IsNotExist(err) {
			t.Fatalf("internal directory does not exist at %s", internalDir)
		}
		if !info.IsDir() {
			t.Fatalf("internal is not a directory at %s", internalDir)
		}
	})

	t.Run("Makefile exists", func(t *testing.T) {
		makefilePath := filepath.Join(serverDir, "Makefile")
		if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
			t.Fatalf("Makefile does not exist at %s", makefilePath)
		}
	})
}

// TestGoModTidy verifies that go mod tidy runs successfully,
// indicating the Go module is properly configured.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestGoModTidy(t *testing.T) {
	serverDir := findServerDir(t)

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = serverDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed: %v\nOutput: %s", err, string(output))
	}
}

// TestMakefileTargets verifies that the Makefile contains the required targets.
//
// Labels: scope:contract loop:g0-work layer:infra
func TestMakefileTargets(t *testing.T) {
	serverDir := findServerDir(t)
	makefilePath := filepath.Join(serverDir, "Makefile")

	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("Failed to read Makefile: %v", err)
	}

	makefileContent := string(content)
	requiredTargets := []string{
		"build:",
		"test:",
		"clean:",
		"tidy:",
	}

	for _, target := range requiredTargets {
		if !containsTarget(makefileContent, target) {
			t.Errorf("Makefile missing required target: %s", target)
		}
	}
}

// findServerDir finds the server directory relative to the test file.
func findServerDir(t *testing.T) string {
	// Get the directory where this test file is located
	_, testFile, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(testFile)

	// Navigate up to server directory (test is in server/internal/)
	dir := testDir
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			t.Fatalf("Could not find server directory from %s", testDir)
		}
		if filepath.Base(dir) == "server" {
			return dir
		}
		dir = parent
	}
}

// containsTarget checks if the Makefile content contains a target definition.
func containsTarget(content, target string) bool {
	// Look for target at start of line (with optional whitespace)
	lines := []rune(content)
	targetRunes := []rune(target)

	for i := 0; i <= len(lines)-len(targetRunes); i++ {
		// Check if we're at start of line (beginning or after newline)
		if i > 0 && lines[i-1] != '\n' {
			continue
		}

		// Check if target matches
		match := true
		for j := 0; j < len(targetRunes); j++ {
			if i+j >= len(lines) || lines[i+j] != targetRunes[j] {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

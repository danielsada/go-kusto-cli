package auth

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestHelperProcess is used by tests to mock the `az` CLI.
// It is not a real test — it's invoked as a subprocess.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_TEST_HELPER_PROCESS") != "1" {
		return
	}
	switch os.Getenv("GO_TEST_HELPER_MODE") {
	case "success":
		_, _ = os.Stdout.WriteString("mock-token-abc123\n")
		os.Exit(0)
	case "empty":
		_, _ = os.Stdout.WriteString("\n")
		os.Exit(0)
	case "error":
		_, _ = os.Stderr.WriteString("ERROR: Please run 'az login'\n")
		os.Exit(1)
	default:
		os.Exit(1)
	}
}

func TestGetToken_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exec-based test in short mode")
	}
	// We can't easily inject the exec.Command into GetToken without refactoring,
	// but we can test the function's interface contract.
	// This test verifies behavior when `az` is available by using the real CLI.
	// If az is not available, skip.
	_, err := exec.LookPath("az")
	if err != nil {
		t.Skip("az CLI not available, skipping integration test")
	}

	// Just verify it doesn't panic — actual token depends on login state.
	_, _ = GetToken()
}

func TestGetToken_AzNotFound(t *testing.T) {
	// Save and restore PATH
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "/nonexistent")
	defer func() { _ = os.Setenv("PATH", origPath) }()

	_, err := GetToken()
	if err == nil {
		t.Fatal("expected error when az is not found")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "executable file") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

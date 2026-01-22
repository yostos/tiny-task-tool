package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing.
// Returns the path to the repository and a cleanup function.
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "ttt-git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commits (errors are non-fatal for tests)
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	_ = cmd.Run()

	// Create initial commit
	testFile := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		_ = os.RemoveAll(dir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	_ = cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = dir
	_ = cmd.Run()

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}

	return dir, cleanup
}

// TestSetRemote verifies that SetRemote() adds a new remote named "origin".
// Spec: docs/specification.md "リモートリポジトリの登録（v0.3.0）" section
func TestSetRemote(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	url := "https://github.com/user/repo.git"
	err := SetRemote(dir, url)
	if err != nil {
		t.Fatalf("SetRemote() error: %v", err)
	}

	// Verify remote was set
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get remote URL: %v", err)
	}

	got := strings.TrimSpace(string(output))
	if got != url {
		t.Errorf("Remote URL = %q, want %q", got, url)
	}
}

// TestSetRemoteUpdate verifies that SetRemote() updates existing remote.
// Spec: docs/specification.md "origin が既に存在する場合は git remote set-url origin <url> で更新"
func TestSetRemoteUpdate(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Set initial remote
	oldURL := "https://github.com/old/repo.git"
	cmd := exec.Command("git", "remote", "add", "origin", oldURL)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add initial remote: %v", err)
	}

	// Update remote
	newURL := "https://github.com/new/repo.git"
	err := SetRemote(dir, newURL)
	if err != nil {
		t.Fatalf("SetRemote() error: %v", err)
	}

	// Verify remote was updated
	cmd = exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get remote URL: %v", err)
	}

	got := strings.TrimSpace(string(output))
	if got != newURL {
		t.Errorf("Remote URL = %q, want %q", got, newURL)
	}
}

// TestHasRemote verifies that HasRemote() correctly detects remote existence.
// Spec: docs/specification.md "リモートリポジトリの登録（v0.3.0）" section
func TestHasRemote(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Initially no remote
	if HasRemote(dir, "origin") {
		t.Error("HasRemote() = true, want false (no remote set)")
	}

	// Add remote
	cmd := exec.Command("git", "remote", "add", "origin", "https://example.com/repo.git")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to add remote: %v", err)
	}

	// Now should have remote
	if !HasRemote(dir, "origin") {
		t.Error("HasRemote() = false, want true (remote was set)")
	}
}

// TestGetCurrentBranch verifies that GetCurrentBranch() returns the current branch name.
// Spec: docs/specification.md "手動同期（v0.3.0）" section - sync uses current branch
func TestGetCurrentBranch(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	branch, err := GetCurrentBranch(dir)
	if err != nil {
		t.Fatalf("GetCurrentBranch() error: %v", err)
	}

	// Default branch is typically "main" or "master"
	if branch != "main" && branch != "master" {
		t.Errorf("GetCurrentBranch() = %q, want 'main' or 'master'", branch)
	}
}

// TestSyncNoRemote verifies that Sync() returns error when no remote is configured.
// Spec: docs/specification.md "リモートが未設定: Error: No remote 'origin' configured."
func TestSyncNoRemote(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	err := Sync(dir)
	if err == nil {
		t.Error("Sync() should return error when no remote is configured")
	}

	// Error message should mention 'origin'
	if err != nil && !strings.Contains(err.Error(), "origin") {
		t.Errorf("Error message should mention 'origin', got: %v", err)
	}
}

// TestSyncPullFailureSkipsToPush verifies that Sync() skips pull and proceeds to push
// when pull fails (e.g., remote branch doesn't exist yet).
// Spec: docs/specification.md "pull失敗（リモートにブランチなし等）: pull をスキップして commit → push を実行"
func TestSyncPullFailureSkipsToPush(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a bare remote repository (no branches yet)
	remoteDir, err := os.MkdirTemp("", "ttt-git-remote-*")
	if err != nil {
		t.Fatalf("Failed to create remote dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(remoteDir) }()

	cmd := exec.Command("git", "init", "--bare")
	cmd.Dir = remoteDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to init bare repo: %v", err)
	}

	// Add remote pointing to bare repo
	if err := SetRemote(dir, remoteDir); err != nil {
		t.Fatalf("SetRemote() error: %v", err)
	}

	// Sync should succeed (pull fails but push should work)
	err = Sync(dir)
	if err != nil {
		t.Errorf("Sync() should succeed on first sync, got error: %v", err)
	}
}

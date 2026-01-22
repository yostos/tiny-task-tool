// Package git provides git operations for ttt.
package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// SetRemote sets or updates the remote URL for origin.
// If origin already exists, it updates the URL using set-url.
func SetRemote(dir, url string) error {
	if HasRemote(dir, "origin") {
		// Update existing remote
		cmd := exec.Command("git", "remote", "set-url", "origin", url)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update remote: %w", err)
		}
	} else {
		// Add new remote
		cmd := exec.Command("git", "remote", "add", "origin", url)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add remote: %w", err)
		}
	}
	return nil
}

// HasRemote checks if a remote with the given name exists.
func HasRemote(dir, name string) bool {
	cmd := exec.Command("git", "remote", "get-url", name)
	cmd.Dir = dir
	return cmd.Run() == nil
}

// GetCurrentBranch returns the current branch name.
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Sync performs pull, commit (if needed), and push.
// Returns an error if no remote 'origin' is configured.
// If pull fails (e.g., remote branch doesn't exist), it skips pull and proceeds to push.
func Sync(dir string) error {
	// Check if remote exists
	if !HasRemote(dir, "origin") {
		return fmt.Errorf("no remote 'origin' configured. Use 'ttt remote <url>' first")
	}

	branch, err := GetCurrentBranch(dir)
	if err != nil {
		return err
	}

	// Pull from remote (skip if fails, e.g., remote branch doesn't exist yet)
	cmd := exec.Command("git", "pull", "origin", branch)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		outputStr := string(output)
		// Check for merge conflict - this is a real error
		if strings.Contains(outputStr, "CONFLICT") {
			return fmt.Errorf("merge conflict detected. Please resolve manually:\n%s", output)
		}
		// Other pull failures (e.g., remote ref not found) - skip and proceed to push
		// This handles the case of first sync when remote branch doesn't exist
	}

	// Check for uncommitted changes
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check status: %w", err)
	}

	// If there are changes, commit them
	if len(strings.TrimSpace(string(output))) > 0 {
		// Stage all changes
		cmd = exec.Command("git", "add", "-A")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stage changes: %w", err)
		}

		// Commit
		cmd = exec.Command("git", "commit", "-m", "Sync changes")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to commit: %w", err)
		}
	}

	// Push to remote
	cmd = exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("push failed: %s", output)
	}

	return nil
}

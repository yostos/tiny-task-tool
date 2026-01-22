package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEnsureRepoFilesCreatesReadme verifies that ensureRepoFiles creates README.md
// when it doesn't exist.
// Spec: docs/specification.md "リポジトリファイルの自動生成（v0.3.0）" section
func TestEnsureRepoFilesCreatesReadme(t *testing.T) {
	dir := t.TempDir()

	err := ensureRepoFiles(dir)
	if err != nil {
		t.Fatalf("ensureRepoFiles() error: %v", err)
	}

	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		t.Error("README.md was not created")
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	// Verify README.md contains required elements per spec
	requiredElements := []string{
		"My Tasks",
		"ttt",
		"tasks.md",
		"archive.md",
		"https://github.com/yostos/tiny-task-tool",
		"Created by ttt",
	}
	for _, elem := range requiredElements {
		if !strings.Contains(string(content), elem) {
			t.Errorf("README.md should contain %q", elem)
		}
	}
}

// TestEnsureRepoFilesCreatesGitignore verifies that ensureRepoFiles creates .gitignore
// when it doesn't exist.
// Spec: docs/specification.md "リポジトリファイルの自動生成（v0.3.0）" section
func TestEnsureRepoFilesCreatesGitignore(t *testing.T) {
	dir := t.TempDir()

	err := ensureRepoFiles(dir)
	if err != nil {
		t.Fatalf("ensureRepoFiles() error: %v", err)
	}

	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore was not created")
	}

	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	// Verify .gitignore contains OS and editor patterns per spec
	requiredPatterns := []string{
		// macOS
		".DS_Store",
		// Windows
		"Thumbs.db",
		// Linux/General
		"*~",
		// Vim
		"*.swp",
		// VS Code
		".vscode/",
	}
	for _, pattern := range requiredPatterns {
		if !strings.Contains(string(content), pattern) {
			t.Errorf(".gitignore should contain %q", pattern)
		}
	}
}

// TestEnsureRepoFilesDoesNotOverwrite verifies that ensureRepoFiles does not
// overwrite existing files.
// Spec: docs/specification.md "存在しない場合は自動生成"
func TestEnsureRepoFilesDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()

	// Create existing files with custom content
	readmePath := filepath.Join(dir, "README.md")
	customReadme := "# Custom README\n"
	if err := os.WriteFile(readmePath, []byte(customReadme), 0644); err != nil {
		t.Fatalf("Failed to create custom README.md: %v", err)
	}

	gitignorePath := filepath.Join(dir, ".gitignore")
	customGitignore := "# Custom gitignore\n"
	if err := os.WriteFile(gitignorePath, []byte(customGitignore), 0644); err != nil {
		t.Fatalf("Failed to create custom .gitignore: %v", err)
	}

	// Run ensureRepoFiles
	err := ensureRepoFiles(dir)
	if err != nil {
		t.Fatalf("ensureRepoFiles() error: %v", err)
	}

	// Verify files were not overwritten
	readmeContent, _ := os.ReadFile(readmePath)
	if string(readmeContent) != customReadme {
		t.Error("README.md was overwritten")
	}

	gitignoreContent, _ := os.ReadFile(gitignorePath)
	if string(gitignoreContent) != customGitignore {
		t.Error(".gitignore was overwritten")
	}
}

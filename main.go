package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yostos/tiny-task-tool/internal/cli"
	"github.com/yostos/tiny-task-tool/internal/config"
	"github.com/yostos/tiny-task-tool/internal/git"
	"github.com/yostos/tiny-task-tool/internal/tui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := cli.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if opts.ShowHelp {
		fmt.Println(cli.Usage())
		return nil
	}

	if opts.ShowVersion {
		fmt.Println(cli.VersionString())
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := ensureWorkingDir(cfg); err != nil {
		return err
	}

	// Handle subcommands
	if opts.RemoteURL != "" {
		return setRemote(cfg, opts.RemoteURL)
	}

	if opts.Sync {
		return syncTasks(cfg)
	}

	if opts.Task != "" {
		return addTask(cfg, opts.Task)
	}

	// TUI mode
	return runTUI(cfg)
}

func ensureWorkingDir(cfg *config.Config) error {
	dir, err := cfg.WorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create working directory: %w", err)
		}

		if err := initGitRepo(dir); err != nil {
			return fmt.Errorf("failed to initialize git repository: %w", err)
		}

		if err := ensureRepoFiles(dir); err != nil {
			return fmt.Errorf("failed to create repository files: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to access working directory: %w", err)
	} else {
		if err := ensureGitRepo(dir); err != nil {
			return fmt.Errorf("failed to ensure git repository: %w", err)
		}
	}

	tasksPath, err := cfg.TasksPath()
	if err != nil {
		return fmt.Errorf("failed to get tasks path: %w", err)
	}

	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		if err := os.WriteFile(tasksPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create tasks file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to access tasks file: %w", err)
	}

	return nil
}

func initGitRepo(dir string) error {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.Run()
}

func ensureGitRepo(dir string) error {
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return initGitRepo(dir)
	}
	return nil
}

func ensureRepoFiles(dir string) error {
	// Create README.md if not exists
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		readme := fmt.Sprintf(`# My Tasks

This repository contains task files managed by [ttt (Tiny Task Tool)](https://github.com/yostos/tiny-task-tool).

## Files

- `+"`tasks.md`"+` - Current tasks
- `+"`archive.md`"+` - Archived completed tasks

## Quick Start

`+"```"+`bash
ttt                    # Launch TUI
ttt -t "Buy milk"      # Add a task
ttt sync               # Sync with remote
`+"```"+`

For more information, visit: https://github.com/yostos/tiny-task-tool

---
Created by ttt %s on %s
`, cli.Version, time.Now().Format("2006-01-02"))

		if err := os.WriteFile(readmePath, []byte(readme), 0644); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
	}

	// Create .gitignore if not exists
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		gitignore := `# macOS
.DS_Store
._*
.Spotlight-V100
.Trashes

# Windows
Thumbs.db
ehthumbs.db
Desktop.ini

# Linux
*~
.directory

# Vim
*.swp
*.swo
.*.swp

# Emacs
*~
\#*\#
.#*

# VS Code
.vscode/

# Sublime Text
*.sublime-workspace

# nano
.*.swp
`
		if err := os.WriteFile(gitignorePath, []byte(gitignore), 0644); err != nil {
			return fmt.Errorf("failed to create .gitignore: %w", err)
		}
	}

	return nil
}

func addTask(cfg *config.Config, task string) error {
	tasksPath, err := cfg.TasksPath()
	if err != nil {
		return fmt.Errorf("failed to get tasks path: %w", err)
	}

	content, err := os.ReadFile(tasksPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}

	taskLine := fmt.Sprintf("- [ ] %s\n", task)

	var newContent string
	if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
		newContent = string(content) + "\n" + taskLine
	} else {
		newContent = string(content) + taskLine
	}

	if err := os.WriteFile(tasksPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	if cfg.Git.AutoCommit {
		if err := gitCommit(cfg, fmt.Sprintf("Add task: %s", task)); err != nil {
			// Don't fail if git commit fails, just log it
			fmt.Fprintf(os.Stderr, "Warning: git commit failed: %v\n", err)
		}
	}

	fmt.Printf("Added: %s\n", task)
	return nil
}

func runTUI(cfg *config.Config) error {
	tasksPath, err := cfg.TasksPath()
	if err != nil {
		return fmt.Errorf("failed to get tasks path: %w", err)
	}

	archivePath, err := cfg.ArchivePath()
	if err != nil {
		return fmt.Errorf("failed to get archive path: %w", err)
	}

	content, err := os.ReadFile(tasksPath)
	if err != nil {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}

	model := tui.NewWithPaths(cfg, string(content), tasksPath, archivePath)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

func gitCommit(cfg *config.Config, message string) error {
	dir, err := cfg.WorkingDir()
	if err != nil {
		return err
	}

	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = dir
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	// Check if there are changes to commit
	diffCmd := exec.Command("git", "diff", "--cached", "--quiet")
	diffCmd.Dir = dir
	if err := diffCmd.Run(); err == nil {
		// No changes to commit
		return nil
	}

	commitMsg := fmt.Sprintf("%s (%s)", message, time.Now().Format("2006-01-02 15:04"))
	commitCmd := exec.Command("git", "commit", "-m", commitMsg)
	commitCmd.Dir = dir
	return commitCmd.Run()
}

func setRemote(cfg *config.Config, url string) error {
	dir, err := cfg.WorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Ensure README.md and .gitignore exist before setting remote
	if err := ensureRepoFiles(dir); err != nil {
		return fmt.Errorf("failed to create repository files: %w", err)
	}

	if err := git.SetRemote(dir, url); err != nil {
		return err
	}

	fmt.Printf("Remote set to: %s\n", url)
	return nil
}

func syncTasks(cfg *config.Config) error {
	dir, err := cfg.WorkingDir()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := git.Sync(dir); err != nil {
		return err
	}

	fmt.Println("Sync completed successfully.")
	return nil
}

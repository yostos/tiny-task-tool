# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ttt (Tiny Task Tool)** is a TUI (Terminal User Interface) application for simple task management, following the "one sheet of paper" philosophy. It provides a minimal, distraction-free way to manage tasks without leaving the terminal.

## Technology Stack

- **Language:** Go (latest stable version)
- **TUI Framework:** charmbracelet/bubbletea + lipgloss + bubbles (Elm Architecture)
- **CLI Parsing:** spf13/pflag
- **Markdown Parsing:** yuin/goldmark
- **Config:** pelletier/go-toml v2
- **Target OS:** Linux, macOS (Windows is low priority)

## Build Commands

```bash
go build -o ttt .
go run .
go test ./...                   # Run all tests
go test -v ./path/to/package    # Run specific package tests
go test -run TestName ./...     # Run specific test
go test -cover ./...            # Run tests with coverage
golangci-lint run              # Run static analysis
golangci-lint run --fix        # Auto-fix issues
```

## Architecture

### Elm Architecture (bubbletea)

The application follows the Elm Architecture pattern:
- **Model:** Application state
- **Update:** State transitions based on messages (key presses, etc.)
- **View:** Render state to terminal output

### Key Design Principles

1. **"One Sheet of Paper"** - Only one file is managed; no file selection via CLI arguments
2. **"Focus on Today"** - No search, no past data browsing
3. **Unix Philosophy** - Do one thing well; editing is delegated to external editors ($EDITOR)
4. **Minimal Features** - Remove rather than add; simplicity is the core value

### Intentionally Excluded Features (Never Implement)

- File path arguments
- Multiple file support
- In-app editing
- Task completion toggle (use external editor)
- Markdown rendering
- Search/filter
- Tags/metadata
- Cloud sync
- AI features

## CLI Interface

```bash
ttt                          # Launch TUI
ttt -t buy milk              # Add task (no TUI)
ttt --task "buy milk"        # Add task with quotes
ttt --help                   # Help
ttt --version                # Version
```

## Task Format

Uses TaskPaper-style completion dates:
```markdown
- [ ] Incomplete task
- [x] Completed task @done(2026-01-18)
```

Completed tasks are archived to `archive.md` after a configurable delay period.

## Git Branch Strategy

GitHub Flow を採用：

```
main (常にリリース可能)
  ├── feature/xxx    # 機能追加
  ├── fix/xxx        # バグ修正
  └── docs/xxx       # ドキュメント更新
```

**ルール：**
- `main` は常にリリース可能な状態を維持
- 作業は必ずブランチを切って行う
- 完了したらmainにsquash merge
- バージョンはタグで管理（`v0.1.0` 等）

**GitHub CLI (gh) が使用可能：**
```bash
gh pr create --title "Add feature" --body "Description"
gh pr list
gh pr merge --squash
gh issue create --title "Bug report" --body "Description"
gh release create v0.1.0 --notes "Release notes"
```

## Documentation Structure

- `docs/concept.md` - Vision and design philosophy (immutable)
- `docs/specification.md` - Functional specifications (mutable)
- `docs/architecture.md` - Technology decisions with rationale (mutable)
- `docs/TODO.md` - Progress tracking

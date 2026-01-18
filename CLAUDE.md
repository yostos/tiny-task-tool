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

## Development Guidelines (MANDATORY)

### Test Requirements

**This is an absolute rule that must never be violated.**

1. **Every Public function/method MUST have test cases**
   - If a function is exported (starts with uppercase), it MUST have corresponding tests
   - No exceptions. No Public function without tests.

2. **Tests as Specification (Tests > Coverage)**
   - **Reading test code alone must explain the specification of the source code**
   - Test names should clearly describe the behavior being tested
   - Use table-driven tests to show various input/output scenarios
   - Coverage percentage is secondary; comprehensiveness of specification is primary
   - If someone reads only the test file, they should understand what the source file does
   - **Every test function MUST have an English comment explaining its intent and purpose**

3. **One source file = One test file (minimum)**
   - Every `.go` source file must have a corresponding `_test.go` file
   - Example: `config.go` → `config_test.go`
   - Example: `model.go` → `model_test.go`

4. **No undocumented Public APIs**
   - If there's no test case, the Public function should not exist
   - Before adding a Public function, write the test first (TDD)

5. **Private functions are tested through Public functions**
   - Private functions (lowercase) are tested indirectly via Public API tests
   - Complex private logic may have direct tests for clarity

6. **Integration tests are also required**
   - Unit tests for individual functions
   - Integration tests for component interactions
   - Place integration tests in `_test.go` files or separate `integration_test.go`

**Example of proper test structure:**
```go
// config.go has these Public functions:
// - Load(), ConfigPath(), ExpandPath(), etc.

// config_test.go must describe ALL of them:
func TestLoad(t *testing.T) { ... }                    // What Load() does
func TestLoadNonExistentConfig(t *testing.T) { ... }   // Edge case
func TestLoadInvalidConfig(t *testing.T) { ... }       // Error case
func TestConfigPath(t *testing.T) { ... }              // What ConfigPath() does
func TestExpandPath(t *testing.T) { ... }              // What ExpandPath() does
// ... every Public function covered
```

### Specification-Driven Testing (CRITICAL)

**Tests must verify specification, not just implementation.**

1. **Always check `docs/specification.md` before writing tests**
   - If the spec says "config file is at ~/.config/ttt/config.toml", the test MUST verify exactly that path
   - Do NOT use loose assertions like `HasSuffix()` or `Contains()` when exact values are specified

2. **Test comments should reference specification**
   ```go
   // TestConfigPath verifies that ConfigPath() returns ~/.config/ttt/config.toml.
   // Spec: docs/specification.md "設定ファイル仕様" section
   func TestConfigPath(t *testing.T) {
       expected := filepath.Join(home, ".config", "ttt", "config.toml")
       // Exact match, not partial match
       if path != expected {
           t.Errorf(...)
       }
   }
   ```

3. **Avoid "implementation-convenient" tests**
   - BAD: `strings.HasSuffix(path, "config.toml")` - passes even if path is wrong
   - GOOD: `path == expectedExactPath` - fails if implementation deviates from spec

4. **When specification has concrete values, test those exact values**
   - File paths, default values, key names, error messages
   - If spec says `delay_days = 2`, test must check for exactly `2`

### Before Committing Code

1. Run `go test ./...` - all tests must pass
2. Run `golangci-lint run` - no lint errors
3. Verify all new Public functions have tests
4. Verify test file exists for every source file
5. **Verify tests match specification in `docs/specification.md`**

## Documentation Structure

- `docs/concept.md` - Vision and design philosophy (immutable)
- `docs/specification.md` - Functional specifications (mutable)
- `docs/architecture.md` - Technology decisions with rationale (mutable)
- `docs/TODO.md` - Progress tracking

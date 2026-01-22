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
- Custom Markdown rendering implementation (library usage is acceptable)
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

Adopting GitHub Flow:

```
main (always releasable)
  ├── feature/xxx    # Feature additions
  ├── fix/xxx        # Bug fixes
  └── docs/xxx       # Documentation updates
```

**Rules:**
- `main` is always kept in a releasable state
- Work is always done on a branch
- Squash merge to main when complete
- Versions are managed with tags (`v0.1.0`, etc.)

**GitHub CLI (gh) is available:**
```bash
gh pr create --title "Add feature" --body "Description"
gh pr list
gh pr merge --squash
gh issue create --title "Bug report" --body "Description"
gh release create v0.1.0 --notes "Release notes"
```

### Bug Fix Workflow (Issue-Driven)

Bug fixes are done starting from GitHub Issues:

1. **Create Issue**: When a bug is found, create an issue with `gh issue create`
2. **Create Branch**: Create a `fix/issue-N` branch
3. **Fix & Test**: Fix using TDD, pass `go test ./...` and `golangci-lint run`
4. **User Verification**: User verifies the fix
5. **Create PR**: Create Pull Request with `gh pr create`
6. **Merge**: Squash merge to main on GitHub

```bash
# Example: Fixing Issue #1
gh issue create --title "Bug: ..." --body "Description"
git checkout -b fix/issue-1
# ... fix work ...
go test ./... && golangci-lint run
gh pr create --title "Fix #1: ..." --body "Closes #1"
# Merge on GitHub
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
   // Spec: docs/specification.md "Configuration File Specification" section
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
- `docs/roadmap.md` - Version plans and future features (mutable)
- `docs/TODO.md` - Progress tracking

### Documentation Format Rules

- **Updated date format**: `Updated: YYYY-MM-DD.` (with trailing period)
  - Example: `Updated: 2026-01-22.`

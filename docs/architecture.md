# ttt - Architecture

Updated: 2026-01-22.

## About This Document

This document describes the technology choices and architecture of ttt (Tiny Task Tool).
It records the decision-making process and rationale for future reference.

- For vision and design philosophy, see [concept.md](concept.md)
- For functional specifications, see [specification.md](specification.md)

## Technology Choices

### Finalized Decisions

#### Programming Language: Go

**Decision Date**: 2026-01-18

**Selection Rationale**:
- Easy distribution as a single binary
- Fast startup speed (aligns with concept.md's "eliminating friction")
- Low learning cost, good compatibility with AI-driven development
- bubbletea (TUI library) is mature

**Considered Alternatives and Rejection Reasons**:

| Candidate | Rejection Reason |
| --------- | ---------------- |
| Rust | High learning cost. Ownership/lifetime concepts are unique; risk of going in circles during AI-driven development when fixing compile errors without understanding. Possibility of rewriting for learning purposes remains open for the future. |
| Python | PEP 668 restricts direct installation to system Python, complicating distribution. Forces additional steps like pipx on users. Startup speed is also slow. |
| Ruby | Distributable via gem install, but the language's momentum is declining. AI training data unlikely to increase, raising concerns about future viability as a platform for learning AI-driven development. |
| TypeScript | High AI generation accuracy, but requires Node.js runtime dependency. Cannot distribute as single binary, violating concept.md's "eliminating friction." |

**Future Prospects**:
- Possible rewrite to Rust for learning purposes
- Would use ratatui + crossterm in that case

#### Configuration File Processing: pelletier/go-toml v2

**Decision Date**: 2026-01-18

**Selection Rationale**:
- TOML 1.0.0 compliant
- Actively maintained
- encoding/json-like API familiar to Go developers
- Strict mode enables typo detection
- 2-5x performance advantage

**Considered Alternatives and Rejection Reasons**:

| Candidate | Rejection Reason |
| --------- | ---------------- |
| BurntSushi/toml | Not maintained (unsupported). Stuck at TOML 0.4. |
| spf13/viper | Feature-rich but heavy. Overkill for this project. |
| knadh/koanf | Lightweight viper alternative, but unnecessary for simple TOML reading. |

#### Markdown Parser: yuin/goldmark

**Decision Date**: 2026-01-18

**Selection Rationale**:
- Actively maintained (updated 2026-01-06)
- CommonMark 0.31.2 compliant
- Default adoption by Hugo (static site generator), 32.2k dependencies track record
- Task list extension support
- No external dependencies (standard library only)
- High extensibility with custom AST, parser, renderer

**Considered Alternatives and Rejection Reasons**:

| Candidate | Rejection Reason |
| --------- | ---------------- |
| russross/blackfriday | No updates since November 2020. Maintenance stopped. |
| gomarkdown/markdown | Significantly inferior performance (about 20x slower). Low update frequency. |
| Custom regex implementation | Patterns become complex when handling heading/link coloring, reducing maintainability. |

#### CLI Argument Parsing: spf13/pflag

**Decision Date**: 2026-01-18

**Selection Rationale**:
- Actively maintained (updated 2025-09-02, v1.0.10)
- 348k projects depend on it
- POSIX-compliant argument parsing
- Can define short form (`-t`) and long form (`--task`) in one line
- Compatible with Go's standard flag package

**Considered Alternatives and Rejection Reasons**:

| Candidate | Rejection Reason |
| --------- | ---------------- |
| flag (standard) | Requires two definitions to define both short and long forms; verbose. |
| spf13/cobra | Features like subcommands are overkill for this project. |
| urfave/cli | Sufficient features but slightly heavier than pflag. |
| alecthomas/kong | Lower recognition, fewer adoption cases than pflag. |

#### TUI Framework: charmbracelet/bubbletea

**Decision Date**: 2026-01-18

**Selection Rationale**:
- Overwhelming adoption (38.4k stars, 17.6k dependent projects)
- Continuous maintenance by Charm company (updated 2025-09-17, v1.3.10)
- Clear state management with Elm architecture, easy to organize code
- Integration with lipgloss (styling) and bubbles (UI components)
- Easy external editor invocation and return (designed for exec.Command integration)
- AI-driven development compatibility: Clear patterns make AI-generated code easy to review

**Considered Alternatives and Rejection Reasons**:

| Candidate | Rejection Reason |
| --------- | ---------------- |
| rivo/tview | Widget-based and somewhat heavy. Overkill for ttt's simple requirements. |
| gdamore/tcell | Low-level API requires much custom implementation, increasing development cost. |

#### Static Analysis: golangci-lint

**Decision Date**: 2026-01-18

**Selection Rationale**:
- De facto standard in the Go community
- Integrates 100+ linters (staticcheck, revive, go vet, errcheck, etc.)
- Fast execution (parallel execution, caching)
- Centralized configuration with `.golangci.yml`
- Official GitHub Actions action available

**Considered Alternatives and Rejection Reasons**:

| Candidate | Rejection Reason |
| --------- | ---------------- |
| revive alone | Style checking only. Usable via golangci-lint. |
| staticcheck alone | High quality but limited features alone. Usable via golangci-lint. |
| go vet alone | Few check items. Included in golangci-lint. |

#### Testing: testing + testify

**Decision Date**: 2026-01-18

**Selection Rationale**:
- Based on Go's standard testing package, with concise assertions via testify
- Improved readability with `assert.Equal`, etc.
- Mock functionality (testify/mock) also available
- Widely adopted combination in the Go community

**Considered Alternatives and Rejection Reasons**:

| Candidate | Rejection Reason |
| --------- | ---------------- |
| testing only | Assertions are verbose (`if got != want` style). Inefficient for TDD. |
| ginkgo + gomega | BDD style has high learning cost; overkill for small projects. |

**Test Automation:**
- During development: Manual (`go test ./...`)
- On PR: Automated via GitHub Actions (tests + golangci-lint)

### Under Consideration

(None at this time)

## Technical Constraints

### Language

- Go (using latest stable version)

### TUI Library

- charmbracelet/bubbletea + lipgloss + bubbles

### Target Environments

- Target OS: Linux, macOS (Windows is low priority)
- Minimum supported terminal: 80x24

## Git Branch Strategy

Adopting GitHub Flow.

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

**Selection Rationale:**
- Prioritizing simplicity for solo development
- Git Flow is too complex (develop/release/hotfix branches unnecessary)
- Trunk Based Development is possible, but GitHub Flow is adopted for easier feature-based management

## Architecture

### Directory Structure

```
ttt/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── go.sum                  # Dependency hashes
├── .golangci.yml           # Linter configuration
├── docs/                   # Documentation
│   ├── concept.md
│   ├── specification.md
│   ├── architecture.md
│   └── TODO.md
└── internal/               # Internal packages
    ├── cli/                # CLI argument parsing
    │   ├── cli.go
    │   └── cli_test.go
    ├── config/             # Configuration file loading
    │   ├── config.go
    │   └── config_test.go
    ├── task/               # Task operations (parsing, adding, archiving)
    └── tui/                # TUI (bubbletea)
```

### Module Structure

| Package | Responsibility |
|---------|----------------|
| `main` | Entry point, initialization, mode branching |
| `internal/cli` | CLI argument parsing (using pflag) |
| `internal/config` | Configuration file loading, default value management |
| `internal/task` | Task file reading/writing, archive processing |
| `internal/tui` | TUI display, key input handling (bubbletea) |

### Data Flow

```
Startup
  │
  ├─ CLI argument parsing
  │    │
  │    ├─ --help/--version → Display and exit
  │    │
  │    └─ --task → Add task → git commit → Exit
  │
  └─ TUI mode
       │
       ├─ Load configuration
       ├─ Ensure working_dir (auto-create, git init)
       ├─ Load tasks.md
       ├─ Auto-add @done (when new completed tasks detected)
       ├─ Auto-archive (when archive.auto=true)
       │
       └─ TUI loop
            ├─ Display
            ├─ Wait for key input
            │    ├─ e → Launch editor → Reload → git commit
            │    ├─ a → Execute archive → git commit
            │    ├─ r → Reload
            │    └─ q → Exit
            └─ Update state
```

## Design Decision Notes

Records responsibilities ttt does not handle and how to address them.

### Mobile Access

**Need**: Want to view tasks.md from mobile

**Decision**: ttt is not involved. Solve with GitHub's existing ecosystem.

**Recommended Method**:
1. Manage `~/.ttt` as a GitHub Private repository
2. Sync to remote with `ttt sync` (v0.3.0)
3. View with GitHub Mobile app (authentication required ensures privacy)

**Rationale**: Following Unix philosophy "do one thing well," mobile viewing is delegated to GitHub. ttt focuses on task management, delegating everything else to existing tools to maintain simplicity.

**Note**: Publishing via GitHub Pages makes "access only for yourself" technically difficult (URLs are accessible even from Private repositories if known). GitHub Mobile app requires authentication, ensuring privacy.

### Git Sync is Manual Only (auto_sync Not Provided)

**Decision Date**: 2026-01-22

**Need**: Sync functionality with remote repository

**Decision**: Provide only manual sync via `ttt sync` command; do not provide `auto_sync` setting.

**Rationale**:
1. **Consideration for offline environments**: Not always online. Displaying errors each time auto-sync fails is a UX problem
2. **Predictable behavior**: User explicitly runs `ttt sync`, so network communication timing is predictable
3. **Conflict handling**: Pull conflicts require manual resolution. Auto-sync conflicts would force users to respond at unexpected times
4. **Maintaining simplicity**: Keep configuration options minimal and behavior simple

**Provided Features**:
- `ttt remote <url>`: Register remote repository as origin
- `ttt sync`: Manual sync (pull → commit → push)

**Not Provided**:
- `git.auto_sync` setting
- Auto-sync on startup/exit
- Auto-push after commit

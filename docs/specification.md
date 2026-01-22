# ttt (Tiny Task Tool) - Specification

Updated: 2026-01-22.

## About This Document

This document defines the functional specifications of ttt (Tiny Task Tool).

- For vision and design philosophy, see [concept.md](concept.md)
- For technology choices and architecture, see [architecture.md](architecture.md)

## Project Name

**ttt** (Tiny Task Tool)

- Short and memorable.
- Clear purpose.
- Easy to type as a command.

```bash
ttt                                    # Launch TUI
ttt -t buy kitchen paper and wasabi    # Add task (no TUI)
ttt --task "buy kitchen paper"         # Add task with quotes
ttt remote <url>                       # Register remote repository (v0.3.0)
ttt sync                               # Manual sync with remote (v0.3.0)
ttt --help                             # Show help
ttt -h                                 # Show help
ttt --version                          # Show version
ttt -v                                 # Show version
```

File specification is not supported. Based on the "one sheet of paper" principle, the file to open is a single fixed file specified in the configuration.

The `-t` (`--task`) option allows adding tasks. If an argument is provided, it's appended as a task to the main file, and the TUI is not launched. This lets you quickly add tasks without leaving the terminal.

## Use Cases

### Typical Daily Workflow

#### 1. Morning Routine

```bash
ttt
```

- Review yesterday's remaining tasks.
- With `archive.auto = true`, completed tasks past the configured delay are automatically archived.
- Press `a` to manually archive at any time.
- Press `e` to open the editor and add today's tasks.
- Close the editor to return to ttt.
- Upon returning, completed tasks automatically get `@done(date)` added.
- Press `q` to exit.

#### 2. During Work

```bash
ttt
```

- Look at tasks and decide next action.
- When a task is done, press `e` to open editor and check it off.
- Add new thoughts as they come.

#### 3. End of Day

```bash
ttt
```

- Review what was completed today.
- Press `e` to open editor and change completed tasks from `- [ ]` to `- [x]`.

## TaskPaper Format Adoption

### Recording Completion Dates

When a task is completed, the date is recorded in TaskPaper format.

```markdown
- [ ] Incomplete task
- [x] Completed task @done(2026-01-18)
- [x] Old completed task @done(2026-01-10)
```

### Archive Timing

Archive execution timing (see "Configuration File Specification" section for details):

- **`a` key**: Manually execute archive at any time
- **Auto-execute on startup**: If `archive.auto = true`, auto-execute on startup
- **After returning from editor**: Auto-archive is not executed (only file reload)

Only completed tasks that have passed the `delay_days` period are archived. This allows completed tasks to remain visible for a while.

### Archive Mechanism

Completed tasks are moved to the archive file (`archive.md`). They are removed from the main file, so they won't appear on screen at startup.

**Archive File Structure**

Sections with `## YYYY-MM-DD` headers are created for each completion date, grouping tasks completed on that date.

```markdown
## 2026-01-20

- [x] Task A @done(2026-01-20)
- [x] Task B @done(2026-01-20)

## 2026-01-19

- [x] Task C @done(2026-01-19)
```

**Characteristics**

- The main file opened by `ttt` doesn't show tasks past the delay period from completion
- If history is needed, `archive.md` can be viewed separately
- Plain text, readable by any tool
- Git provides complete history management

### Hierarchical Tasks (Parent-Child Relationships)

Markdown indentation-based hierarchy is supported.

```markdown
- [ ] Parent task
  - [ ] Child task 1
  - [ ] Child task 2
    - [ ] Grandchild task
```

**Indentation Rules:**
- 2 spaces = 1 level
- Tabs are treated as 2 spaces

**Behavior When Parent Task is Completed:**

When a parent task is completed (`- [x]`), all child tasks are automatically completed as well.

```markdown
# Before editing
- [x] Parent task
  - [ ] Child task 1
  - [ ] Child task 2

# After ttt processing
- [x] Parent task @done(2026-01-19)
  - [x] Child task 1 @done(2026-01-19)
  - [x] Child task 2 @done(2026-01-19)
```

**Behavior During Archive:**

When a parent task becomes archivable, all child tasks and child nodes (including non-task lines) are archived together.
Indentation structure is preserved.

**Important Rules:**
- **If parent is incomplete**: Child tasks/nodes are not archived even if completed
- **If parent is completed but less than delay_days**: Child tasks/nodes are not archived
- **Grouping uses parent's date**: The archive.md section heading (`## YYYY-MM-DD`) uses the parent task's completion date. Child task @done tags are preserved, but displayed under the parent's date section
- **Non-task lines (bullets)**: `- text` format child nodes (without checkbox) are treated as completed and archived with parent

```markdown
# archive.md
## 2026-01-19

- [x] Parent task @done(2026-01-19)
  - [x] Child task 1 @done(2026-01-18)  ← Child's date preserved but under parent's section
  - Memo line                           ← Non-task line archived with parent
```

## Configuration File Specification

### Location

Follows XDG Base Directory specification:

1. If `XDG_CONFIG_HOME` environment variable is set → `$XDG_CONFIG_HOME/ttt/config.toml`
2. If not set → `os.UserConfigDir()/ttt/config.toml`

```
# Example: Linux (XDG_CONFIG_HOME not set, os.UserConfigDir() = /home/foo/.config)
/home/foo/.config/ttt/config.toml

# Example: XDG_CONFIG_HOME=/custom/config
/custom/config/ttt/config.toml
```

### Auto-creation

If the configuration file doesn't exist, it's automatically created with default values on first launch.
If the directory doesn't exist, it's also created automatically.

### Configuration File Structure

```toml
[file]
# Directory for task files
working_dir = "~/.ttt"
# File names are fixed:
#   - tasks.md (main file)
#   - archive.md (archive file)

[archive]
# Execute auto-archive on startup
auto = false
# Days after completion before archiving
delay_days = 2

[editor]
# Editor launch command template
# {file} is replaced with the file path
# If omitted, uses $EDITOR environment variable (auto-appends "{file}")
# Example: command = "vim {file}"
# Example: command = "code --wait {file}"
# Example: command = "emacs -nw {file}"
command = "vim {file}"

[keybindings]
# Alternative keys for ↑↓ (multiple can be specified)
up = ["k", "ctrl+p"]
down = ["j", "ctrl+n"]

# Go to top/bottom
top = ["g", "Home"]
bottom = ["G", "End"]

# Half-page scroll
half_page_up = ["ctrl+u"]
half_page_down = ["ctrl+d"]

# Modifier key notation:
#   - ctrl+<key>: Ctrl key + key (e.g., ctrl+n, ctrl+p)
#   - alt+<key>: Alt key + key (e.g., alt+f, alt+b)
#   - shift+<key>: Shift key + key (e.g., shift+tab)
#   - Case sensitive (G ≠ g)
#   - Function keys: Home, End, PageUp, PageDown, etc.

[git]
# Auto-commit (enabled by default)
# Automatically git commit in background on changes
auto_commit = true
```

### Default Values

When the configuration file doesn't exist, these default values are used:

- `file.working_dir` → `~/.ttt`
- File names (fixed):
  - Main file: `tasks.md`
  - Archive file: `archive.md`
- `archive.auto` → `false`
- `archive.delay_days` → `2`
- `editor.command` → Value of `$EDITOR` environment variable + ` {file}`
  - If `$EDITOR` is not set: `vi {file}`
- `keybindings.up` → `["k"]`
- `keybindings.down` → `["j"]`
- `keybindings.top` → `["g", "Home"]`
- `keybindings.bottom` → `["G", "End"]`
- `keybindings.half_page_up` → `["ctrl+u"]`
- `keybindings.half_page_down` → `["ctrl+d"]`
- `git.auto_commit` → `true`

### Design Rationale

- **Fixed file names**: Reinforces the "one sheet of paper" principle, eliminates file selection complexity
- **Template format**: Allows flexible specification of editor-specific options (`--wait`, `-nw`, etc.)
- **XDG compliance**: Follows standard configuration file placement for Linux/macOS

## Keybinding Specification

ttt is a viewer-type TUI tool without editing capabilities. It provides simple, memorable keybindings.

### Fixed Keybindings (Not Configurable)

The following keys are fixed and cannot be changed in the configuration file:

| Key | Action | Description |
|-----|--------|-------------|
| `↑` | Scroll up one line | Always enabled |
| `↓` | Scroll down one line | Always enabled |
| `e` | Launch editor | Opens tasks.md in configured editor |
| `a` | Execute archive | Archives completed tasks meeting criteria |
| `r` | Reload | Reloads file (automatic after editor exit) |
| `q` | Quit | Exit ttt |
| `?` / `h` | Show help | Display keybinding list as overlay |

### Configurable Keybindings

The following keys can be customized in the configuration file (`[keybindings]`):

| Action | Config Key | Default | Description |
|--------|-----------|---------|-------------|
| Scroll up | `up` | `["k"]` | Alternative to ↑ key |
| Scroll down | `down` | `["j"]` | Alternative to ↓ key |
| Go to top | `top` | `["g", "Home"]` | Go to beginning of file |
| Go to bottom | `bottom` | `["G", "End"]` | Go to end of file |
| Half page up | `half_page_up` | `["ctrl+u"]` | |
| Half page down | `half_page_down` | `["ctrl+d"]` | |

**Customization Example:**

```toml
[keybindings]
# Emacs-style keybindings
up = ["k", "ctrl+p"]
down = ["j", "ctrl+n"]
top = ["alt+<", "Home"]
bottom = ["alt+>", "End"]
```

### Design Rationale

- **Minimal fixed keys**: Only basic operations (↑↓) and function keys (e/a/r/q/?/h) are fixed
- **Flexible scroll keys**: Supports different editor habits (vim/Emacs)
- **less/man-like**: Similar operation feel to existing pager tools
- **No cursor**: Since there's no editing function, cursor concept is unnecessary

## TUI Screen Layout

Based on the "one sheet of paper" concept, only minimal information is displayed.

### Screen Structure

```
┌─────────────────────────────────────────────────┐
│ # Today                                         │
│                                                 │
│ - [ ] Buy milk                                  │
│ - [x] Review PR #123 @done(2026-01-18)          │
│ - [ ] Write documentation                       │
│                                                 │
│ ## Notes                                        │ ← Main area
│                                                 │
│ - Remember to check email                       │
│                                                 │
├─────────────────────────────────────────────────┤
│ ? help | e edit | a archive | q quit   [15/42] │ ← Footer
└─────────────────────────────────────────────────┘
```

### Area Details

#### Header

**None**. Based on the concept of not making users aware of files, no header is displayed.

#### Main Area

- Displays file content as-is (no Markdown rendering)
- Scrollable
- No line numbers displayed

#### Footer (1 line)

**Normal display:**
```
? help | e edit | a archive | q quit    [15/42] ttt v0.1.0
```

- Left side: Key operation hints
- Right side: Scroll position `[current line/total lines]` and version info

### Status Messages

Temporary messages are displayed in the footer (returns to normal display after 3 seconds).

**On startup (when new completed tasks detected):**
```
3 tasks marked as done                      [1/42]
```
Displayed when `- [x]` tasks without `@done(date)` are detected and automatically tagged with `@done(today's date)`.

**On archive execution:**
```
Archived 3 tasks                            [1/39]
```

**When no tasks to archive:**
```
No tasks to archive                         [1/42]
```

### Help Overlay

When pressing `?` or `h` to show help, it's displayed as an overlay in the center of the screen.
Customized keybindings are reflected dynamically.

**Display example with default settings:**

```
┌────────────── Help ──────────────┐
│                                  │
│  ↑/k      Scroll up              │
│  ↓/j      Scroll down            │
│  g/Home   Go to top              │
│  G/End    Go to bottom           │
│  Ctrl+u   Half page up           │
│  Ctrl+d   Half page down         │
│                                  │
│  e        Open editor            │
│  a        Archive                │
│  r        Reload                 │
│                                  │
│  q        Quit                   │
│  ?/h      Help                   │
│                                  │
│  Press any key to close          │
└──────────────────────────────────┘
```

### Colors and Styling

Minimal coloring to maintain simplicity.

| Element | Style |
|---------|-------|
| Incomplete task (`- [ ]`) | Normal display |
| Completed task (`- [x]`) | Gray/dim |
| `@done(date)` | Gray/dim |
| Footer | Inverted |
| Help overlay | With border |

## Error Handling

### Basic Policy

- Non-fatal errors are handled automatically, reducing user effort
- Fatal errors display a message and exit
- Files are edited directly (backups are protected by Git auto-commit)

### On Startup

| Case | Response |
|------|----------|
| No configuration file | Operate with default values |
| working_dir doesn't exist | Auto-create + `git init` |
| tasks.md doesn't exist | Auto-create empty file |
| archive.md doesn't exist | Auto-create on first archive |
| Not a git repository | Auto `git init` |
| Cannot read tasks.md (permission error) | Display error message and exit |
| Configuration file format error | Display error message and exit |

### At Runtime

| Case | Response |
|------|----------|
| Editor not found | Display error message in footer |
| Editor exits abnormally | Reload file and continue |
| Cannot write to archive.md | Display error message in footer |
| File deleted externally | Display error message and exit |

### Error Message Examples

**On startup (fatal error):**
```
Error: Cannot read ~/.ttt/tasks.md: permission denied
```

**At runtime (displayed in footer):**
```
Error: Editor not found: vim                [15/42]
```

## Git Integration

Following the nb command approach, transparent Git integration is provided. Users don't need to be aware of Git operations; version history is automatically saved.

### Auto-commit

Background auto-commit at the following times:

- After editor exit (on file changes)
- After archive execution
- After `@done(date)` addition
- When adding task via `ttt -t`

### Initialization

- Auto `git init` when creating working_dir
- Auto `git init` even if existing directory is not a git repository

### Repository File Auto-generation (v0.3.0)

The following files are auto-generated if they don't exist during working_dir initialization and `ttt remote` execution.

**README.md:**
- Written in English
- Includes brief ttt usage instructions
- Includes creation date and ttt version
- Includes link to ttt project

**README.md content example:**
```markdown
# My Tasks

This repository contains task files managed by [ttt (Tiny Task Tool)](https://github.com/yostos/tiny-task-tool).

## Files

- `tasks.md` - Current tasks
- `archive.md` - Archived completed tasks

## Quick Start

ttt                    # Launch TUI
ttt -t "Buy milk"      # Add a task
ttt sync               # Sync with remote

For more information, visit: https://github.com/yostos/tiny-task-tool

---
Created by ttt vX.Y.Z on YYYY-MM-DD
```

**.gitignore:**

Excludes OS-generated files and editor temporary files.

| Category | Target Files |
|----------|--------------|
| macOS | `.DS_Store`, `._*`, `.Spotlight-V100`, `.Trashes` |
| Windows | `Thumbs.db`, `ehthumbs.db`, `Desktop.ini` |
| Linux | `*~`, `.directory` |
| Vim | `*.swp`, `*.swo`, `.*.swp` |
| Emacs | `*~`, `\#*\#`, `.#*` |
| VS Code | `.vscode/` |
| Sublime Text | `*.sublime-workspace` |
| nano | `.*.swp` |

### Remote Repository Registration (v0.3.0)

Register a remote repository with the `ttt remote <url>` command.

```bash
ttt remote https://github.com/user/my-tasks.git
ttt remote git@github.com:user/my-tasks.git
```

**Behavior:**
- Executes `git remote add origin <url>`
- If origin already exists, updates with `git remote set-url origin <url>`
- On success: Displays `Remote set to: <url>`
- On failure: Displays error message and exits with code 1

### Manual Sync (v0.3.0)

Execute manual sync with the `ttt sync` command.

```bash
ttt sync
```

**Behavior:**
1. `git pull origin <current-branch>` to fetch from remote
   - On first sync (remote branch doesn't exist), skip pull
2. Auto-commit if there are uncommitted changes
3. `git push origin <current-branch>` to push to remote

**Error Handling:**
- Remote not configured: Display `Error: No remote 'origin' configured. Use 'ttt remote <url>' first.`
- Conflict on pull: Display `Error: Merge conflict detected. Please resolve manually.` and output diff with `git diff`
- Pull failure (no branch on remote, etc.): Skip pull and proceed to commit → push
- Push failure: Display error message

**Notes:**
- `auto_sync` setting is not provided. Sync is always manual with `ttt sync`
- No sync functionality from TUI. TUI remains a viewer only
- Safe to use in offline environments

### Configuration

```toml
[git]
auto_commit = true  # Enabled by default, can be disabled with false
```

## Installation Methods (v0.3.0)

### go install

Supports `go install` as Go's standard installation method.

```bash
go install github.com/yostos/tiny-task-tool@latest
```

**Requirements:**
- Go 1.21 or later
- `$GOPATH/bin` or `$HOME/go/bin` must be in PATH

**Version Information:**
- `@latest` installs the latest release
- Tag specification like `@v0.3.0` is also possible

### make install

Use `make install` when building and installing from source.

```bash
# Default: Install to $GOPATH/bin
make install

# Custom directory: Install to PREFIX/bin
make install PREFIX=/usr/local
```

**Makefile Targets:**
- `make build` - Build binary (generates `ttt` in current directory)
- `make install` - Build and install
- `make test` - Run tests
- `make lint` - Static analysis
- `make check` - test + lint
- `make clean` - Remove build artifacts
- `make version` - Show current version

### Homebrew (macOS / Linux)

Supports installation via Homebrew.

```bash
brew install yostos/tap/ttt
```

**Tap Repository:**
- Formula placed in `github.com/yostos/homebrew-tap`
- Formula name: `ttt.rb`

**Formula Contents:**
- Builds with Go (`go build`)
- Version synced with GitHub Release tags
- Supports Darwin (macOS) and Linux

## Scope and Constraints

### Implemented Features (v0.1.0〜v0.2.0)

- [x] Main file viewing (scrollable)
- [x] Completed task detection (`- [x]`)
- [x] Completed task archiving (auto-add `@done(date)`)
- [x] External editor launch (using configuration file template)
- [x] Basic keybindings (↑↓, e, a, r, q, ?/h)
- [x] Task addition from command line (`-t`, `--task`)
- [x] Git auto-commit (transparent version control)
- [x] Hierarchical tasks (parent-child relationships, completion cascade, batch archive)

### Future Features

For features planned for future versions, see [roadmap.md](roadmap.md).


### Resource Constraints

- Solo development (minimize maintenance burden)
- Minimal documentation
  - README.md
  - concept.md (vision and philosophy)
  - specification.md (this document)
  - architecture.md (technology choices and architecture)
  - roadmap.md (future plans)
- Tests are minimal (core functions only)

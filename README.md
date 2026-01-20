# ttt (Tiny Task Tool)

[![CI](https://github.com/yostos/tiny-task-tool/actions/workflows/ci.yml/badge.svg)](https://github.com/yostos/tiny-task-tool/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/yostos/tiny-task-tool)](https://github.com/yostos/tiny-task-tool/releases)

A minimal TUI task manager for the terminal.

## Philosophy

- **One sheet of paper** - One file, no folders, no decisions
- **Focus on today** - No searching, no past data browsing
- **Node-oriented** - Tasks are managed as trees, not lines. When you complete a parent, children follow. When you archive, the whole subtree moves together. This matches how humans naturally think of indented blocks as a single unit.
- **Unix philosophy** - Edit with your favorite editor ($EDITOR)

## Inspiration

This project is inspired by:

- [ephe](https://github.com/unvalley/ephe) - "One clean page to focus your day"
- [aeph](https://github.com/siki-712/aeph) - CLI markdown editor inspired by ephe

I was impressed by aeph's approach of bringing the experience to the command line. However, I felt that these tools had drifted from the original "one sheet of paper" philosophy. This tool is my attempt to stay true to that core principle: one file, no choices, no distractions.

## Installation

### From Source

```bash
git clone https://github.com/yostos/tiny-task-tool.git
cd tiny-task-tool
make install
```

### Requirements

- Go 1.21+
- Linux or macOS

## Usage

### Quick Start

```bash
ttt                    # Launch TUI
ttt -t "buy milk"      # Add task quickly
ttt --help             # Show help
ttt --version          # Show version
```

### Daily Workflow

1. Run `ttt` to view your tasks
2. Press `e` to open your editor and add/edit tasks
3. Mark tasks as done by changing `- [ ]` to `- [x]`
4. Save and close the editor - ttt will automatically add `@done(date)` tags
5. Old completed tasks are automatically archived after a configurable delay

## Task Format

Tasks use a simple markdown checkbox format:

```markdown
- [ ] Incomplete task
- [x] Completed task @done(2026-01-18)
```

### Hierarchical Tasks

Tasks can have children using 2-space indentation:

```markdown
- [x] Parent task @done(2026-01-19)
  - [x] Child task 1 @done(2026-01-19)
  - [x] Child task 2 @done(2026-01-19)
  - Note without checkbox (archived with parent)
```

**Rules:**
- When a parent task is marked complete, all children are automatically completed
- Children are only archived when their parent becomes archivable
- Non-task lines (bullets without checkboxes) are archived together with their parent

## Keybindings

| Key | Action |
|-----|--------|
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |
| `g` / `Home` | Go to top |
| `G` / `End` | Go to bottom |
| `Ctrl+u` | Half page up |
| `Ctrl+d` | Half page down |
| `e` | Open editor |
| `a` | Archive completed tasks |
| `r` | Reload file |
| `q` | Quit |
| `?` / `h` | Show help |

## Configuration

Configuration file location: `~/.config/ttt/config.toml`

```toml
[file]
# Directory for task files
working_dir = "~/.ttt"
# Files: tasks.md (main), archive.md (archive)

[archive]
# Auto-archive on startup
auto = false
# Days to wait before archiving completed tasks
delay_days = 2

[editor]
# Editor command template ({file} is replaced with file path)
command = "vim {file}"

[keybindings]
# Customize navigation keys
up = ["k"]
down = ["j"]
```

### Default Values

If no config file exists, these defaults are used:

- `file.working_dir`: `~/.ttt`
- `archive.auto`: `false`
- `archive.delay_days`: `2`
- `editor.command`: `$EDITOR {file}` (falls back to `vi {file}`)

## Archive

Completed tasks with `@done(date)` tags are automatically archived after `delay_days` have passed.

### Archive Behavior

- Tasks are moved to `archive.md` in the same directory
- Archive is organized by completion date with `## YYYY-MM-DD` headers
- Hierarchical structure is preserved
- Children are always archived with their parent (using parent's completion date for grouping)

### Manual Archive

Press `a` in the TUI to manually trigger archiving.

## Git Integration

ttt automatically manages version control:

- Initializes a git repository in the working directory
- Auto-commits changes after edits, archive operations, and task additions
- Can be disabled with `git.auto_commit = false` in config

## License

MIT

# ttt (Tiny Task Tool)

[![CI](https://github.com/yostos/tiny-task-tool/actions/workflows/ci.yml/badge.svg)](https://github.com/yostos/tiny-task-tool/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/yostos/tiny-task-tool)](https://github.com/yostos/tiny-task-tool/releases)

A digital sheet of paper for your terminal.
Write down what's on your mind, then get back to work.

Born from the frustration of todo apps that ask too many questions.

![demo](images/demo.gif)

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Features](#features)
- [Configuration](#configuration)
- [Git Integration](#git-integration)
- [Why ttt?](#why-ttt)
- [Philosophy](#philosophy)
- [Inspiration](#inspiration)
- [Contributing](#contributing)
- [Built with](#built-with)
- [License](#license)

## Installation

### Using Homebrew (macOS / Linux)

```bash
brew install yostos/tap/ttt

# Upgrade to latest version
brew update && brew upgrade yostos/tap/ttt
```

### Using go install

```bash
go install github.com/yostos/tiny-task-tool@latest
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH.

### From Source

```bash
git clone https://github.com/yostos/tiny-task-tool.git
cd tiny-task-tool
make install
```

By default, `make install` installs to `$GOPATH/bin` or `$HOME/go/bin`. You can specify a custom directory:

```bash
make install PREFIX=/usr/local
```

### Requirements

- Go 1.24+
- Linux or macOS (Windows is not supported)

### Uninstall

```bash
# Homebrew
brew uninstall yostos/tap/ttt

# go install / From Source (may require sudo depending on install location)
which ttt              # Check the install path first
rm "$(which ttt)"
```

## Usage

### Quick Start

```bash
ttt                    # Launch TUI
ttt -t "buy milk"      # Add task quickly
ttt --help             # Show help
ttt --version          # Show version
```

> **Note:** ttt automatically initializes a git repository and auto-commits changes in your working directory (`~/.ttt`). You can disable this with `git.auto_commit = false` in config.

### Daily Workflow

1. Run `ttt` to view your tasks
2. Press `e` to open your editor and add/edit tasks
3. Mark tasks as done by changing `- [ ]` to `- [x]`
4. Save and close the editor - ttt will automatically add `@done(date)` tags
5. Old completed tasks are automatically archived to `archive.md` after a configurable delay

## Features

### Task Format

Tasks use a simple markdown checkbox format:

```markdown
- [ ] Incomplete task
- [ ] Task with deadline @due(2026-02-15)
- [x] Completed task @done(2026-01-18)
```

### Due Dates

Add `@due(YYYY-MM-DD)` to track deadlines. Tasks are color-coded by urgency:

- **Red**: Overdue (past due date)
- **Green**: Due soon (within 7 days by default)
- **White**: Future (beyond threshold)

```markdown
- [ ] Submit report @due(2026-02-15)
- [ ] Review PR @due(2026-02-10)
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

### Keybindings

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

Navigation keys (`k`, `j`, etc.) are customizable via `[keybindings]` in config.

### Display

The TUI features syntax highlighting and a two-line footer:

```
Total: 12  Incomplete: 5  Completed: 7 (3 today)  Overdue: 2
? help | e edit | a archive | q quit              [15/42] v0.4.0
```

**Syntax highlighting:**
- Headings (`#`, `##`, etc.) are shown in cyan/bold
- Completed tasks are shown in gray
- Overdue tasks (`@due` in the past) are shown in red
- Due soon tasks (`@due` within threshold) are shown in green

## Configuration

Configuration file: `~/.config/ttt/config.toml`

```toml
[file]
working_dir = "~/.ttt"

[archive]
auto = false
delay_days = 2

[editor]
command = "vim {file}"

[display]
due_warning_days = 7

[keybindings]
up = ["k"]
down = ["j"]

[git]
auto_commit = true

[colors]
heading = "cyan"
completed = "gray"
overdue = "red"
due_soon = "green"
```

For all configuration options and default values, see [docs/specification.md](docs/specification.md).

## Git Integration

ttt automatically manages version control:

- Initializes a git repository in the working directory
- Auto-commits changes after edits, archive operations, and task additions
- Can be disabled with `git.auto_commit = false` in config

### Remote Sync

You can sync your tasks with a remote repository:

```bash
# Set remote repository (once)
ttt remote https://github.com/username/my-tasks.git

# Sync with remote
ttt sync
```

The `ttt sync` command performs:
1. `git pull` from remote (skipped on first sync)
2. Auto-commit any uncommitted changes
3. `git push` to remote

> **Note:** Your task file may contain personal thoughts and sensitive information. If using remote sync, ensure the repository is **private** to avoid unintentional exposure. Keep in mind that data pushed to git history is difficult to fully remove later.

## Why ttt?

### Comparison

| Aspect | Todo Apps | Note Apps | **ttt** |
| --- | --- | --- | --- |
| Features | Projects, tags, deadlines | Links, search, folders | Only viewing and archiving |
| Files | Managed within app | Multiple files/folders | One Markdown file |
| Editing | Dedicated editor | Dedicated editor | Your editor (Vim, etc.) |
| Environment | GUI (Web/Desktop) | GUI (Web/Desktop) | TUI (Terminal) |
| Learning cost | High | Medium-High | Low (just a few keys) |

### What we don't do

- No file selection or multiple file support
- No in-app editing (use your `$EDITOR`)
- No search, filter, tags, or metadata
- No cloud sync (other than Git)
- No AI features

Having fewer features is the point. If you need those things, there are great tools for that. ttt is for when you just want to write something down and get back to work.

## Philosophy

- **One sheet of paper** - One file, no folders, no decisions
- **Focus on now** - Capture ideas, tasks, and worries the moment they come to mind. No searching, no organizing, just write and clear your head.
- **Node-oriented** - Tasks are managed as trees, not lines. When you complete a parent, children follow. When you archive, the whole subtree moves together. This matches how humans naturally think of indented blocks as a single unit.
- **Unix philosophy** - Edit with your favorite editor ($EDITOR)

## Inspiration

This project is inspired by:

- [ephe](https://github.com/unvalley/ephe) - "One clean page to focus your day"
- [aeph](https://github.com/siki-712/aeph) - CLI markdown editor inspired by ephe

I was impressed by aeph's approach of bringing the experience to the command line. However, I felt that these tools had moved beyond the original "one sheet of paper" philosophy. This tool is my attempt to stay true to that core principle: one file, no choices, no distractions.

## Contributing

Issues and pull requests are welcome on [GitHub](https://github.com/yostos/tiny-task-tool).

## Built with

[Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm.

## License

[MIT License](LICENSE) - Copyright (c) 2026 yostos

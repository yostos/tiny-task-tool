# Demo Assets

This directory contains files for creating demo recordings of ttt.

## Files

| File | Description |
|------|-------------|
| `record.sh` | Script to record demo (handles backup/restore) |
| `demo.tape` | VHS script for recording terminal demo |
| `sample-tasks.md` | Sample task file used in demo |

Generated output is saved to `../images/demo.gif`.

## Prerequisites

- [VHS](https://github.com/charmbracelet/vhs) - Terminal recorder by Charm
- [ttt](https://github.com/yostos/tiny-task-tool) - This tool, installed and in PATH

### Installing VHS

```bash
# macOS
brew install vhs

# Go
go install github.com/charmbracelet/vhs@latest
```

## Usage

### Quick Start (Recommended)

```bash
cd demo
./record.sh
```

This script will:
1. Backup your existing `~/.ttt/tasks.md`
2. Replace it with `sample-tasks.md`
3. Run VHS to record the demo
4. Restore your original `tasks.md`
5. Move generated GIF to `images/demo.gif`

### Manual Steps

If you prefer to run manually:

```bash
# Backup and setup
cp ~/.ttt/tasks.md ~/.ttt/tasks.md.backup
cp sample-tasks.md ~/.ttt/tasks.md

# Record
vhs demo.tape

# Restore
mv ~/.ttt/tasks.md.backup ~/.ttt/tasks.md
```

### 3. Customize

Edit `demo.tape` to modify:
- Output format (`Output demo.gif` or `Output demo.webm`)
- Terminal dimensions (`Set Width`, `Set Height`)
- Theme (`Set Theme`)
- Typing speed (`Set TypingSpeed`)
- Actions and timing

See [VHS documentation](https://github.com/charmbracelet/vhs) for full syntax.

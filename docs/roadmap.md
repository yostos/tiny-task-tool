# ttt - Roadmap

Updated: 2026-02-09.

## Versioning Policy

- Follows Semantic Versioning (SemVer)
- Features added with each minor version
- Features violating concept.md design philosophy will not be added

---

## v0.2.0 (Released)

Hierarchical task support.

- [x] Indentation detection (2 spaces = 1 level)
- [x] Parent task completion → child tasks auto-complete
- [x] Parent task archive → child tasks move together
- [x] Makefile added (version embedding build)
- [x] Version display in footer

---

## v0.3.1 (Released)

Bug fix release.

- [x] Translate HELP dialog from Japanese to English

---

## v0.3.0 (Released)

Git sync functionality and public beta preparation.

### Git Sync Functionality

- [x] `ttt remote <url>` command: Register remote repository
- [x] `ttt sync` command: Manual sync with remote (pull → commit → push)
- [x] Repository file auto-generation (README.md, .gitignore)

**Note:** `auto_sync` setting is not provided (manual sync only)

### Public Beta Preparation

- [x] `go install github.com/yostos/tiny-task-tool@latest` support
- [x] Makefile improvements (go install / PREFIX variable support)
- [x] Homebrew formula creation (`brew install yostos/tap/ttt`)
- [x] Document translation to English (except TODO.md)
- [x] Add installation instructions to README.md

---

## v0.4.0 (Released)

Display improvements and due date support.

### Syntax Highlighting

- [x] Headings (`#`, `##`, etc.) in cyan/bold
- [x] Completed tasks (`- [x]`) in gray
- [x] Due date color coding based on urgency
- [x] Configurable colors via `[colors]` section in config

### Due Date Support (`@due` tag)

- [x] `@due(YYYY-MM-DD)` tag parsing (display only, not auto-managed)
- [x] Overdue tasks highlighted in red
- [x] Due soon tasks (within threshold) in green
- [x] Configurable warning threshold (`display.due_warning_days`)

### Statistics Display

- [x] Footer shows: completed today, incomplete, overdue counts
- [x] Detailed statistics in help overlay (hierarchy breakdown)

---

## v0.5.0 (Under Consideration)

Additional features. Low priority, implementation undecided.

- [ ] Snapshot feature (save today's state)
- [ ] Archive to separate file (selectable in settings)

---

## Not To Be Implemented

Features defined as "what we don't do" in concept.md are not included in the roadmap.

- File specification via command-line arguments
- Multiple file support
- In-file editing functionality
- Task completion toggle (leave to editor)
- Custom Markdown rendering implementation (library usage is acceptable)
- Search/filter functionality
- Tags or metadata
- Cloud sync (other than Git)
- AI features

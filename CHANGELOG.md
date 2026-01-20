# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1] - 2026-01-20

### Fixed

- Help dialog layout broken with Japanese (multibyte) text background (#9)
  - Use display width instead of byte length for character width calculation
  - Add `truncateByDisplayWidth()` and `skipByDisplayWidth()` helper functions

## [0.2.0] - 2026-01-19

### Added

- Hierarchical task support with 2-space indentation
- Parent task completion cascades to all children automatically
- Non-task bullet points (notes) are archived together with parent task
- `ArchiveTask` struct for tracking archive grouping dates
- Comprehensive test coverage for hierarchical archive behavior
- English user guide in README.md
- Makefile for version-embedded builds
- Version display in TUI footer

### Changed

- Archive grouping now uses parent task's completion date for all children
- Children are only archived when their parent becomes archivable
- Updated specification.md with detailed hierarchical task rules

### Removed

- Empty `docs/development-guideline.md` (content is in CLAUDE.md)

## [0.1.0] - 2026-01-18

### Added

- Initial release
- TUI for viewing and scrolling tasks
- External editor integration (`e` key)
- Task completion detection (`- [x]`)
- Automatic `@done(date)` tag addition
- Archive feature for old completed tasks (`a` key)
- Auto-archive on startup option
- File reload (`r` key)
- Help overlay (`?`/`h` key)
- CLI task addition (`-t`/`--task` option)
- Git auto-commit integration
- Configurable keybindings
- TOML configuration file support

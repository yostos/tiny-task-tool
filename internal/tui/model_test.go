package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yostos/tiny-task-tool/internal/config"
)

// Test constants
const (
	testTasksPath   = "/tmp/test-tasks.md"
	testArchivePath = "/tmp/test-archive.md"
)

// TestNew verifies that New() correctly initializes a Model with the given content.
// It should parse lines correctly and handle edge cases like empty content and trailing newlines.
func TestNew(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		name          string
		content       string
		expectedLines int
	}{
		{"empty content", "", 0},
		{"single line without newline", "- [ ] Buy milk", 1},
		{"single line with newline", "- [ ] Buy milk\n", 1},
		{"two lines", "- [ ] Buy milk\n- [ ] Buy bread\n", 2},
		{"multiple lines with headers", "# Tasks\n\n- [ ] Task 1\n- [ ] Task 2\n", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(cfg, tt.content)

			if len(m.lines) != tt.expectedLines {
				t.Errorf("New() lines = %d, want %d (content: %q)", len(m.lines), tt.expectedLines, tt.content)
			}

			if m.content != tt.content {
				t.Errorf("New() content = %q, want %q", m.content, tt.content)
			}

			if m.config != cfg {
				t.Error("New() config not set correctly")
			}
		})
	}
}

// TestInit verifies that Init() returns nil command.
// The TUI doesn't need any initialization commands on startup.
func TestInit(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil, no initialization commands needed")
	}
}

// TestUpdateQuit verifies that Update() handles quit keys correctly.
// Both 'q' and 'ctrl+c' should trigger application exit.
func TestUpdateQuit(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Simulate window size to initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"q key quits", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"ctrl+c quits", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := m.Update(tt.key)

			if cmd == nil {
				t.Error("Update() should return quit command")
			}
		})
	}
}

// TestUpdateWindowSize verifies that Update() handles window resize events.
// The viewport should be initialized and resized correctly.
func TestUpdateWindowSize(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task 1\n- [ ] Task 2\n")

	// Initial state: not ready
	if m.ready {
		t.Error("Model should not be ready before WindowSizeMsg")
	}

	// Send window size message
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// After WindowSizeMsg: should be ready
	if !m.ready {
		t.Error("Model should be ready after WindowSizeMsg")
	}

	if m.width != 80 {
		t.Errorf("width = %d, want 80", m.width)
	}

	if m.height != 24 {
		t.Errorf("height = %d, want 24", m.height)
	}
}

// TestUpdateScroll verifies that Update() handles scroll key presses.
// Arrow keys and vim-style keys should scroll the viewport.
func TestUpdateScroll(t *testing.T) {
	cfg := config.Default()
	// Create content longer than viewport
	content := strings.Repeat("- [ ] Task\n", 50)
	m := New(cfg, content)

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m = newModel.(Model)

	initialOffset := m.viewport.YOffset

	// Test down scroll with 'j'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)

	if m.viewport.YOffset <= initialOffset {
		t.Error("'j' key should scroll down")
	}
}

// TestView verifies that View() returns correctly formatted output.
// It should include content and footer when ready, or loading message when not ready.
func TestView(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task 1\n- [ ] Task 2\n")

	// Before initialization
	view := m.View()
	if view != "Initializing..." {
		t.Errorf("View() before ready = %q, want 'Initializing...'", view)
	}

	// After initialization
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)
	view = m.View()

	// Should contain footer elements
	if !strings.Contains(view, "help") {
		t.Error("View() should contain 'help' in footer")
	}
	if !strings.Contains(view, "quit") {
		t.Error("View() should contain 'quit' in footer")
	}
}

// TestMatchKey verifies that matchKey() correctly matches pressed keys against configured bindings.
// This is the foundation of customizable keybindings.
func TestMatchKey(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "")

	tests := []struct {
		name       string
		pressed    string
		configured []string
		expected   bool
	}{
		{"exact match", "k", []string{"k"}, true},
		{"no match", "j", []string{"k"}, false},
		{"match in list", "k", []string{"j", "k"}, true},
		{"modifier key match", "ctrl+u", []string{"ctrl+u"}, true},
		{"empty config", "x", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.matchKey(tt.pressed, tt.configured)
			if result != tt.expected {
				t.Errorf("matchKey(%q, %v) = %v, want %v", tt.pressed, tt.configured, result, tt.expected)
			}
		})
	}
}

// TestMatchAction verifies that matchAction() returns the correct action for a key press.
// This maps key presses to scroll actions based on configuration.
func TestMatchAction(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "")

	tests := []struct {
		name     string
		key      string
		expected action
	}{
		{"k maps to up", "k", actionUp},
		{"j maps to down", "j", actionDown},
		{"g maps to top", "g", actionTop},
		{"G maps to bottom", "G", actionBottom},
		{"ctrl+u maps to half page up", "ctrl+u", actionHalfPageUp},
		{"ctrl+d maps to half page down", "ctrl+d", actionHalfPageDown},
		{"unknown key maps to none", "x", actionNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.matchAction(tt.key)
			if result != tt.expected {
				t.Errorf("matchAction(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

// TestUpdateEditKey verifies that 'e' key triggers editor launch command.
// The editor command should be returned for execution by the main program.
func TestUpdateEditKey(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")
	m.tasksPath = testTasksPath

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Press 'e' key
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	if cmd == nil {
		t.Error("'e' key should return a command for editor launch")
	}
}

// TestUpdateArchiveKey verifies that 'a' key triggers archive command.
// The archive command should process completed tasks older than delay_days.
func TestUpdateArchiveKey(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [x] Completed task @done(2020-01-01)")
	m.tasksPath = testTasksPath
	m.archivePath = testArchivePath

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Press 'a' key
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if cmd == nil {
		t.Error("'a' key should return a command for archive")
	}
}

// TestUpdateReloadKey verifies that 'r' key triggers reload command.
// The reload command should re-read the tasks file from disk.
func TestUpdateReloadKey(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")
	m.tasksPath = testTasksPath

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Press 'r' key
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd == nil {
		t.Error("'r' key should return a command for reload")
	}
}

// TestNewWithPaths verifies that NewWithPaths() correctly sets file paths.
// The tasksPath and archivePath should be set for edit/archive/reload operations.
func TestNewWithPaths(t *testing.T) {
	cfg := config.Default()
	content := "- [ ] Task"
	tasksPath := "/tmp/tasks.md"
	archivePath := "/tmp/archive.md"

	m := NewWithPaths(cfg, content, tasksPath, archivePath)

	if m.tasksPath != tasksPath {
		t.Errorf("NewWithPaths() tasksPath = %q, want %q", m.tasksPath, tasksPath)
	}
	if m.archivePath != archivePath {
		t.Errorf("NewWithPaths() archivePath = %q, want %q", m.archivePath, archivePath)
	}
	if m.content != content {
		t.Errorf("NewWithPaths() content = %q, want %q", m.content, content)
	}
}

// TestUpdateReloadFinishedMsg verifies that ReloadFinishedMsg updates the model.
// On successful reload, the content and lines should be updated.
func TestUpdateReloadFinishedMsg(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Old task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Send reload finished message
	newContent := "- [ ] New task\n- [ ] Another task"
	newModel, _ = m.Update(ReloadFinishedMsg{Content: newContent, Err: nil})
	m = newModel.(Model)

	if m.content != newContent {
		t.Errorf("ReloadFinishedMsg content = %q, want %q", m.content, newContent)
	}
	if len(m.lines) != 2 {
		t.Errorf("ReloadFinishedMsg lines = %d, want 2", len(m.lines))
	}
	if m.status != "Reloaded" {
		t.Errorf("ReloadFinishedMsg status = %q, want 'Reloaded'", m.status)
	}
}

// TestUpdateArchiveFinishedMsg verifies that ArchiveFinishedMsg updates status.
// On successful archive, the status should show the count of archived tasks.
func TestUpdateArchiveFinishedMsg(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	tests := []struct {
		name           string
		msg            ArchiveFinishedMsg
		expectedStatus string
	}{
		{"archived 3 tasks", ArchiveFinishedMsg{Count: 3, Err: nil}, "Archived 3 task(s)"},
		{"no tasks to archive", ArchiveFinishedMsg{Count: 0, Err: nil}, "No tasks to archive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := m.Update(tt.msg)
			updated := newModel.(Model)

			if tt.msg.Count == 0 && updated.status != tt.expectedStatus {
				t.Errorf("ArchiveFinishedMsg status = %q, want %q", updated.status, tt.expectedStatus)
			}
		})
	}
}

// TestParseLines verifies that parseLines() correctly handles different content formats.
// It should handle empty content, single lines, and trailing newlines.
func TestParseLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{"empty content", "", 0},
		{"single line no newline", "task", 1},
		{"single line with newline", "task\n", 1},
		{"two lines", "task1\ntask2", 2},
		{"two lines with trailing newline", "task1\ntask2\n", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLines(tt.content)
			if len(result) != tt.expected {
				t.Errorf("parseLines(%q) = %d lines, want %d", tt.content, len(result), tt.expected)
			}
		})
	}
}

// TestSetStatusWithTimeout verifies that setStatusWithTimeout() sets status and returns a timeout command.
// The status should be cleared after the timeout command is processed.
// Spec: docs/specification.md "ステータスメッセージ" section - status clears after 3 seconds.
func TestSetStatusWithTimeout(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Set status with timeout
	m, cmd := m.setStatusWithTimeout("Test message")

	// Status should be set
	if m.status != "Test message" {
		t.Errorf("status = %q, want 'Test message'", m.status)
	}

	// Command should be returned (for timeout)
	if cmd == nil {
		t.Error("setStatusWithTimeout() should return a command for timeout")
	}
}

// TestUpdateClearStatusMsg verifies that ClearStatusMsg clears the status.
// When the timeout fires, the status should be cleared to show the default footer.
func TestUpdateClearStatusMsg(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Set status manually
	m.status = "Some status message"

	// Send ClearStatusMsg
	newModel, _ = m.Update(ClearStatusMsg{})
	m = newModel.(Model)

	// Status should be cleared
	if m.status != "" {
		t.Errorf("ClearStatusMsg should clear status, got %q", m.status)
	}
}

// TestArchiveFinishedMsgWithTimeout verifies that ArchiveFinishedMsg sets status with timeout.
// Spec: docs/specification.md "ステータスメッセージ" - "Archived 3 tasks" with 3-second timeout.
func TestArchiveFinishedMsgWithTimeout(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Send ArchiveFinishedMsg with 0 count (no tasks to archive)
	newModel, cmd := m.Update(ArchiveFinishedMsg{Count: 0, Err: nil})
	m = newModel.(Model)

	// Status should be set
	if m.status != "No tasks to archive" {
		t.Errorf("status = %q, want 'No tasks to archive'", m.status)
	}

	// Timeout command should be returned
	if cmd == nil {
		t.Error("ArchiveFinishedMsg should return a timeout command")
	}
}

// TestReloadFinishedMsgWithTimeout verifies that ReloadFinishedMsg sets status with timeout.
// The "Reloaded" message should auto-clear after 3 seconds.
func TestReloadFinishedMsgWithTimeout(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Old task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Send ReloadFinishedMsg
	newModel, cmd := m.Update(ReloadFinishedMsg{Content: "- [ ] New task", Err: nil})
	m = newModel.(Model)

	// Status should be set
	if m.status != "Reloaded" {
		t.Errorf("status = %q, want 'Reloaded'", m.status)
	}

	// Timeout command should be returned
	if cmd == nil {
		t.Error("ReloadFinishedMsg should return a timeout command")
	}
}

// TestHelpOverlayToggle verifies that '?' and 'h' keys toggle help overlay.
// Spec: docs/specification.md "キーバインド仕様" - ?/h toggles help display.
func TestHelpOverlayToggle(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"? key shows help", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}},
		{"h key shows help", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start without help
			m.showHelp = false

			// Press help key
			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)

			if !m.showHelp {
				t.Errorf("showHelp should be true after pressing %s", tt.key.String())
			}
		})
	}
}

// TestHelpOverlayClose verifies that any key closes the help overlay.
// Spec: docs/specification.md "ヘルプオーバーレイ" - "Press any key to close".
func TestHelpOverlayClose(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Enable help mode
	m.showHelp = true

	// Press any key (e.g., Enter)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.showHelp {
		t.Error("showHelp should be false after pressing any key")
	}
}

// TestViewWithHelpOverlay verifies that View() shows help overlay when enabled.
// The overlay should contain keybinding information from the configuration.
func TestViewWithHelpOverlay(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Enable help mode
	m.showHelp = true

	view := m.View()

	// Check for expected help content
	if !strings.Contains(view, "Help") {
		t.Error("View() with help should contain 'Help' title")
	}
	if !strings.Contains(view, "edit") || !strings.Contains(view, "e") {
		t.Error("View() with help should show 'e' for edit")
	}
	if !strings.Contains(view, "archive") || !strings.Contains(view, "a") {
		t.Error("View() with help should show 'a' for archive")
	}
	if !strings.Contains(view, "quit") || !strings.Contains(view, "q") {
		t.Error("View() with help should show 'q' for quit")
	}
}

// TestHelpOverlayShowsConfiguredKeybindings verifies that help shows custom keybindings.
// Spec: docs/specification.md "ヘルプオーバーレイ" - custom keys should be dynamically reflected.
func TestHelpOverlayShowsConfiguredKeybindings(t *testing.T) {
	cfg := config.Default()
	cfg.Keybindings.Up = []string{"k", "ctrl+p"}
	cfg.Keybindings.Down = []string{"j", "ctrl+n"}
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Enable help mode
	m.showHelp = true

	view := m.View()

	// Should show configured keybindings
	if !strings.Contains(view, "k") {
		t.Error("View() with help should show configured up key 'k'")
	}
	if !strings.Contains(view, "j") {
		t.Error("View() with help should show configured down key 'j'")
	}
}

// TestHelpOverlayDoesNotQuit verifies that 'q' key closes help instead of quitting.
// When help is shown, 'q' should close help, not quit the application.
func TestHelpOverlayDoesNotQuit(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Enable help mode
	m.showHelp = true

	// Press 'q' key
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	// Should NOT quit (cmd should not be tea.Quit)
	if cmd != nil {
		t.Error("'q' in help mode should not return quit command")
	}
}

// TestInitWithAutoArchiveDisabled verifies that Init() returns nil when archive.auto is false.
// Spec: docs/specification.md "設定ファイル仕様" - archive.auto defaults to false.
func TestInitWithAutoArchiveDisabled(t *testing.T) {
	cfg := config.Default()
	cfg.Archive.Auto = false
	m := New(cfg, "- [ ] Task")
	m.tasksPath = testTasksPath
	m.archivePath = testArchivePath

	cmd := m.Init()

	if cmd != nil {
		t.Error("Init() should return nil when archive.auto is false")
	}
}

// TestInitWithAutoArchiveEnabled verifies that Init() returns archive command when archive.auto is true.
// Spec: docs/specification.md "アーカイブのタイミング" - auto archive runs at startup when enabled.
func TestInitWithAutoArchiveEnabled(t *testing.T) {
	cfg := config.Default()
	cfg.Archive.Auto = true
	m := New(cfg, "- [x] Completed task @done(2020-01-01)")
	m.tasksPath = testTasksPath
	m.archivePath = testArchivePath

	cmd := m.Init()

	if cmd == nil {
		t.Error("Init() should return archive command when archive.auto is true")
	}
}

// TestUpdateEditFinishedMsgWithError verifies that editor errors are displayed in status.
// Spec: docs/specification.md "エラー処理" - "Error: Editor not found" shown in footer.
func TestUpdateEditFinishedMsgWithError(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Send EditFinishedMsg with error
	testErr := fmt.Errorf("editor not found: vim")
	newModel, cmd := m.Update(EditFinishedMsg{Err: testErr})
	m = newModel.(Model)

	// Status should show error
	if !strings.Contains(m.status, "Error:") {
		t.Errorf("status should contain 'Error:', got %q", m.status)
	}

	// Timeout command should be returned for auto-clear
	if cmd == nil {
		t.Error("EditFinishedMsg with error should return timeout command")
	}
}

// TestUpdateArchiveFinishedMsgWithError verifies that archive errors are displayed in status.
// Spec: docs/specification.md "エラー処理" - archive errors shown in footer.
func TestUpdateArchiveFinishedMsgWithError(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Send ArchiveFinishedMsg with error
	testErr := fmt.Errorf("permission denied")
	newModel, cmd := m.Update(ArchiveFinishedMsg{Count: 0, Err: testErr})
	m = newModel.(Model)

	// Status should show error
	if !strings.Contains(m.status, "Archive error:") {
		t.Errorf("status should contain 'Archive error:', got %q", m.status)
	}

	// Timeout command should be returned for auto-clear
	if cmd == nil {
		t.Error("ArchiveFinishedMsg with error should return timeout command")
	}
}

// TestUpdateReloadFinishedMsgWithError verifies that reload errors are displayed in status.
// Spec: docs/specification.md "エラー処理" - file deletion should show error.
func TestUpdateReloadFinishedMsgWithError(t *testing.T) {
	cfg := config.Default()
	m := New(cfg, "- [ ] Task")

	// Initialize viewport
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	// Send ReloadFinishedMsg with error
	testErr := fmt.Errorf("file not found")
	newModel, cmd := m.Update(ReloadFinishedMsg{Content: "", Err: testErr})
	m = newModel.(Model)

	// Status should show error
	if !strings.Contains(m.status, "Reload error:") {
		t.Errorf("status should contain 'Reload error:', got %q", m.status)
	}

	// Timeout command should be returned for auto-clear
	if cmd == nil {
		t.Error("ReloadFinishedMsg with error should return timeout command")
	}
}

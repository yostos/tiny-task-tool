// Package tui implements the terminal user interface using bubbletea.
package tui

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yostos/tiny-task-tool/internal/cli"
	"github.com/yostos/tiny-task-tool/internal/config"
	"github.com/yostos/tiny-task-tool/internal/task"
)

// statusTimeout is the duration after which status messages auto-clear.
const statusTimeout = 3 * time.Second

// Model represents the TUI application state.
type Model struct {
	config      *config.Config
	content     string
	lines       []string
	viewport    viewport.Model
	ready       bool
	width       int
	height      int
	err         error
	status      string
	tasksPath   string
	archivePath string
	showHelp    bool
}

// New creates a new TUI model.
func New(cfg *config.Config, content string) Model {
	// Count actual lines (trim trailing newline to avoid empty line count)
	trimmed := strings.TrimSuffix(content, "\n")
	var lines []string
	if trimmed == "" {
		lines = []string{}
	} else {
		lines = strings.Split(trimmed, "\n")
	}
	return Model{
		config:  cfg,
		content: content,
		lines:   lines,
	}
}

// NewWithPaths creates a new TUI model with file paths for edit/archive/reload.
func NewWithPaths(cfg *config.Config, content, tasksPath, archivePath string) Model {
	m := New(cfg, content)
	m.tasksPath = tasksPath
	m.archivePath = archivePath
	return m
}

// Init initializes the model.
// Always adds @done tags to completed tasks at startup.
// If archive.auto is enabled, also runs auto-archive.
func (m Model) Init() tea.Cmd {
	if m.config.Archive.Auto {
		return m.archiveCmd()
	}
	return m.addDoneTagsCmd()
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 0
		footerHeight := 1
		verticalMargins := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMargins)
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMargins
		}

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case ClearStatusMsg:
		m.status = ""
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case EditFinishedMsg:
		if msg.Err != nil {
			m, cmd := m.setStatusWithTimeout("Error: " + msg.Err.Error())
			return m, cmd
		}
		// Add @done tags, then reload
		return m, m.addDoneTagsAndReloadCmd()

	case ArchiveFinishedMsg:
		if msg.Err != nil {
			m, cmd := m.setStatusWithTimeout("Archive error: " + msg.Err.Error())
			return m, cmd
		}
		if msg.Count > 0 {
			m.status = "Archived " + strconv.Itoa(msg.Count) + " task(s)"
			// Reload to show updated content, status will be set with timeout after reload
			return m, m.reloadCmd()
		}
		m, cmd := m.setStatusWithTimeout("No tasks to archive")
		return m, cmd

	case ReloadFinishedMsg:
		if msg.Err != nil {
			m, cmd := m.setStatusWithTimeout("Reload error: " + msg.Err.Error())
			return m, cmd
		}
		m.content = msg.Content
		m.lines = parseLines(msg.Content)
		m.viewport.SetContent(msg.Content)
		m, cmd := m.setStatusWithTimeout("Reloaded")
		return m, cmd

	case AddDoneTagsFinishedMsg:
		if msg.Err != nil {
			m, cmd := m.setStatusWithTimeout("Error: " + msg.Err.Error())
			return m, cmd
		}
		if msg.Count > 0 {
			m.status = strconv.Itoa(msg.Count) + " task(s) marked as done"
			// Reload to show updated content, status will be set with timeout after reload
			return m, m.reloadCmd()
		}
		// No tasks modified, just reload
		return m, m.reloadCmd()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// handleKeyPress processes key press events.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// If help overlay is shown, any key closes it
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	// Fixed keybindings (not configurable)
	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up":
		m.viewport.ScrollUp(1)
	case "down":
		m.viewport.ScrollDown(1)
	case "e":
		return m, m.editCmd()
	case "a":
		return m, m.archiveCmd()
	case "r":
		return m, m.reloadCmd()
	case "?", "h":
		m.showHelp = true
		return m, nil
	}

	// Configurable keybindings
	action := m.matchAction(key)
	switch action {
	case actionUp:
		m.viewport.ScrollUp(1)
	case actionDown:
		m.viewport.ScrollDown(1)
	case actionTop:
		m.viewport.GotoTop()
	case actionBottom:
		m.viewport.GotoBottom()
	case actionHalfPageUp:
		m.viewport.HalfPageUp()
	case actionHalfPageDown:
		m.viewport.HalfPageDown()
	}

	return m, nil
}

// action represents a keybinding action.
type action int

const (
	actionNone action = iota
	actionUp
	actionDown
	actionTop
	actionBottom
	actionHalfPageUp
	actionHalfPageDown
)

// matchAction returns the action for the pressed key.
func (m Model) matchAction(key string) action {
	switch {
	case m.matchKey(key, m.config.Keybindings.Up):
		return actionUp
	case m.matchKey(key, m.config.Keybindings.Down):
		return actionDown
	case m.matchKey(key, m.config.Keybindings.Top):
		return actionTop
	case m.matchKey(key, m.config.Keybindings.Bottom):
		return actionBottom
	case m.matchKey(key, m.config.Keybindings.HalfPageUp):
		return actionHalfPageUp
	case m.matchKey(key, m.config.Keybindings.HalfPageDown):
		return actionHalfPageDown
	default:
		return actionNone
	}
}

// matchKey checks if the pressed key matches any of the configured keys.
func (m Model) matchKey(pressed string, configured []string) bool {
	for _, k := range configured {
		if pressed == k {
			return true
		}
	}
	return false
}

// View renders the UI.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	base := m.viewport.View() + "\n" + m.footerView()

	if m.showHelp {
		return m.overlayHelp(base)
	}

	return base
}

// footerView renders the footer bar.
func (m Model) footerView() string {
	style := lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("252")).
		Width(m.width)

	// Left side: key hints or status message
	var left string
	if m.status != "" {
		left = m.status
	} else {
		left = "? help | e edit | a archive | q quit"
	}

	// Right side: scroll position and version
	totalLines := len(m.lines)
	currentLine := m.viewport.YOffset + 1
	if currentLine > totalLines {
		currentLine = totalLines
	}
	if totalLines == 0 {
		totalLines = 1
		currentLine = 1
	}
	position := formatPosition(currentLine, totalLines)
	version := "ttt " + cli.Version
	right := lipgloss.NewStyle().
		Align(lipgloss.Right).
		Render(position + " " + version)

	// Calculate padding
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := m.width - leftWidth - rightWidth
	if padding < 0 {
		padding = 0
	}

	footer := left + strings.Repeat(" ", padding) + right
	return style.Render(footer)
}

func formatPosition(current, total int) string {
	return "[" + itoa(current) + "/" + itoa(total) + "]"
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	idx := len(b)
	for i > 0 {
		idx--
		b[idx] = byte('0' + i%10)
		i /= 10
	}
	return string(b[idx:])
}

// parseLines splits content into lines, handling trailing newlines.
func parseLines(content string) []string {
	trimmed := strings.TrimSuffix(content, "\n")
	if trimmed == "" {
		return []string{}
	}
	return strings.Split(trimmed, "\n")
}

// Message types
type statusMsg string
type errMsg struct{ err error }

// ClearStatusMsg is sent when the status message timeout expires.
type ClearStatusMsg struct{}

// EditFinishedMsg is sent when the editor closes.
type EditFinishedMsg struct{ Err error }

// ArchiveFinishedMsg is sent when archiving completes.
type ArchiveFinishedMsg struct {
	Count int
	Err   error
}

// ReloadFinishedMsg is sent when reload completes.
type ReloadFinishedMsg struct {
	Content string
	Err     error
}

// AddDoneTagsFinishedMsg is sent when adding @done tags completes.
type AddDoneTagsFinishedMsg struct {
	Count int
	Err   error
}

// editCmd returns a command that launches the external editor.
// It uses tea.ExecProcess to suspend the TUI and run the editor.
func (m Model) editCmd() tea.Cmd {
	editorCmd := m.config.EditorCommand(m.tasksPath)
	// Parse the command to get program and args
	parts := strings.Fields(editorCmd)
	if len(parts) == 0 {
		return func() tea.Msg {
			return EditFinishedMsg{Err: nil}
		}
	}
	c := exec.Command(parts[0], parts[1:]...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return EditFinishedMsg{Err: err}
	})
}

// archiveCmd returns a command that archives old completed tasks.
func (m Model) archiveCmd() tea.Cmd {
	tasksPath := m.tasksPath
	archivePath := m.archivePath
	delayDays := m.config.Archive.DelayDays

	return func() tea.Msg {
		// First, add @done tags to newly completed tasks
		_, err := task.ProcessFileWithDoneTags(tasksPath)
		if err != nil {
			return ArchiveFinishedMsg{Count: 0, Err: err}
		}

		// Then archive old completed tasks
		count, err := task.Archive(tasksPath, archivePath, delayDays)
		return ArchiveFinishedMsg{Count: count, Err: err}
	}
}

// reloadCmd returns a command that reloads the tasks file.
func (m Model) reloadCmd() tea.Cmd {
	tasksPath := m.tasksPath

	return func() tea.Msg {
		content, err := task.LoadFile(tasksPath)
		return ReloadFinishedMsg{Content: content, Err: err}
	}
}

// addDoneTagsCmd returns a command that adds @done tags to completed tasks.
func (m Model) addDoneTagsCmd() tea.Cmd {
	tasksPath := m.tasksPath

	return func() tea.Msg {
		count, err := task.ProcessFileWithDoneTags(tasksPath)
		return AddDoneTagsFinishedMsg{Count: count, Err: err}
	}
}

// addDoneTagsAndReloadCmd returns a command that adds @done tags and then reloads.
func (m Model) addDoneTagsAndReloadCmd() tea.Cmd {
	tasksPath := m.tasksPath

	return func() tea.Msg {
		count, err := task.ProcessFileWithDoneTags(tasksPath)
		if err != nil {
			return AddDoneTagsFinishedMsg{Count: 0, Err: err}
		}
		return AddDoneTagsFinishedMsg{Count: count, Err: nil}
	}
}

// setStatusWithTimeout sets the status message and returns a command that clears it after timeout.
func (m Model) setStatusWithTimeout(status string) (Model, tea.Cmd) {
	m.status = status
	return m, tea.Tick(statusTimeout, func(t time.Time) tea.Msg {
		return ClearStatusMsg{}
	})
}

// overlayHelp renders the help overlay on top of the base view.
func (m Model) overlayHelp(base string) string {
	// Build help content with configured keybindings
	upKeys := formatKeys(m.config.Keybindings.Up, "↑")
	downKeys := formatKeys(m.config.Keybindings.Down, "↓")
	topKeys := formatKeys(m.config.Keybindings.Top, "")
	bottomKeys := formatKeys(m.config.Keybindings.Bottom, "")
	halfPageUpKeys := formatKeys(m.config.Keybindings.HalfPageUp, "")
	halfPageDownKeys := formatKeys(m.config.Keybindings.HalfPageDown, "")

	helpLines := []string{
		"",
		"  " + padRight(upKeys, 12) + "上へスクロール",
		"  " + padRight(downKeys, 12) + "下へスクロール",
		"  " + padRight(topKeys, 12) + "先頭へ移動",
		"  " + padRight(bottomKeys, 12) + "末尾へ移動",
		"  " + padRight(halfPageUpKeys, 12) + "半画面上へ",
		"  " + padRight(halfPageDownKeys, 12) + "半画面下へ",
		"",
		"  " + padRight("e", 12) + "エディタ起動",
		"  " + padRight("a", 12) + "アーカイブ実行",
		"  " + padRight("r", 12) + "再読み込み",
		"",
		"  " + padRight("q", 12) + "終了",
		"  " + padRight("?/h", 12) + "ヘルプ",
		"",
		"  Press any key to close",
	}

	helpContent := strings.Join(helpLines, "\n")

	// Style for help overlay box
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Width(36)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Width(32)

	helpBox := helpStyle.Render(titleStyle.Render("Help") + helpContent)

	// Center the help box on screen
	helpWidth := lipgloss.Width(helpBox)
	helpHeight := lipgloss.Height(helpBox)

	x := (m.width - helpWidth) / 2
	y := (m.height - helpHeight) / 2

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Overlay the help box on the base view
	return placeOverlay(x, y, helpBox, base)
}

// formatKeys formats keybindings for display, prepending arrow key if provided.
func formatKeys(keys []string, arrowKey string) string {
	if arrowKey != "" && len(keys) > 0 {
		return arrowKey + "/" + strings.Join(keys, "/")
	}
	if len(keys) > 0 {
		return strings.Join(keys, "/")
	}
	return arrowKey
}

// padRight pads a string to the given display width.
// Uses lipgloss.Width to correctly handle multibyte characters (e.g., Japanese).
func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// placeOverlay places an overlay string on top of a background at the given position.
// Correctly handles multibyte characters by using display width calculations.
func placeOverlay(x, y int, overlay, background string) string {
	bgLines := strings.Split(background, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Ensure background has enough lines
	for len(bgLines) < y+len(overlayLines) {
		bgLines = append(bgLines, "")
	}

	for i, overlayLine := range overlayLines {
		bgIdx := y + i
		if bgIdx < 0 || bgIdx >= len(bgLines) {
			continue
		}

		bgLine := bgLines[bgIdx]

		// Get the part before the overlay (up to x display width)
		beforeOverlay := truncateByDisplayWidth(bgLine, x)

		// Get the part after the overlay
		overlayWidth := lipgloss.Width(overlayLine)
		afterOverlay := skipByDisplayWidth(bgLine, x+overlayWidth)

		bgLines[bgIdx] = beforeOverlay + overlayLine + afterOverlay
	}

	return strings.Join(bgLines, "\n")
}

// truncateByDisplayWidth returns the prefix of s that fits within the given display width.
// If s is shorter than width, it pads with spaces.
func truncateByDisplayWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}

	var result strings.Builder
	currentWidth := 0

	for _, r := range s {
		runeWidth := lipgloss.Width(string(r))
		if currentWidth+runeWidth > width {
			break
		}
		result.WriteRune(r)
		currentWidth += runeWidth
	}

	// Pad with spaces if needed
	if currentWidth < width {
		result.WriteString(strings.Repeat(" ", width-currentWidth))
	}

	return result.String()
}

// skipByDisplayWidth returns the suffix of s after skipping the given display width.
func skipByDisplayWidth(s string, width int) string {
	if width <= 0 {
		return s
	}

	currentWidth := 0
	for i, r := range s {
		if currentWidth >= width {
			return s[i:]
		}
		currentWidth += lipgloss.Width(string(r))
	}

	return ""
}

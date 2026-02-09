// Package display provides syntax highlighting and statistics for task display.
package display

import (
	"strings"
	"testing"
	"time"

	"github.com/yostos/tiny-task-tool/internal/config"
)

// TestNewRenderer verifies that NewRenderer creates a renderer with default styles.
// Spec: docs/specification.md "Colors and Styling (v0.4.0)" section.
func TestNewRenderer(t *testing.T) {
	cfg := config.Default()
	r := NewRenderer(cfg)

	if r == nil {
		t.Fatal("NewRenderer should not return nil")
	}
	if r.cfg == nil {
		t.Error("Renderer should have a config reference")
	}
}

// TestClassifyLineHeading verifies that Markdown headings are classified correctly.
// Spec: Headings (#, ##, etc.) should be styled as LineHeading.
func TestClassifyLineHeading(t *testing.T) {
	cfg := config.Default()
	r := NewRenderer(cfg)
	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		line     string
		expected LineType
	}{
		{"H1 heading", "# Today", LineHeading},
		{"H2 heading", "## Notes", LineHeading},
		{"H3 heading", "### Section", LineHeading},
		{"H6 heading", "###### Deep", LineHeading},
		{"Not a heading (no space)", "#NoSpace", LineNormal},
		{"Not a heading (just hash)", "#", LineNormal},
		{"Empty line", "", LineNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.ClassifyLine(tt.line, today)
			if result != tt.expected {
				t.Errorf("ClassifyLine(%q) = %d, want %d", tt.line, result, tt.expected)
			}
		})
	}
}

// TestClassifyLineTask verifies that tasks are classified correctly.
// Spec: Incomplete tasks (- [ ]) and completed tasks (- [x]) should be styled differently.
func TestClassifyLineTask(t *testing.T) {
	cfg := config.Default()
	r := NewRenderer(cfg)
	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		line     string
		expected LineType
	}{
		{"Incomplete task", "- [ ] Buy milk", LineIncomplete},
		{"Completed task", "- [x] Done task", LineCompleted},
		{"Completed task with done tag", "- [x] Done task @done(2026-02-09)", LineCompleted},
		{"Indented incomplete", "  - [ ] Sub task", LineIncomplete},
		{"Indented completed", "  - [x] Sub done", LineCompleted},
		{"Not a task (bullet only)", "- Regular text", LineNormal},
		{"Not a task (no checkbox)", "  - Note item", LineNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.ClassifyLine(tt.line, today)
			if result != tt.expected {
				t.Errorf("ClassifyLine(%q) = %d, want %d", tt.line, result, tt.expected)
			}
		})
	}
}

// TestClassifyLineDueDate verifies that @due tag is parsed and classified correctly.
// Spec: docs/specification.md "Due Date Display (@due tag) (v0.4.0)" section.
func TestClassifyLineDueDate(t *testing.T) {
	cfg := config.Default()
	cfg.Display.DueWarningDays = 7
	r := NewRenderer(cfg)

	// Reference date: 2026-02-09
	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		line     string
		expected LineType
	}{
		// Overdue (past due date)
		{"Overdue - yesterday", "- [ ] Task @due(2026-02-08)", LineOverdue},
		{"Overdue - last week", "- [ ] Task @due(2026-02-01)", LineOverdue},

		// Due soon (within 7 days)
		{"Due soon - today", "- [ ] Task @due(2026-02-09)", LineDueSoon},
		{"Due soon - tomorrow", "- [ ] Task @due(2026-02-10)", LineDueSoon},
		{"Due soon - in 7 days", "- [ ] Task @due(2026-02-16)", LineDueSoon},

		// Due future (beyond 7 days)
		{"Due future - in 8 days", "- [ ] Task @due(2026-02-17)", LineDueFuture},
		{"Due future - next month", "- [ ] Task @due(2026-03-01)", LineDueFuture},

		// Completed tasks with @due should still be LineCompleted
		{"Completed with due", "- [x] Task @due(2026-02-08) @done(2026-02-07)", LineCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.ClassifyLine(tt.line, today)
			if result != tt.expected {
				t.Errorf("ClassifyLine(%q) = %d, want %d", tt.line, result, tt.expected)
			}
		})
	}
}

// TestColorFromName verifies that color names are converted to lipgloss colors correctly.
// Spec: docs/specification.md "Color Configuration (v0.4.0)" - available color names.
func TestColorFromName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"black", "0"},
		{"red", "1"},
		{"green", "2"},
		{"yellow", "3"},
		{"blue", "4"},
		{"magenta", "5"},
		{"cyan", "6"},
		{"white", "7"},
		{"gray", "8"},
		{"grey", "8"},
		{"bright_red", "9"},
		{"bright_green", "10"},
		{"CYAN", "6"},    // Case insensitive
		{"White", "7"},   // Case insensitive
		{"unknown", "7"}, // Unknown defaults to white
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := string(colorFromName(tt.name))
			if result != tt.expected {
				t.Errorf("colorFromName(%q) = %q, want %q", tt.name, result, tt.expected)
			}
		})
	}
}

// TestRenderLine verifies that lines are rendered with appropriate styles.
// This test checks that rendering doesn't panic and returns non-empty output.
func TestRenderLine(t *testing.T) {
	cfg := config.Default()
	r := NewRenderer(cfg)
	today := time.Now()

	tests := []struct {
		name string
		line string
	}{
		{"Heading", "# Today"},
		{"Incomplete task", "- [ ] Buy milk"},
		{"Completed task", "- [x] Done"},
		{"Normal text", "Some regular text"},
		{"Empty line", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.RenderLine(tt.line, today)
			// Rendered output should contain the original text (possibly with ANSI codes)
			if tt.line != "" && !strings.Contains(result, tt.line) {
				t.Errorf("RenderLine(%q) should contain original text", tt.line)
			}
		})
	}
}

// TestRenderContent verifies that multi-line content is rendered correctly.
func TestRenderContent(t *testing.T) {
	cfg := config.Default()
	r := NewRenderer(cfg)
	today := time.Now()

	content := `# Today

- [ ] Buy milk
- [x] Review PR @done(2026-02-09)

## Notes

- Regular note`

	result := r.RenderContent(content, today)

	// Result should have same number of lines
	inputLines := strings.Split(content, "\n")
	outputLines := strings.Split(result, "\n")
	if len(inputLines) != len(outputLines) {
		t.Errorf("RenderContent changed line count: got %d, want %d", len(outputLines), len(inputLines))
	}
}

// TestCalculateStatistics verifies that task statistics are calculated correctly.
// Spec: docs/specification.md "Footer" and "Statistics format (v0.4.0)" sections.
func TestCalculateStatistics(t *testing.T) {
	cfg := config.Default()
	cfg.Display.DueWarningDays = 7
	r := NewRenderer(cfg)

	// Reference date: 2026-02-09
	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.Local)

	content := `# Today

- [ ] Buy milk
- [ ] Task with due soon @due(2026-02-10)
- [ ] Overdue task @due(2026-02-01)
- [x] Done today @done(2026-02-09)
- [x] Done yesterday @done(2026-02-08)
  - [ ] Subtask incomplete
  - [x] Subtask done @done(2026-02-09)

## Notes

- Regular note`

	stats := r.CalculateStatistics(content, today)

	// Expected values
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"Total", stats.Total, 7},
		{"Incomplete", stats.Incomplete, 4},
		{"Completed", stats.Completed, 3},
		{"CompletedToday", stats.CompletedToday, 2},
		{"Overdue", stats.Overdue, 1},
		{"DueSoon", stats.DueSoon, 1},
		{"RootTasks", stats.RootTasks, 5},
		{"Subtasks", stats.Subtasks, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("stats.%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestIsRootTask verifies root task detection.
func TestIsRootTask(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"Root task", "- [ ] Task", true},
		{"Root task completed", "- [x] Task", true},
		{"Subtask with spaces", "  - [ ] Subtask", false},
		{"Subtask with tab", "\t- [ ] Subtask", false},
		{"Deep subtask", "    - [ ] Deep", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRootTask(tt.line)
			if result != tt.expected {
				t.Errorf("isRootTask(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

// TestRendererWithCustomColors verifies that custom colors from config are applied.
func TestRendererWithCustomColors(t *testing.T) {
	cfg := config.Default()
	cfg.Colors.Heading = "magenta"
	cfg.Colors.Completed = "bright_black"
	cfg.Colors.Overdue = "bright_red"

	r := NewRenderer(cfg)

	if r == nil {
		t.Fatal("NewRenderer should not return nil with custom colors")
	}

	// Verify styles are initialized (we can't easily inspect lipgloss styles,
	// but we can verify rendering doesn't panic)
	today := time.Now()
	result := r.RenderLine("# Heading", today)
	if result == "" {
		t.Error("RenderLine should return non-empty result for heading")
	}
}

// TestDueWarningDaysConfig verifies that due_warning_days configuration is respected.
func TestDueWarningDaysConfig(t *testing.T) {
	// Test with 3-day warning threshold
	cfg := config.Default()
	cfg.Display.DueWarningDays = 3
	r := NewRenderer(cfg)

	today := time.Date(2026, 2, 9, 0, 0, 0, 0, time.Local)

	tests := []struct {
		name     string
		line     string
		expected LineType
	}{
		// Within 3 days = due soon
		{"Due in 1 day", "- [ ] Task @due(2026-02-10)", LineDueSoon},
		{"Due in 3 days", "- [ ] Task @due(2026-02-12)", LineDueSoon},
		// Beyond 3 days = future
		{"Due in 4 days", "- [ ] Task @due(2026-02-13)", LineDueFuture},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.ClassifyLine(tt.line, today)
			if result != tt.expected {
				t.Errorf("ClassifyLine(%q) with 3-day warning = %d, want %d", tt.line, result, tt.expected)
			}
		})
	}
}

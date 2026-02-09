// Package display provides syntax highlighting and statistics for task display.
package display

import (
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/yostos/tiny-task-tool/internal/config"
)

// LineType represents the type of a line for styling purposes.
type LineType int

const (
	// LineNormal is a regular line with no special styling.
	LineNormal LineType = iota
	// LineHeading is a Markdown heading line (starts with #).
	LineHeading
	// LineIncomplete is an incomplete task (- [ ]).
	LineIncomplete
	// LineCompleted is a completed task (- [x]).
	LineCompleted
	// LineOverdue is an incomplete task with a past due date.
	LineOverdue
	// LineDueSoon is an incomplete task with a due date within the warning threshold.
	LineDueSoon
	// LineDueFuture is an incomplete task with a due date beyond the warning threshold.
	LineDueFuture
)

// Statistics holds task statistics for display.
type Statistics struct {
	Total          int // Total number of tasks
	Incomplete     int // Number of incomplete tasks (- [ ])
	Completed      int // Number of completed tasks (- [x])
	CompletedToday int // Number of tasks completed today
	Overdue        int // Number of overdue tasks
	DueSoon        int // Number of tasks due soon (within warning threshold)
	RootTasks      int // Number of root-level tasks (no indentation)
	Subtasks       int // Number of subtasks (with indentation)
}

// Renderer applies syntax highlighting to task content.
type Renderer struct {
	cfg             *config.Config
	headingStyle    lipgloss.Style
	incompleteStyle lipgloss.Style
	completedStyle  lipgloss.Style
	overdueStyle    lipgloss.Style
	dueSoonStyle    lipgloss.Style
	dueFutureStyle  lipgloss.Style
}

// Regular expressions for parsing.
var (
	headingRegex    = regexp.MustCompile(`^(#{1,6})\s`)
	incompleteRegex = regexp.MustCompile(`^(\s*)-\s*\[\s*\]`)
	completedRegex  = regexp.MustCompile(`^(\s*)-\s*\[x\]`)
	dueRegex        = regexp.MustCompile(`@due\((\d{4}-\d{2}-\d{2})\)`)
	doneRegex       = regexp.MustCompile(`@done\((\d{4}-\d{2}-\d{2})\)`)
)

// NewRenderer creates a new Renderer with the given configuration.
func NewRenderer(cfg *config.Config) *Renderer {
	r := &Renderer{cfg: cfg}
	r.initStyles()
	return r
}

// initStyles initializes lipgloss styles from configuration.
func (r *Renderer) initStyles() {
	r.headingStyle = lipgloss.NewStyle().
		Foreground(colorFromName(r.cfg.Colors.Heading)).
		Bold(true)

	r.incompleteStyle = lipgloss.NewStyle().
		Foreground(colorFromName(r.cfg.Colors.Incomplete))

	r.completedStyle = lipgloss.NewStyle().
		Foreground(colorFromName(r.cfg.Colors.Completed))

	r.overdueStyle = lipgloss.NewStyle().
		Foreground(colorFromName(r.cfg.Colors.Overdue))

	r.dueSoonStyle = lipgloss.NewStyle().
		Foreground(colorFromName(r.cfg.Colors.DueSoon))

	r.dueFutureStyle = lipgloss.NewStyle().
		Foreground(colorFromName(r.cfg.Colors.DueFuture))
}

// colorFromName converts a color name to a lipgloss.Color.
func colorFromName(name string) lipgloss.Color {
	colorMap := map[string]string{
		"black":          "0",
		"red":            "1",
		"green":          "2",
		"yellow":         "3",
		"blue":           "4",
		"magenta":        "5",
		"cyan":           "6",
		"white":          "7",
		"gray":           "8",
		"grey":           "8",
		"bright_black":   "8",
		"bright_red":     "9",
		"bright_green":   "10",
		"bright_yellow":  "11",
		"bright_blue":    "12",
		"bright_magenta": "13",
		"bright_cyan":    "14",
		"bright_white":   "15",
	}

	if code, ok := colorMap[strings.ToLower(name)]; ok {
		return lipgloss.Color(code)
	}
	// Default to white if unknown
	return lipgloss.Color("7")
}

// ClassifyLine determines the type of a line for styling.
func (r *Renderer) ClassifyLine(line string, today time.Time) LineType {
	// Check for heading
	if headingRegex.MatchString(line) {
		return LineHeading
	}

	// Check for completed task
	if completedRegex.MatchString(line) {
		return LineCompleted
	}

	// Check for incomplete task
	if incompleteRegex.MatchString(line) {
		// Check for @due tag
		if match := dueRegex.FindStringSubmatch(line); match != nil {
			dueDate, err := time.Parse("2006-01-02", match[1])
			if err == nil {
				return r.classifyByDueDate(dueDate, today)
			}
		}
		return LineIncomplete
	}

	return LineNormal
}

// classifyByDueDate determines the line type based on due date.
func (r *Renderer) classifyByDueDate(dueDate, today time.Time) LineType {
	// Normalize to date only (no time component)
	dueDate = time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, time.Local)
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.Local)

	if dueDate.Before(today) {
		return LineOverdue
	}

	warningDays := r.cfg.Display.DueWarningDays
	warningThreshold := today.AddDate(0, 0, warningDays)

	if dueDate.Before(warningThreshold) || dueDate.Equal(warningThreshold) {
		return LineDueSoon
	}

	return LineDueFuture
}

// RenderLine applies syntax highlighting to a single line.
func (r *Renderer) RenderLine(line string, today time.Time) string {
	lineType := r.ClassifyLine(line, today)
	return r.applyStyle(line, lineType)
}

// applyStyle applies the appropriate style to a line.
func (r *Renderer) applyStyle(line string, lineType LineType) string {
	switch lineType {
	case LineHeading:
		return r.headingStyle.Render(line)
	case LineCompleted:
		return r.completedStyle.Render(line)
	case LineOverdue:
		return r.overdueStyle.Render(line)
	case LineDueSoon:
		return r.dueSoonStyle.Render(line)
	case LineDueFuture:
		return r.dueFutureStyle.Render(line)
	case LineIncomplete:
		return r.incompleteStyle.Render(line)
	default:
		return line
	}
}

// RenderContent applies syntax highlighting to all lines in the content.
func (r *Renderer) RenderContent(content string, today time.Time) string {
	lines := strings.Split(content, "\n")
	rendered := make([]string, len(lines))
	for i, line := range lines {
		rendered[i] = r.RenderLine(line, today)
	}
	return strings.Join(rendered, "\n")
}

// CalculateStatistics computes task statistics from content.
func (r *Renderer) CalculateStatistics(content string, today time.Time) Statistics {
	lines := strings.Split(content, "\n")
	stats := Statistics{}

	todayStr := today.Format("2006-01-02")

	for _, line := range lines {
		// Check for completed task
		if completedRegex.MatchString(line) {
			stats.Total++
			stats.Completed++

			// Check if completed today
			if match := doneRegex.FindStringSubmatch(line); match != nil {
				if match[1] == todayStr {
					stats.CompletedToday++
				}
			}

			// Check indentation for root/subtask
			if isRootTask(line) {
				stats.RootTasks++
			} else {
				stats.Subtasks++
			}
			continue
		}

		// Check for incomplete task
		if incompleteRegex.MatchString(line) {
			stats.Total++
			stats.Incomplete++

			// Check for @due tag
			if match := dueRegex.FindStringSubmatch(line); match != nil {
				dueDate, err := time.Parse("2006-01-02", match[1])
				if err == nil {
					lineType := r.classifyByDueDate(dueDate, today)
					switch lineType {
					case LineOverdue:
						stats.Overdue++
					case LineDueSoon:
						stats.DueSoon++
					}
				}
			}

			// Check indentation for root/subtask
			if isRootTask(line) {
				stats.RootTasks++
			} else {
				stats.Subtasks++
			}
		}
	}

	return stats
}

// isRootTask checks if a task line has no indentation.
func isRootTask(line string) bool {
	return !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t")
}

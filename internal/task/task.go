// Package task handles task detection, completion tagging, and archiving.
package task

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	// TabWidth is the number of spaces a tab character represents for indentation.
	TabWidth = 2
)

var (
	// completedPattern matches completed task lines: "- [x]" or "- [X]"
	completedPattern = regexp.MustCompile(`^\s*-\s*\[[xX]\]`)

	// taskPattern matches any task line: "- [ ]" or "- [x]" (with optional leading whitespace)
	taskPattern = regexp.MustCompile(`^\s*-\s*\[[xX ]\]`)

	// doneTagPattern matches @done(YYYY-MM-DD) format
	doneTagPattern = regexp.MustCompile(`@done\((\d{4}-\d{2}-\d{2})\)`)
)

// ParsedLine represents a line with its hierarchical context.
type ParsedLine struct {
	LineNumber  int    // 0-indexed position in file
	Content     string // Original line content
	Indent      int    // Number of leading spaces (tabs converted to TabWidth spaces)
	IsTask      bool   // Whether this is a task line (- [ ] or - [x])
	IsCompleted bool   // Whether the task is completed
	HasDoneTag  bool   // Whether @done tag exists
}

// TaskTree represents a task with its children for hierarchical operations.
type TaskTree struct {
	Line     *ParsedLine
	Children []*TaskTree
}

// ArchiveTask represents a task to be archived with its grouping metadata.
// GroupDate is used for archive section grouping (parent's completion date).
type ArchiveTask struct {
	Content   string    // Original line content
	GroupDate time.Time // Date to use for archive section grouping
}

// GetIndentLevel returns the number of leading spaces in a line.
// Tab characters are converted to TabWidth spaces.
func GetIndentLevel(line string) int {
	count := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			count++
		case '\t':
			count += TabWidth
		default:
			return count
		}
	}
	return count
}

// IsTask returns true if the line is a task (- [ ] or - [x]).
func IsTask(line string) bool {
	return taskPattern.MatchString(line)
}

// IsCompleted returns true if the line is a completed task (- [x] or - [X]).
func IsCompleted(line string) bool {
	return completedPattern.MatchString(line)
}

// HasDoneTag returns true if the line contains a valid @done(YYYY-MM-DD) tag.
func HasDoneTag(line string) bool {
	return doneTagPattern.MatchString(line)
}

// AddDoneTag adds @done(today) to a completed task if it doesn't already have one.
// Returns the modified line and whether it was changed.
func AddDoneTag(line string) (string, bool) {
	if !IsCompleted(line) {
		return line, false
	}

	if HasDoneTag(line) {
		return line, false
	}

	today := time.Now().Format("2006-01-02")
	return line + " @done(" + today + ")", true
}

// ParseDoneDate extracts the date from a @done(YYYY-MM-DD) tag.
// Returns the parsed date and true if found, zero time and false otherwise.
func ParseDoneDate(line string) (time.Time, bool) {
	matches := doneTagPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return time.Time{}, false
	}

	date, err := time.Parse("2006-01-02", matches[1])
	if err != nil {
		return time.Time{}, false
	}

	return date, true
}

// ParseLines parses content into a slice of ParsedLine structs.
// Each line is annotated with its indent level, task status, and completion state.
func ParseLines(content string) []ParsedLine {
	rawLines := strings.Split(content, "\n")
	result := make([]ParsedLine, len(rawLines))

	for i, line := range rawLines {
		result[i] = ParsedLine{
			LineNumber:  i,
			Content:     line,
			Indent:      GetIndentLevel(line),
			IsTask:      IsTask(line),
			IsCompleted: IsCompleted(line),
			HasDoneTag:  HasDoneTag(line),
		}
	}

	return result
}

// BuildTaskTrees builds a forest of task trees from parsed lines.
// Children are determined by having greater indentation than their parent.
// Non-task lines are ignored for hierarchy building but preserved in content.
func BuildTaskTrees(lines []ParsedLine) []*TaskTree {
	var forest []*TaskTree
	var stack []*TaskTree

	for i := range lines {
		line := &lines[i]
		if !line.IsTask {
			continue
		}

		tree := &TaskTree{Line: line, Children: nil}

		// Pop stack until we find a parent (lower indent)
		for len(stack) > 0 && stack[len(stack)-1].Line.Indent >= line.Indent {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			// This is a root task
			forest = append(forest, tree)
		} else {
			// This is a child of the top of the stack
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, tree)
		}

		stack = append(stack, tree)
	}

	return forest
}

// CascadeCompletion cascades completion status from parent tasks to children.
// When a parent is completed, all children are marked completed with @done(today).
// Returns the modified lines and the count of newly completed tasks.
func CascadeCompletion(lines []ParsedLine, today string) ([]ParsedLine, int) {
	trees := BuildTaskTrees(lines)
	count := 0

	for _, tree := range trees {
		count += cascadeCompletionRecursive(tree, lines, today)
	}

	return lines, count
}

// cascadeCompletionRecursive recursively cascades completion to children.
func cascadeCompletionRecursive(tree *TaskTree, lines []ParsedLine, today string) int {
	count := 0

	if tree.Line.IsCompleted {
		// Cascade to all children
		for _, child := range tree.Children {
			count += markTreeCompleted(child, lines, today)
		}
	}

	// Continue to check children (in case they are independently completed)
	for _, child := range tree.Children {
		count += cascadeCompletionRecursive(child, lines, today)
	}

	return count
}

// markTreeCompleted marks a task and all its descendants as completed.
func markTreeCompleted(tree *TaskTree, lines []ParsedLine, today string) int {
	count := 0
	line := tree.Line

	// Only modify if not already completed
	if !line.IsCompleted {
		// Change [ ] to [x] and add @done
		newContent := strings.Replace(line.Content, "[ ]", "[x]", 1)
		newContent = newContent + " @done(" + today + ")"

		lines[line.LineNumber].Content = newContent
		lines[line.LineNumber].IsCompleted = true
		lines[line.LineNumber].HasDoneTag = true
		count++
	}

	// Recursively mark children
	for _, child := range tree.Children {
		count += markTreeCompleted(child, lines, today)
	}

	return count
}

// ReconstructContent rebuilds content string from ParsedLines.
func ReconstructContent(lines []ParsedLine) string {
	contents := make([]string, len(lines))
	for i, line := range lines {
		contents[i] = line.Content
	}
	return strings.Join(contents, "\n")
}

// ProcessContent adds @done(today) tags to all completed tasks that don't have one.
// It also cascades completion from parent tasks to children.
// Returns the processed content and the count of tasks modified.
func ProcessContent(content string) (string, int) {
	today := time.Now().Format("2006-01-02")
	lines := ParseLines(content)
	count := 0

	// First, cascade completion from parents to children
	lines, cascadeCount := CascadeCompletion(lines, today)
	count += cascadeCount

	// Then, add @done tags to completed tasks that don't have one
	for i := range lines {
		if lines[i].IsCompleted && !lines[i].HasDoneTag {
			lines[i].Content = lines[i].Content + " @done(" + today + ")"
			lines[i].HasDoneTag = true
			count++
		}
	}

	return ReconstructContent(lines), count
}

// FilterArchivable separates tasks into archivable and remaining based on delay_days.
// Tasks completed more than delayDays ago are archivable.
// When a parent task is archivable, all its children (including non-task lines) are archived with it.
// Children cannot be archived independently - they only archive when parent is archivable.
// Returns (archivable tasks with group dates, remaining content as string).
func FilterArchivable(content string, delayDays int) ([]ArchiveTask, string) {
	lines := ParseLines(content)
	trees := BuildTaskTrees(lines)
	cutoff := time.Now().AddDate(0, 0, -delayDays)

	// Mark which line numbers should be archived and their group dates
	archiveSet := make(map[int]bool)
	groupDates := make(map[int]time.Time)

	for _, tree := range trees {
		markArchivableRecursive(tree, cutoff, archiveSet, groupDates, false, time.Time{}, true)
	}

	// Include non-task lines that belong to archived task subtrees
	includeNonTaskChildren(lines, archiveSet, groupDates)

	var archivable []ArchiveTask
	var remaining []string

	for i, line := range lines {
		if archiveSet[i] {
			archivable = append(archivable, ArchiveTask{
				Content:   line.Content,
				GroupDate: groupDates[i],
			})
		} else {
			remaining = append(remaining, line.Content)
		}
	}

	return archivable, strings.Join(remaining, "\n")
}

// includeNonTaskChildren marks non-task lines for archiving when they are children of archived tasks.
// A non-task line is considered a child of a task if it has greater indentation and appears
// between the task and the next task at the same or lesser indentation level.
func includeNonTaskChildren(lines []ParsedLine, archiveSet map[int]bool, groupDates map[int]time.Time) {
	for i := 0; i < len(lines); i++ {
		if !archiveSet[i] || !lines[i].IsTask {
			continue
		}

		// This is an archived task - find and include its non-task children
		parentIndent := lines[i].Indent
		parentGroupDate := groupDates[i]

		for j := i + 1; j < len(lines); j++ {
			childLine := lines[j]

			// Stop if we reach a line with same or lesser indentation
			if childLine.Indent <= parentIndent {
				break
			}

			// If this is a non-task line with greater indentation, include it
			if !childLine.IsTask && !archiveSet[j] {
				archiveSet[j] = true
				groupDates[j] = parentGroupDate
			}
		}
	}
}

// markArchivableRecursive marks a task tree for archiving if the root task is old enough.
// Only root tasks (isRoot=true) can independently qualify for archiving.
// Children are only archived when their parent is archivable.
// groupDates tracks the completion date to use for archive grouping (parent's date).
func markArchivableRecursive(
	tree *TaskTree,
	cutoff time.Time,
	archiveSet map[int]bool,
	groupDates map[int]time.Time,
	parentArchivable bool,
	parentDate time.Time,
	isRoot bool,
) {
	line := tree.Line
	shouldArchive := parentArchivable
	groupDate := parentDate

	// Only root tasks can independently qualify for archiving
	// Children can only be archived via parent
	if isRoot && !shouldArchive && line.IsCompleted && line.HasDoneTag {
		doneDate, found := ParseDoneDate(line.Content)
		if found && doneDate.Before(cutoff) {
			shouldArchive = true
			groupDate = doneDate // Use this task's date for grouping
		}
	}

	if shouldArchive {
		archiveSet[line.LineNumber] = true
		groupDates[line.LineNumber] = groupDate
	}

	// Recursively process children - they are never "root" for archive purposes
	for _, child := range tree.Children {
		markArchivableRecursive(child, cutoff, archiveSet, groupDates, shouldArchive, groupDate, false)
	}
}

// FormatArchiveEntry formats tasks for the archive file, grouped by GroupDate.
// Tasks are grouped under "## YYYY-MM-DD" headers, sorted by date descending.
// Each task's GroupDate determines which section it appears in (typically parent's completion date).
func FormatArchiveEntry(tasks []ArchiveTask) string {
	if len(tasks) == 0 {
		return ""
	}

	// Group tasks by GroupDate
	byDate := make(map[string][]string)
	for _, task := range tasks {
		dateStr := task.GroupDate.Format("2006-01-02")
		byDate[dateStr] = append(byDate[dateStr], task.Content)
	}

	// Sort dates descending
	var dates []string
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	// Build output
	var builder strings.Builder
	for _, date := range dates {
		builder.WriteString("## " + date + "\n\n")
		for _, task := range byDate[date] {
			builder.WriteString(task + "\n")
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// LoadFile reads the content of a file and returns it as a string.
// Returns an error if the file cannot be read.
func LoadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteFile writes content to a file, creating it if it doesn't exist
// or overwriting it if it does.
func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// PrependToFile adds content to the beginning of a file.
// Used for archive entries where newest dates should appear first.
func PrependToFile(path string, content string) error {
	existing, err := LoadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return WriteFile(path, content+existing)
}

// ProcessFileWithDoneTags reads a file, adds @done tags to completed tasks,
// and writes the result back. Returns the count of modified tasks.
func ProcessFileWithDoneTags(path string) (int, error) {
	content, err := LoadFile(path)
	if err != nil {
		return 0, err
	}

	processed, count := ProcessContent(content)
	if count > 0 {
		if err := WriteFile(path, processed); err != nil {
			return 0, err
		}
	}

	return count, nil
}

// Archive moves old completed tasks from the tasks file to the archive file.
// Tasks completed more than delayDays ago are archived.
// Children are only archived when their parent is archivable.
// Returns the count of archived tasks.
func Archive(tasksPath, archivePath string, delayDays int) (int, error) {
	content, err := LoadFile(tasksPath)
	if err != nil {
		return 0, err
	}

	archivableTasks, remaining := FilterArchivable(content, delayDays)
	if len(archivableTasks) == 0 {
		return 0, nil
	}

	// Format archive entry
	archiveEntry := FormatArchiveEntry(archivableTasks)

	// Prepend to archive file
	if err := PrependToFile(archivePath, archiveEntry); err != nil {
		return 0, err
	}

	// Write remaining tasks back
	if err := WriteFile(tasksPath, remaining); err != nil {
		return 0, err
	}

	return len(archivableTasks), nil
}

// Package task handles task detection, completion tagging, and archiving.
package task

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	// completedPattern matches completed task lines: "- [x]" or "- [X]"
	completedPattern = regexp.MustCompile(`^\s*-\s*\[[xX]\]`)

	// doneTagPattern matches @done(YYYY-MM-DD) format
	doneTagPattern = regexp.MustCompile(`@done\((\d{4}-\d{2}-\d{2})\)`)
)

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

// ProcessContent adds @done(today) tags to all completed tasks that don't have one.
// Returns the processed content and the count of tasks modified.
func ProcessContent(content string) (string, int) {
	lines := strings.Split(content, "\n")
	count := 0

	for i, line := range lines {
		newLine, changed := AddDoneTag(line)
		if changed {
			lines[i] = newLine
			count++
		}
	}

	return strings.Join(lines, "\n"), count
}

// FilterArchivable separates tasks into archivable and remaining based on delay_days.
// Tasks completed more than delayDays ago are archivable.
// Returns (archivable lines joined, remaining content).
func FilterArchivable(content string, delayDays int) (string, string) {
	lines := strings.Split(content, "\n")
	cutoff := time.Now().AddDate(0, 0, -delayDays)

	var archivable []string
	var remaining []string

	for _, line := range lines {
		if IsCompleted(line) && HasDoneTag(line) {
			doneDate, found := ParseDoneDate(line)
			if found && doneDate.Before(cutoff) {
				archivable = append(archivable, line)
				continue
			}
		}
		remaining = append(remaining, line)
	}

	return strings.Join(archivable, "\n"), strings.Join(remaining, "\n")
}

// FormatArchiveEntry formats tasks for the archive file, grouped by completion date.
// Tasks are grouped under "## YYYY-MM-DD" headers, sorted by date descending.
func FormatArchiveEntry(tasks []string) string {
	if len(tasks) == 0 {
		return ""
	}

	// Group tasks by date
	byDate := make(map[string][]string)
	for _, task := range tasks {
		date, found := ParseDoneDate(task)
		if found {
			dateStr := date.Format("2006-01-02")
			byDate[dateStr] = append(byDate[dateStr], task)
		}
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
// Returns the count of archived tasks.
func Archive(tasksPath, archivePath string, delayDays int) (int, error) {
	content, err := LoadFile(tasksPath)
	if err != nil {
		return 0, err
	}

	archivable, remaining := FilterArchivable(content, delayDays)
	if archivable == "" {
		return 0, nil
	}

	// Parse archivable tasks into slice
	tasks := strings.Split(archivable, "\n")
	var validTasks []string
	for _, task := range tasks {
		if task != "" {
			validTasks = append(validTasks, task)
		}
	}

	if len(validTasks) == 0 {
		return 0, nil
	}

	// Format archive entry
	archiveEntry := FormatArchiveEntry(validTasks)

	// Prepend to archive file
	if err := PrependToFile(archivePath, archiveEntry); err != nil {
		return 0, err
	}

	// Write remaining tasks back
	if err := WriteFile(tasksPath, remaining); err != nil {
		return 0, err
	}

	return len(validTasks), nil
}

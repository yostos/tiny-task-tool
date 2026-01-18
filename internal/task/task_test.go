package task

import (
	"testing"
	"time"
)

// TestIsCompleted verifies that IsCompleted() correctly identifies completed tasks.
// A task is completed if it matches the pattern "- [x]" (case insensitive for x).
func TestIsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"completed task", "- [x] Buy milk", true},
		{"completed task uppercase", "- [X] Buy milk", true},
		{"incomplete task", "- [ ] Buy milk", false},
		{"completed with done tag", "- [x] Buy milk @done(2026-01-18)", true},
		{"not a task", "Some regular text", false},
		{"empty line", "", false},
		{"heading", "# Tasks", false},
		{"indented completed", "  - [x] Subtask", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCompleted(tt.line)
			if result != tt.expected {
				t.Errorf("IsCompleted(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

// TestHasDoneTag verifies that HasDoneTag() detects the @done(date) tag.
// The tag format is @done(YYYY-MM-DD).
func TestHasDoneTag(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"has done tag", "- [x] Buy milk @done(2026-01-18)", true},
		{"no done tag", "- [x] Buy milk", false},
		{"incomplete task", "- [ ] Buy milk", false},
		{"done tag in middle", "- [x] Task @done(2026-01-18) extra", true},
		{"malformed done tag", "- [x] Task @done(invalid)", false},
		{"empty line", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasDoneTag(tt.line)
			if result != tt.expected {
				t.Errorf("HasDoneTag(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

// TestAddDoneTag verifies that AddDoneTag() adds @done(date) to completed tasks.
// It should only add the tag if the task is completed and doesn't already have one.
func TestAddDoneTag(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	tests := []struct {
		name     string
		line     string
		expected string
		changed  bool
	}{
		{
			"add to completed task",
			"- [x] Buy milk",
			"- [x] Buy milk @done(" + today + ")",
			true,
		},
		{
			"already has done tag",
			"- [x] Buy milk @done(2026-01-15)",
			"- [x] Buy milk @done(2026-01-15)",
			false,
		},
		{
			"incomplete task unchanged",
			"- [ ] Buy milk",
			"- [ ] Buy milk",
			false,
		},
		{
			"non-task unchanged",
			"# Header",
			"# Header",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, changed := AddDoneTag(tt.line)
			if result != tt.expected {
				t.Errorf("AddDoneTag(%q) = %q, want %q", tt.line, result, tt.expected)
			}
			if changed != tt.changed {
				t.Errorf("AddDoneTag(%q) changed = %v, want %v", tt.line, changed, tt.changed)
			}
		})
	}
}

// TestParseDoneDate verifies that ParseDoneDate() extracts the date from @done tag.
// Returns the date and true if found, zero time and false otherwise.
func TestParseDoneDate(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectedDay int // day of month, 0 if not found
		found       bool
	}{
		{"valid done tag", "- [x] Task @done(2026-01-18)", 18, true},
		{"no done tag", "- [x] Task", 0, false},
		{"invalid date", "- [x] Task @done(invalid)", 0, false},
		{"empty line", "", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, found := ParseDoneDate(tt.line)
			if found != tt.found {
				t.Errorf("ParseDoneDate(%q) found = %v, want %v", tt.line, found, tt.found)
			}
			if found && date.Day() != tt.expectedDay {
				t.Errorf("ParseDoneDate(%q) day = %d, want %d", tt.line, date.Day(), tt.expectedDay)
			}
		})
	}
}

// TestProcessContent verifies that ProcessContent() adds @done tags to all
// newly completed tasks in the content. Returns the processed content and
// the count of tasks that were modified.
func TestProcessContent(t *testing.T) {
	input := `# Tasks

- [ ] Incomplete task
- [x] Completed without done
- [x] Already has @done(2026-01-15)
- [x] Another completed
`
	result, count := ProcessContent(input)

	// Should have modified 2 tasks (the two completed without @done)
	if count != 2 {
		t.Errorf("ProcessContent() count = %d, want 2", count)
	}

	// Result should contain @done tags for the modified tasks
	if !containsString(result, "@done(") {
		t.Error("ProcessContent() should add @done tags")
	}

	// Original @done tag should be preserved
	if !containsString(result, "@done(2026-01-15)") {
		t.Error("ProcessContent() should preserve existing @done tags")
	}

	// Incomplete task should remain unchanged
	if !containsString(result, "- [ ] Incomplete task") {
		t.Error("ProcessContent() should not modify incomplete tasks")
	}
}

// TestFilterArchivable verifies that FilterArchivable() correctly identifies
// tasks that should be archived based on the delay_days setting.
func TestFilterArchivable(t *testing.T) {
	// Create dates for testing
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")    // 5 days ago
	recentDate := now.AddDate(0, 0, -1).Format("2006-01-02") // 1 day ago

	content := `# Tasks

- [ ] Incomplete task
- [x] Old completed @done(` + oldDate + `)
- [x] Recent completed @done(` + recentDate + `)
- [x] No done tag
`

	archivable, remaining := FilterArchivable(content, 2) // 2 day delay

	// Old task should be archivable
	if !containsString(archivable, "Old completed") {
		t.Error("FilterArchivable() should include old completed task")
	}

	// Recent task should remain
	if !containsString(remaining, "Recent completed") {
		t.Error("FilterArchivable() should keep recent completed task")
	}

	// Task without done tag should remain
	if !containsString(remaining, "No done tag") {
		t.Error("FilterArchivable() should keep task without @done tag")
	}

	// Incomplete task should remain
	if !containsString(remaining, "Incomplete task") {
		t.Error("FilterArchivable() should keep incomplete tasks")
	}
}

// TestFormatArchiveEntry verifies that FormatArchiveEntry() creates properly
// formatted archive entries grouped by completion date.
func TestFormatArchiveEntry(t *testing.T) {
	tasks := []string{
		"- [x] Task A @done(2026-01-18)",
		"- [x] Task B @done(2026-01-18)",
		"- [x] Task C @done(2026-01-17)",
	}

	result := FormatArchiveEntry(tasks)

	// Should have date headers
	if !containsString(result, "## 2026-01-18") {
		t.Error("FormatArchiveEntry() should include date header for 2026-01-18")
	}
	if !containsString(result, "## 2026-01-17") {
		t.Error("FormatArchiveEntry() should include date header for 2026-01-17")
	}

	// Tasks should be included
	if !containsString(result, "Task A") {
		t.Error("FormatArchiveEntry() should include Task A")
	}
}

// helper function
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// File Operations Tests
// =============================================================================

// TestLoadFile verifies that LoadFile() reads file content correctly.
// It should return the file content as a string, or an error if the file doesn't exist.
func TestLoadFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test-tasks.md"

	content := "- [ ] Task 1\n- [x] Task 2\n"
	if err := WriteFile(testFile, content); err != nil {
		t.Fatalf("WriteFile() setup error: %v", err)
	}

	// Test loading existing file
	result, err := LoadFile(testFile)
	if err != nil {
		t.Fatalf("LoadFile() error: %v", err)
	}
	if result != content {
		t.Errorf("LoadFile() = %q, want %q", result, content)
	}

	// Test loading non-existent file
	_, err = LoadFile(tmpDir + "/nonexistent.md")
	if err == nil {
		t.Error("LoadFile() should return error for non-existent file")
	}
}

// TestWriteFile verifies that WriteFile() writes content to a file correctly.
// It should create the file if it doesn't exist, or overwrite if it does.
func TestWriteFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test-output.md"

	content := "- [ ] New task\n"

	// Write to new file
	err := WriteFile(testFile, content)
	if err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Verify content
	result, err := LoadFile(testFile)
	if err != nil {
		t.Fatalf("LoadFile() verification error: %v", err)
	}
	if result != content {
		t.Errorf("WriteFile() wrote %q, want %q", result, content)
	}

	// Overwrite existing file
	newContent := "- [x] Updated task\n"
	err = WriteFile(testFile, newContent)
	if err != nil {
		t.Fatalf("WriteFile() overwrite error: %v", err)
	}

	result, err = LoadFile(testFile)
	if err != nil {
		t.Fatalf("LoadFile() verification error: %v", err)
	}
	if result != newContent {
		t.Errorf("WriteFile() overwrite wrote %q, want %q", result, newContent)
	}
}

// TestAppendToFile verifies that AppendToFile() adds content to the beginning of a file.
// New content should be prepended, not appended, for archive entries.
func TestAppendToFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/test-archive.md"

	// Write initial content
	initial := "## 2026-01-17\n\n- [x] Old task @done(2026-01-17)\n\n"
	if err := WriteFile(testFile, initial); err != nil {
		t.Fatalf("WriteFile() setup error: %v", err)
	}

	// Prepend new content
	newContent := "## 2026-01-18\n\n- [x] New task @done(2026-01-18)\n\n"
	err := PrependToFile(testFile, newContent)
	if err != nil {
		t.Fatalf("PrependToFile() error: %v", err)
	}

	// Verify new content is at the beginning
	result, err := LoadFile(testFile)
	if err != nil {
		t.Fatalf("LoadFile() verification error: %v", err)
	}

	// New content should come first
	if !containsString(result, "## 2026-01-18") {
		t.Error("PrependToFile() should include new date header")
	}
	if !containsString(result, "## 2026-01-17") {
		t.Error("PrependToFile() should preserve old date header")
	}
}

// TestArchive verifies the complete archive workflow.
// It should move old completed tasks from tasks file to archive file.
func TestArchive(t *testing.T) {
	tmpDir := t.TempDir()
	tasksFile := tmpDir + "/tasks.md"
	archiveFile := tmpDir + "/archive.md"

	// Create dates for testing
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	recentDate := now.AddDate(0, 0, -1).Format("2006-01-02")

	tasksContent := `# Tasks

- [ ] Incomplete task
- [x] Old task @done(` + oldDate + `)
- [x] Recent task @done(` + recentDate + `)
`

	if err := WriteFile(tasksFile, tasksContent); err != nil {
		t.Fatalf("WriteFile() setup error: %v", err)
	}

	// Run archive with 2-day delay
	count, err := Archive(tasksFile, archiveFile, 2)
	if err != nil {
		t.Fatalf("Archive() error: %v", err)
	}

	// Should have archived 1 task (the old one)
	if count != 1 {
		t.Errorf("Archive() count = %d, want 1", count)
	}

	// Verify tasks file no longer contains old task
	remaining, err := LoadFile(tasksFile)
	if err != nil {
		t.Fatalf("LoadFile() tasks error: %v", err)
	}
	if containsString(remaining, "Old task") {
		t.Error("Archive() should remove old task from tasks file")
	}
	if !containsString(remaining, "Recent task") {
		t.Error("Archive() should keep recent task in tasks file")
	}
	if !containsString(remaining, "Incomplete task") {
		t.Error("Archive() should keep incomplete task in tasks file")
	}

	// Verify archive file contains old task
	archived, err := LoadFile(archiveFile)
	if err != nil {
		t.Fatalf("LoadFile() archive error: %v", err)
	}
	if !containsString(archived, "Old task") {
		t.Error("Archive() should add old task to archive file")
	}
	if !containsString(archived, "## "+oldDate) {
		t.Error("Archive() should include date header in archive")
	}
}

// TestArchiveNoTasks verifies Archive() behavior when there are no tasks to archive.
// It should return 0 count and not modify files unnecessarily.
func TestArchiveNoTasks(t *testing.T) {
	tmpDir := t.TempDir()
	tasksFile := tmpDir + "/tasks.md"
	archiveFile := tmpDir + "/archive.md"

	tasksContent := "- [ ] Incomplete task\n- [x] Recent task @done(" + time.Now().Format("2006-01-02") + ")\n"
	if err := WriteFile(tasksFile, tasksContent); err != nil {
		t.Fatalf("WriteFile() setup error: %v", err)
	}

	count, err := Archive(tasksFile, archiveFile, 2)
	if err != nil {
		t.Fatalf("Archive() error: %v", err)
	}

	if count != 0 {
		t.Errorf("Archive() count = %d, want 0", count)
	}
}

// TestProcessFileWithDoneTags verifies that ProcessFileWithDoneTags() adds @done tags
// to completed tasks in the file and saves it.
func TestProcessFileWithDoneTags(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := tmpDir + "/tasks.md"

	content := "- [ ] Incomplete\n- [x] Completed without done\n- [x] Has done @done(2026-01-15)\n"
	if err := WriteFile(testFile, content); err != nil {
		t.Fatalf("WriteFile() setup error: %v", err)
	}

	count, err := ProcessFileWithDoneTags(testFile)
	if err != nil {
		t.Fatalf("ProcessFileWithDoneTags() error: %v", err)
	}

	// Should have modified 1 task
	if count != 1 {
		t.Errorf("ProcessFileWithDoneTags() count = %d, want 1", count)
	}

	// Verify file was updated
	result, err := LoadFile(testFile)
	if err != nil {
		t.Fatalf("LoadFile() verification error: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	if !containsString(result, "@done("+today+")") {
		t.Error("ProcessFileWithDoneTags() should add today's date")
	}
	if !containsString(result, "@done(2026-01-15)") {
		t.Error("ProcessFileWithDoneTags() should preserve existing @done tags")
	}
}

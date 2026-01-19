package task

import (
	"strings"
	"testing"
	"time"
)

// archiveTasksToString converts []ArchiveTask to a joined string of contents.
// This is a test helper for compatibility with tests that expect string output.
func archiveTasksToString(tasks []ArchiveTask) string {
	var contents []string
	for _, task := range tasks {
		contents = append(contents, task.Content)
	}
	return strings.Join(contents, "\n")
}

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

	archivableTasks, remaining := FilterArchivable(content, 2) // 2 day delay
	archivable := archiveTasksToString(archivableTasks)

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
// formatted archive entries grouped by GroupDate.
func TestFormatArchiveEntry(t *testing.T) {
	date18, _ := time.Parse("2006-01-02", "2026-01-18")
	date17, _ := time.Parse("2006-01-02", "2026-01-17")

	tasks := []ArchiveTask{
		{Content: "- [x] Task A @done(2026-01-18)", GroupDate: date18},
		{Content: "- [x] Task B @done(2026-01-18)", GroupDate: date18},
		{Content: "- [x] Task C @done(2026-01-17)", GroupDate: date17},
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

// =============================================================================
// Hierarchy Support Tests (Phase 1)
// =============================================================================

// TestGetIndentLevel verifies indentation calculation for hierarchy detection.
// Tab characters are converted to 2 spaces.
func TestGetIndentLevel(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected int
	}{
		{"no indent", "- [ ] Task", 0},
		{"2 spaces", "  - [ ] Task", 2},
		{"4 spaces", "    - [ ] Task", 4},
		{"tab as 2 spaces", "\t- [ ] Task", 2},
		{"tab + 2 spaces", "\t  - [ ] Task", 4},
		{"empty line", "", 0},
		{"only spaces", "   ", 3},
		{"non-task with indent", "  Some text", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIndentLevel(tt.line)
			if result != tt.expected {
				t.Errorf("GetIndentLevel(%q) = %d, want %d", tt.line, result, tt.expected)
			}
		})
	}
}

// TestIsTask verifies that IsTask() identifies task lines (- [ ] or - [x]).
func TestIsTask(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{"incomplete task", "- [ ] Buy milk", true},
		{"completed task", "- [x] Buy milk", true},
		{"indented incomplete", "  - [ ] Subtask", true},
		{"indented completed", "  - [x] Subtask", true},
		{"not a task heading", "# Tasks", false},
		{"not a task text", "Some regular text", false},
		{"empty line", "", false},
		{"bullet without checkbox", "- Item", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsTask(tt.line)
			if result != tt.expected {
				t.Errorf("IsTask(%q) = %v, want %v", tt.line, result, tt.expected)
			}
		})
	}
}

// TestParseLines verifies content parsing into ParsedLine structs.
// Each line should have correct indent, task status, and completion flags.
func TestParseLines(t *testing.T) {
	content := `# Header
- [ ] Task 1
  - [x] Subtask @done(2026-01-18)
- [x] Task 2
Some text`

	lines := ParseLines(content)

	if len(lines) != 5 {
		t.Fatalf("ParseLines() returned %d lines, want 5", len(lines))
	}

	// Line 0: Header
	if lines[0].IsTask || lines[0].Indent != 0 {
		t.Errorf("Line 0: expected non-task with indent 0, got IsTask=%v Indent=%d", lines[0].IsTask, lines[0].Indent)
	}

	// Line 1: Task 1 (incomplete, no indent)
	if !lines[1].IsTask || lines[1].IsCompleted || lines[1].Indent != 0 {
		t.Errorf("Line 1: expected incomplete task with indent 0")
	}

	// Line 2: Subtask (completed, indent 2, has done tag)
	if !lines[2].IsTask || !lines[2].IsCompleted || lines[2].Indent != 2 || !lines[2].HasDoneTag {
		t.Errorf("Line 2: expected completed task with indent 2 and done tag")
	}

	// Line 3: Task 2 (completed, no indent, no done tag)
	if !lines[3].IsTask || !lines[3].IsCompleted || lines[3].Indent != 0 || lines[3].HasDoneTag {
		t.Errorf("Line 3: expected completed task with indent 0, no done tag")
	}

	// Line 4: Some text (not a task)
	if lines[4].IsTask {
		t.Errorf("Line 4: expected non-task")
	}
}

// TestBuildTaskTrees verifies tree construction from parsed lines.
// Children should be correctly associated with parents based on indentation.
func TestBuildTaskTrees(t *testing.T) {
	content := `- [ ] Parent 1
  - [ ] Child 1.1
  - [ ] Child 1.2
    - [ ] Grandchild 1.2.1
- [ ] Parent 2
  - [ ] Child 2.1`

	lines := ParseLines(content)
	trees := BuildTaskTrees(lines)

	// Should have 2 top-level trees
	if len(trees) != 2 {
		t.Fatalf("BuildTaskTrees() returned %d trees, want 2", len(trees))
	}

	// Parent 1 should have 2 children
	if len(trees[0].Children) != 2 {
		t.Errorf("Parent 1 should have 2 children, got %d", len(trees[0].Children))
	}

	// Child 1.2 should have 1 grandchild
	if len(trees[0].Children) >= 2 && len(trees[0].Children[1].Children) != 1 {
		t.Errorf("Child 1.2 should have 1 grandchild, got %d", len(trees[0].Children[1].Children))
	}

	// Parent 2 should have 1 child
	if len(trees[1].Children) != 1 {
		t.Errorf("Parent 2 should have 1 child, got %d", len(trees[1].Children))
	}
}

// TestBuildTaskTreesWithNonTaskLines verifies that non-task lines don't break hierarchy.
func TestBuildTaskTreesWithNonTaskLines(t *testing.T) {
	content := `- [ ] Parent
Some note
  - [ ] Child`

	lines := ParseLines(content)
	trees := BuildTaskTrees(lines)

	// Should have 1 top-level tree with 1 child
	if len(trees) != 1 {
		t.Fatalf("BuildTaskTrees() returned %d trees, want 1", len(trees))
	}

	if len(trees[0].Children) != 1 {
		t.Errorf("Parent should have 1 child, got %d", len(trees[0].Children))
	}
}

// =============================================================================
// Hierarchy Support Tests (Phase 2 - Cascade Completion)
// =============================================================================

// TestCascadeCompletion verifies parent completion cascades to children.
// When parent is [x], all children should become [x] with @done(date).
func TestCascadeCompletion(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	input := `- [x] Parent task
  - [ ] Child 1
  - [ ] Child 2`

	lines := ParseLines(input)
	result, count := CascadeCompletion(lines, today)

	// Should have cascaded to 2 children
	if count != 2 {
		t.Errorf("CascadeCompletion() count = %d, want 2", count)
	}

	// Children should now be completed
	if !result[1].IsCompleted {
		t.Error("Child 1 should be completed")
	}
	if !result[2].IsCompleted {
		t.Error("Child 2 should be completed")
	}

	// Children content should have [x] and @done
	if !containsString(result[1].Content, "[x]") {
		t.Error("Child 1 content should have [x]")
	}
	if !containsString(result[1].Content, "@done("+today+")") {
		t.Error("Child 1 content should have @done tag")
	}
}

// TestCascadeCompletionDeepNesting verifies cascade works for nested hierarchies.
// Grandchildren should also be completed when grandparent is completed.
func TestCascadeCompletionDeepNesting(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	input := `- [x] Grandparent
  - [ ] Parent
    - [ ] Child`

	lines := ParseLines(input)
	result, count := CascadeCompletion(lines, today)

	// Should cascade to parent and child
	if count != 2 {
		t.Errorf("CascadeCompletion() count = %d, want 2", count)
	}

	// Both should be completed
	if !result[1].IsCompleted || !result[2].IsCompleted {
		t.Error("All descendants should be completed")
	}
}

// TestCascadeCompletionIncompleteParent verifies incomplete parent doesn't cascade.
func TestCascadeCompletionIncompleteParent(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	input := `- [ ] Parent task
  - [ ] Child 1`

	lines := ParseLines(input)
	_, count := CascadeCompletion(lines, today)

	// Should not cascade anything
	if count != 0 {
		t.Errorf("CascadeCompletion() count = %d, want 0", count)
	}
}

// TestCascadeCompletionAlreadyCompleted verifies already completed children aren't double-tagged.
func TestCascadeCompletionAlreadyCompleted(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	input := `- [x] Parent task
  - [x] Already done @done(2026-01-15)`

	lines := ParseLines(input)
	result, count := CascadeCompletion(lines, today)

	// Should not modify already completed child
	if count != 0 {
		t.Errorf("CascadeCompletion() count = %d, want 0", count)
	}

	// Original @done tag should be preserved
	if !containsString(result[1].Content, "@done(2026-01-15)") {
		t.Error("Original @done tag should be preserved")
	}
}

// TestReconstructContent verifies content reconstruction from ParsedLines.
func TestReconstructContent(t *testing.T) {
	input := `# Header
- [ ] Task 1
  - [x] Subtask`

	lines := ParseLines(input)
	result := ReconstructContent(lines)

	if result != input {
		t.Errorf("ReconstructContent() = %q, want %q", result, input)
	}
}

// TestProcessContentWithHierarchy verifies ProcessContent cascades completion.
func TestProcessContentWithHierarchy(t *testing.T) {
	today := time.Now().Format("2006-01-02")

	input := `- [x] Parent
  - [ ] Child 1
  - [ ] Child 2
- [ ] Other task`

	result, count := ProcessContent(input)

	// Should have modified: parent (@done) + 2 children (cascade)
	if count != 3 {
		t.Errorf("ProcessContent() count = %d, want 3", count)
	}

	// Parent should have @done
	if !containsString(result, "- [x] Parent @done("+today+")") {
		t.Error("Parent should have @done tag")
	}

	// Children should be completed with @done
	if !containsString(result, "- [x] Child 1 @done("+today+")") {
		t.Error("Child 1 should be completed with @done")
	}
	if !containsString(result, "- [x] Child 2 @done("+today+")") {
		t.Error("Child 2 should be completed with @done")
	}

	// Other task should remain incomplete
	if !containsString(result, "- [ ] Other task") {
		t.Error("Other task should remain incomplete")
	}
}

// =============================================================================
// Hierarchy Support Tests (Phase 3 - Archive with Hierarchy)
// =============================================================================

// TestFilterArchivableWithHierarchy verifies children are archived with parent.
// When parent is archivable, all children move to archive regardless of state.
func TestFilterArchivableWithHierarchy(t *testing.T) {
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")
	recentDate := now.AddDate(0, 0, -1).Format("2006-01-02")

	content := `- [x] Old parent @done(` + oldDate + `)
  - [x] Old child @done(` + oldDate + `)
- [x] Recent parent @done(` + recentDate + `)
  - [x] Recent child @done(` + recentDate + `)
- [ ] Incomplete task`

	archivableTasks, remaining := FilterArchivable(content, 2)
	archivable := archiveTasksToString(archivableTasks)

	// Old parent and child should be archived together
	if !containsString(archivable, "Old parent") {
		t.Error("Old parent should be archivable")
	}
	if !containsString(archivable, "Old child") {
		t.Error("Old child should be archived with parent")
	}

	// Recent tasks should remain
	if !containsString(remaining, "Recent parent") {
		t.Error("Recent parent should remain")
	}
	if !containsString(remaining, "Recent child") {
		t.Error("Recent child should remain")
	}

	// Incomplete task should remain
	if !containsString(remaining, "Incomplete task") {
		t.Error("Incomplete task should remain")
	}
}

// TestFilterArchivablePreservesIndentation verifies archived tasks keep their indentation.
func TestFilterArchivablePreservesIndentation(t *testing.T) {
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")

	content := `- [x] Parent @done(` + oldDate + `)
  - [x] Child @done(` + oldDate + `)`

	archivableTasks, _ := FilterArchivable(content, 2)
	archivable := archiveTasksToString(archivableTasks)

	// Indentation should be preserved
	if !containsString(archivable, "  - [x] Child") {
		t.Error("Child indentation should be preserved in archive")
	}
}

// TestFilterArchivableDeepNesting verifies deep nesting is handled correctly.
func TestFilterArchivableDeepNesting(t *testing.T) {
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")

	content := `- [x] Grandparent @done(` + oldDate + `)
  - [x] Parent @done(` + oldDate + `)
    - [x] Child @done(` + oldDate + `)`

	archivableTasks, remaining := FilterArchivable(content, 2)
	archivable := archiveTasksToString(archivableTasks)

	// All three should be archived
	if !containsString(archivable, "Grandparent") {
		t.Error("Grandparent should be archivable")
	}
	if !containsString(archivable, "Parent") {
		t.Error("Parent should be archived with grandparent")
	}
	if !containsString(archivable, "Child") {
		t.Error("Child should be archived with grandparent")
	}

	// Remaining should be empty or just newlines
	trimmed := strings.TrimSpace(remaining)
	if trimmed != "" {
		t.Errorf("Remaining should be empty, got %q", trimmed)
	}
}

// TestArchiveWithHierarchy verifies the complete archive workflow with hierarchy.
func TestArchiveWithHierarchy(t *testing.T) {
	tmpDir := t.TempDir()
	tasksFile := tmpDir + "/tasks.md"
	archiveFile := tmpDir + "/archive.md"

	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")

	tasksContent := `- [x] Old parent @done(` + oldDate + `)
  - [x] Old child @done(` + oldDate + `)
- [ ] Incomplete task
`

	if err := WriteFile(tasksFile, tasksContent); err != nil {
		t.Fatalf("WriteFile() setup error: %v", err)
	}

	count, err := Archive(tasksFile, archiveFile, 2)
	if err != nil {
		t.Fatalf("Archive() error: %v", err)
	}

	// Should have archived 2 tasks (parent + child)
	if count != 2 {
		t.Errorf("Archive() count = %d, want 2", count)
	}

	// Verify tasks file
	remaining, _ := LoadFile(tasksFile)
	if containsString(remaining, "Old parent") || containsString(remaining, "Old child") {
		t.Error("Old tasks should be removed from tasks file")
	}
	if !containsString(remaining, "Incomplete task") {
		t.Error("Incomplete task should remain")
	}

	// Verify archive file
	archived, _ := LoadFile(archiveFile)
	if !containsString(archived, "Old parent") {
		t.Error("Old parent should be in archive")
	}
	if !containsString(archived, "Old child") {
		t.Error("Old child should be in archive")
	}
}

// TestChildNotArchivedWhenParentIncomplete verifies that child tasks
// are NOT archived when parent is incomplete, even if child has old @done date.
// Spec: Children should only be archived when their parent is archivable.
func TestChildNotArchivedWhenParentIncomplete(t *testing.T) {
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02") // 5 days ago

	content := `- [ ] Incomplete parent
  - [x] Old child @done(` + oldDate + `)`

	archivableTasks, remaining := FilterArchivable(content, 2)
	archivable := archiveTasksToString(archivableTasks)

	// Child should NOT be archived because parent is incomplete
	if containsString(archivable, "Old child") {
		t.Error("Child with old @done should NOT be archived when parent is incomplete")
	}

	// Both should remain
	if !containsString(remaining, "Incomplete parent") {
		t.Error("Incomplete parent should remain")
	}
	if !containsString(remaining, "Old child") {
		t.Error("Child of incomplete parent should remain")
	}
}

// TestChildNotArchivedWhenParentNotOldEnough verifies that child tasks
// follow parent's archivability, not their own date.
// Spec: Even if child has older @done date, it follows parent's archive status.
func TestChildNotArchivedWhenParentNotOldEnough(t *testing.T) {
	now := time.Now()
	recentDate := now.AddDate(0, 0, -1).Format("2006-01-02") // 1 day ago
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")    // 5 days ago

	content := `- [x] Recent parent @done(` + recentDate + `)
  - [x] Old child @done(` + oldDate + `)`

	archivableTasks, remaining := FilterArchivable(content, 2)
	archivable := archiveTasksToString(archivableTasks)

	// Neither should be archived - parent is too recent
	if containsString(archivable, "Recent parent") {
		t.Error("Recent parent should NOT be archived")
	}
	if containsString(archivable, "Old child") {
		t.Error("Child should NOT be archived when parent is not archivable")
	}

	// Both should remain
	if !containsString(remaining, "Recent parent") {
		t.Error("Recent parent should remain")
	}
	if !containsString(remaining, "Old child") {
		t.Error("Old child should remain with non-archivable parent")
	}
}

// TestFormatArchiveEntryUsesParentDate verifies that child tasks are grouped
// under parent's date in archive, not their own @done date.
// Spec: Archive sections use parent task's completion date for grouping.
func TestFormatArchiveEntryUsesParentDate(t *testing.T) {
	parentDate, _ := time.Parse("2006-01-02", "2026-01-18")
	childDate := "2026-01-15" // Different date than parent

	tasks := []ArchiveTask{
		{Content: "- [x] Parent @done(2026-01-18)", GroupDate: parentDate},
		{Content: "  - [x] Child @done(" + childDate + ")", GroupDate: parentDate}, // Uses parent's date!
	}

	result := FormatArchiveEntry(tasks)

	// Both should be under parent's date section
	if !containsString(result, "## 2026-01-18") {
		t.Error("Archive should have parent's date header")
	}

	// Should NOT have child's date as a separate section
	if containsString(result, "## 2026-01-15") {
		t.Error("Child's @done date should NOT create separate section")
	}

	// Both tasks should be present
	if !containsString(result, "Parent") || !containsString(result, "Child") {
		t.Error("Both tasks should be in archive")
	}
}

// TestChildDoneTagPreserved verifies that child's @done tag is preserved
// even though it's grouped by parent's date.
// Spec: Child's original @done tag remains unchanged in archived content.
func TestChildDoneTagPreserved(t *testing.T) {
	parentDate, _ := time.Parse("2006-01-02", "2026-01-18")
	childDateStr := "2026-01-15"

	tasks := []ArchiveTask{
		{Content: "- [x] Parent @done(2026-01-18)", GroupDate: parentDate},
		{Content: "  - [x] Child @done(" + childDateStr + ")", GroupDate: parentDate},
	}

	result := FormatArchiveEntry(tasks)

	// Child's original @done tag should be preserved
	if !containsString(result, "@done("+childDateStr+")") {
		t.Error("Child's original @done tag should be preserved")
	}
}

// TestNonTaskChildArchivedWithParent verifies that non-task children (plain bullet points)
// are archived together with their completed parent.
// Spec: Non-task lines (- text without checkbox) are treated as completed and archive with parent.
func TestNonTaskChildArchivedWithParent(t *testing.T) {
	now := time.Now()
	oldDate := now.AddDate(0, 0, -5).Format("2006-01-02")

	content := `- [x] Old parent @done(` + oldDate + `)
  - Note line without checkbox
  - Another note`

	archivableTasks, remaining := FilterArchivable(content, 2)
	archivable := archiveTasksToString(archivableTasks)

	// Parent should be archived
	if !containsString(archivable, "Old parent") {
		t.Error("Old parent should be archivable")
	}

	// Non-task children should be archived with parent
	if !containsString(archivable, "Note line without checkbox") {
		t.Error("Non-task child should be archived with parent")
	}
	if !containsString(archivable, "Another note") {
		t.Error("All non-task children should be archived with parent")
	}

	// Nothing should remain (except possibly empty lines)
	trimmed := strings.TrimSpace(remaining)
	if trimmed != "" {
		t.Errorf("Remaining should be empty, got %q", trimmed)
	}
}

// TestNonTaskChildNotArchivedWhenParentIncomplete verifies that non-task children
// are NOT archived when parent is incomplete.
// Spec: Non-task lines follow parent's archive status.
func TestNonTaskChildNotArchivedWhenParentIncomplete(t *testing.T) {
	content := `- [ ] Incomplete parent
  - Note line without checkbox`

	archivableTasks, remaining := FilterArchivable(content, 2)
	archivable := archiveTasksToString(archivableTasks)

	// Nothing should be archived
	if containsString(archivable, "Note line") {
		t.Error("Non-task child should NOT be archived when parent is incomplete")
	}

	// Both should remain
	if !containsString(remaining, "Incomplete parent") {
		t.Error("Incomplete parent should remain")
	}
	if !containsString(remaining, "Note line without checkbox") {
		t.Error("Non-task child of incomplete parent should remain")
	}
}

// =============================================================================
// File Operations Tests
// =============================================================================

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

package cli

import (
	"testing"
)

// TestParseNoArgs verifies that Parse() with no arguments returns default options.
// This is the normal case when launching the TUI mode.
func TestParseNoArgs(t *testing.T) {
	opts, err := Parse([]string{})
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if opts.Task != "" {
		t.Errorf("Task = %q, want empty", opts.Task)
	}
	if opts.ShowHelp {
		t.Error("ShowHelp = true, want false")
	}
	if opts.ShowVersion {
		t.Error("ShowVersion = true, want false")
	}
}

// TestParseHelp verifies that -h and --help flags set ShowHelp to true.
// Both short and long forms should work identically.
func TestParseHelp(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short flag -h", []string{"-h"}},
		{"long flag --help", []string{"--help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse(%v) error: %v", tt.args, err)
			}
			if !opts.ShowHelp {
				t.Errorf("Parse(%v) ShowHelp = false, want true", tt.args)
			}
		})
	}
}

// TestParseVersion verifies that -v and --version flags set ShowVersion to true.
// Both short and long forms should work identically.
func TestParseVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"short flag -v", []string{"-v"}},
		{"long flag --version", []string{"--version"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse(%v) error: %v", tt.args, err)
			}
			if !opts.ShowVersion {
				t.Errorf("Parse(%v) ShowVersion = false, want true", tt.args)
			}
		})
	}
}

// TestParseTask verifies that -t and --task flags correctly capture task text.
// Tasks can be provided with quotes or as multiple arguments.
func TestParseTask(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"short flag with quoted string", []string{"-t", "buy milk"}, "buy milk"},
		{"long flag with quoted string", []string{"--task", "buy milk"}, "buy milk"},
		{"short flag with multiple words", []string{"-t", "buy", "kitchen", "paper"}, "buy kitchen paper"},
		{"long flag with multiple words", []string{"--task", "buy", "kitchen", "paper"}, "buy kitchen paper"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse(%v) error: %v", tt.args, err)
			}
			if opts.Task != tt.expected {
				t.Errorf("Parse(%v) Task = %q, want %q", tt.args, opts.Task, tt.expected)
			}
		})
	}
}

// TestUsage verifies that Usage() returns a non-empty help text.
// The help text should contain essential usage information.
func TestUsage(t *testing.T) {
	usage := Usage()

	if usage == "" {
		t.Error("Usage() returned empty string")
	}

	// Should contain key elements
	expectedPhrases := []string{"ttt", "-t", "--task", "--help", "--version"}
	for _, phrase := range expectedPhrases {
		if !contains(usage, phrase) {
			t.Errorf("Usage() should contain %q", phrase)
		}
	}
}

// TestVersionString verifies that VersionString() returns formatted version info.
// The format should be "ttt version X.Y.Z".
func TestVersionString(t *testing.T) {
	Version = "1.0.0"
	vs := VersionString()
	expected := "ttt version 1.0.0"

	if vs != expected {
		t.Errorf("VersionString() = %q, want %q", vs, expected)
	}
}

// helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

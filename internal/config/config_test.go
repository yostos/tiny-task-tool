package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDefault verifies that Default() returns a Config with all expected default values.
// This ensures new installations work correctly without a config file.
func TestDefault(t *testing.T) {
	cfg := Default()

	// Verify file settings
	if cfg.File.WorkingDir != "~/.ttt" {
		t.Errorf("WorkingDir = %q, want %q", cfg.File.WorkingDir, "~/.ttt")
	}

	// Verify archive settings
	if cfg.Archive.Auto != false {
		t.Errorf("Archive.Auto = %v, want %v", cfg.Archive.Auto, false)
	}
	if cfg.Archive.DelayDays != 2 {
		t.Errorf("Archive.DelayDays = %d, want %d", cfg.Archive.DelayDays, 2)
	}

	// Verify git settings
	if cfg.Git.AutoCommit != true {
		t.Errorf("Git.AutoCommit = %v, want %v", cfg.Git.AutoCommit, true)
	}

	// Verify keybindings
	expectedUp := []string{"k"}
	if len(cfg.Keybindings.Up) != 1 || cfg.Keybindings.Up[0] != "k" {
		t.Errorf("Keybindings.Up = %v, want %v", cfg.Keybindings.Up, expectedUp)
	}
}

// TestConfigDir verifies that ConfigDir() respects XDG_CONFIG_HOME.
// Spec: docs/specification.md "設定ファイル仕様 > 配置場所" section.
// 1. XDG_CONFIG_HOME が設定されている場合 → その値を返す
// 2. 設定されていない場合 → os.UserConfigDir() を返す
func TestConfigDir(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Run("XDG_CONFIG_HOME set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		result, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error: %v", err)
		}
		if result != "/custom/config" {
			t.Errorf("ConfigDir() = %q, want %q", result, "/custom/config")
		}
	})

	// Test with XDG_CONFIG_HOME not set (uses os.UserConfigDir())
	t.Run("XDG_CONFIG_HOME not set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		result, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error: %v", err)
		}
		expected, _ := os.UserConfigDir()
		if result != expected {
			t.Errorf("ConfigDir() = %q, want %q (os.UserConfigDir())", result, expected)
		}
	})
}

// TestConfigPath verifies that ConfigPath() returns the correct config file path.
// Spec: docs/specification.md "設定ファイル仕様 > 配置場所" section.
// 1. XDG_CONFIG_HOME が設定されている場合 → $XDG_CONFIG_HOME/ttt/config.toml
// 2. 設定されていない場合 → os.UserConfigDir()/ttt/config.toml
func TestConfigPath(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Run("XDG_CONFIG_HOME set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		path, err := ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() error: %v", err)
		}
		expected := "/custom/config/ttt/config.toml"
		if path != expected {
			t.Errorf("ConfigPath() = %q, want %q", path, expected)
		}
	})

	// Test with XDG_CONFIG_HOME not set
	t.Run("XDG_CONFIG_HOME not set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		path, err := ConfigPath()
		if err != nil {
			t.Fatalf("ConfigPath() error: %v", err)
		}
		userConfigDir, _ := os.UserConfigDir()
		expected := filepath.Join(userConfigDir, "ttt", "config.toml")
		if path != expected {
			t.Errorf("ConfigPath() = %q, want %q", path, expected)
		}
	})
}

// TestExpandPath verifies that ExpandPath() correctly expands ~ to home directory.
// This is essential for user-friendly path configuration.
func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"tilde path expands to home", "~/.ttt", filepath.Join(home, ".ttt")},
		{"absolute path unchanged", "/absolute/path", "/absolute/path"},
		{"relative path unchanged", "relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExpandPath(tt.input)
			if err != nil {
				t.Errorf("ExpandPath(%q) error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestWorkingDir verifies that WorkingDir() returns the expanded working directory path.
// This method combines the config value with path expansion.
func TestWorkingDir(t *testing.T) {
	cfg := Default()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error: %v", err)
	}

	workDir, err := cfg.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error: %v", err)
	}

	expected := filepath.Join(home, ".ttt")
	if workDir != expected {
		t.Errorf("WorkingDir() = %q, want %q", workDir, expected)
	}
}

// TestTasksPath verifies that TasksPath() returns the correct path to tasks.md.
// The tasks file is always named "tasks.md" within the working directory.
func TestTasksPath(t *testing.T) {
	cfg := Default()

	workDir, err := cfg.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error: %v", err)
	}

	tasksPath, err := cfg.TasksPath()
	if err != nil {
		t.Fatalf("TasksPath() error: %v", err)
	}

	expected := filepath.Join(workDir, "tasks.md")
	if tasksPath != expected {
		t.Errorf("TasksPath() = %q, want %q", tasksPath, expected)
	}
}

// TestArchivePath verifies that ArchivePath() returns the correct path to archive.md.
// The archive file is always named "archive.md" within the working directory.
func TestArchivePath(t *testing.T) {
	cfg := Default()

	workDir, err := cfg.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error: %v", err)
	}

	archivePath, err := cfg.ArchivePath()
	if err != nil {
		t.Fatalf("ArchivePath() error: %v", err)
	}

	expected := filepath.Join(workDir, "archive.md")
	if archivePath != expected {
		t.Errorf("ArchivePath() = %q, want %q", archivePath, expected)
	}
}

// TestEditorCommand verifies that EditorCommand() substitutes {file} placeholder.
// This allows flexible editor configuration with file path injection.
func TestEditorCommand(t *testing.T) {
	tests := []struct {
		name     string
		template string
		filePath string
		expected string
	}{
		{"vim with placeholder", "vim {file}", "/path/to/file.md", "vim /path/to/file.md"},
		{"vscode with wait flag", "code --wait {file}", "/tmp/tasks.md", "code --wait /tmp/tasks.md"},
		{"emacs in terminal", "emacs -nw {file}", "~/notes.md", "emacs -nw ~/notes.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Editor: EditorConfig{Command: tt.template},
			}
			result := cfg.EditorCommand(tt.filePath)
			if result != tt.expected {
				t.Errorf("EditorCommand(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

// TestLoadNonExistentConfig verifies that Load() creates config file with defaults when it doesn't exist.
// Spec: docs/specification.md "設定ファイル仕様 > 自動作成" section.
// 設定ファイルが存在しない場合、初回起動時にデフォルト値で自動作成する。
func TestLoadNonExistentConfig(t *testing.T) {
	// Use temporary directory for test
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configPath := filepath.Join(tmpDir, "ttt", "config.toml")

	// Verify config file doesn't exist yet
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("Config file should not exist before Load()")
	}

	// Load should create the file
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Should return default values
	if cfg.File.WorkingDir != "~/.ttt" {
		t.Errorf("WorkingDir = %q, want %q", cfg.File.WorkingDir, "~/.ttt")
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Load() should create config file when it doesn't exist")
	}
}

// TestLoadCreatesDirectory verifies that Load() creates the config directory if it doesn't exist.
// Spec: docs/specification.md "設定ファイル仕様 > 自動作成" section.
// ディレクトリが存在しない場合も自動的に作成する。
func TestLoadCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := filepath.Join(tmpDir, "ttt")

	// Verify directory doesn't exist yet
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Fatal("Config directory should not exist before Load()")
	}

	// Load should create the directory
	_, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(configDir)
	if os.IsNotExist(err) {
		t.Error("Load() should create config directory when it doesn't exist")
	}
	if err == nil && !info.IsDir() {
		t.Error("Config path should be a directory")
	}
}

// TestLoadExistingConfig verifies that Load() reads existing config file without overwriting.
// 既存の設定ファイルがある場合は読み込むだけで上書きしない。
func TestLoadExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	configDir := filepath.Join(tmpDir, "ttt")
	configPath := filepath.Join(configDir, "config.toml")

	// Create config directory and file with custom value
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll() error: %v", err)
	}
	customConfig := `[file]
working_dir = "~/custom-tasks"
`
	if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	// Load should read existing config
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	// Should return custom value, not default
	if cfg.File.WorkingDir != "~/custom-tasks" {
		t.Errorf("WorkingDir = %q, want %q", cfg.File.WorkingDir, "~/custom-tasks")
	}
}

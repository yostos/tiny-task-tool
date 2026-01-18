// Package config handles configuration loading and defaults for ttt.
package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the application configuration.
type Config struct {
	File        FileConfig        `toml:"file"`
	Archive     ArchiveConfig     `toml:"archive"`
	Editor      EditorConfig      `toml:"editor"`
	Keybindings KeybindingsConfig `toml:"keybindings"`
	Git         GitConfig         `toml:"git"`
}

// FileConfig defines file location settings.
type FileConfig struct {
	WorkingDir string `toml:"working_dir"`
}

// ArchiveConfig defines archive behavior settings.
type ArchiveConfig struct {
	Auto      bool `toml:"auto"`
	DelayDays int  `toml:"delay_days"`
}

// EditorConfig defines editor settings.
type EditorConfig struct {
	Command string `toml:"command"`
}

// KeybindingsConfig defines customizable key bindings.
type KeybindingsConfig struct {
	Up           []string `toml:"up"`
	Down         []string `toml:"down"`
	Top          []string `toml:"top"`
	Bottom       []string `toml:"bottom"`
	HalfPageUp   []string `toml:"half_page_up"`
	HalfPageDown []string `toml:"half_page_down"`
}

// GitConfig defines git integration settings.
type GitConfig struct {
	AutoCommit bool `toml:"auto_commit"`
}

// Fixed file names (not configurable).
const (
	TasksFileName   = "tasks.md"
	ArchiveFileName = "archive.md"
)

// Default returns a Config with default values.
func Default() *Config {
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vi"
	}
	editorCmd += " {file}"

	return &Config{
		File: FileConfig{
			WorkingDir: "~/.ttt",
		},
		Archive: ArchiveConfig{
			Auto:      false,
			DelayDays: 2,
		},
		Editor: EditorConfig{
			Command: editorCmd,
		},
		Keybindings: KeybindingsConfig{
			Up:           []string{"k"},
			Down:         []string{"j"},
			Top:          []string{"g", "Home"},
			Bottom:       []string{"G", "End"},
			HalfPageUp:   []string{"ctrl+u"},
			HalfPageDown: []string{"ctrl+d"},
		},
		Git: GitConfig{
			AutoCommit: true,
		},
	}
}

// ConfigDir returns the config directory.
// Checks XDG_CONFIG_HOME first, falls back to os.UserConfigDir().
func ConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg, nil
	}
	return os.UserConfigDir()
}

// ConfigPath returns the path to the configuration file.
// Uses XDG_CONFIG_HOME if set, otherwise os.UserConfigDir()/ttt/config.toml.
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ttt", "config.toml"), nil
}

// Load reads the configuration from the config file.
// If the file doesn't exist, it creates one with default values.
func Load() (*Config, error) {
	cfg := Default()

	configPath, err := ConfigPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create config file with defaults
			if err := Save(cfg); err != nil {
				return nil, err
			}
			return cfg, nil
		}
		return nil, err
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ExpandPath expands ~ to the user's home directory.
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

// WorkingDir returns the expanded working directory path.
func (c *Config) WorkingDir() (string, error) {
	return ExpandPath(c.File.WorkingDir)
}

// TasksPath returns the full path to the tasks file.
func (c *Config) TasksPath() (string, error) {
	dir, err := c.WorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, TasksFileName), nil
}

// ArchivePath returns the full path to the archive file.
func (c *Config) ArchivePath() (string, error) {
	dir, err := c.WorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ArchiveFileName), nil
}

// EditorCommand returns the editor command with the file path substituted.
func (c *Config) EditorCommand(filePath string) string {
	return strings.ReplaceAll(c.Editor.Command, "{file}", filePath)
}

// Save writes the configuration to the config file.
// Creates the directory if it doesn't exist.
func Save(cfg *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	// Create directory if needed
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal config to TOML
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

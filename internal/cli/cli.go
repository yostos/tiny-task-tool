// Package cli handles command-line argument parsing for ttt.
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// Version is set at build time.
var Version = "dev"

// Options represents parsed command-line options.
type Options struct {
	Task        string
	ShowHelp    bool
	ShowVersion bool
}

// Parse parses command-line arguments and returns Options.
func Parse(args []string) (*Options, error) {
	opts := &Options{}

	fs := pflag.NewFlagSet("ttt", pflag.ContinueOnError)
	fs.StringVarP(&opts.Task, "task", "t", "", "Add a task (TUI is not launched)")
	fs.BoolVarP(&opts.ShowHelp, "help", "h", false, "Show help message")
	fs.BoolVarP(&opts.ShowVersion, "version", "v", false, "Show version")

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, Usage())
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Handle remaining args as additional task text when -t flag is used
	if fs.NArg() > 0 && fs.Changed("task") {
		if opts.Task == "" {
			opts.Task = strings.Join(fs.Args(), " ")
		} else {
			opts.Task = opts.Task + " " + strings.Join(fs.Args(), " ")
		}
	}

	return opts, nil
}

// Usage returns the help text.
func Usage() string {
	return `ttt - Tiny Task Tool

Usage:
  ttt                     Launch TUI
  ttt -t <task>           Add a task (TUI is not launched)
  ttt --task "<task>"     Add a task with quotes

Options:
  -t, --task <text>   Add a task to the task file
  -h, --help          Show this help message
  -v, --version       Show version

Examples:
  ttt                                    # Launch TUI
  ttt -t buy kitchen paper and wasabi    # Add task
  ttt --task "buy kitchen paper"         # Add task with quotes`
}

// VersionString returns the version string.
func VersionString() string {
	return fmt.Sprintf("ttt version %s", Version)
}

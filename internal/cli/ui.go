package cli

import (
	"os"
	"time"

	"github.com/RichardFlp/portmaster/internal/tui"
	"golang.org/x/term"
)

const uiUsage = `Usage: portmaster ui [options]

Options:
  --interval <s>  Refresh interval in seconds (default 2)
`

func (a *app) cmdUI(args []string) int {
	fs := a.flagSet("ui")
	var interval int
	fs.IntVar(&interval, "interval", 0, "")
	ok, code := a.parseFlags(fs, args, uiUsage)
	if !ok {
		return code
	}
	if fs.NArg() > 0 {
		return a.fail(ExitUsage, "ui takes no positional arguments")
	}
	if interval <= 0 {
		interval = a.config.RefreshInterval
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return a.fail(ExitError, "interactive terminal required for the UI")
	}
	err := tui.Run(os.Stdin, os.Stdout, tui.Options{
		Interval:         time.Duration(interval) * time.Second,
		IgnoredProcesses: a.config.IgnoredProcesses,
		IgnoredPorts:     a.config.IgnoredPorts,
		Browser:          a.config.Browser,
	})
	if err != nil {
		return a.fail(ExitError, "ui: %v", err)
	}
	return ExitOK
}

package cli

import (
	"fmt"

	"github.com/RichardFlp/portmaster/internal/browser"
	"github.com/RichardFlp/portmaster/internal/ports"
)

const openUsage = `Usage: portmaster open <port>

Opens the port in the default browser.
`

func (a *app) cmdOpen(args []string) int {
	fs := a.flagSet("open")
	ok, code := a.parseFlags(fs, args, openUsage)
	if !ok {
		return code
	}
	if fs.NArg() != 1 {
		return a.fail(ExitUsage, "open requires exactly one port")
	}
	port, valid := parsePortArg(fs.Arg(0))
	if !valid {
		return a.fail(ExitUsage, "invalid port %q", fs.Arg(0))
	}
	ls, err := ports.List()
	if err != nil {
		return a.fail(ExitError, "scan failed: %v", err)
	}
	var matches []ports.Listener
	for _, l := range ls {
		if l.Port == port {
			matches = append(matches, l)
		}
	}
	if len(matches) == 0 {
		return a.fail(ExitNotFound, "no process is listening on port %d", port)
	}
	url := browser.URLFor(matches)
	if err := browser.Open(url, a.config.Browser); err != nil {
		return a.fail(ExitError, "failed to open browser: %v", err)
	}
	fmt.Fprintf(a.stdout, "Opened %s\n", url)
	return ExitOK
}

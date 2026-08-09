package cli

import (
	"fmt"

	"github.com/RichardFlp/portmaster/internal/browser"
	"github.com/RichardFlp/portmaster/internal/inspect"
	"github.com/RichardFlp/portmaster/internal/ports"
)

const inspectUsage = `Usage: portmaster <port> [options]

Options:
  --json  Output JSON
  --open  Open the port in the browser
`

func (a *app) cmdInspect(args []string) int {
	fs := a.flagSet("inspect")
	var jsonOut, openBrowser bool
	fs.BoolVar(&jsonOut, "json", false, "")
	fs.BoolVar(&openBrowser, "open", false, "")
	ok, code := a.parseFlags(fs, args, inspectUsage)
	if !ok {
		return code
	}
	if fs.NArg() != 1 {
		return a.fail(ExitUsage, "inspect requires exactly one port")
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
	result := inspect.Build(port, matches)
	if jsonOut {
		return a.writeJSON(result)
	}
	fmt.Fprint(a.stdout, result.Text())
	if openBrowser {
		if err := browser.Open(browser.URLFor(matches), a.config.Browser); err != nil {
			return a.fail(ExitError, "failed to open browser: %v", err)
		}
		fmt.Fprintf(a.stdout, "\nOpened %s\n", browser.URLFor(matches))
	}
	return ExitOK
}

package cli

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/RichardFlp/portmaster/internal/output"
	"github.com/RichardFlp/portmaster/internal/ports"
)

const listUsage = `Usage: portmaster list [options]

Options:
  --process <name>  Filter by process name
  --port <n>        Filter by exact port
  --protocol <p>    Filter by protocol (tcp or udp)
  --json            Output JSON
  --quiet           Output port numbers only
`

func (a *app) cmdList(args []string) int {
	fs := a.flagSet("list")
	var jsonOut, quiet bool
	var process, protocol string
	var port int
	fs.BoolVar(&jsonOut, "json", false, "")
	fs.BoolVar(&quiet, "quiet", false, "")
	fs.StringVar(&process, "process", "", "")
	fs.StringVar(&protocol, "protocol", "", "")
	fs.IntVar(&port, "port", 0, "")
	ok, code := a.parseFlags(fs, args, listUsage)
	if !ok {
		return code
	}
	if fs.NArg() > 0 {
		return a.fail(ExitUsage, "list takes no positional arguments")
	}
	if protocol != "" && protocol != "tcp" && protocol != "udp" {
		return a.fail(ExitUsage, "invalid protocol %q, expected tcp or udp", protocol)
	}
	portSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "port" {
			portSet = true
		}
	})
	if portSet && (port < 1 || port > 65535) {
		return a.fail(ExitUsage, "invalid port %d", port)
	}
	ls, err := ports.List()
	if err != nil {
		return a.fail(ExitError, "scan failed: %v", err)
	}
	f := filter{
		process:          process,
		port:             port,
		proto:            protocol,
		ignoredProcesses: a.config.IgnoredProcesses,
		ignoredPorts:     a.config.IgnoredPorts,
	}
	var rows []ports.Listener
	for _, l := range ls {
		if f.match(l) {
			rows = append(rows, l)
		}
	}
	if jsonOut {
		return a.writeJSON(toListenerJSON(rows))
	}
	if quiet {
		for _, l := range rows {
			fmt.Fprintln(a.stdout, l.Port)
		}
		return ExitOK
	}
	if a.config.OutputFormat == "json" {
		return a.writeJSON(toListenerJSON(rows))
	}
	writeListenerTable(a.stdout, rows)
	return ExitOK
}

func (a *app) writeJSON(v any) int {
	if err := output.JSON(a.stdout, v); err != nil {
		return a.fail(ExitError, "failed to write output: %v", err)
	}
	return ExitOK
}

func parsePortArg(raw string) (int, bool) {
	port, err := strconv.Atoi(raw)
	if err != nil || port < 1 || port > 65535 {
		return 0, false
	}
	return port, true
}

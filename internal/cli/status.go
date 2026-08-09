package cli

import (
	"fmt"

	"github.com/RichardFlp/portmaster/internal/ports"
)

const statusUsage = `Usage: portmaster status <port> [options]

Options:
  --json  Output JSON
`

type statusJSON struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
	Status   string `json:"status"`
}

func (a *app) cmdStatus(args []string) int {
	fs := a.flagSet("status")
	var jsonOut bool
	fs.BoolVar(&jsonOut, "json", false, "")
	ok, code := a.parseFlags(fs, args, statusUsage)
	if !ok {
		return code
	}
	if fs.NArg() != 1 {
		return a.fail(ExitUsage, "status requires exactly one port")
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
		if jsonOut {
			return a.writeJSON([]statusJSON{{Port: port, Status: "AVAILABLE"}})
		}
		fmt.Fprintf(a.stdout, "%d -> AVAILABLE\n", port)
		return ExitOK
	}
	var out []statusJSON
	for _, l := range matches {
		status := "LISTENING"
		if l.Protocol == "udp" {
			status = "BOUND"
		}
		out = append(out, statusJSON{Port: port, Protocol: l.Protocol, Status: status})
	}
	if jsonOut {
		return a.writeJSON(out)
	}
	for _, s := range out {
		fmt.Fprintf(a.stdout, "%d -> %s\n", port, s.Status)
	}
	return ExitOK
}

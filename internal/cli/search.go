package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RichardFlp/portmaster/internal/ports"
	"github.com/RichardFlp/portmaster/internal/processes"
)

const searchUsage = `Usage: portmaster search <query> [options]

Matches ports, PIDs, process names, executables, and command lines.

Options:
  --json   Output JSON
  --quiet  Output port numbers only
`

func (a *app) cmdSearch(args []string) int {
	fs := a.flagSet("search")
	var jsonOut, quiet bool
	fs.BoolVar(&jsonOut, "json", false, "")
	fs.BoolVar(&quiet, "quiet", false, "")
	ok, code := a.parseFlags(fs, args, searchUsage)
	if !ok {
		return code
	}
	if fs.NArg() != 1 {
		return a.fail(ExitUsage, "search requires exactly one query")
	}
	query := strings.ToLower(fs.Arg(0))
	if query == "" {
		return a.fail(ExitUsage, "search requires a non-empty query")
	}
	ls, err := ports.List()
	if err != nil {
		return a.fail(ExitError, "scan failed: %v", err)
	}
	details := make(map[int]processes.Info)
	lookup := func(pid int) processes.Info {
		if info, ok := details[pid]; ok {
			return info
		}
		if info, err := processes.Lookup(pid); err == nil {
			details[pid] = info
			return info
		}
		return processes.Info{}
	}
	var matches []ports.Listener
	for _, l := range ls {
		if matchesQuery(l, query, lookup) {
			matches = append(matches, l)
		}
	}
	if jsonOut {
		return a.writeJSON(toListenerJSON(matches))
	}
	if quiet {
		for _, l := range matches {
			fmt.Fprintln(a.stdout, l.Port)
		}
		return ExitOK
	}
	writeListenerTable(a.stdout, matches)
	return ExitOK
}

func matchesQuery(l ports.Listener, query string, lookup func(int) processes.Info) bool {
	if strings.Contains(strings.ToLower(l.Process), query) {
		return true
	}
	if strconv.Itoa(l.Port) == query || strconv.Itoa(l.PID) == query {
		return true
	}
	if l.PID > 0 {
		info := lookup(l.PID)
		if strings.Contains(strings.ToLower(info.Executable), query) ||
			strings.Contains(strings.ToLower(info.Command), query) {
			return true
		}
	}
	return false
}

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/RichardFlp/portmaster/internal/ports"
)

const rangeUsage = `Usage: portmaster range <start>-<end> [options]

Options:
  --free   List available ports instead of occupied ones
  --json   Output JSON
  --quiet  Output port numbers only
`

func (a *app) cmdRange(args []string) int {
	fs := a.flagSet("range")
	var jsonOut, quiet, free bool
	fs.BoolVar(&jsonOut, "json", false, "")
	fs.BoolVar(&quiet, "quiet", false, "")
	fs.BoolVar(&free, "free", false, "")
	ok, code := a.parseFlags(fs, args, rangeUsage)
	if !ok {
		return code
	}
	if fs.NArg() != 1 {
		return a.fail(ExitUsage, "range requires a <start>-<end> argument")
	}
	lo, hi, err := parseRange(fs.Arg(0))
	if err != nil {
		return a.fail(ExitUsage, "%v", err)
	}
	ls, err := ports.List()
	if err != nil {
		return a.fail(ExitError, "scan failed: %v", err)
	}
	occupied := make([]bool, 65536)
	var inRange []ports.Listener
	for _, l := range ls {
		if l.Port >= lo && l.Port <= hi {
			occupied[l.Port] = true
			inRange = append(inRange, l)
		}
	}
	if free {
		var available []int
		for p := lo; p <= hi; p++ {
			if !occupied[p] {
				available = append(available, p)
			}
		}
		if jsonOut {
			return a.writeJSON(available)
		}
		if quiet {
			for _, p := range available {
				fmt.Fprintln(a.stdout, p)
			}
			return ExitOK
		}
		fmt.Fprintf(a.stdout, "Available ports in %d-%d:\n\n", lo, hi)
		for _, p := range available {
			fmt.Fprintln(a.stdout, p)
		}
		return ExitOK
	}
	if jsonOut {
		return a.writeJSON(toListenerJSON(inRange))
	}
	if quiet {
		for _, l := range inRange {
			fmt.Fprintln(a.stdout, l.Port)
		}
		return ExitOK
	}
	if len(inRange) == 0 {
		fmt.Fprintf(a.stdout, "No occupied ports in %d-%d.\n", lo, hi)
		return ExitOK
	}
	writeListenerTable(a.stdout, inRange)
	return ExitOK
}

func parseRange(spec string) (int, int, error) {
	parts := strings.Split(spec, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range %q, expected <start>-<end>", spec)
	}
	lo, err1 := strconv.Atoi(parts[0])
	hi, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("invalid range %q, expected <start>-<end>", spec)
	}
	if lo < 1 || hi > 65535 || lo > hi {
		return 0, 0, fmt.Errorf("invalid range %q, ports must satisfy 1 <= start <= end <= 65535", spec)
	}
	return lo, hi, nil
}

package cli

import (
	"flag"
	"fmt"
	"net"
)

const freeUsage = `Usage: portmaster free [start] [options]

Options:
  --from <n>    Start searching at this port (default 3000)
  --count <n>   Number of ports to return (default 5)
  --json        Output JSON
  --quiet       Output port numbers only
`

func (a *app) cmdFree(args []string) int {
	fs := a.flagSet("free")
	var jsonOut, quiet bool
	var from, count int
	fs.BoolVar(&jsonOut, "json", false, "")
	fs.BoolVar(&quiet, "quiet", false, "")
	fs.IntVar(&from, "from", 0, "")
	fs.IntVar(&count, "count", 0, "")
	ok, code := a.parseFlags(fs, args, freeUsage)
	if !ok {
		return code
	}
	if fs.NArg() > 1 {
		return a.fail(ExitUsage, "free accepts a single start port argument")
	}
	countSet := false
	fromSet := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "count":
			countSet = true
		case "from":
			fromSet = true
		}
	})
	start := from
	if fs.NArg() == 1 {
		v, valid := parsePortArg(fs.Arg(0))
		if !valid {
			return a.fail(ExitUsage, "invalid start port %q", fs.Arg(0))
		}
		start = v
	}
	if fromSet && (from < 1 || from > 65535) {
		return a.fail(ExitUsage, "invalid start port %d", from)
	}
	if start == 0 {
		start = 3000
	}
	if start < 1 || start > 65535 {
		return a.fail(ExitUsage, "invalid start port %d", start)
	}
	n := count
	if n == 0 {
		n = a.config.FreeCount
	}
	if n < 1 || n > 65535 {
		return a.fail(ExitUsage, "invalid count %d", n)
	}
	if fs.NArg() == 1 && !countSet {
		n = 1
	}
	freePorts := findFree(start, n)
	if jsonOut {
		return a.writeJSON(freePorts)
	}
	if quiet {
		for _, p := range freePorts {
			fmt.Fprintln(a.stdout, p)
		}
		return ExitOK
	}
	fmt.Fprintln(a.stdout, "Available ports:")
	fmt.Fprintln(a.stdout)
	for _, p := range freePorts {
		fmt.Fprintln(a.stdout, p)
	}
	return ExitOK
}

func findFree(start, count int) []int {
	var result []int
	for p := start; p <= 65535 && len(result) < count; p++ {
		if tcpFree(p) && udpFree(p) {
			result = append(result, p)
		}
	}
	return result
}

func tcpFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func udpFree(port int) bool {
	pc, err := net.ListenPacket("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	pc.Close()
	return true
}

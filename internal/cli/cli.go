package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/RichardFlp/portmaster/internal/config"
	"github.com/RichardFlp/portmaster/internal/output"
	"github.com/RichardFlp/portmaster/internal/ports"
	"golang.org/x/term"
)

const (
	ExitOK         = 0
	ExitError      = 1
	ExitUsage      = 2
	ExitNotFound   = 3
	ExitPermission = 4
	ExitCancelled  = 5
)

type app struct {
	stdout  io.Writer
	stderr  io.Writer
	stdin   *os.File
	version string
	config  *config.Config
}

func Run(args []string, stdout, stderr io.Writer, stdin *os.File, version string) int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "portmaster: warning: %v\n", err)
		cfg = config.Defaults()
	}
	cfg.Normalize()
	return runWith(args, stdout, stderr, stdin, version, cfg)
}

func runWith(args []string, stdout, stderr io.Writer, stdin *os.File, version string, cfg *config.Config) int {
	if cfg == nil {
		cfg = config.Defaults()
	}
	a := &app{stdout: stdout, stderr: stderr, stdin: stdin, version: version, config: cfg}
	return a.run(args)
}

func (a *app) run(args []string) int {
	if len(args) == 0 {
		return a.cmdList(nil)
	}
	switch args[0] {
	case "list":
		return a.cmdList(args[1:])
	case "status":
		return a.cmdStatus(args[1:])
	case "free":
		return a.cmdFree(args[1:])
	case "kill":
		return a.cmdKill(args[1:])
	case "open":
		return a.cmdOpen(args[1:])
	case "search":
		return a.cmdSearch(args[1:])
	case "range":
		return a.cmdRange(args[1:])
	case "watch":
		return a.cmdWatch(args[1:])
	case "ui":
		return a.cmdUI(args[1:])
	case "version", "--version":
		return a.cmdVersion()
	case "help", "--help", "-h", "-help":
		return a.cmdHelp()
	default:
		if n, err := strconv.Atoi(args[0]); err == nil && n >= 1 && n <= 65535 {
			return a.cmdInspect(args)
		}
		fmt.Fprintf(a.stderr, "portmaster: unknown command %q\n\n", args[0])
		a.cmdHelp()
		return ExitUsage
	}
}

func (a *app) cmdHelp() int {
	fmt.Fprint(a.stdout, `PortMaster - inspect and manage ports and processes

Usage:
  portmaster                 List listening ports
  portmaster <port>          Inspect a port
  portmaster list            List listening ports
  portmaster status <port>   Show whether a port is in use
  portmaster free [start]    Find available ports
  portmaster range <a-b>     Show ports in a range
  portmaster search <query>  Search ports and processes
  portmaster kill <port>     Terminate the process using a port
  portmaster kill --pid N    Terminate a process by PID
  portmaster open <port>     Open a port in the browser
  portmaster watch           Watch for port changes
  portmaster ui              Interactive terminal UI
  portmaster version         Print the version
  portmaster help            Show this help

Run "portmaster <command> --help" for details on a command.
`)
	return ExitOK
}

func (a *app) cmdVersion() int {
	fmt.Fprintf(a.stdout, "portmaster v%s\n", a.version)
	return ExitOK
}

func (a *app) flagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.stderr)
	fs.Usage = func() {}
	return fs
}

func (a *app) parseFlags(fs *flag.FlagSet, args []string, usage string) (bool, int) {
	if err := fs.Parse(reorderArgs(args, fs)); err != nil {
		if err == flag.ErrHelp {
			fmt.Fprint(a.stdout, usage)
			return false, ExitOK
		}
		fmt.Fprint(a.stderr, usage)
		return false, ExitUsage
	}
	return true, ExitOK
}

func reorderArgs(args []string, fs *flag.FlagSet) []string {
	known := make(map[string]bool)
	boolFlags := make(map[string]bool)
	fs.VisitAll(func(f *flag.Flag) {
		known[f.Name] = true
		if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
			boolFlags[f.Name] = true
		}
	})
	var flags, positionals []string
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			name := strings.TrimLeft(arg, "-")
			hasValue := false
			if eq := strings.Index(name, "="); eq >= 0 {
				name = name[:eq]
				hasValue = true
			}
			if known[name] {
				flags = append(flags, arg)
				if !hasValue && !boolFlags[name] && i+1 < len(args) {
					i++
					flags = append(flags, args[i])
				}
				i++
				continue
			}
		}
		positionals = append(positionals, arg)
		i++
	}
	return append(flags, positionals...)
}

func (a *app) fail(code int, format string, args ...any) int {
	fmt.Fprintf(a.stderr, "portmaster: "+format+"\n", args...)
	return code
}

type filter struct {
	process          string
	port             int
	proto            string
	ignoredProcesses []string
	ignoredPorts     []int
}

func (f filter) match(l ports.Listener) bool {
	if f.port > 0 && l.Port != f.port {
		return false
	}
	if f.proto != "" && l.Protocol != f.proto {
		return false
	}
	if f.process != "" && !strings.Contains(strings.ToLower(l.Process), strings.ToLower(f.process)) {
		return false
	}
	return !ports.Ignored(l, f.ignoredProcesses, f.ignoredPorts)
}

func (a *app) applyIgnores(ls []ports.Listener) []ports.Listener {
	f := filter{
		ignoredProcesses: a.config.IgnoredProcesses,
		ignoredPorts:     a.config.IgnoredPorts,
	}
	var out []ports.Listener
	for _, l := range ls {
		if f.match(l) {
			out = append(out, l)
		}
	}
	return out
}

type listenerJSON struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
	PID      int    `json:"pid"`
	Process  string `json:"process"`
}

func toListenerJSON(ls []ports.Listener) []listenerJSON {
	out := make([]listenerJSON, 0, len(ls))
	for _, l := range ls {
		out = append(out, listenerJSON{
			Port:     l.Port,
			Protocol: l.Protocol,
			Address:  l.Address,
			PID:      l.PID,
			Process:  l.Process,
		})
	}
	return out
}

func displayProcess(name string) string {
	if name == "" {
		return "-"
	}
	return name
}

func displayPID(pid int) string {
	if pid == 0 {
		return "-"
	}
	return strconv.Itoa(pid)
}

func outputTable(w io.Writer, header []string, rows [][]string) {
	output.Table(w, header, rows, terminalWidth())
}

func writeListenerTable(w io.Writer, ls []ports.Listener) {
	rows := make([][]string, 0, len(ls))
	for _, l := range ls {
		rows = append(rows, []string{
			strconv.Itoa(l.Port),
			strings.ToUpper(l.Protocol),
			displayPID(l.PID),
			displayProcess(l.Process),
			l.Address,
		})
	}
	outputTable(w, []string{"PORT", "PROTOCOL", "PID", "PROCESS", "ADDRESS"}, rows)
}

func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 0
	}
	return w
}

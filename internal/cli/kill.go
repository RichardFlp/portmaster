package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/RichardFlp/portmaster/internal/ports"
	"github.com/RichardFlp/portmaster/internal/processes"
)

const killUsage = `Usage: portmaster kill <port> [options]
       portmaster kill --pid <n> [options]

Options:
  --force  Skip confirmation
`

type killTarget struct {
	pid     int
	port    int
	process string
	command string
}

func (a *app) cmdKill(args []string) int {
	fs := a.flagSet("kill")
	var force bool
	var pid int
	fs.BoolVar(&force, "force", false, "")
	fs.IntVar(&pid, "pid", 0, "")
	ok, code := a.parseFlags(fs, args, killUsage)
	if !ok {
		return code
	}
	if fs.NArg() > 1 {
		return a.fail(ExitUsage, "kill accepts a single port argument")
	}
	pidSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "pid" {
			pidSet = true
		}
	})
	if pidSet && pid < 1 {
		return a.fail(ExitUsage, "invalid pid %d", pid)
	}
	var targets []killTarget
	if pid > 0 {
		if fs.NArg() > 0 {
			return a.fail(ExitUsage, "--pid cannot be combined with a port")
		}
		info, err := processes.Lookup(pid)
		if err != nil {
			if processes.Exists(pid) {
				return a.fail(ExitPermission, "unable to inspect process %d. Additional permissions may be required.", pid)
			}
			return a.fail(ExitNotFound, "no such process %d", pid)
		}
		targets = []killTarget{{pid: pid, process: info.Name, command: info.Command}}
	} else if fs.NArg() == 1 {
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
		seen := make(map[int]bool)
		for _, l := range matches {
			if l.PID == 0 || seen[l.PID] {
				continue
			}
			seen[l.PID] = true
			t := killTarget{pid: l.PID, port: l.Port, process: l.Process}
			if info, err := processes.Lookup(l.PID); err == nil {
				t.process = info.Name
				t.command = info.Command
			}
			targets = append(targets, t)
		}
		if len(targets) == 0 {
			return a.fail(ExitError, "no owning process found for port %d\nhint: %s", port, noOwnerHint(port))
		}
	} else {
		return a.fail(ExitUsage, "kill requires a port or --pid")
	}
	for _, t := range targets {
		if t.pid == os.Getpid() {
			return a.fail(ExitError, "refusing to terminate portmaster itself")
		}
	}
	for i, t := range targets {
		if len(targets) > 1 {
			fmt.Fprintf(a.stdout, "Target %d:\n\n", i+1)
		} else if t.port > 0 {
			fmt.Fprintf(a.stdout, "Port %d is used by:\n\n", t.port)
		} else {
			fmt.Fprintf(a.stdout, "Process %d:\n\n", t.pid)
		}
		fmt.Fprintf(a.stdout, "PID:     %d\n", t.pid)
		if t.process != "" {
			fmt.Fprintf(a.stdout, "Process: %s\n", t.process)
		}
		if t.command != "" {
			fmt.Fprintf(a.stdout, "Command: %s\n", t.command)
		}
		fmt.Fprintln(a.stdout)
	}
	confirmed := force
	if !force {
		confirmed = a.confirm("Kill this process? [y/N] ")
	}
	if !confirmed {
		return a.fail(ExitCancelled, "aborted by user")
	}
	exit := ExitOK
	for _, t := range targets {
		if err := processes.Kill(t.pid); err != nil {
			code := ExitError
			if isPermissionError(err) {
				code = ExitPermission
				fmt.Fprintln(a.stderr, "Additional permissions may be required.")
			}
			fmt.Fprintf(a.stderr, "portmaster: failed to terminate process %d: %v\n", t.pid, err)
			if exit == ExitOK {
				exit = code
			}
			continue
		}
		fmt.Fprintf(a.stdout, "Terminated process %d (%s).\n", t.pid, t.process)
	}
	return exit
}

func (a *app) confirm(prompt string) bool {
	fmt.Fprint(a.stdout, prompt)
	if a.stdin == nil {
		return false
	}
	reader := bufio.NewReader(a.stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

// noOwnerHint explains why a listener's process could not be identified and
// how to gain the privileges needed to terminate it. This typically happens
// when the listener is owned by another user, for example a system service.
func noOwnerHint(port int) string {
	base := "the process may be owned by another user (e.g. a system service)"
	if runtime.GOOS == "windows" {
		return base + ". Run portmaster from an Administrator terminal to kill it."
	}
	return fmt.Sprintf("%s. Try: sudo portmaster kill %d", base, port)
}

func isPermissionError(err error) bool {
	msg := strings.ToLower(err.Error())
	return errors.Is(err, os.ErrPermission) ||
		errors.Is(err, syscall.EACCES) ||
		strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "access is denied")
}

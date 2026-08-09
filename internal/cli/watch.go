package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/RichardFlp/portmaster/internal/ports"
)

const watchUsage = `Usage: portmaster watch [options]

Options:
  --interval <s>  Refresh interval in seconds (default 2)
  --json          Output JSON events
`

type watchEvent struct {
	Event    string       `json:"event"`
	Listener listenerJSON `json:"listener"`
}

func (a *app) cmdWatch(args []string) int {
	fs := a.flagSet("watch")
	var jsonOut bool
	var interval int
	fs.BoolVar(&jsonOut, "json", false, "")
	fs.IntVar(&interval, "interval", 0, "")
	ok, code := a.parseFlags(fs, args, watchUsage)
	if !ok {
		return code
	}
	if fs.NArg() > 0 {
		return a.fail(ExitUsage, "watch takes no positional arguments")
	}
	if interval <= 0 {
		interval = a.config.WatchInterval
	}
	delay := time.Duration(interval) * time.Second
	prev, err := ports.List()
	if err != nil {
		return a.fail(ExitError, "scan failed: %v", err)
	}
	prev = a.applyIgnores(prev)
	if !jsonOut {
		writeWatchTable(a.stdout, prev)
		fmt.Fprintln(a.stdout)
		fmt.Fprintln(a.stdout, "Watching for changes... (Ctrl+C to quit)")
		fmt.Fprintln(a.stdout)
	} else {
		for _, l := range prev {
			if err := emitWatchJSON(a.stdout, "add", l); err != nil {
				return a.fail(ExitError, "failed to write output: %v", err)
			}
		}
	}
	ticker := time.NewTicker(delay)
	defer ticker.Stop()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)
	for {
		select {
		case <-quit:
			return ExitOK
		case <-ticker.C:
			cur, err := ports.List()
			if err != nil {
				fmt.Fprintf(a.stderr, "portmaster: scan failed: %v\n", err)
				continue
			}
			cur = a.applyIgnores(cur)
			added, removed := ports.Diff(prev, cur)
			for _, l := range added {
				if jsonOut {
					if err := emitWatchJSON(a.stdout, "add", l); err != nil {
						return a.fail(ExitError, "failed to write output: %v", err)
					}
				} else {
					fmt.Fprintf(a.stdout, "+ %s\n", formatEvent(l))
				}
			}
			for _, l := range removed {
				if jsonOut {
					if err := emitWatchJSON(a.stdout, "remove", l); err != nil {
						return a.fail(ExitError, "failed to write output: %v", err)
					}
				} else {
					fmt.Fprintf(a.stdout, "- %s\n", formatEvent(l))
				}
			}
			prev = cur
		}
	}
}

func formatEvent(l ports.Listener) string {
	return ports.Describe(l)
}

func emitWatchJSON(w io.Writer, event string, l ports.Listener) error {
	return json.NewEncoder(w).Encode(watchEvent{
		Event:    event,
		Listener: toListenerJSON([]ports.Listener{l})[0],
	})
}

func writeWatchTable(w io.Writer, ls []ports.Listener) {
	rows := make([][]string, 0, len(ls))
	for _, l := range ls {
		rows = append(rows, []string{
			strconv.Itoa(l.Port),
			displayPID(l.PID),
			displayProcess(l.Process),
		})
	}
	outputTable(w, []string{"PORT", "PID", "PROCESS"}, rows)
}

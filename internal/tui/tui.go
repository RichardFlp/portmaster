package tui

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RichardFlp/portmaster/internal/browser"
	"github.com/RichardFlp/portmaster/internal/inspect"
	"github.com/RichardFlp/portmaster/internal/ports"
	"github.com/RichardFlp/portmaster/internal/processes"
	"golang.org/x/term"
)

type Options struct {
	Interval         time.Duration
	IgnoredProcesses []string
	IgnoredPorts     []int
	Browser          string
}

const defaultKeys = "Up/Down Navigate  Enter Inspect  K Kill  O Open  R Refresh  Q Quit"

type tuiApp struct {
	stdin     *os.File
	stdout    *os.File
	opts      Options
	listeners []ports.Listener
	selected  int
	detail    bool
	confirm   bool
	status    string
	changes   []string
	keys      chan []byte
}

func Run(stdin, stdout *os.File, opts Options) error {
	if !term.IsTerminal(int(stdin.Fd())) || !term.IsTerminal(int(stdout.Fd())) {
		return errors.New("interactive terminal required")
	}
	enableVirtualTerminal()
	oldState, err := term.MakeRaw(int(stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(stdin.Fd()), oldState)
	a := &tuiApp{
		stdin:  stdin,
		stdout: stdout,
		opts:   opts,
		keys:   make(chan []byte, 64),
	}
	if err := a.refresh(); err != nil {
		return err
	}
	fmt.Fprint(stdout, "\x1b[?1049h\x1b[?25l")
	defer fmt.Fprint(stdout, "\x1b[?25h\x1b[?1049l\x1b[0m")
	go a.readKeys()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(quit)
	resize := make(chan os.Signal, 1)
	if sigs := resizeSignals(); len(sigs) > 0 {
		signal.Notify(resize, sigs...)
		defer signal.Stop(resize)
	}
	ticker := time.NewTicker(opts.Interval)
	defer ticker.Stop()
	a.draw()
	for {
		select {
		case <-quit:
			return nil
		case <-resize:
			a.draw()
		case data := <-a.keys:
			if a.handleKeys(data) {
				return nil
			}
			a.draw()
		case <-ticker.C:
			a.refresh()
			a.draw()
		}
	}
}

func (a *tuiApp) readKeys() {
	buf := make([]byte, 64)
	for {
		n, err := a.stdin.Read(buf)
		if n > 0 {
			select {
			case a.keys <- append([]byte(nil), buf[:n]...):
			default:
			}
		}
		if err != nil {
			return
		}
	}
}

func (a *tuiApp) refresh() error {
	ls, err := ports.List()
	if err != nil {
		return err
	}
	var filtered []ports.Listener
	for _, l := range ls {
		if !ports.Ignored(l, a.opts.IgnoredProcesses, a.opts.IgnoredPorts) {
			filtered = append(filtered, l)
		}
	}
	added, removed := ports.Diff(a.listeners, filtered)
	if len(added) > 0 || len(removed) > 0 {
		var lines []string
		for _, l := range added {
			lines = append(lines, "+ "+ports.Describe(l))
		}
		for _, l := range removed {
			lines = append(lines, "- "+ports.Describe(l))
		}
		a.changes = append(a.changes, lines...)
		if len(a.changes) > 3 {
			a.changes = a.changes[len(a.changes)-3:]
		}
		a.status = lines[len(lines)-1]
	}
	a.listeners = filtered
	if len(a.listeners) > 0 && a.selected >= len(a.listeners) {
		a.selected = len(a.listeners) - 1
	}
	if len(a.listeners) == 0 {
		a.selected = 0
	}
	return nil
}

func (a *tuiApp) handleKeys(data []byte) bool {
	if a.confirm {
		a.confirm = false
		switch {
		case len(data) == 1 && (data[0] == 'y' || data[0] == 'Y'):
			a.doKill()
		case len(data) == 1 && (data[0] == 'n' || data[0] == 'N' || data[0] == 27):
			a.status = "Cancelled."
		default:
			a.status = "Press y to kill, n to cancel."
			a.confirm = true
		}
		return false
	}
	if a.detail {
		switch {
		case isEsc(data), isEnter(data):
			a.detail = false
		case isQuit(data):
			return true
		}
		return false
	}
	switch keyOf(data) {
	case "up":
		a.move(-1)
	case "down":
		a.move(1)
	case "enter":
		a.openDetail()
	case "kill":
		a.startConfirm()
	case "open":
		a.openBrowser()
	case "refresh":
		a.refresh()
		a.status = "Refreshed."
	case "quit":
		return true
	}
	return false
}

func (a *tuiApp) move(delta int) {
	if len(a.listeners) == 0 {
		return
	}
	a.selected += delta
	if a.selected < 0 {
		a.selected = 0
	}
	if a.selected >= len(a.listeners) {
		a.selected = len(a.listeners) - 1
	}
}

func (a *tuiApp) openDetail() {
	if len(a.listeners) == 0 {
		return
	}
	a.detail = true
}

func (a *tuiApp) startConfirm() {
	if len(a.listeners) == 0 {
		return
	}
	a.confirm = true
}

func (a *tuiApp) doKill() {
	if len(a.listeners) == 0 {
		return
	}
	l := a.listeners[a.selected]
	if l.PID == 0 {
		a.status = "Unknown PID; nothing to terminate."
		return
	}
	if l.PID == os.Getpid() {
		a.status = "Refusing to terminate portmaster itself."
		return
	}
	if err := processes.Kill(l.PID); err != nil {
		a.status = "Kill failed: " + err.Error()
		return
	}
	a.status = fmt.Sprintf("Terminated PID %d.", l.PID)
	a.refresh()
}

func (a *tuiApp) openBrowser() {
	if len(a.listeners) == 0 {
		return
	}
	l := a.listeners[a.selected]
	url := browser.URLFor([]ports.Listener{l})
	if err := browser.Open(url, a.opts.Browser); err != nil {
		a.status = "Failed to open browser: " + err.Error()
		return
	}
	a.status = "Opened " + url
}

func (a *tuiApp) draw() {
	width, height, err := term.GetSize(int(a.stdout.Fd()))
	if err != nil {
		return
	}
	detail := ""
	if a.detail && len(a.listeners) > 0 {
		l := a.listeners[a.selected]
		detail = inspect.Build(l.Port, []ports.Listener{l}).Text()
	}
	status := a.status
	if a.confirm && len(a.listeners) > 0 {
		l := a.listeners[a.selected]
		status = fmt.Sprintf("Kill %s (PID %d)? [y/N]", displayName(l.Process), l.PID)
	}
	fmt.Fprint(a.stdout, render(frame{
		width:     width,
		height:    height,
		title:     "PortMaster",
		listeners: a.listeners,
		selected:  a.selected,
		detail:    detail,
		status:    status,
		keys:      defaultKeys,
		changes:   a.changes,
	}))
	fmt.Fprint(a.stdout, "\x1b[J")
}

func displayName(name string) string {
	if name == "" {
		return "(unknown)"
	}
	return name
}

func keyOf(data []byte) string {
	if len(data) == 1 {
		switch data[0] {
		case 'q', 'Q', 3:
			return "quit"
		case 'j', 'J':
			return "down"
		case 'k', 'K':
			return "kill"
		case 'i', 'I', 13, 10:
			return "enter"
		case 'r', 'R':
			return "refresh"
		case 'o', 'O':
			return "open"
		}
	}
	if len(data) >= 3 && data[0] == 27 {
		switch data[1] {
		case '[':
			switch data[2] {
			case 'A':
				return "up"
			case 'B':
				return "down"
			case 'C':
				return "right"
			case 'D':
				return "left"
			}
		case 'O':
			switch data[2] {
			case 'A':
				return "up"
			case 'B':
				return "down"
			}
		}
	}
	return ""
}

func isEsc(data []byte) bool {
	return len(data) == 1 && data[0] == 27
}

func isEnter(data []byte) bool {
	return len(data) == 1 && (data[0] == 13 || data[0] == 10)
}

func isQuit(data []byte) bool {
	return len(data) == 1 && (data[0] == 'q' || data[0] == 'Q' || data[0] == 3)
}

package tui

import (
	"strings"
	"testing"

	"github.com/RichardFlp/portmaster/internal/ports"
)

func stripANSI(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		if runes[i] == 0x1b {
			for i < len(runes) && runes[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteRune(runes[i])
	}
	return b.String()
}

func TestRenderTable(t *testing.T) {
	ls := []ports.Listener{
		{Port: 3000, Protocol: "tcp", PID: 18432, Process: "node", Address: "127.0.0.1"},
		{Port: 8080, Protocol: "tcp", PID: 20112, Process: "python", Address: "0.0.0.0"},
		{Port: 5432, Protocol: "tcp", PID: 4212, Process: "postgres", Address: "127.0.0.1"},
	}
	got := render(frame{
		width:     80,
		height:    24,
		title:     "PortMaster",
		listeners: ls,
		selected:  1,
		status:    "status message",
		keys:      defaultKeys,
	})
	for _, want := range []string{
		"PortMaster",
		"3 ports",
		"PORT",
		"PROTOCOL",
		"PROCESS",
		"3000",
		"18432",
		"node",
		"8080",
		"20112",
		"python",
		"status message",
		"Q Quit",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("render missing %q:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "\x1b[7m") {
		t.Error("selected row not highlighted")
	}
	lines := strings.Count(got, "\n") + 1
	if lines != 11 {
		t.Errorf("render produced %d lines, want 11", lines)
	}
	for _, line := range strings.Split(stripANSI(got), "\n") {
		if len([]rune(line)) > 80 {
			t.Errorf("line exceeds width: %d chars", len([]rune(line)))
		}
	}
}

func TestRenderSelectedRow(t *testing.T) {
	ls := []ports.Listener{{Port: 3000, Protocol: "tcp", PID: 1, Process: "node", Address: "127.0.0.1"}}
	got := render(frame{width: 80, height: 24, title: "PortMaster", listeners: ls, selected: 0, keys: defaultKeys})
	if !strings.Contains(got, "\x1b[7m|") {
		t.Errorf("selected row marker missing:\n%s", got)
	}
}

func TestRenderSmallTerminal(t *testing.T) {
	got := render(frame{width: 20, height: 5, title: "PortMaster", keys: defaultKeys})
	if !strings.Contains(got, "too small") {
		t.Errorf("small terminal message missing: %q", got)
	}
	got = render(frame{width: 80, height: 5, title: "PortMaster", keys: defaultKeys})
	if !strings.Contains(got, "too small") {
		t.Errorf("short terminal message missing: %q", got)
	}
}

func TestRenderNarrowTerminal(t *testing.T) {
	ls := []ports.Listener{{Port: 3000, Protocol: "tcp", PID: 18432, Process: "averylongprocessname", Address: "127.0.0.1"}}
	got := render(frame{width: 34, height: 20, title: "PortMaster", listeners: ls, selected: 0, keys: defaultKeys})
	if strings.Contains(stripANSI(got), "averylongprocessname") {
		t.Errorf("long process name not truncated on narrow terminal:\n%s", got)
	}
	for _, line := range strings.Split(stripANSI(got), "\n") {
		if len([]rune(line)) > 34 {
			t.Errorf("line exceeds width 34: %d chars", len([]rune(line)))
		}
	}
}

func TestRenderDetail(t *testing.T) {
	ls := []ports.Listener{{Port: 3000, Protocol: "tcp", PID: 18432, Process: "node", Address: "127.0.0.1"}}
	got := render(frame{
		width:     80,
		height:    24,
		title:     "PortMaster",
		listeners: ls,
		selected:  0,
		detail:    "Port 3000\n\nStatus:   LISTENING\nProtocol: TCP\nAddress:  127.0.0.1:3000",
		keys:      defaultKeys,
	})
	for _, want := range []string{"Port 3000", "LISTENING", "TCP"} {
		if !strings.Contains(got, want) {
			t.Errorf("detail view missing %q:\n%s", want, got)
		}
	}
}

func TestRenderEmpty(t *testing.T) {
	got := render(frame{width: 80, height: 24, title: "PortMaster", keys: defaultKeys})
	if !strings.Contains(got, "0 ports") {
		t.Errorf("empty state missing count:\n%s", got)
	}
	if !strings.Contains(got, "PORT") {
		t.Errorf("empty state missing header:\n%s", got)
	}
}

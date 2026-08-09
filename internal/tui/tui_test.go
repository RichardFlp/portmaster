package tui

import (
	"testing"

	"github.com/RichardFlp/portmaster/internal/ports"
)

func testApp() *tuiApp {
	return &tuiApp{
		listeners: []ports.Listener{
			{Port: 3000, Protocol: "tcp", PID: 1, Process: "node", Address: "127.0.0.1"},
			{Port: 8080, Protocol: "tcp", PID: 2, Process: "python", Address: "0.0.0.0"},
			{Port: 5432, Protocol: "tcp", PID: 3, Process: "postgres", Address: "127.0.0.1"},
		},
		selected: 0,
		keys:     make(chan []byte, 16),
	}
}

func TestHandleKeysNavigation(t *testing.T) {
	a := testApp()
	if a.handleKeys([]byte{'j'}) {
		t.Error("down should not quit")
	}
	if a.selected != 1 {
		t.Errorf("selected = %d, want 1", a.selected)
	}
	if a.handleKeys([]byte{27, '[', 'A'}) {
		t.Error("up should not quit")
	}
	if a.selected != 0 {
		t.Errorf("selected = %d, want 0", a.selected)
	}
	a.selected = 2
	a.handleKeys([]byte{'j'})
	if a.selected != 2 {
		t.Errorf("down past end clamped to %d, want 2", a.selected)
	}
	a.selected = 0
	a.handleKeys([]byte{27, '[', 'A'})
	if a.selected != 0 {
		t.Errorf("up past start clamped to %d, want 0", a.selected)
	}
}

func TestHandleKeysQuit(t *testing.T) {
	a := testApp()
	if !a.handleKeys([]byte{'q'}) {
		t.Error("q should quit")
	}
	a = testApp()
	if !a.handleKeys([]byte{'Q'}) {
		t.Error("Q should quit")
	}
	a = testApp()
	if !a.handleKeys([]byte{3}) {
		t.Error("Ctrl+C should quit")
	}
	a = testApp()
	if a.handleKeys([]byte{27, '[', 'B'}) {
		t.Error("down should not quit")
	}
	if !a.handleKeys([]byte{'q'}) {
		t.Error("q after navigation should quit")
	}
}

func TestHandleKeysDetail(t *testing.T) {
	a := testApp()
	if a.handleKeys([]byte{'i'}) {
		t.Error("i should open detail, not quit")
	}
	if !a.detail {
		t.Error("detail not opened")
	}
	if a.handleKeys([]byte{27}) {
		t.Error("esc should not quit")
	}
	if a.detail {
		t.Error("detail not closed on esc")
	}
	a.handleKeys([]byte{'i'})
	if a.handleKeys([]byte{13}) {
		t.Error("enter should not quit")
	}
	if a.detail {
		t.Error("detail not closed on enter")
	}
}

func TestHandleKeysKillConfirm(t *testing.T) {
	a := testApp()
	a.handleKeys([]byte{'k'})
	if !a.confirm {
		t.Fatal("kill should request confirmation")
	}
	if a.handleKeys([]byte{'n'}) {
		t.Error("declining kill should not quit")
	}
	if a.confirm {
		t.Error("confirm state not cleared")
	}
	if a.status != "Cancelled." {
		t.Errorf("status = %q, want Cancelled.", a.status)
	}
}

func TestHandleKeysKillUnknownPID(t *testing.T) {
	a := testApp()
	a.listeners[0].PID = 0
	a.handleKeys([]byte{'k'})
	a.handleKeys([]byte{'y'})
	if a.status != "Unknown PID; nothing to terminate." {
		t.Errorf("status = %q", a.status)
	}
}

func TestHandleKeysRefresh(t *testing.T) {
	a := testApp()
	a.handleKeys([]byte{'r'})
	if a.status != "Refreshed." {
		t.Errorf("status = %q, want Refreshed.", a.status)
	}
}

func TestKeyOf(t *testing.T) {
	cases := []struct {
		data []byte
		want string
	}{
		{[]byte{'j'}, "down"},
		{[]byte{'k'}, "kill"},
		{[]byte{'K'}, "kill"},
		{[]byte{'i'}, "enter"},
		{[]byte{13}, "enter"},
		{[]byte{'r'}, "refresh"},
		{[]byte{'o'}, "open"},
		{[]byte{'q'}, "quit"},
		{[]byte{27, '[', 'A'}, "up"},
		{[]byte{27, '[', 'B'}, "down"},
		{[]byte{27, 'O', 'A'}, "up"},
		{[]byte{'x'}, ""},
		{[]byte{27}, ""},
	}
	for _, c := range cases {
		if got := keyOf(c.data); got != c.want {
			t.Errorf("keyOf(%v) = %q, want %q", c.data, got, c.want)
		}
	}
}

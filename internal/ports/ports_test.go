package ports

import (
	"net"
	"testing"
	"time"
)

func TestSplitAddress(t *testing.T) {
	cases := []struct {
		raw  string
		addr string
		port int
		ok   bool
	}{
		{"127.0.0.1:3000", "127.0.0.1", 3000, true},
		{"0.0.0.0:80", "0.0.0.0", 80, true},
		{"*:5353", "0.0.0.0", 5353, true},
		{"[::1]:3000", "::1", 3000, true},
		{"[::]:3000", "::", 3000, true},
		{"[0:0:0:0:0:0:0:0]:8080", "::", 8080, true},
		{"[fe80::1]:8080", "fe80::1", 8080, true},
		{"192.168.1.5:22", "192.168.1.5", 22, true},
		{"", "", 0, false},
		{"no-port", "", 0, false},
		{"3000", "", 0, false},
		{"127.0.0.1:0", "", 0, false},
		{"127.0.0.1:99999", "", 0, false},
		{"127.0.0.1:", "", 0, false},
	}
	for _, c := range cases {
		addr, port, ok := splitAddress(c.raw)
		if addr != c.addr || port != c.port || ok != c.ok {
			t.Errorf("splitAddress(%q) = %q, %d, %v; want %q, %d, %v",
				c.raw, addr, port, ok, c.addr, c.port, c.ok)
		}
	}
}

func TestDiff(t *testing.T) {
	prev := []Listener{
		{Port: 3000, Protocol: "tcp", PID: 1},
		{Port: 8080, Protocol: "tcp", PID: 2},
		{Port: 5353, Protocol: "udp", PID: 3},
	}
	cur := []Listener{
		{Port: 3000, Protocol: "tcp", PID: 1},
		{Port: 4200, Protocol: "tcp", PID: 4},
		{Port: 5353, Protocol: "udp", PID: 3},
	}
	added, removed := Diff(prev, cur)
	if len(added) != 1 || added[0].Port != 4200 {
		t.Errorf("added = %v, want port 4200", added)
	}
	if len(removed) != 1 || removed[0].Port != 8080 {
		t.Errorf("removed = %v, want port 8080", removed)
	}
}

func TestDiffStableOnNoChange(t *testing.T) {
	ls := []Listener{
		{Port: 3000, Protocol: "tcp", PID: 1},
		{Port: 8080, Protocol: "tcp", PID: 2},
	}
	added, removed := Diff(ls, ls)
	if len(added) != 0 || len(removed) != 0 {
		t.Errorf("expected no changes, got added=%v removed=%v", added, removed)
	}
}

func TestDescribe(t *testing.T) {
	l := Listener{Port: 4200, Protocol: "tcp", PID: 22014, Process: "node"}
	if got := Describe(l); got != "4200 TCP node PID 22014" {
		t.Errorf("Describe = %q", got)
	}
}

func TestIgnored(t *testing.T) {
	l := Listener{Port: 3000, Protocol: "tcp", Process: "node.exe"}
	if !Ignored(l, nil, []int{3000}) {
		t.Error("Ignored port not detected")
	}
	if !Ignored(l, []string{"NODE.EXE"}, nil) {
		t.Error("Ignored process not detected")
	}
	if Ignored(l, []string{"python"}, []int{8080}) {
		t.Error("non-matching ignore lists should not ignore")
	}
}

func TestListFindsBoundTCP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	waitForListener(t, port, "tcp")
}

func TestListFindsBoundUDP(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer pc.Close()
	port := pc.LocalAddr().(*net.UDPAddr).Port
	waitForListener(t, port, "udp")
}

func waitForListener(t *testing.T, port int, proto string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ls, err := List()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		if containsListener(ls, port, proto) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("List() never reported bound %s port %d", proto, port)
}

func TestListSortedByPort(t *testing.T) {
	ls, err := List()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(ls); i++ {
		if ls[i].Port < ls[i-1].Port {
			t.Errorf("list not sorted at index %d: %v > %v", i, ls[i-1].Port, ls[i].Port)
		}
	}
}

func containsListener(ls []Listener, port int, proto string) bool {
	for _, l := range ls {
		if l.Port == port && l.Protocol == proto {
			return true
		}
	}
	return false
}

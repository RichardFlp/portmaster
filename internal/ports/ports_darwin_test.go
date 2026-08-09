//go:build darwin

package ports

import "testing"

func TestParseLsof(t *testing.T) {
	out := "p1234\ncnode\nn127.0.0.1:3000\n" +
		"p5678\ncpostgres\nn*:5432\n" +
		"p9999\ncchrome\nn[::1]:8080\n" +
		"p1111\ncnode\nn127.0.0.1:5353->8.8.8.8:53\n" +
		"p2222\ncnode\nn*:3000\n"
	ls := parseLsof(out, "tcp")
	if len(ls) != 4 {
		t.Fatalf("parseLsof returned %d listeners, want 4: %v", len(ls), ls)
	}
	check := func(i int, addr string, port int, pid int, name string) {
		l := ls[i]
		if l.Address != addr || l.Port != port || l.PID != pid || l.Process != name {
			t.Errorf("listener %d = %s:%d pid %d name %s, want %s:%d pid %d name %s",
				i, l.Address, l.Port, l.PID, l.Process, addr, port, pid, name)
		}
	}
	check(0, "127.0.0.1", 3000, 1234, "node")
	check(1, "0.0.0.0", 5432, 5678, "postgres")
	check(2, "::1", 8080, 9999, "chrome")
	check(3, "0.0.0.0", 3000, 2222, "node")
}

func TestParseLsofEmpty(t *testing.T) {
	ls := parseLsof("", "udp")
	if len(ls) != 0 {
		t.Errorf("expected no listeners, got %v", ls)
	}
}

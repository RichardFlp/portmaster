//go:build windows

package ports

import "testing"

func TestParseNetstatLine(t *testing.T) {
	cases := []struct {
		line  string
		addr  string
		port  int
		proto string
		pid   int
		ok    bool
	}{
		{"  TCP    0.0.0.0:3000           0.0.0.0:0              LISTENING       18432",
			"0.0.0.0", 3000, "tcp", 18432, true},
		{"  TCP    127.0.0.1:5173         127.0.0.1:0             LISTENING       19281",
			"127.0.0.1", 5173, "tcp", 19281, true},
		{"  TCP    [::]:8080              [::]:0                  LISTENING       20112",
			"::", 8080, "tcp", 20112, true},
		{"  TCP    192.168.1.5:22         0.0.0.0:0               LISTENING       4",
			"192.168.1.5", 22, "tcp", 4, true},
		{"  TCP    [::1]:3000              [::]:0                  LISTENING       18432",
			"::1", 3000, "tcp", 18432, true},
		{"  UDP    0.0.0.0:5353           *:*                                    1234",
			"0.0.0.0", 5353, "udp", 1234, true},
		{"  UDP    [::]:5353              *:*                                    1234",
			"::", 5353, "udp", 1234, true},
		{"  UDP    *:5353                 *:*                                    1234",
			"0.0.0.0", 5353, "udp", 1234, true},
		{"  TCP    0.0.0.0:3000           0.0.0.0:0              ESTABLISHED     18432",
			"", 0, "", 0, false},
		{"  TCP    0.0.0.0:3000           0.0.0.0:0              TIME_WAIT        0",
			"", 0, "", 0, false},
		{"Active Connections", "", 0, "", 0, false},
		{"garbage line", "", 0, "", 0, false},
	}
	for _, c := range cases {
		l, ok := parseNetstatLine(c.line)
		if ok != c.ok {
			t.Errorf("parseNetstatLine(%q) ok = %v, want %v", c.line, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if l.Address != c.addr || l.Port != c.port || l.Protocol != c.proto || l.PID != c.pid {
			t.Errorf("parseNetstatLine(%q) = %s:%d %s pid %d, want %s:%d %s pid %d",
				c.line, l.Address, l.Port, l.Protocol, l.PID, c.addr, c.port, c.proto, c.pid)
		}
	}
}

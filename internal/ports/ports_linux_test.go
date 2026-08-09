//go:build linux

package ports

import "testing"

func TestParseProcNetLine(t *testing.T) {
	cases := []struct {
		line  string
		proto string
		ipv6  bool
		addr  string
		port  int
		ok    bool
	}{
		{
			"  0: 0100007F:0BB8 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 12345 1 0000000000000000 100 0 0 10 0",
			"tcp", false, "127.0.0.1", 3000, true,
		},
		{
			"  1: 00000000:1F90 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 54321 1 0000000000000000 100 0 0 10 0",
			"tcp", false, "0.0.0.0", 8080, true,
		},
		{
			"  0: 00000000000000000000000001000000:0BB8 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 9999 1 0000000000000000 100 0 0 10 0",
			"tcp", true, "::1", 3000, true,
		},
		{
			"  0: 00000000000000000000000000000000:0BB8 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 8888 1 0000000000000000 100 0 0 10 0",
			"tcp", true, "::", 3000, true,
		},
		{
			"  0: 00000000000000000000000000000000:0BB8 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 8888 1 0000000000000000 100 0 0 10 0",
			"tcp", true, "::", 3000, true,
		},
		{
			"  1: 0100007F:0BB8 00000000:0000 01 00000000:00000000 00:00000000 00000000  1000        0 12345 1 0000000000000000 100 0 0 10 0",
			"tcp", false, "", 0, false,
		},
		{
			"  0: 0100007F:14E9 00000000:0000 07 00000000:00000000 00:00000000 00000000  1000        0 7777 1 0000000000000000 100 0 0 10 0",
			"udp", false, "127.0.0.1", 5353, true,
		},
		{
			"  1: 0100007F:14E9 0100007F:0035 07 00000000:00000000 00:00000000 00000000  1000        0 7777 1 0000000000000000 100 0 0 10 0",
			"udp", false, "", 0, false,
		},
		{
			"  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode",
			"tcp", false, "", 0, false,
		},
		{
			"garbage",
			"tcp", false, "", 0, false,
		},
	}
	for _, c := range cases {
		l, inode, ok := parseProcNetLine(c.line, c.proto, c.ipv6)
		if ok != c.ok {
			t.Errorf("parseProcNetLine ok = %v, want %v", ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if l.Address != c.addr || l.Port != c.port {
			t.Errorf("parseProcNetLine = %s:%d, want %s:%d", l.Address, l.Port, c.addr, c.port)
		}
		if inode == 0 {
			t.Error("expected nonzero inode")
		}
	}
}

func TestIsZeroHexAddress(t *testing.T) {
	if !isZeroHexAddress("00000000:0000") {
		t.Error("00000000:0000 should be zero")
	}
	if !isZeroHexAddress("00000000000000000000000000000000:0000") {
		t.Error("all-zero ipv6 should be zero")
	}
	if isZeroHexAddress("0100007F:0035") {
		t.Error("nonzero address reported as zero")
	}
}

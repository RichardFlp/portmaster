package browser

import (
	"testing"

	"github.com/RichardFlp/portmaster/internal/ports"
)

func TestURLFor(t *testing.T) {
	cases := []struct {
		name string
		ls   []ports.Listener
		want string
	}{
		{
			"loopback",
			[]ports.Listener{{Port: 3000, Protocol: "tcp", Address: "127.0.0.1"}},
			"http://127.0.0.1:3000",
		},
		{
			"wildcard",
			[]ports.Listener{{Port: 3000, Protocol: "tcp", Address: "0.0.0.0"}},
			"http://localhost:3000",
		},
		{
			"ipv6 wildcard",
			[]ports.Listener{{Port: 3000, Protocol: "tcp", Address: "::"}},
			"http://localhost:3000",
		},
		{
			"ipv6 loopback",
			[]ports.Listener{{Port: 3000, Protocol: "tcp", Address: "::1"}},
			"http://[::1]:3000",
		},
		{
			"specific address",
			[]ports.Listener{{Port: 3000, Protocol: "tcp", Address: "192.168.1.5"}},
			"http://192.168.1.5:3000",
		},
		{
			"https on 443",
			[]ports.Listener{{Port: 443, Protocol: "tcp", Address: "0.0.0.0"}},
			"https://localhost:443",
		},
		{
			"prefers loopback over specific",
			[]ports.Listener{
				{Port: 3000, Protocol: "tcp", Address: "192.168.1.5"},
				{Port: 3000, Protocol: "tcp", Address: "127.0.0.1"},
			},
			"http://127.0.0.1:3000",
		},
		{
			"ipv6 specific address",
			[]ports.Listener{{Port: 8080, Protocol: "tcp", Address: "fe80::1"}},
			"http://[fe80::1]:8080",
		},
	}
	for _, c := range cases {
		if got := URLFor(c.ls); got != c.want {
			t.Errorf("%s: URLFor = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestURLForEmpty(t *testing.T) {
	if got := URLFor(nil); got != "" {
		t.Errorf("URLFor(nil) = %q, want empty", got)
	}
}

package browser

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/RichardFlp/portmaster/internal/ports"
)

func Open(url, preferred string) error {
	if preferred != "" {
		cmd := exec.Command(preferred, url)
		return cmd.Start()
	}
	return openDefault(url)
}

func URLFor(ls []ports.Listener) string {
	if len(ls) == 0 {
		return ""
	}
	port := ls[0].Port
	scheme := "http"
	if port == 443 {
		scheme = "https"
	}
	host := "localhost"
	for _, l := range ls {
		switch l.Address {
		case "127.0.0.1":
			return fmt.Sprintf("%s://127.0.0.1:%d", scheme, port)
		case "::1":
			return fmt.Sprintf("%s://[::1]:%d", scheme, port)
		}
		if l.Address != "" && l.Address != "0.0.0.0" && l.Address != "::" {
			host = l.Address
		}
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") {
		host = "[" + host + "]"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}

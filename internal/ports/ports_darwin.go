package ports

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func scan() ([]Listener, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var result []Listener
	for _, mode := range []struct {
		args  []string
		proto string
	}{
		{[]string{"-nP", "-w", "-iTCP", "-sTCP:LISTEN", "-F", "pcn"}, "tcp"},
		{[]string{"-nP", "-w", "-iUDP", "-F", "pcn"}, "udp"},
	} {
		out, err := exec.CommandContext(ctx, "lsof", mode.args...).Output()
		if err != nil && len(out) == 0 {
			return nil, err
		}
		result = append(result, parseLsof(string(out), mode.proto)...)
	}
	return result, nil
}

func parseLsof(out, proto string) []Listener {
	var result []Listener
	pid := 0
	name := ""
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			p, err := strconv.Atoi(line[1:])
			if err != nil {
				continue
			}
			pid = p
		case 'c':
			name = line[1:]
		case 'n':
			if strings.Contains(line, "->") {
				continue
			}
			addr, port, ok := splitAddress(line[1:])
			if !ok {
				continue
			}
			result = append(result, Listener{
				Port:     port,
				Protocol: proto,
				Address:  addr,
				PID:      pid,
				Process:  name,
			})
		}
	}
	return result
}

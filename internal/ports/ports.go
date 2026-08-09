package ports

import (
	"sort"
	"strconv"
	"strings"
)

type Listener struct {
	Port     int
	Protocol string
	Address  string
	PID      int
	Process  string
}

func List() ([]Listener, error) {
	ls, err := scan()
	if err != nil {
		return nil, err
	}
	sort.Slice(ls, func(i, j int) bool {
		if ls[i].Port != ls[j].Port {
			return ls[i].Port < ls[j].Port
		}
		return ls[i].Protocol < ls[j].Protocol
	})
	return ls, nil
}

func Diff(prev, cur []Listener) (added, removed []Listener) {
	key := func(l Listener) string {
		return l.Protocol + ":" + strconv.Itoa(l.Port)
	}
	prevSet := make(map[string]Listener, len(prev))
	curSet := make(map[string]Listener, len(cur))
	for _, l := range prev {
		prevSet[key(l)] = l
	}
	for _, l := range cur {
		curSet[key(l)] = l
	}
	for _, l := range cur {
		if _, ok := prevSet[key(l)]; !ok {
			added = append(added, l)
		}
	}
	for _, l := range prev {
		if _, ok := curSet[key(l)]; !ok {
			removed = append(removed, l)
		}
	}
	sort.Slice(added, func(i, j int) bool { return added[i].Port < added[j].Port })
	sort.Slice(removed, func(i, j int) bool { return removed[i].Port < removed[j].Port })
	return added, removed
}

func Describe(l Listener) string {
	parts := []string{strconv.Itoa(l.Port), strings.ToUpper(l.Protocol)}
	if l.Process != "" {
		parts = append(parts, l.Process)
	}
	if l.PID > 0 {
		parts = append(parts, "PID", strconv.Itoa(l.PID))
	}
	return strings.Join(parts, " ")
}

func Ignored(l Listener, processes []string, ports []int) bool {
	for _, p := range ports {
		if l.Port == p {
			return true
		}
	}
	for _, name := range processes {
		if strings.EqualFold(l.Process, name) {
			return true
		}
	}
	return false
}

func splitAddress(raw string) (string, int, bool) {
	if raw == "" {
		return "", 0, false
	}
	addr, portStr, ok := cutHostPort(raw)
	if !ok {
		return "", 0, false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return "", 0, false
	}
	switch addr {
	case "*":
		addr = "0.0.0.0"
	case "[::]":
		addr = "::"
	case "[0:0:0:0:0:0:0:0]":
		addr = "::"
	}
	return strings.Trim(addr, "[]"), port, true
}

func cutHostPort(raw string) (string, string, bool) {
	if strings.HasPrefix(raw, "[") {
		end := strings.LastIndex(raw, "]")
		if end < 0 || end+1 >= len(raw) || raw[end+1] != ':' {
			return "", "", false
		}
		return raw[:end+1], raw[end+2:], true
	}
	idx := strings.LastIndex(raw, ":")
	if idx < 0 || idx == len(raw)-1 {
		return "", "", false
	}
	return raw[:idx], raw[idx+1:], true
}

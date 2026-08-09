package ports

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type socketRecord struct {
	listener Listener
	inode    uint64
}

type socketOwner struct {
	pid  int
	name string
}

func scan() ([]Listener, error) {
	var records []socketRecord
	inodes := make(map[uint64]struct{})
	for _, src := range []struct {
		path  string
		proto string
		ipv6  bool
	}{
		{"/proc/net/tcp", "tcp", false},
		{"/proc/net/tcp6", "tcp", true},
		{"/proc/net/udp", "udp", false},
		{"/proc/net/udp6", "udp", true},
	} {
		if err := parseProcNet(src.path, src.proto, src.ipv6, &records, inodes); err != nil {
			return nil, err
		}
	}
	owners := ownersForInodes(inodes)
	seen := make(map[uint64]bool, len(records))
	out := make([]Listener, 0, len(records))
	for _, r := range records {
		if seen[r.inode] {
			continue
		}
		seen[r.inode] = true
		if o, ok := owners[r.inode]; ok {
			r.listener.PID = o.pid
			r.listener.Process = o.name
		}
		out = append(out, r.listener)
	}
	return out, nil
}

func parseProcNet(path, proto string, ipv6 bool, records *[]socketRecord, inodes map[uint64]struct{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		l, inode, ok := parseProcNetLine(line, proto, ipv6)
		if !ok {
			continue
		}
		inodes[inode] = struct{}{}
		*records = append(*records, socketRecord{listener: l, inode: inode})
	}
	return nil
}

func parseProcNetLine(line, proto string, ipv6 bool) (Listener, uint64, bool) {
	fields := strings.Fields(line)
	if len(fields) < 10 || !strings.HasSuffix(fields[0], ":") {
		return Listener{}, 0, false
	}
	addr, port, ok := parseHexAddress(fields[1], ipv6)
	if !ok || port == 0 {
		return Listener{}, 0, false
	}
	state := fields[3]
	if proto == "tcp" && state != "0A" {
		return Listener{}, 0, false
	}
	if proto == "udp" && (state != "07" || !isZeroHexAddress(fields[2])) {
		return Listener{}, 0, false
	}
	inode, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return Listener{}, 0, false
	}
	return Listener{Port: port, Protocol: proto, Address: addr}, inode, true
}

func parseHexAddress(hex string, ipv6 bool) (string, int, bool) {
	idx := strings.LastIndex(hex, ":")
	if idx < 0 {
		return "", 0, false
	}
	port, err := strconv.ParseUint(hex[idx+1:], 16, 16)
	if err != nil {
		return "", 0, false
	}
	if ipv6 {
		if len(hex[:idx]) != 32 {
			return "", 0, false
		}
		ip := make(net.IP, 16)
		for i := 0; i < 4; i++ {
			group, err := strconv.ParseUint(hex[idx-32+i*8:idx-24+i*8], 16, 32)
			if err != nil {
				return "", 0, false
			}
			ip[i*4] = byte(group)
			ip[i*4+1] = byte(group >> 8)
			ip[i*4+2] = byte(group >> 16)
			ip[i*4+3] = byte(group >> 24)
		}
		return ip.String(), int(port), true
	}
	if len(hex[:idx]) != 8 {
		return "", 0, false
	}
	var b [4]byte
	for i := 0; i < 4; i++ {
		v, err := strconv.ParseUint(hex[idx-8+i*2:idx-6+i*2], 16, 8)
		if err != nil {
			return "", 0, false
		}
		b[i] = byte(v)
	}
	return fmt.Sprintf("%d.%d.%d.%d", b[3], b[2], b[1], b[0]), int(port), true
}

func isZeroHexAddress(hex string) bool {
	idx := strings.LastIndex(hex, ":")
	if idx < 0 {
		return false
	}
	for _, c := range hex[:idx] {
		if c != '0' {
			return false
		}
	}
	for _, c := range hex[idx+1:] {
		if c != '0' {
			return false
		}
	}
	return true
}

func ownersForInodes(inodes map[uint64]struct{}) map[uint64]socketOwner {
	owners := make(map[uint64]socketOwner)
	if len(inodes) == 0 {
		return owners
	}
	procs, err := os.ReadDir("/proc")
	if err != nil {
		return owners
	}
	for _, p := range procs {
		if len(owners) == len(inodes) {
			break
		}
		if !p.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(p.Name())
		if err != nil {
			continue
		}
		fdDir := "/proc/" + p.Name() + "/fd"
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range fds {
			target, err := os.Readlink(fdDir + "/" + fd.Name())
			if err != nil || !strings.HasPrefix(target, "socket:[") {
				continue
			}
			inode, err := strconv.ParseUint(strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]"), 10, 64)
			if err != nil {
				continue
			}
			if _, want := inodes[inode]; !want {
				continue
			}
			if _, done := owners[inode]; done {
				continue
			}
			comm, _ := os.ReadFile("/proc/" + p.Name() + "/comm")
			owners[inode] = socketOwner{pid: pid, name: strings.TrimSpace(string(comm))}
		}
	}
	return owners
}

package ports

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func scan() ([]Listener, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "netstat", "-ano").Output()
	if err != nil {
		return nil, err
	}
	names := processNames()
	var result []Listener
	for _, line := range strings.Split(string(out), "\n") {
		l, ok := parseNetstatLine(line)
		if !ok {
			continue
		}
		l.Process = names[l.PID]
		result = append(result, l)
	}
	return result, nil
}

func parseNetstatLine(line string) (Listener, bool) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return Listener{}, false
	}
	proto := strings.ToLower(fields[0])
	isTCP := proto == "tcp" || proto == "tcpv6"
	isUDP := proto == "udp" || proto == "udpv6"
	if !isTCP && !isUDP {
		return Listener{}, false
	}
	addr, port, ok := splitAddress(fields[1])
	if !ok {
		return Listener{}, false
	}
	pid := 0
	if isTCP {
		if len(fields) < 5 {
			return Listener{}, false
		}
		state := strings.ToUpper(fields[3])
		if state != "LISTENING" && state != "LISTEN" {
			return Listener{}, false
		}
		pid, _ = strconv.Atoi(fields[4])
	} else {
		pid, _ = strconv.Atoi(fields[3])
	}
	return Listener{Port: port, Protocol: proto[:3], Address: addr, PID: pid}, true
}

func processNames() map[int]string {
	snapshot, err := syscall.CreateToolhelp32Snapshot(0x2, 0)
	if err != nil {
		return map[int]string{}
	}
	defer syscall.CloseHandle(snapshot)
	names := make(map[int]string)
	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	err = syscall.Process32First(snapshot, &entry)
	for err == nil {
		names[int(entry.ProcessID)] = syscall.UTF16ToString(entry.ExeFile[:])
		err = syscall.Process32Next(snapshot, &entry)
	}
	return names
}

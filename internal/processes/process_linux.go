package processes

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

func lookup(pid int) (Info, error) {
	info := Info{PID: pid}
	stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return info, err
	}
	raw := string(stat)
	end := strings.LastIndex(raw, ")")
	if end < 0 || end+1 >= len(raw) {
		return info, fmt.Errorf("invalid stat data for pid %d", pid)
	}
	rest := strings.Fields(raw[end+1:])
	if len(rest) >= 2 {
		if ppid, perr := strconv.Atoi(rest[1]); perr == nil {
			info.ParentPID = ppid
		}
	}
	if len(rest) >= 20 {
		if ticks, terr := strconv.ParseInt(rest[19], 10, 64); terr == nil {
			if btime, bok := bootTime(); bok {
				info.Started = time.Unix(btime+ticks/clockTicksPerSecond(), 0)
			}
		}
	}
	info.Name = strings.TrimSpace(readFile(fmt.Sprintf("/proc/%d/comm", pid)))
	info.Command = cmdline(pid)
	info.Executable = readlink(fmt.Sprintf("/proc/%d/exe", pid))
	return info, nil
}

func cmdline(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return ""
	}
	parts := strings.Split(string(data), "\x00")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, " ")
}

const atClktck = 17

func clockTicksPerSecond() int64 {
	data, err := os.ReadFile("/proc/self/auxv")
	if err != nil {
		return 100
	}
	word := int(unsafe.Sizeof(uintptr(0)))
	for i := 0; i+word*2 <= len(data); i += word * 2 {
		if nativeUint(data[i:i+word], word) != atClktck {
			continue
		}
		if hz := nativeUint(data[i+word:i+word*2], word); hz > 0 {
			return int64(hz)
		}
	}
	return 100
}

func nativeUint(b []byte, word int) uint64 {
	if word == 8 {
		return binary.NativeEndian.Uint64(b)
	}
	return uint64(binary.NativeEndian.Uint32(b))
}

func bootTime() (int64, bool) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, false
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "btime ") {
			v, err := strconv.ParseInt(strings.TrimSpace(line[len("btime "):]), 10, 64)
			if err == nil {
				return v, true
			}
		}
	}
	return 0, false
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func readlink(path string) string {
	v, err := os.Readlink(path)
	if err != nil {
		return ""
	}
	return v
}

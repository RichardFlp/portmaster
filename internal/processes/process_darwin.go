package processes

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var errNoProcess = errors.New("no such process")

func lookup(pid int) (Info, error) {
	out, err := exec.Command("ps", "-o", "pid=,ppid=,comm=,args=,lstart=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return Info{}, err
	}
	line := strings.TrimSpace(string(out))
	if line == "" {
		return Info{}, errNoProcess
	}
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return Info{}, errNoProcess
	}
	info := Info{PID: pid}
	info.ParentPID, _ = strconv.Atoi(fields[1])
	name := fields[2]
	info.Name = filepath.Base(name)
	if strings.Contains(name, "/") {
		info.Executable = name
	}
	info.Command = strings.Join(fields[3:len(fields)-5], " ")
	if started, err := time.Parse("Mon Jan _2 15:04:05 2006", strings.Join(fields[len(fields)-5:], " ")); err == nil {
		info.Started = started
	}
	return info, nil
}

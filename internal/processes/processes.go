package processes

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Info struct {
	PID        int
	Name       string
	Executable string
	Command    string
	ParentPID  int
	ParentName string
	Started    time.Time
}

func Lookup(pid int) (Info, error) {
	if pid <= 0 {
		return Info{}, errors.New("invalid pid " + strconv.Itoa(pid))
	}
	info, err := lookup(pid)
	if err != nil {
		return info, err
	}
	if info.ParentPID > 0 {
		if parent, perr := lookup(info.ParentPID); perr == nil {
			info.ParentName = parent.Name
		}
	}
	return info, nil
}

func Kill(pid int) error {
	if pid == os.Getpid() {
		return errors.New("refusing to terminate the current process")
	}
	return kill(pid)
}

func Exists(pid int) bool {
	return exists(pid)
}

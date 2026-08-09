//go:build linux || darwin

package processes

import "syscall"

func kill(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

func exists(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

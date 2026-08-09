//go:build linux || darwin

package tui

import (
	"os"
	"syscall"
)

func resizeSignals() []os.Signal {
	return []os.Signal{syscall.SIGWINCH}
}

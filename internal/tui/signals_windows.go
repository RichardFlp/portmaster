//go:build windows

package tui

import "os"

func resizeSignals() []os.Signal {
	return nil
}

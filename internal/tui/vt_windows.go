//go:build windows

package tui

import (
	"syscall"
	"unsafe"
)

const (
	enableVirtualTerminalInput      = 0x0200
	enableVirtualTerminalProcessing = 0x0004
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func enableVirtualTerminal() {
	setConsoleMode(syscall.STD_INPUT_HANDLE, enableVirtualTerminalInput)
	setConsoleMode(syscall.STD_OUTPUT_HANDLE, enableVirtualTerminalProcessing)
}

func setConsoleMode(handle int, add uint32) {
	h, err := syscall.GetStdHandle(handle)
	if err != nil {
		return
	}
	var mode uint32
	r1, _, _ := procGetConsoleMode.Call(uintptr(h), uintptr(unsafe.Pointer(&mode)))
	if r1 == 0 {
		return
	}
	procSetConsoleMode.Call(uintptr(h), uintptr(mode|add))
}

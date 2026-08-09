package processes

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	processTerminate        = 0x0001
	processVMRead           = 0x0010
	processQueryInformation = 0x0400
	processQueryLimitedInfo = 0x1000
)

var (
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	ntdll                     = syscall.NewLazyDLL("ntdll.dll")
	procQueryFullProcessImage = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetProcessTimes       = kernel32.NewProc("GetProcessTimes")
	procNtQueryInformation    = ntdll.NewProc("NtQueryInformationProcess")
)

func kill(pid int) error {
	h, err := syscall.OpenProcess(processTerminate, false, uint32(pid))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(h)
	return syscall.TerminateProcess(h, 1)
}

func exists(pid int) bool {
	h, err := syscall.OpenProcess(processQueryLimitedInfo, false, uint32(pid))
	if err != nil {
		return false
	}
	syscall.CloseHandle(h)
	return true
}

func lookup(pid int) (Info, error) {
	info := Info{PID: pid}
	snapshot, err := syscall.CreateToolhelp32Snapshot(0x2, 0)
	if err == nil {
		defer syscall.CloseHandle(snapshot)
		var entry syscall.ProcessEntry32
		entry.Size = uint32(unsafe.Sizeof(entry))
		err = syscall.Process32First(snapshot, &entry)
		for err == nil {
			if int(entry.ProcessID) == pid {
				info.Name = syscall.UTF16ToString(entry.ExeFile[:])
				info.ParentPID = int(entry.ParentProcessID)
				break
			}
			err = syscall.Process32Next(snapshot, &entry)
		}
	}
	h, err := syscall.OpenProcess(processQueryInformation|processVMRead, false, uint32(pid))
	if err != nil {
		return info, err
	}
	defer syscall.CloseHandle(h)
	if exe := executablePath(h); exe != "" {
		info.Executable = exe
	}
	if cmd := commandLine(h); cmd != "" {
		info.Command = cmd
	}
	if start, ok := processStartTime(h); ok {
		info.Started = start
	}
	return info, nil
}

func executablePath(h syscall.Handle) string {
	buf := make([]uint16, 1024)
	size := uint32(len(buf))
	r1, _, _ := procQueryFullProcessImage.Call(
		uintptr(h),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r1 == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:size])
}

func commandLine(h syscall.Handle) string {
	var pbi processBasicInformation
	status, _, _ := procNtQueryInformation.Call(
		uintptr(h),
		0,
		uintptr(unsafe.Pointer(&pbi)),
		unsafe.Sizeof(pbi),
		0,
	)
	if status != 0 || pbi.PebBaseAddress == 0 {
		return ""
	}
	paramsAddr, err := readMemoryUintptr(h, pbi.PebBaseAddress+0x20)
	if err != nil || paramsAddr == 0 {
		return ""
	}
	var us unicodeString
	if err := readMemory(h, paramsAddr+0x70, unsafe.Pointer(&us), unsafe.Sizeof(us)); err != nil {
		return ""
	}
	if us.Length == 0 || us.Buffer == 0 {
		return ""
	}
	buf := make([]byte, us.Length)
	if err := readMemory(h, us.Buffer, unsafe.Pointer(&buf[0]), uintptr(len(buf))); err != nil {
		return ""
	}
	return syscall.UTF16ToString(unsafe.Slice((*uint16)(unsafe.Pointer(&buf[0])), us.Length/2))
}

func processStartTime(h syscall.Handle) (time.Time, bool) {
	var created, exited, kernel, user syscall.Filetime
	r1, _, _ := procGetProcessTimes.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&created)),
		uintptr(unsafe.Pointer(&exited)),
		uintptr(unsafe.Pointer(&kernel)),
		uintptr(unsafe.Pointer(&user)),
	)
	if r1 == 0 {
		return time.Time{}, false
	}
	nanos := int64(created.HighDateTime)<<32 + int64(created.LowDateTime)
	const unixEpochAsFileTime = 116444736000000000
	return time.Unix(0, (nanos-unixEpochAsFileTime)*100), true
}

func readMemoryUintptr(h syscall.Handle, addr uintptr) (uintptr, error) {
	var out uintptr
	if err := readMemory(h, addr, unsafe.Pointer(&out), unsafe.Sizeof(out)); err != nil {
		return 0, err
	}
	return out, nil
}

func readMemory(h syscall.Handle, addr uintptr, dest unsafe.Pointer, size uintptr) error {
	return windows.ReadProcessMemory(windows.Handle(h), addr, (*byte)(dest), size, nil)
}

type processBasicInformation struct {
	ExitStatus                   int32
	PebBaseAddress               uintptr
	AffinityMask                 uintptr
	BasePriority                 int32
	UniqueProcessID              uintptr
	InheritedFromUniqueProcessID uintptr
}

type unicodeString struct {
	Length uint16
	MaxLen uint16
	Buffer uintptr
}

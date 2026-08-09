package processes

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestLookupSelf(t *testing.T) {
	info, err := Lookup(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if info.PID != os.Getpid() {
		t.Errorf("PID = %d, want %d", info.PID, os.Getpid())
	}
	if info.Name == "" {
		t.Error("Name is empty for the current process")
	}
}

func TestLookupInvalidPID(t *testing.T) {
	if _, err := Lookup(0); err == nil {
		t.Error("Lookup(0) succeeded, want error")
	}
	if _, err := Lookup(-1); err == nil {
		t.Error("Lookup(-1) succeeded, want error")
	}
}

func TestLookupNonexistentPID(t *testing.T) {
	pid := impossiblePID()
	if _, err := Lookup(pid); err == nil {
		t.Errorf("Lookup(%d) succeeded, want error", pid)
	}
}

func TestKillRefusesSelf(t *testing.T) {
	if err := Kill(os.Getpid()); err == nil {
		t.Error("Kill(self) succeeded, want error")
	}
}

func TestKillChildProcess(t *testing.T) {
	cmd := sleepingChild()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pid := cmd.Process.Pid
	defer func() {
		if cmd.ProcessState == nil {
			cmd.Process.Kill()
		}
		cmd.Process.Release()
	}()
	if !Exists(pid) {
		t.Fatalf("child %d should exist", pid)
	}
	if err := Kill(pid); err != nil {
		t.Fatalf("Kill(%d): %v", pid, err)
	}
	if err := waitForExit(cmd, 5*time.Second); err != nil {
		t.Fatalf("child did not terminate after kill: %v", err)
	}
	cmd.Process.Release()
	waitForNotExists(t, pid, 5*time.Second)
}

func waitForNotExists(t *testing.T, pid int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !Exists(pid) {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Error("child still exists after kill")
}

func TestKillNonexistentPID(t *testing.T) {
	if err := Kill(impossiblePID()); err == nil {
		t.Error("Kill of nonexistent pid succeeded, want error")
	}
}

func sleepingChild() *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	return cmd
}

func waitForExit(cmd *exec.Cmd, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return os.ErrDeadlineExceeded
	}
}

func impossiblePID() int {
	if runtime.GOOS == "windows" {
		return 1 << 30
	}
	return 1<<31 - 1
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	time.Sleep(30 * time.Second)
	os.Exit(0)
}

package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/RichardFlp/portmaster/internal/config"
	"github.com/RichardFlp/portmaster/internal/ports"
)

func execCLI(t *testing.T, args []string, stdin *os.File, cfg *config.Config) (int, string, string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	if cfg == nil {
		cfg = config.Defaults()
	}
	code := runWith(args, &out, &errBuf, stdin, "test", cfg)
	return code, out.String(), errBuf.String()
}

func bindTCP(t *testing.T) (int, io.Closer) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln.Addr().(*net.TCPAddr).Port, ln
}

func unusedPort(t *testing.T) int {
	t.Helper()
	ls, err := ports.List()
	if err != nil {
		t.Fatal(err)
	}
	used := make(map[int]bool)
	for _, l := range ls {
		used[l.Port] = true
	}
	for p := 20000; p <= 60000; p++ {
		if used[p] {
			continue
		}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err == nil {
			ln.Close()
			return p
		}
	}
	t.Fatal("no free port found")
	return 0
}

func pipeStdin(t *testing.T, input string) *os.File {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close(); w.Close() })
	go func() {
		w.WriteString(input)
		w.Close()
	}()
	return r
}

func TestRunListJSON(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, errOut := execCLI(t, []string{"list", "--json"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d, stderr = %s", code, errOut)
	}
	var ls []map[string]any
	if err := json.Unmarshal([]byte(out), &ls); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	found := false
	for _, l := range ls {
		if int(l["port"].(float64)) == port {
			found = true
			if int(l["pid"].(float64)) != os.Getpid() {
				t.Errorf("pid = %v, want %d", l["pid"], os.Getpid())
			}
			if l["protocol"] != "tcp" {
				t.Errorf("protocol = %v, want tcp", l["protocol"])
			}
			break
		}
	}
	if !found {
		t.Errorf("bound port %d not in JSON output", port)
	}
}

func TestRunListQuiet(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, _ := execCLI(t, []string{"list", "--quiet"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		t.Fatal("empty quiet output")
	}
	found := false
	for _, f := range fields {
		if _, err := strconv.Atoi(f); err != nil {
			t.Errorf("non-numeric token in quiet output: %q", f)
		}
		if f == strconv.Itoa(port) {
			found = true
		}
	}
	if !found {
		t.Errorf("bound port %d not in quiet output", port)
	}
}

func TestRunListFilters(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, _ := execCLI(t, []string{"list", "--port", strconv.Itoa(port)}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("--port filter missing %d:\n%s", port, out)
	}
	code, out, _ = execCLI(t, []string{"list", "--protocol", "udp"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if strings.Contains(out, "\nTCP") || strings.Contains(out, "  TCP  ") {
		t.Errorf("--protocol udp included tcp rows:\n%s", out)
	}
	code, out, _ = execCLI(t, []string{"list", "--process", "definitely-not-a-process-name"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("--process filter matched our listener:\n%s", out)
	}
	code, _, errOut := execCLI(t, []string{"list", "--protocol", "bogus"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("invalid protocol code = %d, want %d (%s)", code, ExitUsage, errOut)
	}
}

func TestRunListIgnoresConfigured(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	cfg := config.Defaults()
	cfg.IgnoredPorts = []int{port}
	code, out, _ := execCLI(t, []string{"list", "--quiet"}, nil, cfg)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("ignored port %d still listed:\n%s", port, out)
	}
}

func TestRunListConfigJSONFormat(t *testing.T) {
	cfg := config.Defaults()
	cfg.OutputFormat = "json"
	code, out, _ := execCLI(t, []string{"list"}, nil, cfg)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	var ls []map[string]any
	if err := json.Unmarshal([]byte(out), &ls); err != nil {
		t.Errorf("config json output invalid: %v\n%s", err, out)
	}
}

func TestRunInspect(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, errOut := execCLI(t, []string{strconv.Itoa(port)}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d (%s)", code, errOut)
	}
	for _, want := range []string{
		fmt.Sprintf("Port %d", port),
		"LISTENING",
		"TCP",
		"127.0.0.1",
		strconv.Itoa(os.Getpid()),
	} {
		if !strings.Contains(out, want) {
			t.Errorf("inspect output missing %q:\n%s", want, out)
		}
	}
}

func TestRunInspectJSON(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, _ := execCLI(t, []string{strconv.Itoa(port), "--json"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	var result struct {
		Port      int `json:"port"`
		Listeners []struct {
			Status string `json:"status"`
			PID    int    `json:"pid"`
		} `json:"listeners"`
	}
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if result.Port != port {
		t.Errorf("port = %d, want %d", result.Port, port)
	}
	if len(result.Listeners) == 0 {
		t.Fatal("no listeners in inspect json")
	}
	if result.Listeners[0].Status != "LISTENING" {
		t.Errorf("status = %q, want LISTENING", result.Listeners[0].Status)
	}
	if result.Listeners[0].PID != os.Getpid() {
		t.Errorf("pid = %d, want %d", result.Listeners[0].PID, os.Getpid())
	}
}

func TestRunInspectNotFound(t *testing.T) {
	port := unusedPort(t)
	code, _, errOut := execCLI(t, []string{strconv.Itoa(port)}, nil, nil)
	if code != ExitNotFound {
		t.Errorf("code = %d, want %d (%s)", code, ExitNotFound, errOut)
	}
}

func TestRunStatus(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, _ := execCLI(t, []string{"status", strconv.Itoa(port)}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, "LISTENING") {
		t.Errorf("status output missing LISTENING:\n%s", out)
	}
	freePort := unusedPort(t)
	code, out, _ = execCLI(t, []string{"status", strconv.Itoa(freePort)}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, "AVAILABLE") {
		t.Errorf("status output missing AVAILABLE:\n%s", out)
	}
}

func TestRunStatusJSON(t *testing.T) {
	freePort := unusedPort(t)
	code, out, _ := execCLI(t, []string{"status", strconv.Itoa(freePort), "--json"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	var result []map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(result) != 1 || result[0]["status"] != "AVAILABLE" {
		t.Errorf("status json = %s", out)
	}
}

func TestRunFree(t *testing.T) {
	code, out, _ := execCLI(t, []string{"free", "--from", "20000", "--count", "3"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	var nums []int
	for _, line := range strings.Split(out, "\n") {
		if n, err := strconv.Atoi(strings.TrimSpace(line)); err == nil {
			nums = append(nums, n)
		}
	}
	if len(nums) != 3 {
		t.Fatalf("got %d ports, want 3:\n%s", len(nums), out)
	}
	for _, n := range nums {
		if n < 20000 {
			t.Errorf("port %d below start", n)
		}
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", n))
		if err != nil {
			t.Errorf("port %d reported free but not bindable: %v", n, err)
		}
		ln.Close()
	}
}

func TestRunFreePositional(t *testing.T) {
	code, out, _ := execCLI(t, []string{"free", "20000"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	var nums []int
	for _, line := range strings.Split(out, "\n") {
		if n, err := strconv.Atoi(strings.TrimSpace(line)); err == nil {
			nums = append(nums, n)
		}
	}
	if len(nums) != 1 {
		t.Errorf("free with positional start returned %d ports, want 1:\n%s", len(nums), out)
	}
}

func TestRunFreeJSON(t *testing.T) {
	code, out, _ := execCLI(t, []string{"free", "--from", "20000", "--count", "2", "--json"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	var nums []int
	if err := json.Unmarshal([]byte(out), &nums); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(nums) != 2 {
		t.Errorf("got %d ports, want 2", len(nums))
	}
}

func TestRunFreeQuiet(t *testing.T) {
	code, out, _ := execCLI(t, []string{"free", "--from", "20000", "--count", "2", "--quiet"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if strings.Contains(out, "Available") {
		t.Errorf("quiet free output contains header:\n%s", out)
	}
}

func TestRunFreeUsage(t *testing.T) {
	code, _, _ := execCLI(t, []string{"free", "abc"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("invalid start code = %d, want %d", code, ExitUsage)
	}
	code, _, _ = execCLI(t, []string{"free", "--count", "-1"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("invalid count code = %d, want %d", code, ExitUsage)
	}
}

func TestRunRange(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	lo, hi := port-10, port+10
	if lo < 1 {
		lo = 1
	}
	spec := fmt.Sprintf("%d-%d", lo, hi)
	code, out, _ := execCLI(t, []string{"range", spec}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("range output missing occupied port %d:\n%s", port, out)
	}
	code, out, _ = execCLI(t, []string{"range", spec, "--free"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("range --free included occupied port %d:\n%s", port, out)
	}
	code, out, _ = execCLI(t, []string{"range", spec, "--json"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	var ls []map[string]any
	if err := json.Unmarshal([]byte(out), &ls); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out)
	}
	if len(ls) == 0 {
		t.Error("range json empty for occupied port")
	}
}

func TestRunRangeUsage(t *testing.T) {
	for _, bad := range []string{"abc-def", "3000", "4000-3000", "0-100", "1-65536", "-", "3000-4000-5000"} {
		code, _, _ := execCLI(t, []string{"range", bad}, nil, nil)
		if code != ExitUsage {
			t.Errorf("range %q code = %d, want %d", bad, code, ExitUsage)
		}
	}
}

func TestRunSearch(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, _ := execCLI(t, []string{"search", strconv.Itoa(port)}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("search by port missing %d:\n%s", port, out)
	}
	code, out, _ = execCLI(t, []string{"search", strconv.Itoa(os.Getpid())}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("search by pid missing bound port %d:\n%s", port, out)
	}
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	code, out, _ = execCLI(t, []string{"search", name}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("search by name %q missing bound port %d:\n%s", name, port, out)
	}
	code, out, _ = execCLI(t, []string{"search", "zzz-no-such-process-zzz"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("no-match search found our port:\n%s", out)
	}
}

func TestRunKill(t *testing.T) {
	port, child := startHelperListener(t)
	stdin := pipeStdin(t, "n\n")
	code, out, errOut := execCLI(t, []string{"kill", strconv.Itoa(port)}, stdin, nil)
	if code != ExitCancelled {
		t.Fatalf("declined kill code = %d, want %d (%s)", code, ExitCancelled, errOut)
	}
	if !strings.Contains(out, "Kill this process?") {
		t.Errorf("confirmation prompt missing:\n%s", out)
	}
	if !strings.Contains(out, "PID:") {
		t.Errorf("target info missing:\n%s", out)
	}
	if child.ProcessState != nil && child.ProcessState.Exited() {
		t.Fatal("child terminated after declined kill")
	}
	code, _, errOut = execCLI(t, []string{"kill", strconv.Itoa(port), "--force"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("forced kill code = %d (%s)", code, errOut)
	}
	if err := waitForChildExit(child, 5*time.Second); err != nil {
		t.Fatalf("child did not terminate after forced kill: %v", err)
	}
	waitForPortGone(t, port, 5*time.Second)
	code, _, errOut = execCLI(t, []string{"kill", strconv.Itoa(port), "--force"}, nil, nil)
	if code != ExitNotFound {
		t.Errorf("kill of dead port code = %d, want %d (%s)", code, ExitNotFound, errOut)
	}
}

func TestRunKillPID(t *testing.T) {
	_, child := startHelperListener(t)
	pid := child.Process.Pid
	stdin := pipeStdin(t, "y\n")
	code, out, errOut := execCLI(t, []string{"kill", "--pid", strconv.Itoa(pid)}, stdin, nil)
	if code != ExitOK {
		t.Fatalf("code = %d (%s)", code, errOut)
	}
	if !strings.Contains(out, "Terminated process") {
		t.Errorf("missing termination message:\n%s", out)
	}
	if err := waitForChildExit(child, 5*time.Second); err != nil {
		t.Fatalf("child did not terminate: %v", err)
	}
}

func TestRunKillRefusesSelf(t *testing.T) {
	code, _, errOut := execCLI(t, []string{"kill", "--pid", strconv.Itoa(os.Getpid()), "--force"}, nil, nil)
	if code != ExitError {
		t.Errorf("self kill code = %d, want %d (%s)", code, ExitError, errOut)
	}
	if !strings.Contains(errOut, "refusing") {
		t.Errorf("self kill message missing:\n%s", errOut)
	}
}

func TestNoOwnerHint(t *testing.T) {
	hint := noOwnerHint(9050)
	if runtime.GOOS == "windows" {
		if !strings.Contains(hint, "Administrator") {
			t.Errorf("windows hint missing Administrator: %q", hint)
		}
	} else {
		if !strings.Contains(hint, "sudo portmaster kill 9050") {
			t.Errorf("unix hint missing sudo command: %q", hint)
		}
	}
	if !strings.Contains(hint, "system service") {
		t.Errorf("hint missing explanation: %q", hint)
	}
}

func TestRunKillUsage(t *testing.T) {
	code, _, _ := execCLI(t, []string{"kill"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("kill without args code = %d, want %d", code, ExitUsage)
	}
	code, _, _ = execCLI(t, []string{"kill", "abc"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("kill abc code = %d, want %d", code, ExitUsage)
	}
	code, _, _ = execCLI(t, []string{"kill", "--pid", "abc"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("kill --pid abc code = %d, want %d", code, ExitUsage)
	}
	code, _, _ = execCLI(t, []string{"kill", "--pid", "99999"}, nil, nil)
	if code != ExitNotFound {
		t.Errorf("kill --pid 99999 code = %d, want %d", code, ExitNotFound)
	}
}

func TestRunOpenNotFound(t *testing.T) {
	port := unusedPort(t)
	code, _, errOut := execCLI(t, []string{"open", strconv.Itoa(port)}, nil, nil)
	if code != ExitNotFound {
		t.Errorf("code = %d, want %d (%s)", code, ExitNotFound, errOut)
	}
}

func TestRunVersion(t *testing.T) {
	code, out, _ := execCLI(t, []string{"version"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, "portmaster vtest") {
		t.Errorf("version output = %q", out)
	}
}

func TestRunHelp(t *testing.T) {
	code, out, _ := execCLI(t, []string{"help"}, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	for _, want := range []string{"Usage", "portmaster list", "portmaster kill", "portmaster ui"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	code, _, errOut := execCLI(t, []string{"frobnicate"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("code = %d, want %d (%s)", code, ExitUsage, errOut)
	}
	code, _, _ = execCLI(t, []string{"0"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("port 0 code = %d, want %d", code, ExitUsage)
	}
}

func TestRunListUsage(t *testing.T) {
	code, _, _ := execCLI(t, []string{"list", "extra"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("code = %d, want %d", code, ExitUsage)
	}
	code, _, _ = execCLI(t, []string{"list", "--port", "abc"}, nil, nil)
	if code != ExitUsage {
		t.Errorf("bad --port code = %d, want %d", code, ExitUsage)
	}
}

func TestRunEmptyIsList(t *testing.T) {
	port, closer := bindTCP(t)
	defer closer.Close()
	code, out, _ := execCLI(t, nil, nil, nil)
	if code != ExitOK {
		t.Fatalf("code = %d", code)
	}
	if !strings.Contains(out, strconv.Itoa(port)) {
		t.Errorf("default list missing bound port %d", port)
	}
}

func TestFormatEvent(t *testing.T) {
	l := ports.Listener{Port: 4200, Protocol: "tcp", PID: 22014, Process: "node"}
	if got := formatEvent(l); got != "4200 TCP node PID 22014" {
		t.Errorf("formatEvent = %q", got)
	}
	l.PID = 0
	if got := formatEvent(l); got != "4200 TCP node" {
		t.Errorf("formatEvent without pid = %q", got)
	}
}

func TestFilter(t *testing.T) {
	f := filter{process: "node", port: 3000, proto: "tcp"}
	if !f.match(ports.Listener{Port: 3000, Protocol: "tcp", Process: "node.exe"}) {
		t.Error("filter should match")
	}
	if f.match(ports.Listener{Port: 3000, Protocol: "udp", Process: "node.exe"}) {
		t.Error("protocol mismatch should not match")
	}
	if f.match(ports.Listener{Port: 8080, Protocol: "tcp", Process: "node.exe"}) {
		t.Error("port mismatch should not match")
	}
	if f.match(ports.Listener{Port: 3000, Protocol: "tcp", Process: "python.exe"}) {
		t.Error("process mismatch should not match")
	}
}

func TestParseRange(t *testing.T) {
	lo, hi, err := parseRange("3000-4000")
	if err != nil {
		t.Fatal(err)
	}
	if lo != 3000 || hi != 4000 {
		t.Errorf("parseRange = %d-%d", lo, hi)
	}
	for _, bad := range []string{"3000", "abc-def", "4000-3000", "0-100", "1-65536", "", "-", "3000-4000-5000"} {
		if _, _, err := parseRange(bad); err == nil {
			t.Errorf("parseRange(%q) succeeded, want error", bad)
		}
	}
}

func TestReorderArgs(t *testing.T) {
	fs := flag.NewFlagSet("x", flag.ContinueOnError)
	var force, jsonOut, free bool
	var pid int
	fs.BoolVar(&force, "force", false, "")
	fs.BoolVar(&jsonOut, "json", false, "")
	fs.BoolVar(&free, "free", false, "")
	fs.IntVar(&pid, "pid", 0, "")
	cases := []struct {
		in   []string
		want []string
	}{
		{[]string{"3000", "--force"}, []string{"--force", "3000"}},
		{[]string{"--pid", "1234", "--force"}, []string{"--pid", "1234", "--force"}},
		{[]string{"3000", "--json"}, []string{"--json", "3000"}},
		{[]string{"3000-4000", "--free"}, []string{"--free", "3000-4000"}},
		{[]string{"--force", "3000"}, []string{"--force", "3000"}},
		{[]string{"3000"}, []string{"3000"}},
	}
	for _, c := range cases {
		got := strings.Join(reorderArgs(c.in, fs), " ")
		want := strings.Join(c.want, " ")
		if got != want {
			t.Errorf("reorderArgs(%v) = %q, want %q", c.in, got, want)
		}
	}
}

func TestParsePortArg(t *testing.T) {
	if p, ok := parsePortArg("3000"); !ok || p != 3000 {
		t.Errorf("parsePortArg(3000) = %d, %v", p, ok)
	}
	for _, bad := range []string{"0", "65536", "-1", "abc", ""} {
		if _, ok := parsePortArg(bad); ok {
			t.Errorf("parsePortArg(%q) succeeded", bad)
		}
	}
}

func waitForPortGone(t *testing.T, port int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ls, err := ports.List()
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}
		found := false
		for _, l := range ls {
			if l.Port == port {
				found = true
				break
			}
		}
		if !found {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("port %d still visible %v after process exit", port, timeout)
}

func waitForChildExit(cmd *exec.Cmd, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return os.ErrDeadlineExceeded
	}
}

func startHelperListener(t *testing.T) (int, *exec.Cmd) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperListener")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_LISTENER=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
			cmd.Wait()
		}
	})
	portStr, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil {
		t.Fatal(err)
	}
	return port, cmd
}

func TestHelperListener(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_LISTENER") != "1" {
		return
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stdout, ln.Addr().(*net.TCPAddr).Port)
	time.Sleep(60 * time.Second)
	os.Exit(0)
}

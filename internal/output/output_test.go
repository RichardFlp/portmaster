package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	header := []string{"PORT", "PID"}
	rows := [][]string{{"3000", "18432"}, {"8080", "20112"}}
	if err := Table(&buf, header, rows, 0); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"PORT", "PID", "3000", "18432", "8080", "20112"} {
		if !strings.Contains(got, want) {
			t.Errorf("table output missing %q:\n%s", want, got)
		}
	}
}

func TestTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	if err := Table(&buf, []string{"PORT", "PID"}, nil, 0); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "PORT") {
		t.Error("header missing from empty table")
	}
}

func TestTableTruncatesToWidth(t *testing.T) {
	var buf bytes.Buffer
	header := []string{"PROCESS"}
	rows := [][]string{{"averyveryverylongprocessname"}}
	if err := Table(&buf, header, rows, 12); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, "averyveryverylongprocessname") {
		t.Errorf("long name not truncated:\n%s", got)
	}
	if !strings.Contains(got, "...") {
		t.Errorf("truncation marker missing:\n%s", got)
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("hello", 5); got != "hello" {
		t.Errorf("Truncate(hello,5) = %q", got)
	}
	if got := Truncate("hello world", 5); got != "he..." {
		t.Errorf("Truncate(hello world,5) = %q", got)
	}
	if got := Truncate("hello world", 8); got != "hello..." {
		t.Errorf("Truncate(hello world,8) = %q", got)
	}
	if got := Truncate("hello world", 2); got != "he" {
		t.Errorf("Truncate(hello world,2) = %q", got)
	}
	if got := Truncate("日本語テキスト", 6); got != "日本語..." {
		t.Errorf("Truncate multibyte = %q", got)
	}
}

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, map[string]int{"port": 3000}); err != nil {
		t.Fatal(err)
	}
	var v map[string]int
	if err := json.Unmarshal(buf.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v["port"] != 3000 {
		t.Errorf("port = %d, want 3000", v["port"])
	}
	if !strings.HasPrefix(buf.String(), "{") || !strings.HasSuffix(buf.String(), "}\n") {
		t.Errorf("unexpected JSON output: %q", buf.String())
	}
}

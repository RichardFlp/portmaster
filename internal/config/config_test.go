package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.RefreshInterval != 2 {
		t.Errorf("RefreshInterval = %d, want 2", cfg.RefreshInterval)
	}
	if cfg.WatchInterval != 2 {
		t.Errorf("WatchInterval = %d, want 2", cfg.WatchInterval)
	}
	if cfg.FreeCount != 5 {
		t.Errorf("FreeCount = %d, want 5", cfg.FreeCount)
	}
	if cfg.OutputFormat != "table" {
		t.Errorf("OutputFormat = %q, want table", cfg.OutputFormat)
	}
}

func TestLoadFromMissingFile(t *testing.T) {
	cfg, err := LoadFrom(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FreeCount != 5 {
		t.Errorf("FreeCount = %d, want 5", cfg.FreeCount)
	}
}

func TestLoadFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := `{"refresh_interval": 7, "ignored_processes": ["node"], "ignored_ports": [3000], "free_count": 3}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RefreshInterval != 7 {
		t.Errorf("RefreshInterval = %d, want 7", cfg.RefreshInterval)
	}
	if len(cfg.IgnoredProcesses) != 1 || cfg.IgnoredProcesses[0] != "node" {
		t.Errorf("IgnoredProcesses = %v", cfg.IgnoredProcesses)
	}
	if len(cfg.IgnoredPorts) != 1 || cfg.IgnoredPorts[0] != 3000 {
		t.Errorf("IgnoredPorts = %v", cfg.IgnoredPorts)
	}
	if cfg.FreeCount != 3 {
		t.Errorf("FreeCount = %d, want 3", cfg.FreeCount)
	}
}

func TestLoadFromInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadFrom(path); err == nil {
		t.Error("LoadFrom of invalid JSON succeeded, want error")
	}
}

func TestNormalize(t *testing.T) {
	cfg := &Config{RefreshInterval: -1, WatchInterval: 0, FreeCount: -3, OutputFormat: "xml"}
	cfg.Normalize()
	if cfg.RefreshInterval != 2 || cfg.WatchInterval != 2 || cfg.FreeCount != 5 || cfg.OutputFormat != "table" {
		t.Errorf("normalize produced %+v", cfg)
	}
}

func TestPathUnderConfigDir(t *testing.T) {
	dir, err := os.UserConfigDir()
	if err != nil {
		t.Skip(err)
	}
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "portmaster", "config.json")
	if path != want {
		t.Errorf("Path = %q, want %q", path, want)
	}
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	RefreshInterval  int      `json:"refresh_interval"`
	WatchInterval    int      `json:"watch_interval"`
	OutputFormat     string   `json:"output_format"`
	IgnoredProcesses []string `json:"ignored_processes"`
	IgnoredPorts     []int    `json:"ignored_ports"`
	Browser          string   `json:"browser"`
	FreeCount        int      `json:"free_count"`
}

func Defaults() *Config {
	return &Config{
		RefreshInterval: 2,
		WatchInterval:   2,
		OutputFormat:    "table",
		FreeCount:       5,
	}
}

func (c *Config) Normalize() {
	if c.RefreshInterval <= 0 {
		c.RefreshInterval = 2
	}
	if c.WatchInterval <= 0 {
		c.WatchInterval = 2
	}
	if c.FreeCount <= 0 {
		c.FreeCount = 5
	}
	if c.OutputFormat != "table" && c.OutputFormat != "json" {
		c.OutputFormat = "table"
	}
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return Defaults(), err
	}
	return LoadFrom(path)
}

func LoadFrom(path string) (*Config, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "portmaster", "config.json"), nil
}

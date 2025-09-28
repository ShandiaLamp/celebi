package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Connection struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

type Config struct {
	Default     string                `json:"default"`
	Connections map[string]Connection `json:"connections"`
}

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".celebi", "config.json")
}

// 读取配置
func Load() (*Config, error) {
	path := getConfigPath()
	file, err := os.ReadFile(path)
	if err != nil {
		return &Config{Connections: make(map[string]Connection)}, nil
	}
	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// 保存配置
func Save(cfg *Config) error {
	dir := filepath.Dir(getConfigPath())
	os.MkdirAll(dir, 0700)
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(getConfigPath(), data, 0600)
}

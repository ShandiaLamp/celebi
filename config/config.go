package config

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
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

func GetConnection(name string) (*Connection, error) {
	cfg, err := Load()
	if err != nil {
		return nil, fmt.Errorf("无法加载配置: %w", err)
	}
	if name == "" {
		if cfg.Default == "" {
			return nil, fmt.Errorf("既没有指定连接也没有配置默认连接")
		} else {
			name = cfg.Default
		}
	}
	conn, ok := cfg.Connections[name]
	if !ok {
		return nil, fmt.Errorf("连接'%s'不存在", name)
	}
	return &conn, nil
}

func TestConnection(driver, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer db.Close()

	// Ping 测试
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}
	return nil
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 表示持久化的应用配置（当前仅包含 DeepSeek API Key）
// 可根据需要在此结构体中添加更多字段。
//
// 保存路径：$HOME/.tomato_clock_config.json
// Windows 上 HOME 环境变量通常对应用户目录。
//
// 若文件不存在，Load 将返回 (nil, os.ErrNotExist)。
// Save 会在同一路径创建/覆盖文件。

type Config struct {
	APIKey string `json:"api_key"`
}

// configPath 返回配置文件完整路径。
func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tomato_clock_config.json"), nil
}

// Load 从默认路径加载配置。
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save 将给定 APIKey 写入配置文件（若文件不存在则创建）。
func Save(apiKey string) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	cfg := Config{APIKey: apiKey}

	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	if err := json.NewEncoder(f).Encode(&cfg); err != nil {
		f.Close()
		return err
	}
	f.Close()
	// 原子替换
	return os.Rename(tmpPath, path)
}

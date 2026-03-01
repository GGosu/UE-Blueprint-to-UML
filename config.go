package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Host      string
	Port      int
	MaxBodyKB int64
	Version   string
}

func LoadConfig(path string) (Config, error) {
	cfg := Config{}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	flat := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, " #"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.Trim(strings.TrimSpace(line[idx+1:]), `"'`)
		if key != "" && val != "" {
			flat[key] = val
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if v, ok := flat["host"]; ok {
		cfg.Host = v
	}
	if v, ok := flat["port"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Port = n
		}
	}
	if v, ok := flat["max_body_kb"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.MaxBodyKB = n
		}
	}
	if v, ok := flat["version"]; ok && v != "" {
		cfg.Version = v
	}

	return cfg, nil
}

func (c Config) Addr() string {
	host := c.Host
	if host == "localhost" {
		host = "127.0.0.1"
	}
	return fmt.Sprintf("%s:%d", host, c.Port)
}

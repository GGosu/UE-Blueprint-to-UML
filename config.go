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
	cfg := Config{
		Port:      8080,
		MaxBodyKB: 25600,
	}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	flat := map[string]string{}
	var currentSection string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Basic section detection: if line starts with no indent and ends with ":"
		if !strings.HasPrefix(rawLine, " ") && !strings.HasPrefix(rawLine, "\t") && strings.HasSuffix(trimmed, ":") {
			currentSection = strings.TrimSuffix(trimmed, ":")
			continue
		}

		line := trimmed
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
			if currentSection != "" {
				flat[currentSection+"."+key] = val
			} else {
				flat[key] = val
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if v, ok := flat["server.host"]; ok {
		cfg.Host = v
	}
	if v, ok := flat["server.port"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.Port = n
		}
	}
	if v, ok := flat["limits.max_body_kb"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.MaxBodyKB = n
		}
	}
	if v, ok := flat["app.version"]; ok && v != "" {
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

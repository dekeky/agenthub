package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	DefaultServerFilename = "agenthub-server.toml"
	DefaultCLIFilename    = "agenthub-cli.toml"
)

type Server struct {
	Addr        string `toml:"addr"`
	StorageDir  string `toml:"storage_dir"`
	UploadToken string `toml:"upload_token"`
}

type CLI struct {
	URL         string `toml:"url"`
	UploadToken string `toml:"upload_token"`
}

// LoadServer reads server settings from agenthub-server.toml.
func LoadServer(configPath string) (Server, error) {
	path, err := resolvePath(configPath, DefaultServerFilename, "AGENTHUB_SERVER_CONFIG")
	if err != nil {
		return Server{}, err
	}
	var s Server
	if err := readTOML(path, &s); err != nil {
		return Server{}, err
	}
	return applyServerDefaults(s, filepath.Dir(path)), nil
}

// LoadCLI reads CLI settings from agenthub-cli.toml.
func LoadCLI(configPath string) (CLI, error) {
	path, err := resolvePath(configPath, DefaultCLIFilename, "AGENTHUB_CLI_CONFIG")
	if err != nil {
		return CLI{}, err
	}
	var c CLI
	if err := readTOML(path, &c); err != nil {
		return CLI{}, err
	}
	return applyCLIDefaults(c), nil
}

func readTOML(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if err := toml.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	return nil
}

func resolvePath(configPath, defaultName, envKey string) (string, error) {
	if p := strings.TrimSpace(configPath); p != "" {
		return filepath.Clean(p), nil
	}
	if p := strings.TrimSpace(os.Getenv(envKey)); p != "" {
		return filepath.Clean(p), nil
	}
	if _, err := os.Stat(defaultName); err == nil {
		abs, err := filepath.Abs(defaultName)
		if err != nil {
			return defaultName, nil
		}
		return abs, nil
	}
	wd, _ := os.Getwd()
	if wd == "" {
		wd = "."
	}
	example := strings.TrimSuffix(defaultName, ".toml") + ".example.toml"
	return filepath.Join(wd, defaultName), fmt.Errorf("config file not found (looked for %s; copy %s to %s)", defaultName, example, defaultName)
}

func applyServerDefaults(s Server, configDir string) Server {
	if strings.TrimSpace(s.Addr) == "" {
		s.Addr = ":9093"
	}
	if strings.TrimSpace(s.StorageDir) == "" {
		s.StorageDir = filepath.Join(configDir, "storage")
	} else if !filepath.IsAbs(s.StorageDir) {
		s.StorageDir = filepath.Join(configDir, s.StorageDir)
	}
	s.Addr = strings.TrimSpace(s.Addr)
	s.StorageDir = filepath.Clean(s.StorageDir)
	s.UploadToken = strings.TrimSpace(s.UploadToken)
	return s
}

func applyCLIDefaults(c CLI) CLI {
	if strings.TrimSpace(c.URL) == "" {
		c.URL = "http://localhost:8080"
	}
	c.URL = strings.TrimRight(strings.TrimSpace(c.URL), "/")
	c.UploadToken = strings.TrimSpace(c.UploadToken)
	return c
}

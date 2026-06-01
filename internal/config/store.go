package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Current  string             `yaml:"current"`
	Profiles map[string]Profile `yaml:"profiles"`
}

type Profile struct {
	Domain         string            `yaml:"domain"`
	Organization   string            `yaml:"organization"`
	Region         string            `yaml:"region"`
	Output         string            `yaml:"output"`
	Safety         Safety            `yaml:"safety"`
	RepoProjectMap map[string]string `yaml:"repoProjectMap"`
}

type Safety struct {
	ConfirmWrites bool `yaml:"confirmWrites"`
}

type Store struct {
	path         string
	beforeRename func() error
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() (Config, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return emptyConfig(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return emptyConfig(), nil
	}

	cfg := emptyConfig()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	normalize(&cfg)
	return cfg, nil
}

func (s *Store) Save(cfg Config) error {
	normalize(&cfg)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp config: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp config: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp config: %w", err)
	}

	if s.beforeRename != nil {
		if err := s.beforeRename(); err != nil {
			return err
		}
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("publish config: %w", err)
	}
	_ = syncDir(dir)
	return nil
}

func emptyConfig() Config {
	return Config{Profiles: map[string]Profile{}}
}

func normalize(cfg *Config) {
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	for name, profile := range cfg.Profiles {
		if profile.RepoProjectMap == nil {
			profile.RepoProjectMap = map[string]string{}
		}
		cfg.Profiles[name] = profile
	}
}

func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

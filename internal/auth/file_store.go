package auth

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type FileTokenStore struct {
	path         string
	beforeRename func() error
}

func NewFileTokenStore(path string) *FileTokenStore {
	return &FileTokenStore{path: path}
}

func (s *FileTokenStore) Backend() string {
	return "file"
}

func (s *FileTokenStore) Save(profile, token string) error {
	tokens, err := s.loadAll()
	if err != nil {
		return err
	}
	tokens[profile] = token
	return s.saveAll(tokens)
}

func (s *FileTokenStore) Load(profile string) (string, bool, error) {
	tokens, err := s.loadAll()
	if err != nil {
		return "", false, err
	}
	token, ok := tokens[profile]
	return token, ok, nil
}

func (s *FileTokenStore) Delete(profile string) error {
	tokens, err := s.loadAll()
	if err != nil {
		return err
	}
	delete(tokens, profile)
	return s.saveAll(tokens)
}

func (s *FileTokenStore) loadAll() (map[string]string, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read token store: %w", err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return map[string]string{}, nil
	}
	tokens := map[string]string{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&tokens); err != nil {
		return nil, fmt.Errorf("parse token store: %w", err)
	}
	return tokens, nil
}

func (s *FileTokenStore) saveAll(tokens map[string]string) error {
	data, err := yaml.Marshal(tokens)
	if err != nil {
		return fmt.Errorf("marshal token store: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create token store dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tokens-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp token store: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp token store: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp token store: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp token store: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp token store: %w", err)
	}

	if s.beforeRename != nil {
		if err := s.beforeRename(); err != nil {
			return err
		}
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		return fmt.Errorf("publish token store: %w", err)
	}
	_ = syncDir(dir)
	return nil
}

func syncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

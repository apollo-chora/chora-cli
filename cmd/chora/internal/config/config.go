// Package config handles CLI credential storage and configuration.
// Credentials are stored in a user-scoped file with 0600 permissions.
// API keys are NEVER logged or displayed in full.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const credentialFile = "credentials.json"

// ErrNotAuthenticated indicates no stored credentials were found.
var ErrNotAuthenticated = errors.New("not authenticated — run chora auth login")

// Credentials holds the stored API key and gateway URL.
type Credentials struct {
	APIKey     string `json:"api_key"`
	GatewayURL string `json:"gateway_url"`
}

// String returns a safe representation that masks the API key.
func (c *Credentials) String() string {
	masked := maskAPIKey(c.APIKey)
	return fmt.Sprintf("gateway=%s key=%s", c.GatewayURL, masked)
}

// maskAPIKey shows only the prefix and masks the rest.
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	prefix := key[:8]
	return prefix + "****"
}

// FileStore persists credentials to a file on disk.
type FileStore struct {
	dir string
}

// NewFileStore creates a FileStore at the given directory.
func NewFileStore(dir string) *FileStore {
	return &FileStore{dir: dir}
}

// DefaultDir returns the default config directory (~/.config/chora).
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".chora"
	}
	return filepath.Join(home, ".config", "chora")
}

// Save writes credentials to disk with 0600 permissions.
func (s *FileStore) Save(cred *Credentials) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}

	path := filepath.Join(s.dir, credentialFile)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return nil
}

// Load reads credentials from disk.
func (s *FileStore) Load() (*Credentials, error) {
	path := filepath.Join(s.dir, credentialFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotAuthenticated
		}
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	var cred Credentials
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	if cred.APIKey == "" {
		return nil, ErrNotAuthenticated
	}

	return &cred, nil
}

// Clear removes stored credentials.
func (s *FileStore) Clear() error {
	path := filepath.Join(s.dir, credentialFile)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials: %w", err)
	}
	return nil
}

// SanitizeForLog removes or masks sensitive data from a string.
func SanitizeForLog(s string) string {
	// Mask any sk_live_ or sk_test_ prefixed keys.
	for _, prefix := range []string{"sk_live_", "sk_test_"} {
		if idx := strings.Index(s, prefix); idx >= 0 {
			end := idx + len(prefix)
			for end < len(s) && s[end] != ' ' && s[end] != '"' && s[end] != '\'' {
				end++
			}
			s = s[:idx] + maskAPIKey(s[idx:end]) + s[end:]
		}
	}
	return s
}

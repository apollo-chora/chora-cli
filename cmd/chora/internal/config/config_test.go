package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadCredentials(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	cred := &Credentials{
		APIKey:     "sk_live_test123",
		GatewayURL: "https://gateway.chora.example.com",
	}

	err := store.Save(cred)
	require.NoError(t, err)

	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, cred.APIKey, loaded.APIKey)
	assert.Equal(t, cred.GatewayURL, loaded.GatewayURL)
}

func TestLoadCredentials_NotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	_, err := store.Load()
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestClearCredentials(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	cred := &Credentials{
		APIKey:     "sk_live_test123",
		GatewayURL: "https://gateway.chora.example.com",
	}
	require.NoError(t, store.Save(cred))

	err := store.Clear()
	require.NoError(t, err)

	_, err = store.Load()
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestFilePermissions(t *testing.T) {
	dir := t.TempDir()
	store := NewFileStore(dir)

	cred := &Credentials{APIKey: "sk_live_secret", GatewayURL: "https://gw.example.com"}
	require.NoError(t, store.Save(cred))

	info, err := os.Stat(filepath.Join(dir, credentialFile))
	require.NoError(t, err)
	// File should be owner-only readable (0600).
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestAPIKeyNeverInString(t *testing.T) {
	cred := &Credentials{APIKey: "sk_live_supersecret", GatewayURL: "https://gw.example.com"}
	s := cred.String()
	assert.NotContains(t, s, "sk_live_supersecret")
	assert.Contains(t, s, "sk_live_****")
}

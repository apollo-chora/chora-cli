package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SetsAPIKeyHeader(t *testing.T) {
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-API-Key")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer srv.Close()

	c := New(srv.URL, "sk_live_test123")
	_, err := c.Get(context.Background(), "/api/v1/health")
	require.NoError(t, err)
	assert.Equal(t, "sk_live_test123", gotHeader)
}

func TestClient_GetAtoms(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/atoms", r.URL.Path)
		assert.Equal(t, "Mathematics", r.URL.Query().Get("topic"))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]string{
				{"id": "atom-1", "title": "Algebra Basics"},
				{"id": "atom-2", "title": "Calculus Intro"},
			},
			"next_cursor": "cursor_abc",
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "sk_live_test")
	resp, err := c.Get(context.Background(), "/api/v1/atoms?topic=Mathematics")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "IAM_UNAUTHENTICATED",
				"message": "invalid or expired API key",
			},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "sk_live_invalid")
	resp, err := c.Get(context.Background(), "/api/v1/atoms")
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestClient_PostAtomCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/atoms", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.Header().Set("Location", "/api/v1/atoms/new-atom-id")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "new-atom-id"})
	}))
	defer srv.Close()

	c := New(srv.URL, "sk_live_test")
	body := []byte(`{"title":"Test Atom","atom_type":"multiple_choice"}`)
	resp, err := c.Post(context.Background(), "/api/v1/atoms", body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "/api/v1/atoms/new-atom-id", resp.Header.Get("Location"))
}

func TestClient_TenantInfo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/tenants/current", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "tenant-123",
			"name":    "Acme School",
			"add_ons": []string{"familiar", "analytics", "singapore_gov"},
		})
	}))
	defer srv.Close()

	c := New(srv.URL, "sk_live_test")
	resp, err := c.Get(context.Background(), "/api/v1/tenants/current")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

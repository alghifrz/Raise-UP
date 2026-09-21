package storage_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/diuk/raiseup/internal/storage"
)

func TestUploadUsesServiceCredentialsAndReturnsPublicURL(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.Header.Get("apikey") != "secret" ||
			r.Header.Get("Authorization") != "Bearer secret" {
			t.Error("missing service credentials")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "image" {
			t.Errorf("body = %q", body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := storage.NewClientWithHTTP(server.URL, "secret", server.Client())
	publicURL, err := client.Upload(
		context.Background(),
		"gallery",
		"2026/09/a photo.webp",
		"image/webp",
		strings.NewReader("image"),
	)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	if gotPath != "/storage/v1/object/gallery/2026/09/a%20photo.webp" {
		t.Fatalf("path = %q", gotPath)
	}
	want := server.URL + "/storage/v1/object/public/gallery/2026/09/a%20photo.webp"
	if publicURL != want {
		t.Fatalf("public URL = %q, want %q", publicURL, want)
	}
}

func TestDeleteSendsObjectPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"prefixes":["2026/09/photo.jpg"]}` {
			t.Errorf("body = %s", body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := storage.NewClientWithHTTP(server.URL, "secret", server.Client())
	if err := client.Delete(context.Background(), "gallery", "2026/09/photo.jpg"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestSecretKeyIsNotSentAsBearerJWT(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("apikey") != "sb_secret_test" {
			t.Error("missing secret API key")
		}
		if r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected Authorization header: %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := storage.NewClientWithHTTP(server.URL, "sb_secret_test", server.Client())
	if err := client.Delete(context.Background(), "gallery", "photo.jpg"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestStorageErrorIncludesStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bucket missing", http.StatusNotFound)
	}))
	defer server.Close()

	client := storage.NewClientWithHTTP(server.URL, "secret", server.Client())
	_, err := client.Upload(
		context.Background(), "gallery", "photo.jpg", "image/jpeg", strings.NewReader("x"),
	)
	if err == nil || !strings.Contains(err.Error(), "status 404") {
		t.Fatalf("unexpected error: %v", err)
	}
}

package mcpclient

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/idtoken"
)

// NewAuthClient creates an HTTP client that attaches a Google ID token
// to every request for Cloud Run IAM auth. The audience should be the
// Cloud Run service URL. If audience is empty, returns a plain
// http.Client (dev mode).
func NewAuthClient(audience string) (*http.Client, error) {
	if audience == "" {
		return http.DefaultClient, nil
	}

	// Try the idtoken library first (works with service account credentials
	// and GCE metadata). Falls back to gcloud CLI for authorized_user
	// credentials (i.e. `gcloud auth application-default login`).
	client, err := idtoken.NewClient(context.Background(), audience)
	if err == nil {
		return client, nil
	}
	log.Printf("idtoken.NewClient failed (%v), falling back to gcloud CLI", err)

	return &http.Client{
		Transport: &gcloudIDTokenTransport{audience: audience},
	}, nil
}

// gcloudIDTokenTransport attaches a Google ID token obtained via the
// gcloud CLI to each outgoing request. Tokens are cached until near
// expiry (gcloud tokens are valid for ~1h).
type gcloudIDTokenTransport struct {
	audience string

	mu    sync.Mutex
	token string
	expAt time.Time
}

func (t *gcloudIDTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	tok, err := t.getToken()
	if err != nil {
		return nil, fmt.Errorf("gcloud id-token: %w", err)
	}
	req2 := req.Clone(req.Context())
	req2.Header.Set("Authorization", "Bearer "+tok)
	return http.DefaultTransport.RoundTrip(req2)
}

// gcloudPath returns the path to the gcloud binary, searching PATH and then
// known install locations (e.g. ~/google-cloud-sdk/bin) that may not be
// present when Claude Code spawns plugin processes with a minimal environment.
func gcloudPath() (string, error) {
	if p, err := exec.LookPath("gcloud"); err == nil {
		return p, nil
	}
	candidates := []string{
		filepath.Join(os.Getenv("HOME"), "google-cloud-sdk", "bin", "gcloud"),
		"/usr/lib/google-cloud-sdk/bin/gcloud",
		"/snap/bin/gcloud",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("gcloud not found in PATH or known install locations")
}

func (t *gcloudIDTokenTransport) getToken() (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.token != "" && time.Now().Before(t.expAt) {
		return t.token, nil
	}

	gcloud, err := gcloudPath()
	if err != nil {
		return "", err
	}

	// Try with --audiences first (works with service accounts).
	// Fall back without it (required for authorized_user credentials).
	cmd := exec.Command(gcloud, "auth", "print-identity-token",
		"--audiences="+t.audience)
	out, err := cmd.Output()
	if err != nil {
		cmd = exec.Command(gcloud, "auth", "print-identity-token")
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("gcloud auth print-identity-token failed: %w", err)
		}
	}
	t.token = strings.TrimSpace(string(out))
	t.expAt = time.Now().Add(55 * time.Minute) // tokens valid ~1h, refresh at 55m
	return t.token, nil
}

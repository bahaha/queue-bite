package main

import (
	"net/http"
	"testing"

	"queue-bite/internal/testutils"
)

func TestServerConfStartup(t *testing.T) {
	t.Parallel()

	ts := server.NewTestServer(t, map[string]string{})
	ts.Run(t, run)

	// Test
	resp, err := http.Get(ts.URL("/assets/css/output.css"))
	if err != nil {
		t.Fatalf("failed to get static assets: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 for static asset, got %d", resp.StatusCode)
	}
}

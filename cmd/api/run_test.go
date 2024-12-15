package main

import (
	"net/http"
	"testing"

	"queue-bite/internal/testutils"
	"queue-bite/pkg/utils"
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

func TestServerWithRedis(t *testing.T) {
	t.Parallel()

	ts := server.NewTestServer(t, map[string]string{}, server.WithRedisImage("WAITLIST_REDIS"))
	ts.Run(t, run)

	health, err := http.Get(ts.URL("/healthz"))
	if err != nil {
		t.Fatalf("failed to check health of server components: %v", err)
	}
	defer health.Body.Close()
	if health.StatusCode != http.StatusOK {
		t.Fatalf("`/healthz` endpoint expected status 200 for static asset, got %d", health.StatusCode)
	}

	stats, err := utils.DecodeBody[map[string]map[string]string](health)
	if err != nil {
		t.Errorf("could not parse stats of server components: %v", err)
	}

	if redis, ok := stats["redis"]; ok {
		if status, ok := redis["status"]; ok {
			if status != "up" {
				t.Errorf("expected redis status `up` but instead %s", status)
			}
		} else {
			t.Errorf("missing `status` metrics of redis status payload")
		}
	} else {
		t.Errorf("missing health stats of redis")
	}
}

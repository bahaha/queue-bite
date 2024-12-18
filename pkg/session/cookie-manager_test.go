package session

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCookieEncryptDecrypt(t *testing.T) {
	t.Parallel()

	m, err := NewCookieManager("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("failed to create cookie manager with encrypt key: %v", err)
	}

	type Payload struct {
		UserID string
		No     int
	}

	payload := &Payload{
		UserID: "u5566",
		No:     56,
	}
	cfg := NewCookieConfig("test_cookie", "example.com")
	w := httptest.NewRecorder()

	if err := m.SetCookie(w, *cfg, payload); err != nil {
		t.Fatalf("failed to set cookie: %v", err)
	}

	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.AddCookie(cookies[0])

	var decoded Payload
	if err := m.GetCookie(req, cfg, &decoded); err != nil {
		t.Fatalf("failed to get cookie: %v", err)
	}

	assert.Equal(t, payload.UserID, decoded.UserID)
	assert.Equal(t, payload.No, decoded.No)
}

func TestCookieExpiration(t *testing.T) {
	t.Parallel()

	t.Run("default ttl is 30min", func(t *testing.T) {
		cfg := NewCookieConfig("test_expiration", "example.com")
		cfg.Fixed = time.Now()
		expectedTTL := 30 * time.Minute
		assert.Equal(t, expectedTTL, cfg.GetMaxAge())
		assert.WithinDuration(t, cfg.Fixed.Add(expectedTTL), cfg.GetExpiration(), time.Second)
	})

	t.Run("custom ttl", func(t *testing.T) {
		cfg := NewCookieConfig("test_expiration", "example.com").WithTTL(5 * time.Minute)
		cfg.Fixed = time.Now()
		expectedTTL := 5 * time.Minute
		assert.Equal(t, expectedTTL, cfg.GetMaxAge())
		assert.WithinDuration(t, cfg.Fixed.Add(expectedTTL), cfg.GetExpiration(), time.Second)
	})

	t.Run("custom expiration", func(t *testing.T) {
		now := time.Now()
		cfg := NewCookieConfig("test_expiration", "example.com").
			WithExpiration(func() time.Time { return now.Add(10 * time.Minute) })
		cfg.Fixed = now

		assert.InDelta(t, 10*time.Minute.Milliseconds(), cfg.GetMaxAge().Milliseconds(), 1)
		assert.WithinDuration(t, now.Add(10*time.Minute), cfg.GetExpiration(), time.Millisecond)
	})
}

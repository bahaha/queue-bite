package utils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAcceptLanguageParser(t *testing.T) {
	t.Parallel()

	t.Run("empty header", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest("GET", "/", nil)

		langs := CollectAcceptLanguages(req)
		assert.Equal(t, []string{"*"}, langs)
	})

	t.Run("single language", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "en-US")

		langs := CollectAcceptLanguages(req)
		assert.Equal(t, []string{"en-US"}, langs)

	})

	t.Run("multiple language without q", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "en-US,ja,zh_TW")

		langs := CollectAcceptLanguages(req)
		assert.Equal(t, []string{"en-US", "ja", "zh_TW"}, langs)

	})

	t.Run("multiple language with q", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "en-US,ja;q=0.8,zh_TW;q=0.5")

		langs := CollectAcceptLanguages(req)
		assert.Equal(t, []string{"en-US", "ja", "zh_TW"}, langs)

	})

	t.Run("unordered q value", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "en-US,ja;q=0.5,zh_TW;q=0.8")

		langs := CollectAcceptLanguages(req)
		assert.Equal(t, []string{"en-US", "zh_TW", "ja"}, langs)

	})

	t.Run("handle whitespaces", func(t *testing.T) {
		t.Parallel()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Language", "en-US, ja; q=0.8,zh_TW;  q=0.5")

		langs := CollectAcceptLanguages(req)
		assert.Equal(t, []string{"en-US", "ja", "zh_TW"}, langs)

	})
}

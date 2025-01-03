package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func Encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("json encoding error: %w", err)
	}
	return nil
}

func Decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("json decoding error: %w", err)
	}
	return v, nil
}

func DecodeBody[T any](r *http.Response) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("json decoding error: %w", err)
	}
	return v, nil
}

// CollectAcceptLanguages parses the Accept-Language header from a HTTP request
// and returns a slice of language tags sorted by their quality values.
//
// The function processes the Accept-Language header according to RFC 2616,
// handling language tags and their q-values.
//
// It returns language tags ordered from highest to lowest preference.
//
// Example:
//
//	// Accept-Language: en-US,ja;q=0.9,zh_TW;q=0.5
//	langs := CollectAcceptLanguages(r)
//	// langs = []string{"en-US", "ja", "zh_TW"}
func CollectAcceptLanguages(r *http.Request) []string {
	acceptLanguage := r.Header.Get("Accept-Language")
	if acceptLanguage == "" {
		return []string{"*"}
	}

	parts := strings.Split(acceptLanguage, ",")

	type locale struct {
		lang string
		q    float64
	}

	langQs := make([]locale, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		langParts := strings.Split(part, ";")
		lang := strings.TrimSpace(langParts[0])
		q := 1.0

		if len(langParts) > 1 {
			qPart := strings.TrimSpace(langParts[1])
			if strings.HasPrefix(qPart, "q=") {
				if val, err := strconv.ParseFloat(qPart[2:], 64); err == nil {
					q = val
				}
			}
		}

		langQs = append(langQs, locale{lang, q})
	}

	sort.Slice(langQs, func(i, j int) bool {
		return langQs[i].q > langQs[j].q
	})

	result := make([]string, 0, len(langQs))
	for _, lq := range langQs {
		result = append(result, lq.lang)
	}

	return result
}

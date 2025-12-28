package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

var xssPolicy = bluemonday.UGCPolicy()

func XSSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Only inspect JSON bodies
		if r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			next.ServeHTTP(w, r)
			return
		}

		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || mediaType != "application/json" {
			next.ServeHTTP(w, r)
			return
		}

		// Read body with safety
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if len(bytes.TrimSpace(bodyBytes)) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			next.ServeHTTP(w, r)
			return
		}

		var payload interface{}
		if err := json.Unmarshal(bodyBytes, &payload); err != nil {
			http.Error(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}

		sanitized := sanitizeJSON(payload)

		safeBody, err := json.Marshal(sanitized)
		if err != nil {
			http.Error(w, "Failed to sanitize body", http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(safeBody))
		r.ContentLength = int64(len(safeBody))

		next.ServeHTTP(w, r)
	})
}

func sanitizeJSON(data interface{}) interface{} {
	switch v := data.(type) {

	case map[string]interface{}:
		for key, val := range v {
			v[key] = sanitizeJSON(val)
		}
		return v

	case []interface{}:
		for i, val := range v {
			v[i] = sanitizeJSON(val)
		}
		return v

	case string:
		return sanitizeString(v)

	default:
		// numbers, booleans, null stay untouched
		return v
	}
}

func sanitizeString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return xssPolicy.Sanitize(s)
}

package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyHandler_Success(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/import/prometheus", r.URL.Path)
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		assert.Equal(t, "metric_name 42", string(body))

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte("ok"))
		assert.NoError(t, err)
	}))
	defer backend.Close()

	handler := NewProxyHandler(backend.Client(), backend.URL, "")

	req := httptest.NewRequest(http.MethodPost, "/api/metrics", strings.NewReader("metric_name 42"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "ok", rec.Body.String())
}

func TestProxyHandler_PathSuffix(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/import/prometheus/extra/path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := NewProxyHandler(backend.Client(), backend.URL, "")

	req := httptest.NewRequest(http.MethodPost, "/api/metrics/extra/path", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestProxyHandler_AuthRequired(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := NewProxyHandler(backend.Client(), backend.URL, "secret-key")

	req := httptest.NewRequest(http.MethodPost, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestProxyHandler_AuthSuccess(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := NewProxyHandler(backend.Client(), backend.URL, "secret-key")

	req := httptest.NewRequest(http.MethodPost, "/api/metrics", nil)
	req.Header.Set("Authorization", "secret-key")
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestProxyHandler_MethodNotAllowed(t *testing.T) {
	handler := NewProxyHandler(http.DefaultClient, "http://localhost", "")

	for _, method := range []string{http.MethodGet, http.MethodDelete, http.MethodPatch} {
		req := httptest.NewRequest(method, "/api/metrics", nil)
		rec := httptest.NewRecorder()
		handler(rec, req)

		require.Equal(t, http.StatusMethodNotAllowed, rec.Code, "method %s should not be allowed", method)
	}
}

func TestProxyHandler_PutAllowed(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	handler := NewProxyHandler(backend.Client(), backend.URL, "")

	req := httptest.NewRequest(http.MethodPut, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestProxyHandler_BackendError(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("backend error"))
		assert.NoError(t, err)
	}))
	defer backend.Close()

	handler := NewProxyHandler(backend.Client(), backend.URL, "")

	req := httptest.NewRequest(http.MethodPost, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.Equal(t, "backend error", rec.Body.String())
}

func TestProxyHandler_BackendUnreachable(t *testing.T) {
	handler := NewProxyHandler(http.DefaultClient, "http://localhost:99999", "")

	req := httptest.NewRequest(http.MethodPost, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	require.Equal(t, http.StatusBadGateway, rec.Code)
}

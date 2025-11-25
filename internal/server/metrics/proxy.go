package metrics

import (
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func NewProxyHandler(victoriaURL, authKey string) http.HandlerFunc {
	client := &http.Client{}

	return func(w http.ResponseWriter, r *http.Request) {
		rl := zap.L().With(
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("path", r.URL.Path),
		)
		rl.Info("Received metrics push request")

		if authKey != "" && r.Header.Get("Authorization") != authKey {
			rl.Warn("Unauthorized metrics push attempt")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			rl.Warn("Invalid method")
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		pathSuffix := strings.TrimPrefix(r.URL.Path, "/api/metrics")
		targetURL := victoriaURL + "/api/v1/import/prometheus" + pathSuffix

		proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			rl.Error("Failed to create proxy request", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		proxyReq.Header.Set("Content-Type", r.Header.Get("Content-Type"))
		proxyReq.Header.Set("Content-Encoding", r.Header.Get("Content-Encoding"))

		resp, err := client.Do(proxyReq)
		if err != nil {
			rl.Error("Failed to proxy metrics to VictoriaMetrics", zap.Error(err))
			http.Error(w, "Failed to forward metrics", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			rl.Error("Failed to copy response from VictoriaMetrics", zap.Error(err))
		}
	}
}

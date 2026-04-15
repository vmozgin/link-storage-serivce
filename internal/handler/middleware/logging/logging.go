package logging

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}

	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		rw := &responseWriter{ResponseWriter: w}

		next.ServeHTTP(rw, r)

		slog.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"request_body", string(bodyBytes),
			"status", rw.status,
			"duration", time.Since(start),
		)
	})
}

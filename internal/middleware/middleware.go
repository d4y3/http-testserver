package middleware

import (
	"net/http"
	"time"

	"github.com/d4y3/http-testserver/internal/logger"
)

type responseWriterWrapper struct {
	http.ResponseWriter

	status int
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func LoggingMiddleware(l *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.LogRequest(r)

		rw := &responseWriterWrapper{ResponseWriter: w}

		start := time.Now()

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		resp := &http.Response{
			Status:     http.StatusText(rw.status),
			StatusCode: rw.status,
			Header:     w.Header(),
		}

		l.LogResponse(resp, duration)
	})
}

package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

func InitializeLogger(level string) error {
	var slogLevel slog.Level
	err := slogLevel.UnmarshalText([]byte(strings.ToUpper(level)))
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, opts)
	jsonLogger := slog.New(jsonHandler)

	slog.SetDefault(jsonLogger)
	return nil
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		slog.Info(
			"request",
			"uri", uri,
			"method", method,
			"status", responseData.status,
			"duration", duration.String(),
			"size", responseData.size,
		)

	}
	return http.HandlerFunc(logFn)
}

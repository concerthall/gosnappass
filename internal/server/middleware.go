package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/exp/slog"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type contextKey string

var requestIDContextKey contextKey = "requestID"

// addRequestIDMW adds a request UUID to the context.
func addRequestIDMW(next http.Handler) http.Handler {
	requestID := uuid.New()
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		},
	)
}

// newInjectLoggerMW injects logger into the request context.
func newInjectLoggerMW(logger *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			new := slog.NewContext(r.Context(), logger)
			r = r.WithContext(new)
			next.ServeHTTP(w, r)
		})
	}
}

// logRequestMW logs requests with the structured logger found in r.Context().
func logRequestMW(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger := slog.FromContext(r.Context())
			sr := newStatusRecorder(w)
			t := time.Now()
			next.ServeHTTP(sr, r)
			logger.Info("request served",
				"duration_microseconds", time.Since(t).Microseconds(),
				"method", r.Method,
				"path", r.URL.String(),
				"status", sr.status,
				"requestID", r.Context().Value(requestIDContextKey),
			)
		},
	)
}

// newDatabasePingMW generates a middleware that will ping the database and fail if it's not healthy
// or otherwise return next.
func newDatabasePingMW(pingDB func() error) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		if err := pingDB(); err != nil {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "could not reach database")
			})
		}

		return next
	}
}

// ensure statusRecorder implements http.ResponseWriter.
var _ http.ResponseWriter = &statusRecorder{}

// newStatusRecorder returns a statusRecorder with w and the status preset
// to 200 (OK), which aligns with the default behavior of http Handlers that
// do not explicitly write an HTTP header.
func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		status:     200,
		respWriter: w,
	}
}

// statusRecorder is a ResponseWriter that stores the WriteHeader input value
// before passing it to w.
type statusRecorder struct {
	status     int // same as http.statusCode
	respWriter http.ResponseWriter
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.respWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	return r.respWriter.Write(data)
}

func (r *statusRecorder) Header() http.Header {
	return r.respWriter.Header()
}

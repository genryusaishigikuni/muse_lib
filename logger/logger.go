package logger

import (
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}

func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func New(log *slog.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		log := log.With(
			slog.String("component", "middleware/logger"),
		)

		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := log.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("request_id", r.Header.Get("X-Request-ID")),
			)

			ww := &responseWriterWrapper{ResponseWriter: w}
			t1 := time.Now()

			defer func() {
				entry.Info("request completed",
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}

// responseWriterWrapper is used to capture the response status and number of bytes written
type responseWriterWrapper struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

func (r *responseWriterWrapper) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseWriterWrapper) Write(b []byte) (int, error) {
	bytes, err := r.ResponseWriter.Write(b)
	r.bytesWritten += bytes
	return bytes, err
}

func (r *responseWriterWrapper) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *responseWriterWrapper) BytesWritten() int {
	return r.bytesWritten
}

package logging

import (
	"context"
	"log/slog"
	"net/http"
)

type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type requestInfoKey struct{}

type requestInfo struct {
	remoteAddr    string
	method        string
	path          string
	contentLength int64
}

func (r *requestInfo) newLogger(logger *slog.Logger) *slog.Logger {
	return logger.With(
		"req.remote_addr", r.remoteAddr,
		"req.method", r.method,
		"req.path", r.path,
		"req.content_length", r.contentLength,
	)
}

func NewMiddleware(logger Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, requestInfoKey{}, &requestInfo{
			remoteAddr:    r.RemoteAddr,
			method:        r.Method,
			path:          r.URL.Path,
			contentLength: r.ContentLength,
		})
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

type slogLogger struct {
	bareLogger *slog.Logger
}

func NewSlogLogger(logger *slog.Logger) *slogLogger {
	return &slogLogger{bareLogger: logger}
}

func (l *slogLogger) logger(ctx context.Context) *slog.Logger {
	req, ok := ctx.Value(requestInfoKey{}).(*requestInfo)
	if ok {
		return req.newLogger(l.bareLogger)
	}
	return l.bareLogger
}

func (l *slogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.logger(ctx).Debug(msg, args...)
}

func (l *slogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.logger(ctx).Info(msg, args...)
}

func (l *slogLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.logger(ctx).Warn(msg, args...)
}

func (l *slogLogger) Error(ctx context.Context, msg string, args ...any) {
	l.logger(ctx).Error(msg, args...)
}

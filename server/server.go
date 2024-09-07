package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/chriskuehl/fluffy/server/assets"
	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage"
)

func NewConfig() *config.Config {
	return &config.Config{
		StorageBackend: &storage.FilesystemBackend{
			ObjectRoot: filepath.Join("tmp", "object"),
			HTMLRoot:   filepath.Join("tmp", "html"),
		},
		Branding:                "fluffy",
		AbuseContactEmail:       "abuse@example.com",
		MaxUploadBytes:          1024 * 1024 * 10, // 10 MiB
		MaxMultipartMemoryBytes: 1024 * 1024 * 10, // 10 MiB
		HomeURL:                 url.URL{Scheme: "http", Host: "localhost:8080"},
		ObjectURLPattern:        url.URL{Scheme: "http", Host: "localhost:8080", Path: "/dev/object/{path}"},
		HTMLURLPattern:          url.URL{Scheme: "http", Host: "localhost:8080", Path: "/dev/html/{path}"},
		ForbiddenFileExtensions: make(map[string]struct{}),
		Host:                    "127.0.0.1",
		Port:                    8080,
	}
}

func handleHealthz(logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r.Context(), "healthz")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok\n"))
	}
}

type cspNonceKey struct{}

func newCSPMiddleware(conf *config.Config, next http.Handler) http.Handler {
	objectURLBase := conf.ObjectURLPattern
	objectURLBase.Path = ""
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		nonceBytes := make([]byte, 16)
		if _, err := rand.Read(nonceBytes); err != nil {
			panic("failed to generate nonce: " + err.Error())
		}
		nonce := hex.EncodeToString(nonceBytes)
		ctx = context.WithValue(ctx, cspNonceKey{}, nonce)
		csp := fmt.Sprintf(
			"default-src %s; script-src https://ajax.googleapis.com 'nonce-%s' %[1]s; style-src https://fonts.googleapis.com %[1]s; font-src https://fonts.gstatic.com %[1]s",
			objectURLBase.String(),
			nonce,
		)
		w.Header().Set("Content-Security-Policy", csp)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func addRoutes(
	mux *http.ServeMux,
	conf *config.Config,
	logger logging.Logger,
) error {
	mux.HandleFunc("GET /healthz", handleHealthz(logger))
	if handler, err := handleIndex(conf, logger); err != nil {
		return fmt.Errorf("handleIndex: %w", err)
	} else {
		mux.Handle("GET /{$}", handler)
	}
	if handler, err := handleUploadHistory(conf, logger); err != nil {
		return fmt.Errorf("handleUploadHistory: %w", err)
	} else {
		mux.Handle("GET /upload-history", handler)
	}
	mux.Handle("POST /upload", handleUpload(conf, logger))
	mux.Handle("GET /dev/static/", assets.HandleDevStatic(conf, logger))
	mux.Handle("GET /dev/storage/{type}/", storage.HandleDevStorage(conf, logger))
	return nil
}

func NewServer(
	logger logging.Logger,
	conf *config.Config,
) (http.Handler, error) {
	if errs := conf.Validate(); len(errs) > 0 {
		return nil, errors.New("invalid config: " + strings.Join(errs, ", "))
	}
	mux := http.NewServeMux()
	if err := addRoutes(mux, conf, logger); err != nil {
		return nil, fmt.Errorf("adding routes: %w", err)
	}
	var handler http.Handler = mux
	handler = newCSPMiddleware(conf, handler)
	handler = logging.NewMiddleware(logger, handler)
	handler = http.TimeoutHandler(handler, time.Second/2, "timeout")
	return handler, nil
}

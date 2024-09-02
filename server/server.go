package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage"
)

type Config struct {
	// Site configuration.
	StorageBackend          storage.Backend
	Branding                string
	CustomFooterHTML        template.HTML
	AbuseContactEmail       string
	MaxUploadBytes          int64
	MaxMultipartMemoryBytes int64
	HomeURL                 url.URL
	ObjectURLPattern        url.URL
	HTMLURLPattern          url.URL
	ForbiddenFileExtensions map[string]struct{}

	// Runtime options.
	DevMode bool
	Version string
}

func (c *Config) Validate() []string {
	var errs []string
	if c.Branding == "" {
		errs = append(errs, "Branding must not be empty")
	}
	if c.AbuseContactEmail == "" {
		errs = append(errs, "AbuseContactEmail must not be empty")
	}
	if c.MaxUploadBytes <= 0 {
		errs = append(errs, "MaxUploadBytes must be greater than 0")
	}
	if c.MaxMultipartMemoryBytes <= 0 {
		errs = append(errs, "MaxMultipartMemoryBytes must be greater than 0")
	}
	if strings.HasSuffix(c.HomeURL.Path, "/") {
		errs = append(errs, "HomeURL must not end with a slash")
	}
	if !strings.Contains(c.ObjectURLPattern.Path, "%s") {
		errs = append(errs, "ObjectURLPattern must contain a '%s' placeholder")
	}
	if !strings.Contains(c.HTMLURLPattern.Path, "%s") {
		errs = append(errs, "HTMLURLPattern must contain a '%s' placeholder")
	}
	for ext := range c.ForbiddenFileExtensions {
		if strings.HasPrefix(ext, ".") {
			errs = append(errs, "ForbiddenFileExtensions should not start with a dot: "+ext)
		}
	}
	if c.Version == "" {
		errs = append(errs, "Version must not be empty")
	}
	return errs
}

func NewConfig() *Config {
	return &Config{
		StorageBackend: &storage.FilesystemBackend{
			ObjectRoot: filepath.Join("tmp", "object"),
			HTMLRoot:   filepath.Join("tmp", "html"),
		},
		Branding:                "fluffy",
		AbuseContactEmail:       "abuse@example.com",
		MaxUploadBytes:          1024 * 1024 * 10, // 10 MiB
		MaxMultipartMemoryBytes: 1024 * 1024 * 10, // 10 MiB
		HomeURL:                 url.URL{Scheme: "http", Host: "localhost:8080"},
		ObjectURLPattern:        url.URL{Scheme: "http", Host: "localhost:8080", Path: "/dev/object/%s"},
		HTMLURLPattern:          url.URL{Scheme: "http", Host: "localhost:8080", Path: "/dev/html/%s"},
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

func newCSPMiddleware(config *Config, next http.Handler) http.Handler {
	objectURLBase := config.ObjectURLPattern
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
	config *Config,
	logger logging.Logger,
) error {
	mux.HandleFunc("GET /healthz", handleHealthz(logger))
	if handler, err := handleIndex(config, logger); err != nil {
		return fmt.Errorf("handleIndex: %w", err)
	} else {
		mux.Handle("GET /{$}", handler)
	}
	if handler, err := handleUploadHistory(config, logger); err != nil {
		return fmt.Errorf("handleUploadHistory: %w", err)
	} else {
		mux.Handle("GET /upload-history", handler)
	}
	mux.Handle("POST /upload", handleUpload(config, logger))
	mux.Handle("GET /dev/static/", handleDevStatic(config, logger))
	mux.Handle("GET /dev/storage/{type}/", handleDevStorage(config, logger))
	return nil
}

func NewServer(
	logger logging.Logger,
	config *Config,
) (http.Handler, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, errors.New("invalid config: " + strings.Join(errs, ", "))
	}
	mux := http.NewServeMux()
	if err := addRoutes(mux, config, logger); err != nil {
		return nil, fmt.Errorf("adding routes: %w", err)
	}
	var handler http.Handler = mux
	handler = newCSPMiddleware(config, handler)
	handler = logging.NewMiddleware(logger, handler)
	return handler, nil
}

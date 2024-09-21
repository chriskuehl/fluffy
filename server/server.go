package server

import (
	"embed"
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
	"github.com/chriskuehl/fluffy/server/security"
	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/server/views"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var assetsFS embed.FS

func NewConfig() (*config.Config, error) {
	a, err := assets.LoadAssets(&assetsFS)
	if err != nil {
		return nil, fmt.Errorf("loading assets: %w", err)
	}
	return &config.Config{
		StorageBackend: &storage.FilesystemBackend{
			FileRoot: filepath.Join("tmp", "file"),
			HTMLRoot: filepath.Join("tmp", "html"),
		},
		Branding:                "fluffy",
		AbuseContactEmail:       "abuse@example.com",
		MaxUploadBytes:          1024 * 1024 * 10, // 10 MiB
		MaxMultipartMemoryBytes: 1024 * 1024 * 10, // 10 MiB
		HomeURL:                 &url.URL{Scheme: "http", Host: "localhost:8080"},
		FileURLPattern:          &url.URL{Scheme: "http", Host: "localhost:8080", Path: "/dev/storage/file/:key:"},
		HTMLURLPattern:          &url.URL{Scheme: "http", Host: "localhost:8080", Path: "/dev/storage/html/:key:"},
		ForbiddenFileExtensions: make(map[string]struct{}),
		Host:                    "127.0.0.1",
		Port:                    8080,
		GlobalTimeout:           60 * time.Second,
		Assets:                  a,
		Templates:               &config.Templates{FS: &templatesFS},
	}, nil
}

func addRoutes(
	mux *http.ServeMux,
	conf *config.Config,
	logger logging.Logger,
) error {
	mux.HandleFunc("GET /healthz", views.HandleHealthz(logger))
	if handler, err := views.HandleIndex(conf, logger); err != nil {
		return fmt.Errorf("handleIndex: %w", err)
	} else {
		mux.Handle("GET /{$}", handler)
	}
	if handler, err := views.HandleUploadHistory(conf, logger); err != nil {
		return fmt.Errorf("handleUploadHistory: %w", err)
	} else {
		mux.Handle("GET /upload-history", handler)
	}
	mux.Handle("POST /upload", views.HandleUpload(conf, logger))
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
	handler = security.NewCSPMiddleware(conf, handler)
	handler = logging.NewMiddleware(logger, handler)
	handler = http.TimeoutHandler(handler, conf.GlobalTimeout, "global timeout")
	return handler, nil
}

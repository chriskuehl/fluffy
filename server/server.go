package server

import (
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/logging"
)

type Config struct {
	// Site configuration.
	Branding                 string
	CustomFooterHTML         template.HTML
	AbuseContactEmail        string
	MaxUploadBytes           int64
	HomeURL                  string
	ObjectURLPattern         string
	HTMLURLPattern           string
	DisallowedFileExtensions []string

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
	if c.HomeURL == "" {
		errs = append(errs, "HomeURL must not be empty")
	}
	if strings.HasSuffix(c.HomeURL, "/") {
		errs = append(errs, "HomeURL must not end with a slash")
	}
	if c.ObjectURLPattern == "" {
		errs = append(errs, "ObjectURLPattern must not be empty")
	}
	if !strings.Contains(c.ObjectURLPattern, "%s") {
		errs = append(errs, "ObjectURLPattern must contain a %s")
	}
	if c.HTMLURLPattern == "" {
		errs = append(errs, "HTMLURLPattern must not be empty")
	}
	if !strings.Contains(c.HTMLURLPattern, "%s") {
		errs = append(errs, "HTMLURLPattern must contain a %s")
	}
	for _, ext := range c.DisallowedFileExtensions {
		if strings.HasPrefix(ext, ".") {
			errs = append(errs, "DisallowedFileExtensions should not start with a dot: "+ext)
		}
	}
	if c.Version == "" {
		errs = append(errs, "Version must not be empty")
	}
	return errs
}

func NewConfig() *Config {
	return &Config{
		Branding:          "fluffy",
		AbuseContactEmail: "abuse@example.com",
		MaxUploadBytes:    1024 * 1024 * 10, // 10 MiB
		HomeURL:           "http://localhost:8080",
		ObjectURLPattern:  "http://localhost:8080/dev/object/%s",
		HTMLURLPattern:    "http://localhost:8080/dev/html/%s",
	}
}

func handleHealthz(logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r.Context(), "healthz")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok\n"))
	}
}

func addRoutes(
	mux *http.ServeMux,
	config *Config,
	logger logging.Logger,
) {
	mux.HandleFunc("GET /healthz", handleHealthz(logger))
	mux.Handle("GET /{$}", handleIndex(config, logger))
	mux.Handle("GET /dev/static/", handleStatic(config, logger))
}

func NewServer(
	logger logging.Logger,
	config *Config,
) (http.Handler, error) {
	if errs := config.Validate(); len(errs) > 0 {
		return nil, errors.New("invalid config: " + strings.Join(errs, ", "))
	}
	mux := http.NewServeMux()
	addRoutes(mux, config, logger)
	var handler http.Handler = mux
	handler = logging.NewMiddleware(logger, handler)
	return handler, nil
}

package server

import (
	"net/http"

	"github.com/chriskuehl/fluffy/server/logging"
)

type Config struct{}

func handleHealthz(logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r.Context(), "healthz")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok\n"))
	}
}

func addRoutes(
	mux *http.ServeMux,
	logger logging.Logger,
) {
	mux.HandleFunc("/healthz", handleHealthz(logger))
}

func NewServer(
	logger logging.Logger,
	config *Config,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger)
	var handler http.Handler = mux
	handler = logging.NewMiddleware(logger, handler)
	return handler
}

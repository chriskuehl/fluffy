package assets

import (
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
)

func HandleDevStatic(config *config.Config, logger logging.Logger) http.HandlerFunc {
	if !config.DevMode {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Warn(r.Context(), "assets cannot be served from the server in production")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Assets cannot be served from the server in production.\n"))
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		strippedReq := r.Clone(r.Context())
		strippedReq.URL.Path = strings.TrimPrefix(strippedReq.URL.Path, "/dev")
		http.FileServer(http.FS(assetsFS)).ServeHTTP(w, strippedReq)
	}
}

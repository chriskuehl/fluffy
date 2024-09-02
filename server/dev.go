package server

import (
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage"
)

func handleDevStorage(config *Config, logger logging.Logger) http.HandlerFunc {
	if !config.DevMode {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Warn(r.Context(), "storage cannot be served from the server in production")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Storage cannot be served from the server in production.\n"))
		}
	}

	storageBackend, ok := config.StorageBackend.(*storage.FilesystemBackend)
	if !ok {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Error(r.Context(), "storage cannot be served from the server in dev mode if not using the filesystem backend")
			w.WriteHeader(http.StatusNotImplemented)
			w.Write([]byte("Storage cannot be served from the server in dev mode if not using the filesystem backend.\n"))
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		strippedReq := r.Clone(r.Context())
		strippedReq.URL.Path = strings.TrimPrefix(strippedReq.URL.Path, "/dev/storage")

		var root string

		if strings.HasPrefix(strippedReq.URL.Path, "/object/") {
			root = storageBackend.ObjectRoot
			strippedReq.URL.Path = strings.TrimPrefix(strippedReq.URL.Path, "/object/")
		} else if strings.HasPrefix(strippedReq.URL.Path, "/html/") {
			root = storageBackend.HTMLRoot
			strippedReq.URL.Path = strings.TrimPrefix(strippedReq.URL.Path, "/html/")
		} else {
			logger.Error(r.Context(), "invalid storage path", "path", strippedReq.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Invalid storage path.\n"))
			return
		}

		http.FileServer(http.Dir(root)).ServeHTTP(w, strippedReq)
	}
}

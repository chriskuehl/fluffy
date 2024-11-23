package views

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/meta"
)

func HandleUploadHistory(conf *config.Config, logger logging.Logger) (http.HandlerFunc, error) {
	extensions, err := iconExtensionsJS(conf)
	if err != nil {
		return nil, fmt.Errorf("iconExtensions: %w", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		m, err := meta.NewMeta(r.Context(), conf, meta.PageConfig{
			ID: "upload-history",
		})
		if err != nil {
			logger.Error(r.Context(), "creating meta", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		data := struct {
			Meta           *meta.Meta
			IconExtensions template.JS
		}{
			Meta:           m,
			IconExtensions: extensions,
		}
		buf := bytes.Buffer{}
		if err := conf.Templates.UploadHistory.ExecuteTemplate(&buf, "upload-history.html", data); err != nil {
			logger.Error(r.Context(), "executing template", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
		}
	}, nil
}

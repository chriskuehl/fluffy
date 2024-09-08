package views

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/highlighting"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/meta"
)

func HandleIndex(conf *config.Config, logger logging.Logger) (http.HandlerFunc, error) {
	extensions, err := iconExtensionsJS(conf)
	if err != nil {
		return nil, fmt.Errorf("iconExtensions: %w", err)
	}
	tmpl := conf.Templates.Must("index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		extraHTMLClasses := []string{}
		text, ok := r.URL.Query()["text"]
		if ok {
			extraHTMLClasses = append(extraHTMLClasses, "start-on-paste")
		}

		m, err := meta.NewMeta(r.Context(), conf, meta.PageConfig{
			ID:               "index",
			ExtraHTMLClasses: extraHTMLClasses,
		})
		if err != nil {
			logger.Error(r.Context(), "creating meta", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := struct {
			Meta           *meta.Meta
			UILanguagesMap map[string]string
			IconExtensions template.JS
			Text           string
		}{
			Meta:           m,
			UILanguagesMap: highlighting.UILanguagesMap,
			IconExtensions: extensions,
			Text:           strings.Join(text, ""),
		}
		buf := bytes.Buffer{}
		if err := tmpl.ExecuteTemplate(&buf, "index.html", data); err != nil {
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

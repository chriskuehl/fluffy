package server

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"html/template"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/logging"
)

//go:embed templates/*
var templatesFS embed.FS
var templates = template.Must(template.New("").ParseFS(templatesFS, "templates/*.html"))

type pageConfig struct {
	ID               string
	ExtraHTMLClasses []string
}

func (p pageConfig) HTMLClasses() string {
	return "page-" + p.ID + " " + strings.Join(p.ExtraHTMLClasses, " ")
}

type renderContext struct {
	Config     *Config
	PageConfig pageConfig

	Nonce string
}

func NewRenderContext(config *Config, pc pageConfig) renderContext {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		panic("failed to generate nonce: " + err.Error())
	}

	return renderContext{
		Config:     config,
		PageConfig: pc,

		Nonce: hex.EncodeToString(nonce),
	}
}

func handleIndex(config *Config, logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		data := struct {
			RenderContext renderContext
		}{
			RenderContext: NewRenderContext(config, pageConfig{
				ID:               "index",
				ExtraHTMLClasses: []string{"blah", "blarg"},
			}),
		}
		buf := bytes.Buffer{}
		if err := templates.ExecuteTemplate(&buf, "index.html", data); err != nil {
			logger.Error(r.Context(), "executing template", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
		}
	}
}

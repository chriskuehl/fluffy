package server

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/highlighting"
	"github.com/chriskuehl/fluffy/server/logging"
)

//go:embed templates/*
var templatesFS embed.FS

type pageConfig struct {
	ID               string
	ExtraHTMLClasses []string
}

func (p pageConfig) HTMLClasses() string {
	return "page-" + p.ID + " " + strings.Join(p.ExtraHTMLClasses, " ")
}

type meta struct {
	Config     *Config
	PageConfig pageConfig
	Nonce      string
}

func NewMeta(config *Config, pc pageConfig) meta {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		panic("failed to generate nonce: " + err.Error())
	}

	return meta{
		Config:     config,
		PageConfig: pc,
		Nonce:      hex.EncodeToString(nonce),
	}
}

func pageTemplate(name string) *template.Template {
	return template.Must(template.New("").ParseFS(templatesFS, "templates/include/*.html", "templates/"+name))
}

func handleIndex(config *Config, logger logging.Logger) (http.HandlerFunc, error) {
	extensionToURL := make(map[string]string)
	for _, ext := range mimeExtensions {
		extensionToURL[ext] = config.AssetURL("img/mime/small/" + ext + ".png")
	}
	json, err := json.Marshal(extensionToURL)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mime extensions to JSON: %w", err)
	}
	extensions := template.JS(json)

	tmpl := pageTemplate("index.html")

	return func(w http.ResponseWriter, r *http.Request) {
		extraHTMLClasses := []string{}
		text, ok := r.URL.Query()["text"]
		if ok {
			extraHTMLClasses = append(extraHTMLClasses, "start-on-paste")
		}

		data := struct {
			Meta           meta
			UILanguagesMap map[string]string
			IconExtensions template.JS
			Text           string
		}{
			Meta: NewMeta(config, pageConfig{
				ID:               "index",
				ExtraHTMLClasses: extraHTMLClasses,
			}),
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

func handleUploadHistory(config *Config, logger logging.Logger) (http.HandlerFunc, error) {
	extensionToURL := make(map[string]string)
	for _, ext := range mimeExtensions {
		extensionToURL[ext] = config.AssetURL("img/mime/small/" + ext + ".png")
	}
	json, err := json.Marshal(extensionToURL)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal mime extensions to JSON: %w", err)
	}
	extensions := template.JS(json)

	tmpl := pageTemplate("upload-history.html")

	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Meta           meta
			IconExtensions template.JS
		}{
			Meta: NewMeta(config, pageConfig{
				ID: "upload-history",
			}),
			IconExtensions: extensions,
		}
		buf := bytes.Buffer{}
		if err := tmpl.ExecuteTemplate(&buf, "upload-history.html", data); err != nil {
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

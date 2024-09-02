package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"

	"github.com/chriskuehl/fluffy/server/highlighting"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/server/uploads"
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

func NewMeta(ctx context.Context, config *Config, pc pageConfig) meta {
	nonce, ok := ctx.Value(cspNonceKey{}).(string)
	if !ok {
		panic("no nonce in context")
	}
	return meta{
		Config:     config,
		PageConfig: pc,
		Nonce:      nonce,
	}
}

func pageTemplate(name string) *template.Template {
	return template.Must(template.New("").ParseFS(templatesFS, "templates/include/*.html", "templates/"+name))
}

func iconExtensions(config *Config) (template.JS, error) {
	extensionToURL := make(map[string]string)
	for _, ext := range mimeExtensions {
		extensionToURL[ext] = config.AssetURL("img/mime/small/" + ext + ".png")
	}
	json, err := json.Marshal(extensionToURL)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mime extensions to JSON: %w", err)
	}
	return template.JS(json), nil
}

func handleIndex(config *Config, logger logging.Logger) (http.HandlerFunc, error) {
	extensions, err := iconExtensions(config)
	if err != nil {
		return nil, fmt.Errorf("iconExtensions: %w", err)
	}
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
			Meta: NewMeta(r.Context(), config, pageConfig{
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
	extensions, err := iconExtensions(config)
	if err != nil {
		return nil, fmt.Errorf("iconExtensions: %w", err)
	}
	tmpl := pageTemplate("upload-history.html")

	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Meta           meta
			IconExtensions template.JS
		}{
			Meta: NewMeta(r.Context(), config, pageConfig{
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

func handleUpload(config *Config, logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(config.MaxMultipartMemoryBytes)
		if err != nil {
			logger.Error(r.Context(), "parsing form", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to parse multipart form.\n"))
			return
		}

		_, json := r.URL.Query()["json"]
		if _, ok := r.MultipartForm.Value["json"]; ok {
			json = true
		}
		fmt.Printf("json: %v\n", json)

		objs := []storage.Object{}

		for _, fileHeader := range r.MultipartForm.File["file"] {
			file, err := fileHeader.Open()
			if err != nil {
				logger.Error(r.Context(), "opening file", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to open file.\n"))
				return
			}
			defer file.Close()

			// TODO: check file size (keep in mind fileHeader.Size might be a lie?)
			//    -- but maybe not? since Go buffers it first?
			key, err := uploads.SanitizeUploadName(fileHeader.Filename, config.ForbiddenFileExtensions)
			if err != nil {
				// TODO: handle ErrForbiddenExtension and return useful error
				logger.Error(r.Context(), "sanitizing upload name", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to sanitize upload name.\n"))
				return
			}

			obj := storage.Object{
				Key:    *key,
				Reader: file,
			}
			objs = append(objs, obj)
		}

		results := make(chan error, len(objs))
		var wg sync.WaitGroup
		wg.Add(len(objs))
		for _, obj := range objs {
			go func() {
				defer wg.Done()
				if err := config.StorageBackend.StoreObject(r.Context(), obj); err != nil {
					logger.Error(r.Context(), "storing object", "error", err)
					results <- err
				} else {
					results <- nil
				}
			}()
		}
		wg.Wait()

		hadError := false
		for i := 0; i < len(objs); i++ {
			if err := <-results; err != nil {
				logger.Error(r.Context(), "storing object", "error", err)
				hadError = true
			}
		}

		if hadError {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to store object.\n"))
			return
		}

		logger.Info(r.Context(), "uploaded", "objects", len(objs))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Uploaded.\n"))
	}
}

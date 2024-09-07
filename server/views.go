package server

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/assets"
	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/highlighting"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage/storagedata"
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

func pageTemplate(name string) *template.Template {
	return template.Must(template.New("").ParseFS(templatesFS, "templates/include/*.html", "templates/"+name))
}

func iconExtensions(config *config.Config) (template.JS, error) {
	extensionToURL := make(map[string]string)
	for _, ext := range assets.MimeExtensions() {
		url, err := assets.AssetURL(config, "img/mime/small/"+ext+".png")
		if err != nil {
			return "", fmt.Errorf("failed to get asset URL for %q: %w", ext, err)
		}
		extensionToURL[ext] = url
	}
	json, err := json.Marshal(extensionToURL)
	if err != nil {
		return "", fmt.Errorf("failed to marshal mime extensions to JSON: %w", err)
	}
	return template.JS(json), nil
}

func handleIndex(config *config.Config, logger logging.Logger) (http.HandlerFunc, error) {
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

func handleUploadHistory(config *config.Config, logger logging.Logger) (http.HandlerFunc, error) {
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

type UploadResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func handleUpload(config *config.Config, logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jsonError := func(statusCode int, msg string) {
			w.WriteHeader(statusCode)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(UploadResponse{
				Success: false,
				Error:   msg,
			})
		}

		err := r.ParseMultipartForm(config.MaxMultipartMemoryBytes)
		if err != nil {
			logger.Error(r.Context(), "parsing multipart form", "error", err)
			jsonError(http.StatusBadRequest, "Could not parse multipart form.")
			return
		}

		_, json := r.URL.Query()["json"]
		if _, ok := r.MultipartForm.Value["json"]; ok {
			json = true
		}
		fmt.Printf("json: %v\n", json)

		objs := []storagedata.Object{}

		fmt.Printf("files: %v\n", r.MultipartForm.File["file"])

		for _, fileHeader := range r.MultipartForm.File["file"] {
			fmt.Printf("file: %v\n", fileHeader.Filename)
			file, err := fileHeader.Open()
			if err != nil {
				logger.Error(r.Context(), "opening file", "error", err)
				jsonError(http.StatusInternalServerError, "Could not open file.")
				return
			}
			defer file.Close()

			// TODO: check file size (keep in mind fileHeader.Size might be a lie?)
			//    -- but maybe not? since Go buffers it first?
			key, err := uploads.SanitizeUploadName(fileHeader.Filename, config.ForbiddenFileExtensions)
			if err != nil {
				if errors.Is(err, uploads.ErrForbiddenExtension) {
					logger.Info(r.Context(), "forbidden extension", "filename", fileHeader.Filename)
					jsonError(
						http.StatusBadRequest,
						fmt.Sprintf("Sorry, %q has a forbidden file extension.", fileHeader.Filename),
					)
					return
				}
				logger.Error(r.Context(), "sanitizing upload name", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to sanitize upload name.\n"))
				return
			}

			obj := storagedata.Object{
				Key:    key.String(),
				Reader: file,
			}
			objs = append(objs, obj)
		}

		fmt.Printf("objs: %v\n", objs)

		if len(objs) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No files uploaded.\n"))
			return
		}

		errs := uploads.UploadObjects(r.Context(), logger, config, objs)

		if len(errs) > 0 {
			logger.Error(r.Context(), "uploading objects failed", "errors", errs)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to store object.\n"))
			return
		}

		logger.Info(r.Context(), "uploaded", "objects", len(objs))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Uploaded.\n"))
	}
}

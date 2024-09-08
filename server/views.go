package server

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/assets"
	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/highlighting"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage/storagedata"
	"github.com/chriskuehl/fluffy/server/uploads"
	"github.com/chriskuehl/fluffy/server/utils"
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

func iconExtensions(conf *config.Config) (template.JS, error) {
	extensionToURL := make(map[string]string)
	for _, ext := range assets.MimeExtensions() {
		url, err := assets.AssetURL(conf, "img/mime/small/"+ext+".png")
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

func handleIndex(conf *config.Config, logger logging.Logger) (http.HandlerFunc, error) {
	extensions, err := iconExtensions(conf)
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
			Meta: NewMeta(r.Context(), conf, pageConfig{
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

func handleUploadHistory(conf *config.Config, logger logging.Logger) (http.HandlerFunc, error) {
	extensions, err := iconExtensions(conf)
	if err != nil {
		return nil, fmt.Errorf("iconExtensions: %w", err)
	}
	tmpl := pageTemplate("upload-history.html")

	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Meta           meta
			IconExtensions template.JS
		}{
			Meta: NewMeta(r.Context(), conf, pageConfig{
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

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type userError struct {
	code    int
	message string
}

func (e userError) Error() string {
	return e.message
}

func (e userError) output(w http.ResponseWriter) {
	w.WriteHeader(e.code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(errorResponse{
		Success: false,
		Error:   e.message,
	})
}

type uploadedFile struct {
	Bytes int64  `json:"bytes"`
	Raw   string `json:"raw"`
	Paste string `json:"paste,omitempty"`
}

type uploadResponse struct {
	errorResponse
	Redirect      string                  `json:"redirect"`
	Metadata      string                  `json:"metadata"`
	UploadedFiles map[string]uploadedFile `json:"uploadedFiles"`
}

func objectFromFileHeader(
	ctx context.Context,
	conf *config.Config,
	logger logging.Logger,
	fileHeader *multipart.FileHeader,
) (*storagedata.Object, error) {
	file, err := fileHeader.Open()
	if err != nil {
		logger.Error(ctx, "opening file", "error", err)
		return nil, userError{http.StatusBadRequest, "Could not open file."}
	}
	defer file.Close()

	if fileHeader.Size > conf.MaxUploadBytes {
		logger.Info(ctx, "file too large", "size", fileHeader.Size)
		return nil, userError{
			http.StatusBadRequest,
			fmt.Sprintf("File is too large; max size is %s.", utils.FormatBytes(conf.MaxUploadBytes)),
		}
	}

	key, err := uploads.SanitizeUploadName(fileHeader.Filename, conf.ForbiddenFileExtensions)
	if err != nil {
		if errors.Is(err, uploads.ErrForbiddenExtension) {
			logger.Info(ctx, "forbidden extension", "filename", fileHeader.Filename)
			return nil, userError{http.StatusBadRequest, fmt.Sprintf("Sorry, %q has a forbidden file extension.", fileHeader.Filename)}
		}
		logger.Error(ctx, "sanitizing upload name", "error", err)
		return nil, userError{http.StatusInternalServerError, "Failed to sanitize upload name."}
	}

	return &storagedata.Object{
		Key:    key.String(),
		Reader: file,
		Bytes:  fileHeader.Size,
	}, nil
}

func handleUpload(conf *config.Config, logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(conf.MaxMultipartMemoryBytes)
		if err != nil {
			logger.Error(r.Context(), "parsing multipart form", "error", err)
			userError{http.StatusBadRequest, "Could not parse multipart form."}.output(w)
			return
		}

		_, jsonResponse := r.URL.Query()["json"]
		if _, ok := r.MultipartForm.Value["json"]; ok {
			jsonResponse = true
		}

		objs := []storagedata.Object{}

		for _, fileHeader := range r.MultipartForm.File["file"] {
			obj, err := objectFromFileHeader(r.Context(), conf, logger, fileHeader)
			if err != nil {
				userErr, ok := err.(userError)
				if !ok {
					logger.Error(r.Context(), "unexpected error", "error", err)
					userErr = userError{http.StatusInternalServerError, "An unexpected error occurred."}
				}
				userErr.output(w)
				return
			}
			objs = append(objs, *obj)
		}

		if len(objs) == 0 {
			logger.Info(r.Context(), "no files uploaded")
			userError{http.StatusBadRequest, "No files uploaded."}.output(w)
			return
		}

		errs := uploads.UploadObjects(r.Context(), logger, conf, objs)

		if len(errs) > 0 {
			logger.Error(r.Context(), "uploading objects failed", "errors", errs)
			userError{http.StatusInternalServerError, "Failed to store object."}.output(w)
			return
		}

		logger.Info(r.Context(), "uploaded", "objects", len(objs))

		redirect := conf.ObjectURL(objs[0].Key).String()

		if jsonResponse {
			uploadedFiles := make(map[string]uploadedFile, len(objs))
			for _, obj := range objs {
				uploadedFiles[obj.Key] = uploadedFile{
					Bytes: obj.Bytes,
					Raw:   conf.ObjectURL(obj.Key).String(),
					// TODO: Paste for text files
				}
			}

			resp := uploadResponse{
				errorResponse: errorResponse{
					Success: true,
				},
				Redirect:      redirect,
				Metadata:      "TODO",
				UploadedFiles: uploadedFiles,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		} else {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
		}
	}
}

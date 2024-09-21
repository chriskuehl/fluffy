package views

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/meta"
	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/server/uploads"
	"github.com/chriskuehl/fluffy/server/utils"
)

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
	UploadedFiles map[string]uploadedFile `json:"uploaded_files"`
}

// storedFileFromFileHeader creates a StoredFile from a multipart.FileHeader.
//
// Note: The *caller* is responsible for closing the returned StoredFile.
func storedFileFromFileHeader(
	ctx context.Context,
	conf *config.Config,
	logger logging.Logger,
	fileHeader *multipart.FileHeader,
) (config.StoredFile, error) {
	file, err := fileHeader.Open()
	if err != nil {
		logger.Error(ctx, "opening file", "error", err)
		return nil, userError{http.StatusBadRequest, "Could not open file."}
	}

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

	probablyText, err := uploads.ProbablyText(file)
	if err != nil {
		logger.Error(ctx, "determining if file is text", "error", err)
		return nil, userError{http.StatusInternalServerError, "Failed to determine if file is text."}
	}

	mimeType := uploads.DetermineMIMEType(
		fileHeader.Filename,
		fileHeader.Header.Get("Content-Type"),
		probablyText,
	)

	name := "file"
	if fileHeader.Filename != "" {
		name = fileHeader.Filename
	}

	return storage.NewStoredFile(
		file,
		storage.WithKey(key.String()),
		storage.WithName(name),
		storage.WithMIMEType(mimeType),
		storage.WithContentDisposition(
			uploads.DetermineContentDisposition(
				fileHeader.Filename,
				mimeType,
				probablyText,
			),
		),
	), nil
}

func HandleUpload(conf *config.Config, logger logging.Logger) http.HandlerFunc {
	uploadDetailsTmpl := conf.Templates.Must("upload-details.html")

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

		files := []config.StoredFile{}

		for _, fileHeader := range r.MultipartForm.File["file"] {
			file, err := storedFileFromFileHeader(r.Context(), conf, logger, fileHeader)
			if err != nil {
				userErr, ok := err.(userError)
				if !ok {
					logger.Error(r.Context(), "unexpected error", "error", err)
					userErr = userError{http.StatusInternalServerError, "An unexpected error occurred."}
				}
				userErr.output(w)
				return
			}
			defer file.Close()
			files = append(files, file)
		}

		if len(files) == 0 {
			logger.Info(r.Context(), "no files uploaded")
			userError{http.StatusBadRequest, "No files uploaded."}.output(w)
			return
		}

		// Metadata
		metadataKey, err := uploads.GenUniqueObjectKey()
		if err != nil {
			logger.Error(r.Context(), "generating unique object key", "error", err)
			userError{http.StatusInternalServerError, "Failed to generate unique object key."}.output(w)
			return
		}

		metadata, err := uploads.NewUploadMetadata(conf, files)
		if err != nil {
			logger.Error(r.Context(), "creating metadata", "error", err)
			userError{http.StatusInternalServerError, "Failed to create metadata."}.output(w)
			return
		}

		var metadataJSON bytes.Buffer
		if err := json.NewEncoder(&metadataJSON).Encode(metadata); err != nil {
			logger.Error(r.Context(), "encoding metadata", "error", err)
			userError{http.StatusInternalServerError, "Failed to encode metadata."}.output(w)
			return
		}

		metadataFile := storage.NewStoredFile(
			utils.NopReadSeekCloser(bytes.NewReader(metadataJSON.Bytes())),
			storage.WithKey(metadataKey+".json"),
			storage.WithMIMEType("application/json"),
		)
		metadataURL := conf.FileURL(metadataFile.Key())

		// Upload details HTML page
		uploadDetailsKey, err := uploads.GenUniqueObjectKey()
		if err != nil {
			logger.Error(r.Context(), "generating unique object key", "error", err)
			userError{http.StatusInternalServerError, "Failed to generate unique object key."}.output(w)
			return
		}

		uploadDetailsMeta, err := meta.NewMeta(r.Context(), conf, meta.PageConfig{
			ID: "upload-details",
		})
		if err != nil {
			logger.Error(r.Context(), "creating meta", "error", err)
			userError{http.StatusInternalServerError, "Failed to create response."}.output(w)
			return
		}

		var uploadDetails bytes.Buffer
		uploadDetailsData := struct {
			Meta *meta.Meta
		}{
			Meta: uploadDetailsMeta,
		}
		if err := uploadDetailsTmpl.ExecuteTemplate(&uploadDetails, "upload-details.html", uploadDetailsData); err != nil {
			logger.Error(r.Context(), "executing template", "error", err)
			userError{http.StatusInternalServerError, "Failed to create response."}.output(w)
			return
		}
		uploadDetailsHTML := storage.NewStoredHTML(
			utils.NopReadSeekCloser(bytes.NewReader(uploadDetails.Bytes())),
			storage.WithKey(uploadDetailsKey+".html"),
		)

		// Update metadata and links for everything we're about to update.
		uploadHTMLs := []config.StoredHTML{uploadDetailsHTML}
		uploadFiles := append([]config.StoredFile{metadataFile}, files...)
		links := make([]*url.URL, 0, len(uploadFiles)+len(uploadHTMLs))
		for _, file := range uploadFiles {
			links = append(links, conf.FileURL(file.Key()))
		}
		for _, html := range uploadHTMLs {
			links = append(links, conf.HTMLURL(html.Key()))
		}

		for i := range uploadHTMLs {
			uploadHTMLs[i] = storage.UpdatedStoredHTML(
				uploadHTMLs[i],
				storage.WithMetadataURL(metadataURL),
				storage.WithLinks(links),
			)
		}
		for i := range uploadFiles {
			uploadFiles[i] = storage.UpdatedStoredFile(
				uploadFiles[i],
				storage.WithMetadataURL(metadataURL),
				storage.WithLinks(links),
			)
		}

		errs := uploads.UploadObjects(r.Context(), logger, conf, uploadFiles, uploadHTMLs)

		if len(errs) > 0 {
			logger.Error(r.Context(), "uploading objects failed", "errors", errs)
			userError{http.StatusInternalServerError, "Failed to store objects."}.output(w)
			return
		}

		logger.Info(r.Context(), "uploaded", "files", len(uploadFiles), "htmls", len(uploadHTMLs))

		redirect := conf.HTMLURL(uploadDetailsHTML.Key()).String()

		if jsonResponse {
			uploadedFiles := make(map[string]uploadedFile, len(files))
			for _, file := range files {
				bytes, err := utils.FileSizeBytes(file)
				if err != nil {
					logger.Error(r.Context(), "getting file size", "error", err)
					userError{http.StatusInternalServerError, "Failed to get file size."}.output(w)
					return
				}
				uploadedFiles[file.Name()] = uploadedFile{
					Bytes: bytes,
					Raw:   conf.FileURL(file.Key()).String(),
					// TODO: Paste for text files
				}
			}

			resp := uploadResponse{
				errorResponse: errorResponse{
					Success: true,
				},
				Redirect:      redirect,
				Metadata:      metadataURL.String(),
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

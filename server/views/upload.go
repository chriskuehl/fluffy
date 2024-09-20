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

// objectFromFileHeader creates a StoredObject from a multipart.FileHeader.
//
// Note: The *caller* is responsible for closing the returned object.
func objectFromFileHeader(
	ctx context.Context,
	conf *config.Config,
	logger logging.Logger,
	fileHeader *multipart.FileHeader,
) (config.StoredObject, error) {
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

	return storage.NewStoredObject(
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

		objs := []config.StoredObject{}

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
			defer obj.Close()
			objs = append(objs, obj)
		}

		if len(objs) == 0 {
			logger.Info(r.Context(), "no files uploaded")
			userError{http.StatusBadRequest, "No files uploaded."}.output(w)
			return
		}

		metadataKey, err := uploads.GenUniqueObjectID()
		if err != nil {
			logger.Error(r.Context(), "generating unique object ID", "error", err)
			userError{http.StatusInternalServerError, "Failed to generate unique object ID."}.output(w)
			return
		}

		metadata, err := uploads.NewUploadMetadata(conf, objs)
		if err != nil {
			logger.Error(r.Context(), "creating metadata", "error", err)
			userError{http.StatusInternalServerError, "Failed to create metadata."}.output(w)
			return
		}

		// Convert the metadata to JSON.
		var metadataJSON bytes.Buffer
		if err := json.NewEncoder(&metadataJSON).Encode(metadata); err != nil {
			logger.Error(r.Context(), "encoding metadata", "error", err)
			userError{http.StatusInternalServerError, "Failed to encode metadata."}.output(w)
			return
		}

		metadataObject := storage.NewStoredObject(
			utils.NopReadSeekCloser(bytes.NewReader(metadataJSON.Bytes())),
			storage.WithKey(metadataKey+".json"),
			storage.WithName("metadata.json"),
			storage.WithMIMEType("application/json"),
		)
		metadataURL := conf.ObjectURL(metadataObject.Key())

		// uploadObjs includes extra objects like metadata, auto-generated pastes, etc. which
		// shouldn't be in the returned JSON.
		uploadObjs := []config.StoredObject{metadataObject}
		for _, obj := range objs {
			uploadObjs = append(uploadObjs, obj)
		}
		links := make([]*url.URL, len(uploadObjs))
		for i, obj := range uploadObjs {
			links[i] = conf.ObjectURL(obj.Key())
		}

		for i := range uploadObjs {
			uploadObjs[i] = storage.UpdatedStoredObject(
				uploadObjs[i],
				storage.WithMetadataURL(metadataURL),
				storage.WithLinks(links),
			)
		}

		errs := uploads.UploadObjects(r.Context(), logger, conf, uploadObjs)

		if len(errs) > 0 {
			logger.Error(r.Context(), "uploading objects failed", "errors", errs)
			userError{http.StatusInternalServerError, "Failed to store object."}.output(w)
			return
		}

		logger.Info(r.Context(), "uploaded", "objects", len(objs))

		redirect := conf.ObjectURL(objs[0].Key()).String()

		if jsonResponse {
			uploadedFiles := make(map[string]uploadedFile, len(objs))
			for _, obj := range objs {
				bytes, err := utils.FileSizeBytes(obj)
				if err != nil {
					logger.Error(r.Context(), "getting file size", "error", err)
					userError{http.StatusInternalServerError, "Failed to get file size."}.output(w)
					return
				}
				uploadedFiles[obj.Name()] = uploadedFile{
					Bytes: bytes,
					Raw:   conf.ObjectURL(obj.Key()).String(),
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

package views

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/highlighting"
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

type uploadResponse struct {
	errorResponse
	Redirect      string                  `json:"redirect"`
	Metadata      string                  `json:"metadata"`
	UploadedFiles map[string]uploadedFile `json:"uploaded_files"`
}

type uploadedFile struct {
	Bytes    int64                `json:"bytes"`
	Raw      string               `json:"raw"`
	Paste    string               `json:"paste,omitempty"`
	Language uploadedFileLanguage `json:"language,omitempty"`
	NumLines int                  `json:"num_lines,omitempty"`
}

type uploadedFileLanguage struct {
	Title string `json:"title"`
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

	name := "file"
	if fileHeader.Filename != "" {
		name = fileHeader.Filename
	}

	key, err := uploads.SanitizeUploadName(name, conf.ForbiddenFileExtensions)
	if err != nil {
		if errors.Is(err, uploads.ErrForbiddenExtension) {
			logger.Info(ctx, "forbidden extension", "filename", name)
			return nil, userError{http.StatusBadRequest, fmt.Sprintf("Sorry, %q has a forbidden file extension.", name)}
		}
		logger.Error(ctx, "sanitizing upload name", "error", err)
		return nil, userError{http.StatusInternalServerError, "Failed to sanitize upload name."}
	}

	probablyText, err := uploads.ProbablyText(file)
	if err != nil {
		logger.Error(ctx, "determining if file is text", "error", err)
		return nil, userError{http.StatusInternalServerError, "Failed to determine if file is text."}
	}

	mimeType := uploads.DetermineMIMEType(name, fileHeader.Header.Get("Content-Type"), probablyText)

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
		metadata, err := uploads.NewUploadMetadata(conf, files)
		if err != nil {
			logger.Error(r.Context(), "creating metadata", "error", err)
			userError{http.StatusInternalServerError, "Failed to create metadata."}.output(w)
			return
		}

		// Upload details HTML page
		uploadDetailsKey, err := uploads.GenUniqueObjectKey()
		if err != nil {
			logger.Error(r.Context(), "generating unique object key", "error", err)
			userError{http.StatusInternalServerError, "Failed to generate unique object key."}.output(w)
			return
		}

		uploadDetailsMeta, err := meta.NewMeta(r.Context(), conf, meta.PageConfig{
			ID:       "upload-details",
			IsStatic: true,
		})
		if err != nil {
			logger.Error(r.Context(), "creating meta", "error", err)
			userError{http.StatusInternalServerError, "Failed to create response."}.output(w)
			return
		}

		var uploadDetails bytes.Buffer

		type uploadedFileData struct {
			config.StoredFile
			IsImage bool
			Bytes   int64
		}
		uploadDetailsData := struct {
			Meta          *meta.Meta
			UploadedFiles []uploadedFileData
			MetadataURL   string
		}{
			Meta:          uploadDetailsMeta,
			UploadedFiles: make([]uploadedFileData, 0, len(files)),
			MetadataURL:   metadata.URL(conf).String(),
		}
		for _, file := range files {
			// TODO: this is ridiculous, we've done this a bajillion times now
			bytes, err := utils.FileSizeBytes(file)
			if err != nil {
				logger.Error(r.Context(), "getting file size", "error", err)
				userError{http.StatusInternalServerError, "Failed to get file size."}.output(w)
				return
			}
			uploadDetailsData.UploadedFiles = append(uploadDetailsData.UploadedFiles, uploadedFileData{
				StoredFile: file,
				IsImage:    uploads.IsImageMIME(file.MIMEType()),
				Bytes:      bytes,
			})
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

		// Upload
		errs := uploads.UploadObjects(r.Context(), logger, conf, files, []config.StoredHTML{uploadDetailsHTML}, metadata)
		if len(errs) > 0 {
			logger.Error(r.Context(), "uploading objects failed", "errors", errs)
			userError{http.StatusInternalServerError, "Failed to store objects."}.output(w)
			return
		}

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
				Metadata:      metadata.URL(conf).String(),
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

func normalizeFormText(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

func unifiedDiff(text1, text2 string) string {
	return "TODO: unifiedDiff"
}

func normalizeTextAndLanguage(text, diffText1, diffText2, language string) (string, string) {
	if language == "diff-between-two-texts" {
		return unifiedDiff(normalizeFormText(diffText1), normalizeFormText(diffText2)), "diff"
	} else {
		return normalizeFormText(text), language
	}
}

func HandlePaste(conf *config.Config, logger logging.Logger) http.HandlerFunc {
	pasteTmpl := conf.Templates.Must("paste.html")

	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			// cli versions 2.0.0 through 2.2.0 send multipart/form-data, so we need to support it forever.
			err := r.ParseMultipartForm(conf.MaxMultipartMemoryBytes)
			if err != nil {
				logger.Error(r.Context(), "parsing multipart form", "error", err)
				userError{http.StatusBadRequest, "Could not parse multipart form."}.output(w)
				return
			}
		} else {
			err := r.ParseForm()
			if err != nil {
				logger.Error(r.Context(), "parsing form", "error", err)
				userError{http.StatusBadRequest, "Could not parse form."}.output(w)
				return
			}
		}

		_, jsonResponse := r.URL.Query()["json"]
		if _, ok := r.Form["json"]; ok {
			jsonResponse = true
		}
		fmt.Printf("jsonResponse: %v\n", jsonResponse)

		text, language := normalizeTextAndLanguage(
			r.Form.Get("text"),
			r.Form.Get("diff1"),
			r.Form.Get("diff2"),
			r.Form.Get("language"),
		)
		highlighter := highlighting.GuessHighlighterForPaste(text, language, "TODO filename")

		// Raw paste
		rawKey, err := uploads.GenUniqueObjectKey()
		if err != nil {
			logger.Error(r.Context(), "generating unique object key", "error", err)
			userError{http.StatusInternalServerError, "Failed to generate unique object key."}.output(w)
			return
		}

		rawFile := storage.NewStoredFile(
			utils.NopReadSeekCloser(strings.NewReader(text)),
			storage.WithKey(rawKey+".txt"),
			storage.WithMIMEType("text/plain"),
		)

		// Metadata
		// TODO: update for pastes
		metadata, err := uploads.NewUploadMetadata(conf, []config.StoredFile{rawFile})
		if err != nil {
			logger.Error(r.Context(), "creating metadata", "error", err)
			userError{http.StatusInternalServerError, "Failed to create metadata."}.output(w)
			return
		}

		metadata.URL(conf).String()

		// Paste HTML page
		pasteKey, err := uploads.GenUniqueObjectKey()
		if err != nil {
			logger.Error(r.Context(), "generating unique object key", "error", err)
			userError{http.StatusInternalServerError, "Failed to generate unique object key."}.output(w)
			return
		}

		pasteMeta, err := meta.NewMeta(r.Context(), conf, meta.PageConfig{
			ID:               "paste",
			IsStatic:         true,
			ExtraHTMLClasses: highlighter.ExtraHTMLClasses(),
		})
		if err != nil {
			logger.Error(r.Context(), "creating meta", "error", err)
			userError{http.StatusInternalServerError, "Failed to create response."}.output(w)
			return
		}

		var paste bytes.Buffer

		// TODO: calculate this properly
		mapping := make([][]int, 0)
		for i := 0; i < len(strings.Split(text, "\n")); i++ {
			mapping = append(mapping, []int{i})
		}

		texts := highlighter.GenerateTexts(text)
		primaryText := texts[0]

		pasteData := struct {
			Meta        *meta.Meta
			MetadataURL string
			// localStorage variable name for preferred style (either "preferredStyle" or
			// "preferredStyleTerminal").
			PreferredStyleVar string
			// Default style to use if no preferred style is set in the user's localStorage.
			DefaultStyle    highlighting.Style
			CopyAndEditText string
			RawURL          string
			Styles          []highlighting.StyleCategory
			Highlighter     highlighting.Highlighter
			Texts           []*highlighting.Text
		}{
			Meta:              pasteMeta,
			MetadataURL:       metadata.URL(conf).String(),
			PreferredStyleVar: "preferredStyle",
			DefaultStyle:      highlighting.DefaultStyle,
			// TODO: does this need any transformation?
			// TODO: this is not used in the template yet
			CopyAndEditText: text,
			RawURL:          conf.FileURL(rawFile.Key()).String(),
			Styles:          highlighting.Styles,
			Highlighter:     highlighter,
			Texts:           texts,
		}

		// Terminal output gets its own preferred theme setting since many people
		// seem to prefer a dark background for terminal output, but a light
		// background for regular code.
		if highlighter.RenderAsTerminal() {
			pasteData.PreferredStyleVar = "preferredStyleTerminal"
			pasteData.DefaultStyle = highlighting.DefaultDarkStyle
		}

		if highlighter.RenderAsDiff() {
			pasteMeta.PageConf.ExtraHTMLClasses = append(pasteMeta.PageConf.ExtraHTMLClasses, "diff-side-by-side")
		}

		if err := pasteTmpl.ExecuteTemplate(&paste, "paste.html", pasteData); err != nil {
			logger.Error(r.Context(), "executing template", "error", err)
			userError{http.StatusInternalServerError, "Failed to create response."}.output(w)
			return
		}
		pasteHTML := storage.NewStoredHTML(
			utils.NopReadSeekCloser(bytes.NewReader(paste.Bytes())),
			storage.WithKey(pasteKey+".html"),
		)

		// Upload
		errs := uploads.UploadObjects(r.Context(), logger, conf, []config.StoredFile{rawFile}, []config.StoredHTML{pasteHTML}, metadata)
		if len(errs) > 0 {
			logger.Error(r.Context(), "uploading objects failed", "errors", errs)
			userError{http.StatusInternalServerError, "Failed to store objects."}.output(w)
			return
		}

		redirect := conf.HTMLURL(pasteHTML.Key()).String()

		if jsonResponse {
			bytes, err := utils.FileSizeBytes(rawFile)
			if err != nil {
				logger.Error(r.Context(), "getting file size", "error", err)
				userError{http.StatusInternalServerError, "Failed to get file size."}.output(w)
				return
			}

			resp := uploadResponse{
				errorResponse: errorResponse{
					Success: true,
				},
				Redirect: redirect,
				Metadata: metadata.URL(conf).String(),
				UploadedFiles: map[string]uploadedFile{
					"paste": {
						// To be least confusing, we return size of the raw file. This may not
						// match the rendered text if the highlighter applied any transformations.
						Bytes: bytes,
						Raw:   conf.FileURL(rawFile.Key()).String(),
						Paste: conf.HTMLURL(pasteHTML.Key()).String(),
						Language: uploadedFileLanguage{
							Title: highlighter.Name(),
						},
						// This may not match the number of lines in the raw text but since this
						// field is generally used for display purposes (e.g. link previews), it's
						// probably better to provide the number of lines in the rendered text
						// here.
						NumLines: len(primaryText.LineNumberMapping),
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		} else {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
		}
	}
}

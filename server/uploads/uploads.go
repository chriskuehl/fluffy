package uploads

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"mime"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/server/utils"
)

const (
	storedFileNameLength = 32
	storedFileNameChars  = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789"

	mimeGenericBinary = "application/octet-stream"
	mimeGenericText   = "text/plain"
)

var (
	ErrForbiddenExtension = errors.New("forbidden extension")

	// MIME types which are allowed to be presented as detected.
	// TODO: I think we actually only need to prevent text/html (and any HTML
	// variants like XHTML)?
	mimeAllowlist = map[string]struct{}{
		"application/javascript": {},
		"application/json":       {},
		"application/pdf":        {},
		"application/x-ruby":     {},
		"text/css":               {},
		"text/plain":             {},
		"text/x-python":          {},
		"text/x-sh":              {},
	}
	mimePrefixAllowlist = []string{
		"audio/",
		"image/",
		"video/",
	}
	inlineDisplayMIMEAllowlist = map[string]struct{}{
		"application/pdf": {},
	}
	inlineDisplayMIMEPrefixAllowlist = []string{
		"audio/",
		"image/",
		"video/",
	}
	imageMIMEAllowlist = map[string]struct{}{
		"image/gif":     {},
		"image/jpeg":    {},
		"image/png":     {},
		"image/svg+xml": {},
		"image/tiff":    {},
		"image/webp":    {},
	}
)

// GenUniqueObjectKey returns a random string for use as object key.
func GenUniqueObjectKey() (string, error) {
	var s strings.Builder
	for i := 0; i < storedFileNameLength; i++ {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(storedFileNameChars))))
		if err != nil {
			return "", fmt.Errorf("generating random number: %w", err)
		}
		if !r.IsInt64() {
			return "", errors.New("random number is not an int64")
		}
		s.WriteByte(storedFileNameChars[r.Int64()])
	}
	return s.String(), nil
}

type SanitizedKey struct {
	UniqueID  string
	Extension string
}

func (s SanitizedKey) String() string {
	return s.UniqueID + s.Extension
}

func SanitizeUploadName(name string, forbiddenExtensions map[string]struct{}) (*SanitizedKey, error) {
	name = strings.ReplaceAll(name, string(filepath.Separator), "/")
	name = name[strings.LastIndex(name, "/")+1:]
	id, err := GenUniqueObjectKey()
	if err != nil {
		return nil, fmt.Errorf("generating unique object key: %w", err)
	}
	lowercaseName := strings.ToLower(name)
	for ext := range forbiddenExtensions {
		if strings.HasSuffix(lowercaseName, "."+ext) || strings.Contains(lowercaseName, "."+ext+".") {
			return nil, ErrForbiddenExtension
		}
	}
	return &SanitizedKey{
		UniqueID:  id,
		Extension: utils.HumanFileExtension(name),
	}, nil
}

func UploadObjects(
	ctx context.Context,
	logger logging.Logger,
	conf *config.Config,
	files []config.StoredFile,
	htmls []config.StoredHTML,
	metadata *uploadMetadata,
) []error {
	// TODO: Consider consolidating file uploads and HTML uploads somehow.
	metadataFile, err := metadata.StoredFile()
	if err != nil {
		return []error{fmt.Errorf("creating metadata file: %w", err)}
	}

	files = append(files, metadataFile)
	links := make([]*url.URL, 0, len(files)+len(htmls))
	for _, file := range files {
		links = append(links, conf.FileURL(file.Key()))
	}
	for _, html := range htmls {
		links = append(links, conf.HTMLURL(html.Key()))
	}
	for i := range files {
		files[i] = storage.UpdatedStoredFile(
			files[i],
			storage.WithMetadataURL(metadata.URL(conf)),
			storage.WithLinks(links),
		)
	}
	for i := range htmls {
		htmls[i] = storage.UpdatedStoredHTML(
			htmls[i],
			storage.WithMetadataURL(metadata.URL(conf)),
			storage.WithLinks(links),
		)
	}

	logger.Info(ctx, "uploading", "files", len(files), "htmls", len(htmls))

	results := make(chan error, len(files)+len(htmls))
	for i, file := range files {
		go func() {
			// TODO: add file_size once we have it easily accessible here.
			// TODO: provide way to get a sub-logger with these set?
			logParams := []interface{}{
				"file_key", file.Key(),
				"file_name", file.Name(),
				"file_mime_type", file.MIMEType(),
				"file_upload_index", fmt.Sprintf("%d/%d", i+1, len(files)),
			}

			logger.Info(ctx, "storing file", logParams...)
			err := conf.StorageBackend.StoreFile(ctx, file)
			if err != nil {
				logger.Error(ctx, "storing file", append(logParams, "error", err)...)
			} else {
				logger.Info(ctx, "successfully stored file", logParams...)
			}
			results <- err
		}()
	}
	for i, html := range htmls {
		go func() {
			// TODO: add html_size once we have it easily accessible here.
			// TODO: provide way to get a sub-logger with these set?
			logParams := []interface{}{
				"html_key", html.Key(),
				"html_upload_index", fmt.Sprintf("%d/%d", i+1, len(htmls)),
			}

			logger.Info(ctx, "storing HTML", logParams...)
			err := conf.StorageBackend.StoreHTML(ctx, html)
			if err != nil {
				logger.Error(ctx, "storing HTML", append(logParams, "error", err)...)
			} else {
				logger.Info(ctx, "successfully stored HTML", logParams...)
			}
			results <- err
		}()
	}

	errs := make([]error, 0, len(files)+len(htmls))
	for i := 0; i < len(files)+len(htmls); i++ {
		select {
		case err := <-results:
			if err != nil {
				logger.Error(ctx, "storing object", "error", err)
				errs = append(errs, err)
			}
		case <-ctx.Done():
			logger.Error(ctx, "context done while storing objects", "ctx.Err", ctx.Err())
			return []error{ctx.Err()}
		}
	}

	// TODO: unique request ID to tie these together?
	if len(errs) == 0 {
		logger.Info(ctx, "sucessfully uploaded", "files", len(files), "htmls", len(htmls))
	} else {
		logger.Error(ctx, "upload failed with errors",
			"success_count", len(files)+len(htmls)-len(errs),
			"error_count", len(errs),
			"files", len(files),
			"htmls", len(htmls),
		)
	}

	return errs
}

func calculateTextChars() map[byte]struct{} {
	ret := make(map[byte]struct{})
	for i := 7; i <= 13; i++ {
		ret[byte(i)] = struct{}{}
	}
	for i := 0x20; i < 0x7F; i++ {
		ret[byte(i)] = struct{}{}
	}
	for i := 0x80; i < 0x100; i++ {
		ret[byte(i)] = struct{}{}
	}
	return ret
}

var textChars map[byte]struct{} = calculateTextChars()

// ProbablyText returns whether the first KB of the reader seems to be text.
//
// This is roughly based on libmagic's binary/text detection:
// https://github.com/file/file/blob/df74b09b9027676088c797528edcaae5a9ce9ad0/src/encoding.c#L203-L228
func ProbablyText(reader io.ReadSeeker) (isText bool, err error) {
	cur, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, fmt.Errorf("seeking: %w", err)
	}
	defer func() {
		if _, seekErr := reader.Seek(cur, io.SeekStart); err == nil && seekErr != nil {
			err = fmt.Errorf("seeking back: %w", seekErr)
		}
	}()
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("reading: %w", err)
	}
	for _, b := range buf[:n] {
		if _, ok := textChars[b]; !ok {
			return false, nil
		}
	}
	return true, nil
}

type SanitizedMIMEType string

func isAllowedMIMEType(mimeType string) bool {
	if _, ok := mimeAllowlist[mimeType]; ok {
		return true
	}
	for _, prefix := range mimePrefixAllowlist {
		if strings.HasPrefix(string(mimeType), prefix) {
			return true
		}
	}
	return false
}

func DetermineMIMEType(filename string, contentType string, probablyText bool) string {
	// Prefer the Content-Type from the multipart form if it's set to something non-generic (and
	// allowed).
	if contentType != "" && contentType != mimeGenericBinary && contentType != mimeGenericText && isAllowedMIMEType(contentType) {
		return contentType
	}

	if ext := filepath.Ext(filename); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			if isAllowedMIMEType(mimeType) {
				return mimeType
			}
		}
	}

	if probablyText {
		return mimeGenericText
	} else {
		return mimeGenericBinary
	}
}

func isInlineDisplayMIME(mimeType string) bool {
	if _, ok := inlineDisplayMIMEAllowlist[mimeType]; ok {
		return true
	}
	for _, prefix := range inlineDisplayMIMEPrefixAllowlist {
		if strings.HasPrefix(mimeType, prefix) {
			return true
		}
	}
	return false
}

func IsImageMIME(mimeType string) bool {
	_, ok := imageMIMEAllowlist[mimeType]
	return ok
}

func DetermineContentDisposition(filename string, mimeType string, probablyText bool) string {
	renderType := "attachment"
	if probablyText || isInlineDisplayMIME(mimeType) {
		renderType = "inline"
	}
	return fmt.Sprintf(`%s; filename="%s"; filename*=utf-8''%s`,
		renderType,
		strings.ReplaceAll(filename, `"`, ""),
		url.PathEscape(filename),
	)
}

type uploadType string

const (
	UploadTypeFile uploadType = "file"
	// TODO: add UploadTypePaste once paste support is added.
)

type UploadedFile struct {
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
	Raw   string `json:"raw"`
	Paste string `json:"paste,omitempty"`
}

type UploadMetadataFile struct {
	ServerVersion string         `json:"server_version"`
	Timestamp     int64          `json:"timestamp"`
	UploadType    uploadType     `json:"upload_type"`
	UploadedFiles []UploadedFile `json:"uploaded_files"`
}

type uploadMetadata struct {
	key  string
	File UploadMetadataFile
	// TODO: add PasteDetails once paste support is added.
}

func (m *uploadMetadata) URL(conf *config.Config) *url.URL {
	return conf.FileURL(m.key)
}

func (m *uploadMetadata) StoredFile() (config.StoredFile, error) {
	var metadataJSON bytes.Buffer
	if err := json.NewEncoder(&metadataJSON).Encode(m); err != nil {
		return nil, fmt.Errorf("encoding metadata JSON: %w", err)
	}
	return storage.NewStoredFile(
		utils.NopReadSeekCloser(bytes.NewReader(metadataJSON.Bytes())),
		storage.WithKey(m.key),
		storage.WithMIMEType("application/json"),
	), nil
}

func NewUploadMetadata(conf *config.Config, files []config.StoredFile) (*uploadMetadata, error) {
	// TODO: probably make this same function work for pastes with additional arguments.
	key, err := GenUniqueObjectKey()
	if err != nil {
		return nil, fmt.Errorf("generating unique object key: %w", err)
	}
	ret := uploadMetadata{
		key: key + ".json",
		File: UploadMetadataFile{
			ServerVersion: conf.Version,
			Timestamp:     time.Now().Unix(),
			// TODO: set this to UploadTypePaste once paste support is added.
			UploadType:    UploadTypeFile,
			UploadedFiles: make([]UploadedFile, 0, len(files)),
		},
	}
	for _, file := range files {
		// TODO: consider calculating this once and storing it, since it's used in multiple places
		// and it keeps forcing introduction of "impossible" errors in return types that must
		// nevertheless be handled.
		bytes, err := utils.FileSizeBytes(file)
		if err != nil {
			return nil, fmt.Errorf("getting file size: %w", err)
		}
		ret.File.UploadedFiles = append(ret.File.UploadedFiles, UploadedFile{
			Name:  file.Name(),
			Bytes: bytes,
			Raw:   conf.FileURL(file.Key()).String(),
		})
	}
	return &ret, nil

}

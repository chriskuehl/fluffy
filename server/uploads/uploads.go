package uploads

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"mime"
	"path/filepath"
	"strings"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
)

const (
	storedFileNameLength = 32
	storedFileNameChars  = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789"

	mimeGenericBinary = "application/octet-stream"
	mimeGenericText   = "text/plain"
)

var (
	ErrForbiddenExtension = fmt.Errorf("forbidden extension")

	// Extensions that traditionally wrap another file extension.
	wrapperExtensions = map[string]struct{}{
		"bz2": {},
		"gz":  {},
		"xz":  {},
		"zst": {},
	}

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
)

func GenUniqueObjectID() (string, error) {
	var s strings.Builder
	for i := 0; i < storedFileNameLength; i++ {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(storedFileNameChars))))
		if err != nil {
			return "", fmt.Errorf("generating random number: %w", err)
		}
		if !r.IsInt64() {
			return "", fmt.Errorf("random number is not an int64")
		}
		s.WriteByte(storedFileNameChars[r.Int64()])
	}
	return s.String(), nil
}

func extractExtension(name string) string {
	fullExt := ""
	for strings.Contains(name, ".") {
		ext := filepath.Ext(name)
		name = strings.TrimSuffix(name, ext)
		if ext == "." {
			// Don't add ".", but keep processing any additional extensions.
			continue
		}
		fullExt = ext + fullExt
		if _, ok := wrapperExtensions[strings.TrimPrefix(ext, ".")]; !ok {
			return fullExt
		}
	}
	return fullExt
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
	id, err := GenUniqueObjectID()
	if err != nil {
		return nil, fmt.Errorf("generating unique object ID: %w", err)
	}
	ext := extractExtension(name)
	for _, extPart := range strings.Split(ext, ".") {
		if _, ok := forbiddenExtensions[extPart]; ok {
			return nil, ErrForbiddenExtension
		}
	}
	return &SanitizedKey{
		UniqueID:  id,
		Extension: ext,
	}, nil
}

func UploadObjects(
	ctx context.Context,
	logger logging.Logger,
	conf *config.Config,
	objs []config.StoredObject,
) []error {
	results := make(chan error, len(objs))
	for _, obj := range objs {
		go func() {
			err := conf.StorageBackend.StoreObject(ctx, obj)
			if err != nil {
				logger.Error(ctx, "storing object", "obj", obj, "error", err)
			} else {
				logger.Info(ctx, "successfully stored object", "obj", obj)
			}
			results <- err
		}()
	}

	errs := make([]error, 0, len(objs))
	for i := 0; i < len(objs); i++ {
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
	if err != nil {
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

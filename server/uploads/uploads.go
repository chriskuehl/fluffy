package uploads

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/logging"
	"github.com/chriskuehl/fluffy/server/storage/storagedata"
)

const (
	storedFileNameLength = 32
	storedFileNameChars  = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789"
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
)

func genUniqueObjectID() (string, error) {
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
	id, err := genUniqueObjectID()
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
	config *config.Config,
	objs []storagedata.Object,
) []error {
	results := make(chan error, len(objs))
	for _, obj := range objs {
		go func() {
			err := config.StorageBackend.StoreObject(ctx, obj)
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

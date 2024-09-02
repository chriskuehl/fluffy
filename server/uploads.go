package server

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"path/filepath"
	"strings"
)

const (
	storedFileNameLength = 32
	storedFileNameChars  = "bcdfghjklmnpqrstvwxzBCDFGHJKLMNPQRSTVWXZ0123456789"
)

// Extensions that traditionally wrap another file extension.
var wrapperExtensions = map[string]struct{}{
	"bz2": {},
	"gz":  {},
	"xz":  {},
	"zst": {},
}

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
		fullExt = ext + fullExt
		if _, ok := wrapperExtensions[strings.TrimPrefix(ext, ".")]; !ok {
			return fullExt
		}
	}
	return fullExt
}

package utils

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Extensions that traditionally wrap another file extension.
var wrapperExtensions = map[string]struct{}{
	"bz2": {},
	"gz":  {},
	"xz":  {},
	"zst": {},
}

// HumanFileExtension returns the "human" file extension of a file name.
//
// This function tries to mimic what a human would consider the file extension rather than always
// returning just the last extension. For example, "file.tar.gz" would return ".tar.gz" instead of
// just ".gz".
//
// Files with no extension will return an empty string.
//
// This function should not be used for any kind of validation purposes.
func HumanFileExtension(filename string) string {
	fullExt := ""
	for strings.Contains(filename, ".") {
		ext := filepath.Ext(filename)
		filename = strings.TrimSuffix(filename, ext)
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

func Pluralize(s string, n int64) string {
	if n == 1 {
		return s
	}
	return s + "s"
}

func FormatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case bytes >= gb:
		// Note: Using this pattern instead of printf with %.1f to avoid rounding up which can lead
		// to weird results like 1024.0 MiB instead of 1.0 GiB.
		n, rem := bytes/gb, bytes%gb
		return fmt.Sprintf("%d.%d GiB", n, rem*10/gb)
	case bytes >= mb:
		n, rem := bytes/mb, bytes%mb
		return fmt.Sprintf("%d.%d MiB", n, rem*10/mb)
	case bytes >= kb:
		n, rem := bytes/kb, bytes%kb
		return fmt.Sprintf("%d.%d KiB", n, rem*10/kb)
	default:
		return fmt.Sprintf("%d %s", bytes, Pluralize("byte", bytes))
	}
}

// Convert a ReadSeeker to a ReadSeekCloser with a no-op Close method.
//
// Like io.NopCloser, but for ReadSeeker instead of just Reader.
func NopReadSeekCloser(r io.ReadSeeker) io.ReadSeekCloser {
	return nopReadSeekCloser{r}
}

type nopReadSeekCloser struct {
	io.ReadSeeker
}

func (nopReadSeekCloser) Close() error { return nil }

func FileSizeBytes(r io.ReadSeeker) (int64, error) {
	prev, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, fmt.Errorf("seeking to current position: %w", err)
	}
	_, err = r.Seek(0, io.SeekStart)
	if err != nil {
		return 0, fmt.Errorf("seeking to start: %w", err)
	}
	end, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, fmt.Errorf("seeking to current position: %w", err)
	}
	if _, err := r.Seek(prev, io.SeekStart); err != nil {
		return 0, fmt.Errorf("seeking back to original position: %w", err)
	}
	return end, nil
}

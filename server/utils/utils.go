package utils

import (
	"fmt"
	"io"
)

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
	end, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, fmt.Errorf("seeking to current position: %w", err)
	}
	if _, err := r.Seek(prev, io.SeekStart); err != nil {
		return 0, fmt.Errorf("seeking back to original position: %w", err)
	}
	return end, nil
}

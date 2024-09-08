package utils

import (
	"fmt"
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

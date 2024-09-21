package uploads

import (
	"testing"
)

func TestGetUniqueObjectKey(t *testing.T) {
	got, err := GenUniqueObjectKey()
	if err != nil {
		t.Fatalf("GenUniqueObjectKey() error = %v", err)
	}
	wantLength := 32
	if len(got) != wantLength {
		t.Fatalf("got GenUniqueObjectKey() = %q, want len() = %d", got, wantLength)
	}
}

func TestIsAllowedMIMEType(t *testing.T) {
	tests := map[string]bool{
		"application/javascript": true,
		"audio/mpeg":             true,
		"font/woff":              false,
		"image/png":              true,
		"text/html":              false,
		"text/plain":             true,
	}
	for mimeType, want := range tests {
		t.Run(mimeType, func(t *testing.T) {
			if got := isAllowedMIMEType(mimeType); got != want {
				t.Errorf("got isAllowedMIMEType(%q) = %t, want %t", mimeType, got, want)
			}
		})
	}
}

func TestIsInlineDisplayMIME(t *testing.T) {
	tests := map[string]bool{
		"application/pdf":  true,
		"application/json": false,
		"image/png":        true,
		"text/plain":       false,
	}
	for mimeType, want := range tests {
		t.Run(mimeType, func(t *testing.T) {
			if got := isInlineDisplayMIME(mimeType); got != want {
				t.Errorf("got isInlineDisplayMIME(%q) = %t, want %t", mimeType, got, want)
			}
		})
	}
}

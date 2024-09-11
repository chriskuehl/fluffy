package uploads

import (
	"testing"
)

func TestGetUniqueObjectID(t *testing.T) {
	got, err := GenUniqueObjectID()
	if err != nil {
		t.Fatalf("genUniqueObjectID() error = %v", err)
	}
	wantLength := 32
	if len(got) != wantLength {
		t.Fatalf("got genUniqueObjectID() = %q, want len() = %d", got, wantLength)
	}
}

func TestExtractExtension(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{
			name: "no extension",
			in:   "file",
			want: "",
		},
		{
			name: "regular extension",
			in:   "file.txt",
			want: ".txt",
		},
		{
			name: "wrapped extension only",
			in:   "file.gz",
			want: ".gz",
		},
		{
			name: "wrapped extension after regular extension",
			in:   "file.tar.gz",
			want: ".tar.gz",
		},
		{
			name: "multiple wrapped extensions",
			in:   "file.tar.gz.bz2",
			want: ".tar.gz.bz2",
		},
		{
			name: "multiple wrapped extensions with a regular extension",
			in:   "file.txt.tar.gz.bz2",
			want: ".tar.gz.bz2",
		},
		{
			// Kind of nonsense, just making sure it doesn't remove more than it should.
			name: "wrapped extensions before regular extension",
			in:   "file.tar.gz.txt",
			want: ".txt",
		},
		{
			name: ". only",
			in:   ".",
			want: "",
		},
		{
			name: "multiple wrapped extensions with empty extensions",
			in:   "file.txt.tar.gz....bz2",
			want: ".tar.gz.bz2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractExtension(tt.in); got != tt.want {
				t.Errorf("got extractExtension(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
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

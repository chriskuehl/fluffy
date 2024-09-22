package uploads_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/server/uploads"
	"github.com/chriskuehl/fluffy/server/utils"
	"github.com/chriskuehl/fluffy/testfunc"
)

func TestSanitizeUploadName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *uploads.SanitizedKey
		wantErr error
	}{
		{
			name: "simple name",
			in:   "file.txt",
			want: &uploads.SanitizedKey{
				Extension: ".txt",
			},
		},
		{
			name: "dangerous name",
			in:   "path/to/../../etc/resolv.conf",
			want: &uploads.SanitizedKey{
				Extension: ".conf",
			},
		},
		{
			name: "dangerous extension",
			in:   "resolv.conf/../../etc/passwd",
			want: &uploads.SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "dangerous name with windows path separators",
			in:   "path\\to\\..\\..\\etc\\resolv.conf",
			want: &uploads.SanitizedKey{
				Extension: ".conf",
			},
		},
		{
			name: "dangerous extension with windows path separators",
			in:   "resolv.conf\\..\\..\\etc/passwd",
			want: &uploads.SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "empty name",
			in:   "",
			want: &uploads.SanitizedKey{
				Extension: "",
			},
		},
		{
			name: ".. only",
			in:   "..",
			want: &uploads.SanitizedKey{
				Extension: "",
			},
		},
		{
			name: ". only",
			in:   ".",
			want: &uploads.SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "/ only",
			in:   "/",
			want: &uploads.SanitizedKey{
				Extension: "",
			},
		},
		{
			name: "/../ only",
			in:   "/../",
			want: &uploads.SanitizedKey{
				Extension: "",
			},
		},
		{
			name:    "forbidden extension",
			in:      "file.exe",
			wantErr: uploads.ErrForbiddenExtension,
		},
		{
			name:    "forbidden extension with caps",
			in:      "file.EXE",
			wantErr: uploads.ErrForbiddenExtension,
		},
		{
			name:    "forbidden extension before wrapped extension",
			in:      "file.exe.gz",
			wantErr: uploads.ErrForbiddenExtension,
		},
		{
			name:    "forbidden extension before wrapped extension with ..",
			in:      "file.Exe..gz",
			wantErr: uploads.ErrForbiddenExtension,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := uploads.SanitizeUploadName(
				tt.in,
				map[string]struct{}{
					"exe": {},
				},
			)
			if err != tt.wantErr {
				t.Fatalf("sanitizeUploadName(%q) error = %v, want %v", tt.in, err, tt.wantErr)
			}
			if tt.want != nil && got != nil {
				tt.want.UniqueID = got.UniqueID
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("sanitizeUploadName(%q) mismatch (-want +got):\n%s", tt.in, diff)
			}
		})
	}
}

func TestUploadObjects(t *testing.T) {
	t.Parallel()
	logger := testfunc.NewMemoryLogger()
	storageBackend := testfunc.NewMemoryStorageBackend()
	conf := testfunc.NewConfig(testfunc.WithStorageBackend(storageBackend))
	files := []config.StoredFile{
		storage.NewStoredFile(
			utils.NopReadSeekCloser(bytes.NewReader([]byte("hello, world"))),
			storage.WithKey("file.txt"),
			storage.WithName("file.txt"),
		),
	}

	metadata, err := uploads.NewUploadMetadata(conf, files)
	if err != nil {
		t.Fatalf("generating metadata: %s", err)
	}

	errs := uploads.UploadObjects(
		context.Background(),
		logger,
		conf,
		files,
		[]config.StoredHTML{
			storage.NewStoredHTML(
				utils.NopReadSeekCloser(strings.NewReader("<html>hello, world</html>")),
				storage.WithKey("file.html"),
			),
		},
		metadata,
	)

	if len(errs) != 0 {
		t.Fatalf("UploadObjects() = %v, want no errors", errs)
	}

	file, ok := storageBackend.Files["file.txt"]
	if !ok {
		t.Fatalf("file not stored")
	}

	fileBuf := new(strings.Builder)
	if _, err := io.Copy(fileBuf, file); err != nil {
		t.Fatalf("reading stored file: %v", err)
	}
	got := fileBuf.String()
	want := "hello, world"
	if got != want {
		t.Fatalf("stored file = %q, want %q", got, want)
	}

	html, ok := storageBackend.HTMLs["file.html"]
	if !ok {
		t.Fatalf("HTML not stored")
	}

	htmlBuf := new(strings.Builder)
	if _, err := io.Copy(htmlBuf, html); err != nil {
		t.Fatalf("reading stored HTML: %v", err)
	}
	got = htmlBuf.String()
	want = "<html>hello, world</html>"
	if got != want {
		t.Fatalf("stored HTML = %q, want %q", got, want)
	}
}

func TestProbablyText(t *testing.T) {
	// Adopted from https://github.com/pre-commit/identify/blob/52ba50e2a234147d85320b6e1cff065b30377020/tests/identify_test.py#L216
	tests := map[string]bool{
		"hello world": true,
		"":            true,
		"éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ": true,
		`¯\_(ツ)_/¯`: true,
		"♪┏(・o･)┛♪┗ ( ･o･) ┓♪┏ ( ) ┛♪┗ (･o･ ) ┓♪": true,
		"\xe9\xf3\xf1\xe5": true, // "éóñå".encode('latin1')

		"hello world\x00":              false,
		"\x7f\x45\x4c\x46\x02\x01\x01": false, // first few bytes of /bin/bash
		"\x43\x92\xd9\x0f\xaf\x32\x2c": false, // some /dev/urandom output
	}
	for in, want := range tests {
		t.Run(in, func(t *testing.T) {
			got, err := uploads.ProbablyText(strings.NewReader(in))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != want {
				t.Fatalf("ProbablyText(%q) = %v, want %v", in, got, want)
			}
		})
	}
}

func TestDetermineMIMEType(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		contentType  string
		probablyText bool
		want         string
	}{
		{
			name:         "all empty metadata and text",
			filename:     "",
			contentType:  "",
			probablyText: true,
			want:         "text/plain",
		},
		{
			name:         "all empty metadata and binary",
			filename:     "",
			contentType:  "",
			probablyText: false,
			want:         "application/octet-stream",
		},
		{
			name:         "prefers contentType over filename",
			filename:     "image.png",
			contentType:  "application/json",
			probablyText: true,
			want:         "application/json",
		},
		{
			name:         "prefers filename over contentType if contentType disallowed",
			filename:     "image.png",
			contentType:  "text/html",
			probablyText: true,
			want:         "image/png",
		},
		{
			name:         "prefers filename over contentType if contentType missing",
			filename:     "image.png",
			contentType:  "",
			probablyText: true,
			want:         "image/png",
		},
		{
			name:         "ignores filename and contentType if both disallowed",
			filename:     "index.html",
			contentType:  "text/html",
			probablyText: true,
			want:         "text/plain",
		},
		{
			name:         "unrecognized file extension",
			filename:     "file.whatisthisextension",
			contentType:  "",
			probablyText: true,
			want:         "text/plain",
		},
		{
			name:         "no file extension",
			filename:     "file",
			contentType:  "",
			probablyText: true,
			want:         "text/plain",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uploads.DetermineMIMEType(tt.filename, tt.contentType, tt.probablyText); got != tt.want {
				t.Errorf("got determineMIMEType(%q, %q, %t) = %q, want %q", tt.filename, tt.contentType, tt.probablyText, got, tt.want)
			}
		})
	}
}

func TestDetermineContentDisposition(t *testing.T) {
	tests := []struct {
		name         string
		filename     string
		mimeType     string
		probablyText bool
		want         string
	}{
		{
			name:         "text file",
			filename:     "file.txt",
			mimeType:     "text/plain",
			probablyText: true,
			want:         `inline; filename="file.txt"; filename*=utf-8''file.txt`,
		},
		{
			name:         "binary file with inline mime",
			filename:     "image.png",
			mimeType:     "image/png",
			probablyText: false,
			want:         `inline; filename="image.png"; filename*=utf-8''image.png`,
		},
		{
			name:         "binary file with random mime",
			filename:     "file",
			mimeType:     "application/octet-stream",
			probablyText: false,
			want:         `attachment; filename="file"; filename*=utf-8''file`,
		},
		{
			name:         "special characters",
			filename:     "file with spaces and éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ",
			mimeType:     "text/plain",
			probablyText: true,
			want:         `inline; filename="file with spaces and éóñəå  ⊂(◉‿◉)つ(ノ≥∇≤)ノ"; filename*=utf-8''file%20with%20spaces%20and%20%C3%A9%C3%B3%C3%B1%C9%99%C3%A5%20%20%E2%8A%82%28%E2%97%89%E2%80%BF%E2%97%89%29%E3%81%A4%28%E3%83%8E%E2%89%A5%E2%88%87%E2%89%A4%29%E3%83%8E`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uploads.DetermineContentDisposition(tt.filename, tt.mimeType, tt.probablyText); got != tt.want {
				t.Errorf("got determineContentDisposition(%q, %q, %t) = %q, want %q", tt.filename, tt.mimeType, tt.probablyText, got, tt.want)
			}
		})
	}
}

func TestNewUploadMetadata(t *testing.T) {
	tests := []struct {
		name  string
		files []config.StoredFile
		want  uploads.UploadMetadataFile
	}{
		{
			name: "several_files",
			files: []config.StoredFile{
				storage.NewStoredFile(
					utils.NopReadSeekCloser(bytes.NewReader([]byte("abcd"))),
					storage.WithKey("aaaaaa"),
					storage.WithName("file"),
				),
				storage.NewStoredFile(
					utils.NopReadSeekCloser(bytes.NewReader([]byte("abcd"))),
					storage.WithKey("bbbbbb.png"),
					storage.WithName("image.png"),
				),
				storage.NewStoredFile(
					utils.NopReadSeekCloser(bytes.NewReader([]byte("abcd"))),
					storage.WithKey("cccccc.txt"),
					storage.WithName("text.txt"),
				),
			},
			want: uploads.UploadMetadataFile{
				ServerVersion: "(test)",
				UploadType:    uploads.UploadTypeFile,
				UploadedFiles: []uploads.UploadedFile{
					{
						Name:  "file",
						Bytes: 4,
						Raw:   "http://localhost:8080/dev/storage/file/aaaaaa",
					},
					{
						Name:  "image.png",
						Bytes: 4,
						Raw:   "http://localhost:8080/dev/storage/file/bbbbbb.png",
					},
					{
						Name:  "text.txt",
						Bytes: 4,
						Raw:   "http://localhost:8080/dev/storage/file/cccccc.txt",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conf := testfunc.NewConfig()
			metadata, err := uploads.NewUploadMetadata(conf, tt.files)
			if err != nil {
				t.Fatalf("newUploadMetadata() error = %v", err)
			}
			got := metadata.File
			tt.want.Timestamp = got.Timestamp
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("newUploadMetadata() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

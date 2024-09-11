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
			name:    "forbidden extension before wrapped extension",
			in:      "file.exe.gz",
			wantErr: uploads.ErrForbiddenExtension,
		},
		{
			name:    "forbidden extension before wrapped extension with ..",
			in:      "file.exe..gz",
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
	logger := testfunc.NewMemoryLogger()
	storageBackend := testfunc.NewMemoryStorageBackend()

	errs := uploads.UploadObjects(
		context.Background(),
		logger,
		testfunc.NewConfig(
			testfunc.WithStorageBackend(storageBackend),
		),
		[]config.StoredObject{
			&storage.StoredObject{
				BaseStoredObject: storage.BaseStoredObject{
					ObjKey:        "file.txt",
					ObjReadCloser: io.NopCloser(bytes.NewReader([]byte("hello, world"))),
				},
			},
		},
	)

	if len(errs) != 0 {
		t.Fatalf("UploadObjects() = %v, want no errors", errs)
	}

	obj, ok := storageBackend.Objects["file.txt"]
	if !ok {
		t.Fatalf("Object not stored")
	}

	buf := new(strings.Builder)
	readCloser, err := obj.ReadCloser()
	if err != nil {
		t.Fatalf("getting read closer: %v", err)
	}
	defer readCloser.Close()
	_, err = io.Copy(buf, readCloser)
	if err != nil {
		t.Fatalf("reading stored object: %v", err)
	}
	got := buf.String()
	want := "hello, world"
	if got != want {
		t.Fatalf("stored object = %q, want %q", got, want)
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

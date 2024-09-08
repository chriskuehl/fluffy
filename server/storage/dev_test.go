package storage_test

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/testfunc"
)

func TestDevStorageDev(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		wantContentType string
		wantContent     string
	}{
		{
			name:            "object",
			url:             "http://localhost:%d/dev/storage/object/test.txt",
			wantContentType: "text/plain; charset=utf-8",
			wantContent:     "test content\n",
		},
		{
			name:            "html",
			url:             "http://localhost:%d/dev/storage/html/test.html",
			wantContentType: "text/html; charset=utf-8",
			wantContent:     "<html>test content</html>\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()

			objectRoot := filepath.Join(tmp, "object")
			if err := os.MkdirAll(objectRoot, 0755); err != nil {
				t.Fatalf("failed to create object root: %v", err)
			}
			if err := os.WriteFile(filepath.Join(objectRoot, "test.txt"), []byte("test content\n"), 0644); err != nil {
				t.Fatalf("failed to write object file: %v", err)
			}

			htmlRoot := filepath.Join(tmp, "html")
			if err := os.MkdirAll(htmlRoot, 0755); err != nil {
				t.Fatalf("failed to create html root: %v", err)
			}
			if err := os.WriteFile(filepath.Join(htmlRoot, "test.html"), []byte("<html>test content</html>\n"), 0644); err != nil {
				t.Fatalf("failed to write html file: %v", err)
			}

			conf := testfunc.NewConfig()
			conf.DevMode = true
			conf.StorageBackend = &storage.FilesystemBackend{
				ObjectRoot: objectRoot,
				HTMLRoot:   htmlRoot,
			}
			ts := testfunc.RunningServer(t, conf)
			defer ts.Cleanup()

			resp, err := http.Get(fmt.Sprintf(tt.url, ts.Port))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if resp.Header.Get("Content-Type") != tt.wantContentType {
				t.Fatalf("unexpected content type: got %q, want %q", resp.Header.Get("Content-Type"), tt.wantContentType)
			}
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			body := string(bodyBytes)
			if diff := cmp.Diff(tt.wantContent, body); diff != "" {
				t.Fatalf("unexpected body (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDevStorageProd(t *testing.T) {
	urls := map[string]string{
		"object": "http://localhost:%d/dev/storage/object/test.txt",
		"html":   "http://localhost:%d/dev/storage/html/test.html",
	}
	for name, url := range urls {
		t.Run(name, func(t *testing.T) {
			ts := testfunc.RunningServer(t, testfunc.NewConfig())
			defer ts.Cleanup()

			resp, err := http.Get(fmt.Sprintf(url, ts.Port))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
			}
		})
	}
}

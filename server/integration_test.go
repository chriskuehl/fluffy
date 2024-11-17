package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/chriskuehl/fluffy/testfunc"
)

type objectType int

const (
	objectTypeFile objectType = iota
	objectTypeHTML objectType = iota
)

func TestIntegrationUpload(t *testing.T) {
	for _, tt := range testfunc.StorageBackends {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			storage := tt.StorageFactory(t)
			conf := testfunc.NewConfig(
				testfunc.WithStorageBackend(storage.Backend),
			)
			ts := testfunc.RunningServer(t, conf)
			defer ts.Cleanup()

			postBody := new(bytes.Buffer)
			writer := multipart.NewWriter(postBody)
			part, err := writer.CreateFormFile("file", "test.txt")
			if err != nil {
				t.Fatalf("creating form file: %v", err)
			}
			if _, err = part.Write([]byte("test")); err != nil {
				t.Fatalf("writing to form file: %v", err)
			}
			if err := writer.Close(); err != nil {
				t.Fatalf("closing writer: %v", err)
			}

			resp, err := http.Post(
				fmt.Sprintf("http://localhost:%d/upload?json", ts.Port),
				writer.FormDataContentType(),
				postBody,
			)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()

			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			body := string(bodyBytes)

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status code: got %d, want %d\nBody:\n%s", resp.StatusCode, http.StatusOK, body)
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Fatalf("unexpected content type: got %s, want application/json", resp.Header.Get("Content-Type"))
			}

			var result struct {
				Success       bool   `json:"success"`
				Metadata      string `json:"metadata"`
				Redirect      string `json:"redirect"`
				UploadedFiles map[string]struct {
					// TODO: verify the paste by reading the "paste" key here once paste support is
					// added.
					Raw string `json:"raw"`
				} `json:"uploaded_files"`
			}
			if err := json.Unmarshal(bodyBytes, &result); err != nil {
				t.Fatalf("unmarshaling error response: %v", err)
			}

			if !result.Success {
				t.Fatalf("unexpected success: got %v, want true", result.Success)
			}

			wantLenUploadedFiles := 1
			if len(result.UploadedFiles) != wantLenUploadedFiles {
				t.Fatalf(
					"unexpected number of uploaded files: got %d, want %d",
					len(result.UploadedFiles),
					wantLenUploadedFiles,
				)
			}

			uploadDetailsURL, err := url.ParseRequestURI(result.Redirect)
			if err != nil {
				t.Fatalf("parsing redirect URL: %v", err)
			}

			rawURL, err := url.ParseRequestURI(result.UploadedFiles["test.txt"].Raw)
			if err != nil {
				t.Fatalf("parsing raw URL: %v", err)
			}
			metadataURL, err := url.ParseRequestURI(result.Metadata)
			if err != nil {
				t.Fatalf("parsing metadata URL: %v", err)
			}

			links := []*url.URL{rawURL, metadataURL, uploadDetailsURL}

			storage.AssertFile(t, testfunc.KeyFromURL(rawURL.String()), &testfunc.StoredObject{
				Content:            "test",
				MIMEType:           "text/plain",
				ContentDisposition: `inline; filename="test.txt"; filename*=utf-8''test.txt`,
				Links:              testfunc.CanonicalizeLinks(links),
				MetadataURL:        metadataURL.String(),
			})

			uploadDetails := storage.AssertHTML(t, testfunc.KeyFromURL(uploadDetailsURL.String()), &testfunc.StoredObject{
				Content:            testfunc.DoNotCompareContentSentinel,
				MIMEType:           "text/html; charset=utf-8",
				ContentDisposition: "inline",
				Links:              testfunc.CanonicalizeLinks(links),
				MetadataURL:        metadataURL.String(),
			})
			if !strings.Contains(uploadDetails.Content, "<html") {
				t.Fatalf("upload details missing <html> tag:\n%s", uploadDetails.Content)
			}
		})
	}
}

package views_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/chriskuehl/fluffy/server/uploads"
	"github.com/chriskuehl/fluffy/testfunc"
)

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

var httpClientNoRedirects = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func addFile(t *testing.T, writer *multipart.Writer, filename, content string) {
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("creating form file: %v", err)
	}
	if _, err = part.Write([]byte(content)); err != nil {
		t.Fatalf("writing to form file: %v", err)
	}
}

func TestUpload(t *testing.T) {
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
			addFile(t, writer, "test.txt", "test\n")
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

			rawURL, err := url.ParseRequestURI(result.UploadedFiles["test.txt"].Raw)
			if err != nil {
				t.Fatalf("parsing raw URL: %v", err)
			}
			uploadDetailsURL, err := url.ParseRequestURI(result.Redirect)
			if err != nil {
				t.Fatalf("parsing redirect URL: %v", err)
			}
			metadataURL, err := url.ParseRequestURI(result.Metadata)
			if err != nil {
				t.Fatalf("parsing metadata URL: %v", err)
			}

			links := []*url.URL{rawURL, uploadDetailsURL, metadataURL}

			// Raw file
			storage.AssertFile(t, testfunc.KeyFromURL(rawURL.String()), &testfunc.StoredObject{
				Content:            "test\n",
				MIMEType:           "text/plain",
				ContentDisposition: `inline; filename="test.txt"; filename*=utf-8''test.txt`,
				Links:              testfunc.CanonicalizeLinks(links),
				MetadataURL:        metadataURL.String(),
			})

			// Upload details
			uploadDetails := storage.AssertHTML(t, testfunc.KeyFromURL(uploadDetailsURL.String()), &testfunc.StoredObject{
				Content:            testfunc.DoNotCompareContentSentinel,
				MIMEType:           "text/html; charset=utf-8",
				ContentDisposition: "inline",
				Links:              testfunc.CanonicalizeLinks(links),
				MetadataURL:        metadataURL.String(),
			})

			parsed, err := testfunc.ParseUploadDetails(uploadDetails.Content)
			if err != nil {
				t.Fatalf("parsing upload details: %v", err)
			}

			pf := parsed.Files["test.txt"]

			want := &testfunc.ParsedUploadDetailsFile{
				Name:              "test.txt",
				Icon:              "txt.png",
				Size:              "5 bytes",
				DirectLinkFileKey: testfunc.KeyFromURL(rawURL.String()),
				PasteLinkHTMLKey:  "TODO_PASTE_URL",
			}
			if diff := cmp.Diff(want, pf); diff != "" {
				t.Fatalf("unexpected upload details entry (-want +got):\n%s", diff)
			}

			// Metadata
			metadata := storage.AssertFile(t, testfunc.KeyFromURL(metadataURL.String()), &testfunc.StoredObject{
				Content:            testfunc.DoNotCompareContentSentinel,
				MIMEType:           "application/json",
				ContentDisposition: "",
				Links:              testfunc.CanonicalizeLinks(links),
				MetadataURL:        metadataURL.String(),
			})

			var gotMetadata uploads.UploadMetadataFile
			if err := json.Unmarshal([]byte(metadata.Content), &gotMetadata); err != nil {
				t.Fatalf("unmarshaling metadata: %v", err)
			}

			wantMetadata := uploads.UploadMetadataFile{
				ServerVersion: conf.Version,
				Timestamp:     gotMetadata.Timestamp,
				UploadType:    "file",
				UploadedFiles: []uploads.UploadedFile{
					{
						Name:  "test.txt",
						Bytes: 5,
						Raw:   rawURL.String(),
					},
				},
			}
			if diff := cmp.Diff(wantMetadata, gotMetadata); diff != "" {
				t.Fatalf("unexpected metadata (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUploadNoJSON(t *testing.T) {
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
			addFile(t, writer, "test.txt", "test\n")
			if err := writer.Close(); err != nil {
				t.Fatalf("closing writer: %v", err)
			}

			resp, err := httpClientNoRedirects.Post(
				fmt.Sprintf("http://localhost:%d/upload", ts.Port),
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

			if resp.StatusCode != http.StatusSeeOther {
				t.Fatalf("unexpected status code: got %d, want %d\nBody:\n%s", resp.StatusCode, http.StatusSeeOther, body)
			}

			redirect := resp.Header.Get("Location")
			uploadDetailsKey := testfunc.KeyFromURL(redirect)

			// Upload details
			uploadDetails, err := storage.GetHTML(uploadDetailsKey)
			if err != nil {
				t.Fatalf("getting upload details: %v", err)
			}

			parsed, err := testfunc.ParseUploadDetails(uploadDetails.Content)
			if err != nil {
				t.Fatalf("parsing upload details: %v", err)
			}

			pf := parsed.Files["test.txt"]

			want := &testfunc.ParsedUploadDetailsFile{
				Name:              "test.txt",
				Icon:              "txt.png",
				Size:              "5 bytes",
				DirectLinkFileKey: testfunc.KeyFromURL(parsed.Files["test.txt"].DirectLinkFileKey),
				PasteLinkHTMLKey:  "TODO_PASTE_URL",
			}
			if diff := cmp.Diff(want, pf); diff != "" {
				t.Fatalf("unexpected upload details entry (-want +got):\n%s", diff)
			}

			// Raw file
			storage.AssertFile(t, testfunc.KeyFromURL(parsed.Files["test.txt"].DirectLinkFileKey), &testfunc.StoredObject{
				Content:            "test\n",
				MIMEType:           "text/plain",
				ContentDisposition: `inline; filename="test.txt"; filename*=utf-8''test.txt`,
				Links:              uploadDetails.Links,
				MetadataURL:        uploadDetails.MetadataURL,
			})

			// Metadata
			metadata := storage.AssertFile(t, testfunc.KeyFromURL(parsed.MetadataURL), &testfunc.StoredObject{
				Content:            testfunc.DoNotCompareContentSentinel,
				MIMEType:           "application/json",
				ContentDisposition: "",
				Links:              uploadDetails.Links,
				MetadataURL:        uploadDetails.MetadataURL,
			})

			var gotMetadata uploads.UploadMetadataFile
			if err := json.Unmarshal([]byte(metadata.Content), &gotMetadata); err != nil {
				t.Fatalf("unmarshaling metadata: %v", err)
			}

			wantMetadata := uploads.UploadMetadataFile{
				ServerVersion: conf.Version,
				Timestamp:     gotMetadata.Timestamp,
				UploadType:    "file",
				UploadedFiles: []uploads.UploadedFile{
					{
						Name:  "test.txt",
						Bytes: 5,
						Raw:   conf.FileURL(parsed.Files["test.txt"].DirectLinkFileKey).String(),
					},
				},
			}
			if diff := cmp.Diff(wantMetadata, gotMetadata); diff != "" {
				t.Fatalf("unexpected metadata (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUploadNoMultipart(t *testing.T) {
	t.Parallel()
	ts := testfunc.RunningServer(t, testfunc.NewConfig())
	defer ts.Cleanup()

	resp, err := http.Post(fmt.Sprintf("http://localhost:%d/upload", ts.Port), "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestUploadTooLarge(t *testing.T) {
	t.Parallel()
	conf := testfunc.NewConfig()
	conf.MaxUploadBytes = 1
	ts := testfunc.RunningServer(t, conf)
	defer ts.Cleanup()

	postBody := new(bytes.Buffer)
	writer := multipart.NewWriter(postBody)
	addFile(t, writer, "test.txt", "test\n")
	if err := writer.Close(); err != nil {
		t.Fatalf("closing writer: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/upload", ts.Port),
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

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status code: got %d, want %d\nBody:\n%s", resp.StatusCode, http.StatusBadRequest, body)
	}

	var errorResp errorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err != nil {
		t.Fatalf("unmarshaling error response: %v", err)
	}

	want := errorResponse{
		Success: false,
		Error:   "File is too large; max size is 1 byte.",
	}
	if diff := cmp.Diff(want, errorResp); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}
}

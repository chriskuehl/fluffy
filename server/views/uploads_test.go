package views_test

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

	"github.com/google/go-cmp/cmp"

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
	t.Parallel()
	storage := testfunc.NewMemoryStorageBackend()
	conf := testfunc.NewConfig(testfunc.WithStorageBackend(storage))
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
	uploadDetails := storage.HTMLs[uploadDetailsKey]
	var uploadDetailsHTML strings.Builder
	if _, err := io.Copy(&uploadDetailsHTML, uploadDetails); err != nil {
		t.Errorf("copying upload details HTML: %v", err)
	}

	parsed, err := testfunc.ParseUploadDetails(uploadDetailsHTML.String())
	if err != nil {
		t.Fatalf("parsing upload details: %v", err)
	}

	pf, ok := parsed.Files["test.txt"]
	if !ok {
		t.Fatalf("file not found in upload details")
	}

	wantName := "test.txt"
	if pf.Name != wantName {
		t.Fatalf("unexpected name: got %q, want %q", pf.Name, wantName)
	}

	wantIcon := "txt.png"
	if pf.Icon != wantIcon {
		t.Fatalf("unexpected icon: got %q, want %q", pf.Icon, wantIcon)
	}

	wantSize := "5 bytes"
	if pf.Size != wantSize {
		t.Fatalf("unexpected size: got %q, want %q", pf.Size, wantSize)
	}

	storedFile := storage.Files[pf.DirectLinkFileKey]

	if storedFile.Name() != wantName {
		t.Fatalf("unexpected stored file name: got %q, want %q", storedFile.Name(), wantName)
	}

	if storedFile.Key() != pf.DirectLinkFileKey {
		t.Fatalf("unexpected stored file key: got %q, want %q", storedFile.Key(), pf.DirectLinkFileKey)
	}

	rawURL := conf.FileURL(pf.DirectLinkFileKey)
	uploadDetailsURL := conf.HTMLURL(testfunc.KeyFromURL(redirect))
	metadataURL := conf.FileURL( // TODO: where do we get this? need to validate the contents, too.

	links := []*url.URL{rawURL, metadataURL, uploadDetailsURL}

	// TODO: links
	// TODO: metadata URL

	wantContentDisposition := `inline; filename="test.txt"; filename*=utf-8''test.txt`
	if storedFile.ContentDisposition() != wantContentDisposition {
		t.Fatalf("unexpected content disposition: got %q, want %q", storedFile.ContentDisposition(), wantContentDisposition)
	}

	wantMIMEType := "text/plain"
	if storedFile.MIMEType() != wantMIMEType {
		t.Fatalf("unexpected MIME type: got %q, want %q", storedFile.MIMEType(), wantMIMEType)
	}

	// TODO: add assertion about auto-paste once implemented
	// TODO: assert metadata was uploaded
}

func TestUploadJSON(t *testing.T) {
	t.Parallel()
	ts := testfunc.RunningServer(t, testfunc.NewConfig())
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

	var errorResp errorResponse
	if err := json.Unmarshal(bodyBytes, &errorResp); err != nil {
		t.Fatalf("unmarshaling error response: %v", err)
	}

	want := errorResponse{
		Success: true,
	}
	if diff := cmp.Diff(want, errorResp); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}

	// TODO: add assertions based on the redirect location to ensure files were actually uploaded
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

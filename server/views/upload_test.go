package views_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
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
	ts := testfunc.RunningServer(t, testfunc.NewConfig())
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

	// TODO: add assertions based on the redirect location to ensure files were actually uploaded
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

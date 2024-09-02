package server_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/chriskuehl/fluffy/testfunc"
)

func TestIndex(t *testing.T) {
	ts := testfunc.RunningServer(t, testfunc.NewConfig())
	defer ts.Cleanup()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/", ts.Port))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(bodyBytes)

	wantContent := []string{
		`<html class="page-index`,
		`<title>fluffy</title>`,
		`var icons = {"7z":"http://`,
		`var maxUploadSize =  10485760 ;`,
		`<a class="report" href="mailto:abuse@example.com">Report Abuse</a>`,
	}
	for _, want := range wantContent {
		if !strings.Contains(body, want) {
			t.Errorf("unexpected body: wanted to find %q but it is not present\nBody:\n%s", want, body)
		}
	}
}

func TestUploadHistory(t *testing.T) {
	ts := testfunc.RunningServer(t, testfunc.NewConfig())
	defer ts.Cleanup()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/upload-history", ts.Port))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(bodyBytes)

	wantContent := []string{
		`<html class="page-upload-history`,
		`<h2>Upload History</h2>`,
	}
	for _, want := range wantContent {
		if !strings.Contains(body, want) {
			t.Errorf("unexpected body: wanted to find %q but it is not present\nBody:\n%s", want, body)
		}
	}
}

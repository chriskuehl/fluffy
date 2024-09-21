package views_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/chriskuehl/fluffy/testfunc"
)

func TestUploadHistory(t *testing.T) {
	t.Parallel()
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
		`const icons = {"7z":"http://`,
	}
	for _, want := range wantContent {
		if !strings.Contains(body, want) {
			t.Errorf("unexpected body: wanted to find %q but it is not present\nBody:\n%s", want, body)
		}
	}
}

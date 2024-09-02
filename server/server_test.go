package server_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/chriskuehl/fluffy/testfunc"
)

func TestHealthz(t *testing.T) {
	ts := testfunc.RunningServer(t, testfunc.NewConfig())
	defer ts.Cleanup()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", ts.Port))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "ok\n"
	body := string(bodyBytes)
	if body != want {
		t.Errorf("unexpected body: got %q, want %q", body, want)
	}
}

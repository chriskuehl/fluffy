package assets_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/chriskuehl/fluffy/testfunc"
)

func TestDevStaticDev(t *testing.T) {
	conf := testfunc.NewConfig()
	conf.DevMode = true
	ts := testfunc.RunningServer(t, conf)
	defer ts.Cleanup()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/dev/static/img/favicon.ico", ts.Port))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	want := int64(1150)
	if resp.ContentLength != want {
		t.Fatalf("unexpected content length: got %d, want %d", resp.ContentLength, want)
	}
}

func TestDevStaticProd(t *testing.T) {
	ts := testfunc.RunningServer(t, testfunc.NewConfig())
	defer ts.Cleanup()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/dev/static/img/favicon.ico", ts.Port))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

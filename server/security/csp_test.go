package security_test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/chriskuehl/fluffy/testfunc"
)

var cspRegexp = regexp.MustCompile(
	strings.Join(
		[]string{
			`default-src 'self' https://fancy-cdn.com`,
			`script-src https://ajax.googleapis.com 'nonce-(?P<nonce>[^']+)' https://fancy-cdn.com`,
			`style-src 'self' https://fonts.googleapis.com https://fancy-cdn.com`,
			`font-src https://fonts.gstatic.com https://fancy-cdn.com`,
		},
		"; ",
	),
)

func TestCSP(t *testing.T) {
	conf := testfunc.NewConfig()
	conf.ObjectURLPattern = &url.URL{Scheme: "https", Host: "fancy-cdn.com", Path: ":path:"}
	ts := testfunc.RunningServer(t, conf)
	defer ts.Cleanup()

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", ts.Port))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	header := resp.Header.Get("Content-Security-Policy")
	matches := cspRegexp.FindStringSubmatch(header)
	if matches == nil {
		t.Fatalf("unexpected CSP header: got %q", header)
	}
	nonce := matches[cspRegexp.SubexpIndex("nonce")]
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body := string(bodyBytes)

	want := fmt.Sprintf(`<script nonce="%s">`, nonce)
	if !strings.Contains(body, want) {
		t.Fatalf("expected body to contain %q\nBody:\n%s", want, body)
	}
}

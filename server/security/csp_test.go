package security_test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/testfunc"
)

var (
	cspRegexp = regexp.MustCompile(
		strings.Join(
			[]string{
				`default-src 'self' https://fancy-cdn.com`,
				`script-src https://ajax.googleapis.com https://fancy-cdn.com 'nonce-(?P<nonce>[^']+)'`,
				`style-src 'self' https://fonts.googleapis.com https://fancy-cdn.com`,
				`font-src https://fonts.gstatic.com https://fancy-cdn.com`,
			},
			"; ",
		),
	)

	cspDevStaticFileRegexp = regexp.MustCompile(
		strings.Join(
			[]string{
				`default-src 'self' https://fancy-cdn.com \*`,
				`script-src https://ajax.googleapis.com https://fancy-cdn.com 'unsafe-inline'`,
				`style-src 'self' https://fonts.googleapis.com https://fancy-cdn.com`,
				`font-src https://fonts.gstatic.com https://fancy-cdn.com`,
			},
			"; ",
		),
	)
)

func TestCSP(t *testing.T) {
	t.Parallel()
	conf := testfunc.NewConfig()
	conf.FileURLPattern = &url.URL{Scheme: "https", Host: "fancy-cdn.com", Path: ":key:"}
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

func TestCSPDevStaticFile(t *testing.T) {
	tests := []struct {
		name       string
		devMode    bool
		want       *regexp.Regexp
		wantStatus int
	}{
		{
			name:       "dev_mode_has_unsafe_inline",
			devMode:    true,
			want:       cspDevStaticFileRegexp,
			wantStatus: http.StatusOK,
		},
		// This is kind of a silly test since we don't serve uploaded files in prod mode anyway,
		// but this at least verifies the 404 page doesn't have a weird CSP header ¯\_(ツ)_/¯.
		{
			name:       "prod_mode_has_no_unsafe_inline",
			devMode:    false,
			want:       cspRegexp,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			if err := os.WriteFile(filepath.Join(root, "test.html"), []byte("test"), 0644); err != nil {
				t.Fatalf("writing file: %s", err)
			}

			conf := testfunc.NewConfig(
				testfunc.WithStorageBackend(&storage.FilesystemBackend{
					FileRoot: root,
					HTMLRoot: root,
				}),
				testfunc.WithDevMode(tt.devMode),
			)

			conf.FileURLPattern = &url.URL{Scheme: "https", Host: "fancy-cdn.com", Path: ":key:"}
			ts := testfunc.RunningServer(t, conf)
			defer ts.Cleanup()

			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/dev/storage/html/test.html", ts.Port))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("unexpected status code: got %d, want %d", resp.StatusCode, tt.wantStatus)
			}

			header := resp.Header.Get("Content-Security-Policy")
			if header == "" {
				t.Fatalf("expected CSP header")
			}
			if !tt.want.MatchString(header) {
				t.Fatalf("unexpected CSP header: got %q", header)
			}
		})
	}
}

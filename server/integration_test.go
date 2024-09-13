package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/storage"
	"github.com/chriskuehl/fluffy/testfunc"
)

type objectType int

const (
	objectTypeObject objectType = iota
	objectTypeHTML   objectType = iota
)

type CanonicalizedLinks string

func canonicalizeLinks(links []*url.URL) CanonicalizedLinks {
	// Sort the links first to ensure a consistent order.
	urls := make([]string, len(links))
	for i, link := range links {
		urls[i] = link.String()
	}
	sort.Strings(urls)
	return CanonicalizedLinks(strings.Join(urls, " :: "))
}

type storedObject struct {
	Content            string
	MIMEType           string
	ContentDisposition string
	Links              CanonicalizedLinks
	MetadataURL        string
}

func keyFromURL(u *url.URL) string {
	s := u.String()
	return s[strings.LastIndex(s, "/")+1:]
}

func TestIntegration(t *testing.T) {
	tests := []struct {
		name                   string
		config                 func(t *testing.T) *config.Config
		getObject              func(objType objectType, conf *config.Config, key string) (*storedObject, error)
		stripUnsupportedFields func(obj *storedObject)
	}{
		{
			name: "memory_storage_backend",
			config: func(t *testing.T) *config.Config {
				t.Helper()
				return testfunc.NewConfig(
					testfunc.WithStorageBackend(
						testfunc.NewMemoryStorageBackend(),
					),
				)
			},
			getObject: func(objType objectType, conf *config.Config, key string) (*storedObject, error) {
				storageBackend := conf.StorageBackend.(*testfunc.MemoryStorageBackend)
				var obj config.BaseStoredObject
				if objType == objectTypeObject {
					if o, ok := storageBackend.Objects[key]; ok {
						obj = o
					}
				} else {
					if o, ok := storageBackend.HTML[key]; ok {
						obj = o
					}
				}
				if obj == nil {
					return nil, fmt.Errorf("object %q not found", key)
				}
				var content strings.Builder
				if _, err := io.Copy(&content, obj); err != nil {
					return nil, fmt.Errorf("copying object: %w", err)
				}
				return &storedObject{
					Content:            content.String(),
					MIMEType:           obj.MIMEType(),
					ContentDisposition: obj.ContentDisposition(),
					Links:              canonicalizeLinks(obj.Links()),
					MetadataURL:        obj.MetadataURL().String(),
				}, nil
			},
		},
		{
			name: "filesystem_storage_backend",
			config: func(t *testing.T) *config.Config {
				t.Helper()
				htmlRoot := t.TempDir()
				objectRoot := t.TempDir()
				return testfunc.NewConfig(
					testfunc.WithStorageBackend(&storage.FilesystemBackend{
						ObjectRoot: objectRoot,
						HTMLRoot:   htmlRoot,
					}),
				)
			},
			getObject: func(objType objectType, conf *config.Config, key string) (*storedObject, error) {
				storageBackend := conf.StorageBackend.(*storage.FilesystemBackend)
				var path string
				if objType == objectTypeObject {
					path = filepath.Join(storageBackend.ObjectRoot, key)
				} else {
					path = filepath.Join(storageBackend.HTMLRoot, key)
				}

				contents, err := os.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("opening file: %w", err)
				}

				return &storedObject{
					Content: string(contents),
				}, nil
			},
			stripUnsupportedFields: func(obj *storedObject) {
				obj.MIMEType = ""
				obj.ContentDisposition = ""
				obj.Links = ""
				obj.MetadataURL = ""
			},
		},
		{
			name: "s3_storage_backend",
			config: func(t *testing.T) *config.Config {
				t.Helper()
				backend, err := storage.NewS3Backend(
					"fake-region",
					"fake-bucket",
					"object/",
					"html/",
					func(awsCfg aws.Config, optFn func(*s3.Options)) storage.S3Client {
						return testfunc.NewFakeS3Client()
					},
				)
				if err != nil {
					t.Fatalf("constructing backend: %v", err)
				}
				return testfunc.NewConfig(testfunc.WithStorageBackend(backend))
			},
			getObject: func(objType objectType, conf *config.Config, key string) (*storedObject, error) {
				storageBackend := conf.StorageBackend.(*storage.S3Backend)
				client := storageBackend.Client.(*testfunc.FakeS3Client)

				var path string
				if objType == objectTypeObject {
					path = storageBackend.ObjectKeyPrefix + key
				} else {
					path = storageBackend.HTMLKeyPrefix + key
				}

				contents, ok := client.Objects[path]
				if !ok {
					return nil, fmt.Errorf("object %q not found", key)
				}

				links := strings.Split(contents.Metadata["fluffy-links"], "; ")
				linkURLs := make([]*url.URL, len(links))
				for i, link := range links {
					u, err := url.ParseRequestURI(link)
					if err != nil {
						return nil, fmt.Errorf("parsing link %q: %w", link, err)
					}
					linkURLs[i] = u
				}

				return &storedObject{
					Content:            string(contents.Contents),
					MIMEType:           contents.ContentType,
					ContentDisposition: contents.ContentDisposition,
					Links:              canonicalizeLinks(linkURLs),
					MetadataURL:        contents.Metadata["fluffy-metadata"],
				}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			conf := tt.config(t)
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

			// TODO: `redirect` is actually supposed to be a redirect to an HTML page. Update this
			// to verify `redirect` once this is in place.

			rawURL, err := url.ParseRequestURI(result.UploadedFiles["test.txt"].Raw)
			if err != nil {
				t.Fatalf("parsing raw URL: %v", err)
			}
			metadataURL, err := url.ParseRequestURI(result.Metadata)
			if err != nil {
				t.Fatalf("parsing metadata URL: %v", err)
			}

			key := keyFromURL(rawURL)

			obj, err := tt.getObject(objectTypeObject, conf, key)
			if err != nil {
				t.Fatalf("getting object: %v", err)
			}

			links := []*url.URL{rawURL, metadataURL}

			want := &storedObject{
				Content:            "test",
				MIMEType:           "text/plain",
				ContentDisposition: `inline; filename="test.txt"; filename*=utf-8''test.txt`,
				Links:              canonicalizeLinks(links),
				MetadataURL:        obj.MetadataURL,
			}
			if tt.stripUnsupportedFields != nil {
				tt.stripUnsupportedFields(want)
			}
			if diff := cmp.Diff(want, obj); diff != "" {
				t.Fatalf("unexpected object (-want +got):\n%s", diff)
			}
		})
	}
}

package testfunc

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/go-cmp/cmp"

	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/storage"
)

// DoNotCompareContentSentinel is a sentinel value that indicates that the content should not be
// compared when using AssertFile or AssertHTML.
const DoNotCompareContentSentinel = "DO_NOT_COMPARE_CONTENT"

// StoredObject is a test wrapper around a stored object.
type StoredObject struct {
	Content            string
	MIMEType           string
	ContentDisposition string
	Links              CanonicalizedLinks
	MetadataURL        string
}

// StorageBackend is a test wrapper around a storage backend.
type StorageBackend struct {
	Backend                config.StorageBackend
	GetFile                func(key string) (*StoredObject, error)
	GetHTML                func(key string) (*StoredObject, error)
	StripUnsupportedFields func(obj *StoredObject)
}

func (b *StorageBackend) assertObject(t *testing.T, got *StoredObject, want *StoredObject) *StoredObject {
	t.Helper()
	if b.StripUnsupportedFields != nil {
		b.StripUnsupportedFields(want)
	}
	if want.Content == DoNotCompareContentSentinel {
		want.Content = got.Content
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected object (-want +got):\n%s", diff)
	}
	return got
}

func (b *StorageBackend) AssertFile(t *testing.T, key string, want *StoredObject) *StoredObject {
	t.Helper()
	got, err := b.GetFile(key)
	if err != nil {
		t.Fatalf("getting file %q: %v", key, err)
	}
	return b.assertObject(t, got, want)
}

func (b *StorageBackend) AssertHTML(t *testing.T, key string, want *StoredObject) *StoredObject {
	t.Helper()
	got, err := b.GetHTML(key)
	if err != nil {
		t.Fatalf("getting HTML %q: %v", key, err)
	}
	return b.assertObject(t, got, want)
}

type StorageFactory func(*testing.T) StorageBackend

// StorageBackends is a list of storage backends available for testing. This is intended to be used
// in a table test.
var StorageBackends = []struct {
	Name           string
	StorageFactory StorageFactory
}{
	{
		Name:           "memory_backend",
		StorageFactory: memoryBackend,
	},
	{
		Name:           "filesystem_backend",
		StorageFactory: filesystemBackend,
	},
	{
		Name:           "s3_backend",
		StorageFactory: s3Backend,
	},
}

func AddStorageBackends[T any](tests map[string]T) []struct {
	Name           string
	StorageFactory StorageFactory
	T              T
} {
	ret := make([]struct {
		Name           string
		StorageFactory StorageFactory
		T              T
	}, len(tests)*len(StorageBackends))
	i := 0
	for testName, test := range tests {
		for _, tt := range StorageBackends {
			ret[i] = struct {
				Name           string
				StorageFactory StorageFactory
				T              T
			}{
				Name:           fmt.Sprintf("%s/%s", testName, tt.Name),
				StorageFactory: tt.StorageFactory,
				T:              test,
			}
			i++
		}
	}
	return ret
}

func memoryBackend(t *testing.T) StorageBackend {
	backend := NewMemoryStorageBackend()

	getObject := func(obj config.BaseStoredObject) (*StoredObject, error) {
		var content strings.Builder
		if _, err := io.Copy(&content, obj); err != nil {
			return nil, fmt.Errorf("copying object: %w", err)
		}
		return &StoredObject{
			Content:            content.String(),
			MIMEType:           obj.MIMEType(),
			ContentDisposition: obj.ContentDisposition(),
			Links:              CanonicalizeLinks(obj.Links()),
			MetadataURL:        obj.MetadataURL().String(),
		}, nil
	}

	return StorageBackend{
		Backend: backend,
		GetFile: func(key string) (*StoredObject, error) {
			file, ok := backend.Files[key]
			if !ok {
				return nil, fmt.Errorf("file %q not found", key)
			}
			return getObject(file)
		},
		GetHTML: func(key string) (*StoredObject, error) {
			html, ok := backend.HTMLs[key]
			if !ok {
				return nil, fmt.Errorf("HTML %q not found", key)
			}
			return getObject(html)
		},
	}
}

func filesystemBackend(t *testing.T) StorageBackend {
	getObject := func(root string, key string) (*StoredObject, error) {
		path := filepath.Join(root, key)
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		return &StoredObject{
			Content: string(contents),
		}, nil
	}

	backend := &storage.FilesystemBackend{
		FileRoot: t.TempDir(),
		HTMLRoot: t.TempDir(),
	}

	return StorageBackend{
		Backend: backend,
		GetFile: func(key string) (*StoredObject, error) {
			return getObject(backend.FileRoot, key)
		},
		GetHTML: func(key string) (*StoredObject, error) {
			return getObject(backend.HTMLRoot, key)
		},
		StripUnsupportedFields: func(obj *StoredObject) {
			obj.MIMEType = ""
			obj.ContentDisposition = ""
			obj.Links = ""
			obj.MetadataURL = ""
		},
	}
}

func s3Backend(t *testing.T) StorageBackend {
	client := NewFakeS3Client()
	backend, err := storage.NewS3Backend(
		"fake-region",
		"fake-bucket",
		"file/",
		"html/",
		func(awsCfg aws.Config, optFn func(*s3.Options)) storage.S3Client {
			return client
		},
	)
	if err != nil {
		t.Fatalf("constructing backend: %v", err)
	}

	getObject := func(path string) (*StoredObject, error) {
		contents, ok := client.Objects[path]
		if !ok {
			return nil, fmt.Errorf("object %q not found", path)
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

		return &StoredObject{
			Content:            string(contents.Contents),
			MIMEType:           contents.ContentType,
			ContentDisposition: contents.ContentDisposition,
			Links:              CanonicalizeLinks(linkURLs),
			MetadataURL:        contents.Metadata["fluffy-metadata"],
		}, nil
	}

	return StorageBackend{
		Backend: backend,
		GetFile: func(key string) (*StoredObject, error) {
			return getObject(backend.FileKeyPrefix + key)
		},
		GetHTML: func(key string) (*StoredObject, error) {
			return getObject(backend.HTMLKeyPrefix + key)
		},
	}
}

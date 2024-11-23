package config

import (
	"context"
	"embed"
	"html/template"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/chriskuehl/fluffy/server/utils"
)

type BaseStoredObject interface {
	io.ReadSeekCloser
	Key() string
	Links() []*url.URL
	MetadataURL() *url.URL
	ContentDisposition() string
	MIMEType() string
}

// StoredFile represents a file to be stored.
type StoredFile interface {
	BaseStoredObject
	// Name returns the human-readable, non-sanitized, non-unique name of the file.
	Name() string
}

// StoredHTML represents an HTML object to be stored. This object should be stored in such a way
// that it can be served to clients with a text/html MIME type. Additional properties
// (Content-Disposition, links, etc.) may not be available with all storage backends.
type StoredHTML interface {
	BaseStoredObject
}

type StorageBackend interface {
	// StoreFile stores the given file object. This file object should be stored in such a way that
	// it is never served as rendered HTML, even if the uploaded file happens to be HTML.
	// Additional properties (custom MIME type, Content-Disposition, etc.) may also be stored, but
	// support varies by storage backend.
	StoreFile(ctx context.Context, file StoredFile) error

	// StoreHTML stores the given HTML object. This HTML object should be stored in such a way that
	// it can be served to clients with a text/html MIME type.
	StoreHTML(ctx context.Context, html StoredHTML) error

	// Validate returns a list of errors if the storage backend's configuration is invalid.
	Validate() []string
}

type Assets struct {
	FS *embed.FS
	// Hashes is a map of file paths to their SHA-256 hashes.
	Hashes map[string]string
	// MIMEExtensions is a set of all the mime extensions, without dot, e.g. "png", "jpg".
	MIMEExtensions map[string]struct{}
}

type Templates struct {
	// Pages
	Index         *template.Template
	Paste         *template.Template
	UploadDetails *template.Template
	UploadHistory *template.Template

	// Includes
	InlineJS *template.Template
}

func NewTemplates(fs *embed.FS) *Templates {
	funcs := template.FuncMap{
		"plusOne": func(i int) int {
			return i + 1
		},
		"pluralize": utils.Pluralize[int],
	}
	mustTemplate := func(name string) *template.Template {
		return template.Must(template.New("").Funcs(funcs).ParseFS(fs, "templates/include/*.html", "templates/"+name))
	}
	return &Templates{
		Index:         mustTemplate("index.html"),
		Paste:         mustTemplate("paste.html"),
		UploadDetails: mustTemplate("upload-details.html"),
		UploadHistory: mustTemplate("upload-history.html"),

		InlineJS: mustTemplate("include/inline-js.html"),
	}
}

type Config struct {
	// Site configuration.
	StorageBackend          StorageBackend
	Branding                string
	CustomFooterHTML        template.HTML
	AbuseContactEmail       string
	MaxUploadBytes          int64
	MaxMultipartMemoryBytes int64
	HomeURL                 *url.URL
	FileURLPattern          *url.URL
	HTMLURLPattern          *url.URL
	// ForbiddenFileExtensions is the set of file extensions that are not allowed to be uploaded.
	// The extensions should not start with a dot, but may contain one if trying to match multiple
	// extensions, e.g. "tar.gz".
	ForbiddenFileExtensions map[string]struct{}
	Host                    string
	Port                    uint
	GlobalTimeout           time.Duration

	// Runtime options, cannot be set via config.
	DevMode   bool
	Version   string
	Assets    *Assets
	Templates *Templates
}

func (conf *Config) Validate() []string {
	var errs []string
	if conf.StorageBackend == nil {
		errs = append(errs, "StorageBackend must not be nil")
	} else {
		errs = append(errs, conf.StorageBackend.Validate()...)
	}
	if conf.Branding == "" {
		errs = append(errs, "Branding must not be empty")
	}
	if conf.AbuseContactEmail == "" {
		errs = append(errs, "AbuseContactEmail must not be empty")
	}
	if conf.MaxUploadBytes <= 0 {
		errs = append(errs, "MaxUploadBytes must be greater than 0")
	}
	if conf.MaxMultipartMemoryBytes <= 0 {
		errs = append(errs, "MaxMultipartMemoryBytes must be greater than 0")
	}
	if conf.HomeURL == nil {
		errs = append(errs, "HomeURL must not be nil")
	} else if strings.HasSuffix(conf.HomeURL.Path, "/") {
		errs = append(errs, "HomeURL must not end with a slash")
	}
	if conf.FileURLPattern == nil {
		errs = append(errs, "FileURLPattern must not be nil")
	} else if !strings.Contains(conf.FileURLPattern.Path, ":key:") {
		errs = append(errs, "FileURLPattern must contain a ':key:' placeholder")
	}
	if conf.HTMLURLPattern == nil {
		errs = append(errs, "HTMLURLPattern must not be nil")
	} else if !strings.Contains(conf.HTMLURLPattern.Path, ":key:") {
		errs = append(errs, "HTMLURLPattern must contain a ':key:' placeholder")
	}
	if conf.ForbiddenFileExtensions == nil {
		errs = append(errs, "ForbiddenFileExtensions must not be nil")
	}
	for ext := range conf.ForbiddenFileExtensions {
		if strings.HasPrefix(ext, ".") {
			errs = append(errs, "ForbiddenFileExtensions should not start with a dot: "+ext)
		}
		if strings.ToLower(ext) != ext {
			errs = append(errs, "ForbiddenFileExtensions should be lowercase: "+ext)
		}
	}
	if conf.Version == "" {
		errs = append(errs, "Version must not be empty")
	}
	if conf.Assets == nil {
		errs = append(errs, "Assets must not be nil")
	}
	if conf.Templates == nil {
		errs = append(errs, "Templates must not be nil")
	}
	return errs
}

// FileURL returns a URL for the given stored file.
func (conf *Config) FileURL(key string) *url.URL {
	url := *conf.FileURLPattern
	url.Path = strings.Replace(url.Path, ":key:", key, -1)
	return &url
}

// HTMLURL returns a URL for the given stored HTML.
func (conf *Config) HTMLURL(key string) *url.URL {
	url := *conf.HTMLURLPattern
	url.Path = strings.Replace(url.Path, ":key:", key, -1)
	return &url
}

package config

import (
	"context"
	"embed"
	"html/template"
	"io"
	"net/url"
	"strings"
	"time"
)

type BaseStoredObject interface {
	io.ReadSeekCloser
	Key() string
	Links() []*url.URL
	MetadataURL() *url.URL
	ContentDisposition() string
	MIMEType() string
}

type StoredObject interface {
	BaseStoredObject
}

type StoredHTML interface {
	BaseStoredObject
}

type StorageBackend interface {
	StoreObject(ctx context.Context, obj StoredObject) error
	StoreHTML(ctx context.Context, html StoredHTML) error
	Validate() []string
}

type Assets struct {
	FS             *embed.FS
	Hashes         map[string]string
	MIMEExtensions map[string]struct{}
}

type Templates struct {
	FS *embed.FS
}

func (t *Templates) Must(name string) *template.Template {
	return template.Must(template.New("").ParseFS(t.FS, "templates/include/*.html", "templates/"+name))
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
	ObjectURLPattern        *url.URL
	HTMLURLPattern          *url.URL
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
	if conf.ObjectURLPattern == nil {
		errs = append(errs, "ObjectURLPattern must not be nil")
	} else if !strings.Contains(conf.ObjectURLPattern.Path, ":path:") {
		errs = append(errs, "ObjectURLPattern must contain a ':path:' placeholder")
	}
	if conf.HTMLURLPattern == nil {
		errs = append(errs, "HTMLURLPattern must not be nil")
	} else if !strings.Contains(conf.HTMLURLPattern.Path, ":path:") {
		errs = append(errs, "HTMLURLPattern must contain a ':path:' placeholder")
	}
	if conf.ForbiddenFileExtensions == nil {
		errs = append(errs, "ForbiddenFileExtensions must not be nil")
	}
	for ext := range conf.ForbiddenFileExtensions {
		if strings.HasPrefix(ext, ".") {
			errs = append(errs, "ForbiddenFileExtensions should not start with a dot: "+ext)
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

func (conf *Config) ObjectURL(path string) *url.URL {
	url := *conf.ObjectURLPattern
	url.Path = strings.Replace(url.Path, ":path:", path, -1)
	return &url
}

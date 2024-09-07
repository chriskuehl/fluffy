package config

import (
	"context"
	"html/template"
	"net/url"
	"strings"

	"github.com/chriskuehl/fluffy/server/storage/storagedata"
)

type StorageBackend interface {
	StoreObject(ctx context.Context, obj storagedata.Object) error
	StoreHTML(ctx context.Context, obj storagedata.Object) error
}

type Config struct {

	// Site configuration.
	StorageBackend          StorageBackend
	Branding                string
	CustomFooterHTML        template.HTML
	AbuseContactEmail       string
	MaxUploadBytes          int64
	MaxMultipartMemoryBytes int64
	HomeURL                 url.URL
	ObjectURLPattern        url.URL
	HTMLURLPattern          url.URL
	ForbiddenFileExtensions map[string]struct{}

	// Runtime options.
	Host    string
	Port    uint
	DevMode bool
	Version string
}

func (conf *Config) Validate() []string {
	var errs []string
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
	if strings.HasSuffix(conf.HomeURL.Path, "/") {
		errs = append(errs, "HomeURL must not end with a slash")
	}
	if !strings.Contains(conf.ObjectURLPattern.Path, "{path}") {
		errs = append(errs, "ObjectURLPattern must contain a '{path}' placeholder")
	}
	if !strings.Contains(conf.HTMLURLPattern.Path, "{path}") {
		errs = append(errs, "HTMLURLPattern must contain a '{path}' placeholder")
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
	return errs
}

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

func (c *Config) Validate() []string {
	var errs []string
	if c.Branding == "" {
		errs = append(errs, "Branding must not be empty")
	}
	if c.AbuseContactEmail == "" {
		errs = append(errs, "AbuseContactEmail must not be empty")
	}
	if c.MaxUploadBytes <= 0 {
		errs = append(errs, "MaxUploadBytes must be greater than 0")
	}
	if c.MaxMultipartMemoryBytes <= 0 {
		errs = append(errs, "MaxMultipartMemoryBytes must be greater than 0")
	}
	if strings.HasSuffix(c.HomeURL.Path, "/") {
		errs = append(errs, "HomeURL must not end with a slash")
	}
	if !strings.Contains(c.ObjectURLPattern.Path, "%s") {
		errs = append(errs, "ObjectURLPattern must contain a '%s' placeholder")
	}
	if !strings.Contains(c.HTMLURLPattern.Path, "%s") {
		errs = append(errs, "HTMLURLPattern must contain a '%s' placeholder")
	}
	for ext := range c.ForbiddenFileExtensions {
		if strings.HasPrefix(ext, ".") {
			errs = append(errs, "ForbiddenFileExtensions should not start with a dot: "+ext)
		}
	}
	if c.Version == "" {
		errs = append(errs, "Version must not be empty")
	}
	return errs
}

package storagedata

import (
	"io"
	"net/url"
)

type Object struct {
	Key         string
	Links       []*url.URL
	MetadataURL *url.URL
	Reader      io.Reader
	Bytes       int64
}

package storage

import (
	"io"
	"net/url"
)

type BaseStoredObject struct {
	ObjKey         string
	ObjLinks       []*url.URL
	ObjMetadataURL *url.URL
	ObjReadCloser  io.ReadCloser
	ObjBytes       int64
}

func (b *BaseStoredObject) Key() string {
	return b.ObjKey
}

func (b *BaseStoredObject) Links() []*url.URL {
	return b.ObjLinks
}

func (b *BaseStoredObject) MetadataURL() *url.URL {
	return b.ObjMetadataURL
}

func (b *BaseStoredObject) ReadCloser() (io.ReadCloser, error) {
	// TODO: only open this now
	return b.ObjReadCloser, nil
}

func (b *BaseStoredObject) Bytes() int64 {
	return b.ObjBytes
}

type StoredObject struct {
	BaseStoredObject
	ObjMIMEType string
}

func (o *StoredObject) MIMEType() string {
	return o.ObjMIMEType
}

type StoredHTML struct {
	BaseStoredObject
}

func (h *StoredHTML) MIMEType() string {
	return "text/html; charset=utf-8"
}

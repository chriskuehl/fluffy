package storage

import (
	"io"
	"net/url"

	"github.com/chriskuehl/fluffy/server/config"
)

type baseStoredObject struct {
	io.ReadSeekCloser
	key         string
	links       []*url.URL
	metadataURL *url.URL
}

func (b *baseStoredObject) Key() string {
	return b.key
}

func (b *baseStoredObject) Links() []*url.URL {
	return b.links
}

func (b *baseStoredObject) MetadataURL() *url.URL {
	return b.metadataURL
}

type StoredObjectOption interface {
	applyToStoredObject(*storedObject)
}

type StoredHTMLOption interface {
	applyToStoredHTML(*storedHTML)
}

type BaseStoredObjectOption interface {
	StoredObjectOption
	StoredHTMLOption
}

type baseStoredObjectOption func(*baseStoredObject)

func (o baseStoredObjectOption) applyToStoredObject(obj *storedObject) {
	o(&obj.baseStoredObject)
}

func (o baseStoredObjectOption) applyToStoredHTML(obj *storedHTML) {
	o(&obj.baseStoredObject)
}

func WithKey(key string) baseStoredObjectOption {
	return func(o *baseStoredObject) {
		o.key = key
	}
}

func WithLinks(links []*url.URL) baseStoredObjectOption {
	return func(o *baseStoredObject) {
		o.links = links
	}
}

func WithMetadataURL(metadataURL *url.URL) baseStoredObjectOption {
	return func(o *baseStoredObject) {
		o.metadataURL = metadataURL
	}
}

type storedObject struct {
	baseStoredObject
	mimeType           string
	contentDisposition string
}

type storedObjectOption func(*storedObject)

func (o storedObjectOption) applyToStoredObject(obj *storedObject) {
	o(obj)
}

func WithMIMEType(mimeType string) storedObjectOption {
	return func(o *storedObject) {
		o.mimeType = mimeType
	}
}

func WithContentDisposition(contentDisposition string) storedObjectOption {
	return func(o *storedObject) {
		o.contentDisposition = contentDisposition
	}
}

func NewStoredObject(readSeekCloser io.ReadSeekCloser, opts ...StoredObjectOption) config.StoredObject {
	ret := &storedObject{
		baseStoredObject: baseStoredObject{
			ReadSeekCloser: readSeekCloser,
		},
	}
	for _, opt := range opts {
		opt.applyToStoredObject(ret)
	}
	return ret
}

func UpdatedStoredObject(obj config.StoredObject, opts ...StoredObjectOption) config.StoredObject {
	ret := &storedObject{
		baseStoredObject: baseStoredObject{
			ReadSeekCloser: obj,
			key:            obj.Key(),
			links:          obj.Links(),
			metadataURL:    obj.MetadataURL(),
		},
		mimeType:           obj.MIMEType(),
		contentDisposition: obj.ContentDisposition(),
	}
	for _, opt := range opts {
		opt.applyToStoredObject(ret)
	}
	return ret
}

func (o *storedObject) MIMEType() string {
	return o.mimeType
}

func (o *storedObject) ContentDisposition() string {
	return o.contentDisposition
}

type storedHTML struct {
	baseStoredObject
}

func NewStoredHTML(key string, readSeekCloser io.ReadSeekCloser, opts ...StoredHTMLOption) config.StoredHTML {
	ret := &storedHTML{
		baseStoredObject: baseStoredObject{
			ReadSeekCloser: readSeekCloser,
		},
	}
	for _, opt := range opts {
		opt.applyToStoredHTML(ret)
	}
	return ret
}

func (h *storedHTML) MIMEType() string {
	return "text/html; charset=utf-8"
}

func (h *storedHTML) ContentDisposition() string {
	return "inline"
}

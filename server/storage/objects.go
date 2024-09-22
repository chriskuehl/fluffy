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

type StoredFileOption interface {
	applyToStoredFile(*storedFile)
}

type StoredHTMLOption interface {
	applyToStoredHTML(*storedHTML)
}

type BaseStoredObjectOption interface {
	StoredFileOption
	StoredHTMLOption
}

type baseStoredObjectOption func(*baseStoredObject)

func (o baseStoredObjectOption) applyToStoredFile(obj *storedFile) {
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

type storedFile struct {
	baseStoredObject
	mimeType           string
	contentDisposition string
	name               string
}

type storedFileOption func(*storedFile)

func (o storedFileOption) applyToStoredFile(file *storedFile) {
	o(file)
}

func WithMIMEType(mimeType string) storedFileOption {
	return func(o *storedFile) {
		o.mimeType = mimeType
	}
}

func WithContentDisposition(contentDisposition string) storedFileOption {
	return func(o *storedFile) {
		o.contentDisposition = contentDisposition
	}
}

func WithName(name string) storedFileOption {
	return func(o *storedFile) {
		o.name = name
	}
}

func NewStoredFile(readSeekCloser io.ReadSeekCloser, opts ...StoredFileOption) config.StoredFile {
	ret := &storedFile{
		baseStoredObject: baseStoredObject{
			ReadSeekCloser: readSeekCloser,
		},
	}
	for _, opt := range opts {
		opt.applyToStoredFile(ret)
	}
	return ret
}

func UpdatedStoredFile(file config.StoredFile, opts ...StoredFileOption) config.StoredFile {
	ret := &storedFile{
		baseStoredObject: baseStoredObject{
			ReadSeekCloser: file,
			key:            file.Key(),
			links:          file.Links(),
			metadataURL:    file.MetadataURL(),
		},
		mimeType:           file.MIMEType(),
		contentDisposition: file.ContentDisposition(),
		name:               file.Name(),
	}
	for _, opt := range opts {
		opt.applyToStoredFile(ret)
	}
	return ret
}

func (o *storedFile) MIMEType() string {
	return o.mimeType
}

func (o *storedFile) ContentDisposition() string {
	return o.contentDisposition
}

func (o *storedFile) Name() string {
	return o.name
}

type storedHTML struct {
	baseStoredObject
}

func NewStoredHTML(readSeekCloser io.ReadSeekCloser, opts ...StoredHTMLOption) config.StoredHTML {
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

func UpdatedStoredHTML(html config.StoredHTML, opts ...StoredHTMLOption) config.StoredHTML {
	ret := &storedHTML{
		baseStoredObject: baseStoredObject{
			ReadSeekCloser: html,
			key:            html.Key(),
			links:          html.Links(),
			metadataURL:    html.MetadataURL(),
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

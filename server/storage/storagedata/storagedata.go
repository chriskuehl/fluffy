package storagedata

import (
	"io"
)

type Object struct {
	Key         string
	Links       []string
	MetadataURL string
	Reader      io.Reader
	Bytes       int64
}

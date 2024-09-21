package testfunc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/chriskuehl/fluffy/server/config"
)

type MemoryStorageBackend struct {
	Files map[string]config.StoredFile
	HTMLs map[string]config.StoredHTML
	mu    sync.Mutex
}

func (b *MemoryStorageBackend) StoreFile(ctx context.Context, obj config.StoredFile) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Files[obj.Key()] = obj
	return nil
}

func (b *MemoryStorageBackend) StoreHTML(ctx context.Context, obj config.StoredHTML) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.HTMLs[obj.Key()] = obj
	return nil
}

func (b *MemoryStorageBackend) Validate() []string {
	return nil
}

func NewMemoryStorageBackend() *MemoryStorageBackend {
	return &MemoryStorageBackend{
		Files: make(map[string]config.StoredFile),
		HTMLs: make(map[string]config.StoredHTML),
	}
}

type StoredS3Object struct {
	Bucket             string
	Key                string
	Contents           []byte
	Metadata           map[string]string
	ContentDisposition string
	ContentType        string
	ACL                string
}

type FakeS3Client struct {
	Objects map[string]StoredS3Object
	mu      sync.Mutex
}

func (f *FakeS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, params.Body); err != nil {
		return nil, fmt.Errorf("copying object: %w", err)
	}
	f.Objects[*params.Key] = StoredS3Object{
		Bucket:             *params.Bucket,
		Key:                *params.Key,
		Contents:           buf.Bytes(),
		Metadata:           params.Metadata,
		ContentDisposition: *params.ContentDisposition,
		ContentType:        *params.ContentType,
		ACL:                string(params.ACL),
	}
	return &s3.PutObjectOutput{}, nil
}

func NewFakeS3Client() *FakeS3Client {
	return &FakeS3Client{
		Objects: make(map[string]StoredS3Object),
	}
}

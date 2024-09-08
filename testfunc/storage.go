package testfunc

import (
	"context"
	"sync"

	"github.com/chriskuehl/fluffy/server/storage/storagedata"
)

type MemoryStorageBackend struct {
	Objects map[string]storagedata.Object
	HTML    map[string]storagedata.Object
	mu      sync.Mutex
}

func (b *MemoryStorageBackend) StoreObject(ctx context.Context, obj storagedata.Object) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Objects[obj.Key] = obj
	return nil
}

func (b *MemoryStorageBackend) StoreHTML(ctx context.Context, obj storagedata.Object) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.HTML[obj.Key] = obj
	return nil
}

func NewMemoryStorageBackend() *MemoryStorageBackend {
	return &MemoryStorageBackend{
		Objects: make(map[string]storagedata.Object),
		HTML:    make(map[string]storagedata.Object),
	}
}

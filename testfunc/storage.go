package testfunc

import (
	"context"
	"sync"

	"github.com/chriskuehl/fluffy/server/config"
)

type MemoryStorageBackend struct {
	Objects map[string]config.StoredObject
	HTML    map[string]config.StoredHTML
	mu      sync.Mutex
}

func (b *MemoryStorageBackend) StoreObject(ctx context.Context, obj config.StoredObject) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Objects[obj.Key()] = obj
	return nil
}

func (b *MemoryStorageBackend) StoreHTML(ctx context.Context, obj config.StoredHTML) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.HTML[obj.Key()] = obj
	return nil
}

func (b *MemoryStorageBackend) Validate() []string {
	return nil
}

func NewMemoryStorageBackend() *MemoryStorageBackend {
	return &MemoryStorageBackend{
		Objects: make(map[string]config.StoredObject),
		HTML:    make(map[string]config.StoredHTML),
	}
}

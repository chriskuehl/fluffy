package testfunc

import (
	"context"
	"sync"

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

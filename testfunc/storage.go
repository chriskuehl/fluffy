package testfunc

import (
	"context"

	"github.com/chriskuehl/fluffy/server/storage/storagedata"
)

type MemoryStorageBackend struct {
	Objects map[string]storagedata.Object
	HTML    map[string]storagedata.Object
}

func (b *MemoryStorageBackend) StoreObject(ctx context.Context, obj storagedata.Object) error {
	b.Objects[obj.Key] = obj
	return nil
}

func (b *MemoryStorageBackend) StoreHTML(ctx context.Context, obj storagedata.Object) error {
	b.HTML[obj.Key] = obj
	return nil
}

func NewMemoryStorageBackend() *MemoryStorageBackend {
	return &MemoryStorageBackend{
		Objects: make(map[string]storagedata.Object),
		HTML:    make(map[string]storagedata.Object),
	}
}

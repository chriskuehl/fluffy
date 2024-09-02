package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Object struct {
	Key         string
	Links       []string
	MetadataURL string
	Reader      io.Reader
}

func (o Object) Validate() error {
	if o.Key == "" {
		return errors.New("key must not be empty")
	}
	if filepath.Clean(o.Key) != o.Key {
		return errors.New("key contains invalid characters")
	}
	return nil
}

type Backend interface {
	StoreObject(ctx context.Context, obj Object) error
	StoreHTML(ctx context.Context, obj Object) error
}

type FilesystemBackend struct {
	ObjectRoot string
	HTMLRoot   string
}

func absPath(path string) (string, error) {
	p, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("getting absolute path: %w", err)
	}
	p, err = filepath.EvalSymlinks(p)
	if err != nil {
		return "", fmt.Errorf("evaluating symlinks: %w", err)
	}
	return p, nil
}

func (b *FilesystemBackend) store(root string, obj Object) error {
	if err := obj.Validate(); err != nil {
		return fmt.Errorf("validating object: %w", err)
	}
	realRoot, err := absPath(root)
	if err != nil {
		return fmt.Errorf("getting real root: %w", err)
	}

	parentPath, err := absPath(filepath.Join(root, filepath.Dir(obj.Key)))
	if err != nil {
		return fmt.Errorf("getting parent path: %w", err)
	}

	if !strings.HasPrefix(parentPath+string(filepath.Separator), realRoot+string(filepath.Separator)) {
		return fmt.Errorf("parent path %q is outside of root %q", parentPath, realRoot)
	}

	path := filepath.Join(parentPath, filepath.Base(obj.Key))
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, obj.Reader); err != nil {
		return fmt.Errorf("copying file: %w", err)
	}
	return nil
}

func (b *FilesystemBackend) StoreObject(ctx context.Context, obj Object) error {
	return b.store(b.ObjectRoot, obj)
}

func (b *FilesystemBackend) StoreHTML(ctx context.Context, obj Object) error {
	return b.store(b.HTMLRoot, obj)
}

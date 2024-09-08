package assets

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/chriskuehl/fluffy/server/config"
)

var mimeExtensions = []string{}

func LoadAssets(assetsFS *embed.FS) (*config.Assets, error) {
	assets := config.Assets{
		FS:     assetsFS,
		Hashes: map[string]string{},
		// MimeExtensions is a set of all the mime extensions, without dot, e.g. "png", "jpg".
		MimeExtensions: map[string]struct{}{},
	}

	if err := fs.WalkDir(assetsFS, "static", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		f, err := assetsFS.Open(path)
		if err != nil {
			return fmt.Errorf("opening %q: %w", path, err)
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return fmt.Errorf("hashing %q: %w", path, err)
		}
		assets.Hashes[path] = hex.EncodeToString(h.Sum(nil))
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking assets: %w", err)
	}

	if err := fs.WalkDir(assetsFS, "static/img/mime/small", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".png") {
			return nil
		}

		name := filepath.Base(path)
		assets.MimeExtensions[name[:len(name)-len(".png")]] = struct{}{}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking mime icons: %w", err)
	}

	return &assets, nil
}

// AssetObjectPath returns the path to the asset in the object store.
//
// Keep in mind the object may not exist yet depending on when this function is called.
func assetObjectPath(path, hash string) string {
	return filepath.Join("static", hash, path)
}

// AssetURL returns the URL to the asset.
//
// In development mode, this will return a URL served by the fluffy server itself. In production,
// this will return a URL to the object store.
func AssetURL(conf *config.Config, path string) (string, error) {
	if conf.DevMode {
		url := conf.HomeURL
		url.Path = "/dev/static/" + path
		return url.String(), nil
	}

	hash, ok := conf.Assets.Hashes[filepath.Join("static", path)]
	if !ok {
		return "", fmt.Errorf("asset not found: %s", path)
	}
	return conf.ObjectURL(assetObjectPath(path, hash)).String(), nil
}

// AssetAsString returns the contents of the asset as a string.
func AssetAsString(conf *config.Config, path string) (string, error) {
	data, err := fs.ReadFile(conf.Assets.FS, filepath.Join("static", path))
	if err != nil {
		return "", fmt.Errorf("reading asset: %w", err)
	}
	return string(data), nil
}

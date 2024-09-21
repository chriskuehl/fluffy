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

func LoadAssets(assetsFS *embed.FS) (*config.Assets, error) {
	assets := config.Assets{
		FS:             assetsFS,
		Hashes:         map[string]string{},
		MIMEExtensions: map[string]struct{}{},
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
		assets.MIMEExtensions[name[:len(name)-len(".png")]] = struct{}{}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walking mime icons: %w", err)
	}

	return &assets, nil
}

// assetKey returns the key for an asset in the file store.
//
// Keep in mind the asset may not exist yet in the file store depending on when this function is
// called.
func assetKey(path, hash string) string {
	return filepath.Join("static", hash, path)
}

// AssetURL returns the URL to the asset.
//
// In development mode, this will return a URL served by the fluffy server itself. In production,
// this will return a URL to the file store.
func AssetURL(conf *config.Config, path string) (string, error) {
	if conf.DevMode {
		url := *conf.HomeURL
		url.Path = "/dev/static/" + path
		return url.String(), nil
	}

	hash, ok := conf.Assets.Hashes[filepath.Join("static", path)]
	if !ok {
		return "", fmt.Errorf("asset not found: %s", path)
	}
	return conf.FileURL(assetKey(path, hash)).String(), nil
}

// AssetAsString returns the contents of the asset as a string.
func AssetAsString(conf *config.Config, path string) (string, error) {
	data, err := fs.ReadFile(conf.Assets.FS, filepath.Join("static", path))
	if err != nil {
		return "", fmt.Errorf("reading asset: %w", err)
	}
	return string(data), nil
}

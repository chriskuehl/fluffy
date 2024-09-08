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

//go:embed static/*
var assetsFS embed.FS
var assetHash = make(map[string]string)

var mimeExtensions = []string{}

func init() {
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
		assetHash[path] = hex.EncodeToString(h.Sum(nil))
		return nil
	}); err != nil {
		panic("loading asset hashes: " + err.Error())
	}

	if err := fs.WalkDir(assetsFS, "static/img/mime/small", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".png") {
			return nil
		}

		name := filepath.Base(path)
		mimeExtensions = append(mimeExtensions, name[:len(name)-len(".png")])
		return nil
	}); err != nil {
		panic("loading mime extensions: " + err.Error())
	}
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

	hash, ok := assetHash[filepath.Join("static", path)]
	if !ok {
		return "", fmt.Errorf("asset not found: %s", path)
	}
	url := conf.ObjectURLPattern
	url.Path = strings.Replace(url.Path, ":path:", assetObjectPath(path, hash), -1)
	return url.String(), nil
}

// AssetAsString returns the contents of the asset as a string.
func AssetAsString(path string) (string, error) {
	data, err := fs.ReadFile(assetsFS, filepath.Join("static", path))
	if err != nil {
		return "", fmt.Errorf("reading asset: %w", err)
	}
	return string(data), nil
}

// MimeExtensions returns a list of all the mime extensions, without dot, e.g. "png", "jpg".
func MimeExtensions() []string {
	return mimeExtensions
}

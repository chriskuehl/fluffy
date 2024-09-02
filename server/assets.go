package server

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/chriskuehl/fluffy/server/logging"
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

func (c *Config) AssetObjectPath(path, hash string) string {
	return filepath.Join("static", hash, path)
}

func (c *Config) AssetURL(path string) string {
	if c.DevMode {
		return c.HomeURL + "/dev/static/" + path
	}

	hash, ok := assetHash[filepath.Join("static", path)]
	if !ok {
		panic("asset not found: " + path)
	}
	return fmt.Sprintf(c.ObjectURLPattern, c.AssetObjectPath(path, hash))
}

func (c *Config) AssetAsString(path string) string {
	data, err := fs.ReadFile(assetsFS, filepath.Join("static", path))
	if err != nil {
		panic("asset not found: " + path)
	}
	return string(data)
}

func (m meta) InlineJS(path string) template.HTML {
	var buf bytes.Buffer
	data := struct {
		meta
		Content template.JS
	}{
		meta:    m,
		Content: template.JS(m.Config.AssetAsString(path)),
	}
	if err := templates.ExecuteTemplate(&buf, "inline-js.html", data); err != nil {
		panic("executing template: " + err.Error())
	} else {
		return template.HTML(buf.String())
	}
}

func handleStatic(config *Config, logger logging.Logger) http.HandlerFunc {
	if !config.DevMode {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Warn(r.Context(), "assets cannot be served from the server in production")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Assets cannot be served from the server in production.\n"))
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		strippedReq := r.Clone(r.Context())
		strippedReq.URL.Path = strippedReq.URL.Path[len("/dev"):]
		http.FileServer(http.FS(assetsFS)).ServeHTTP(w, strippedReq)
	}
}

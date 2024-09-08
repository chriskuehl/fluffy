package meta

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/chriskuehl/fluffy/server/assets"
	"github.com/chriskuehl/fluffy/server/config"
	"github.com/chriskuehl/fluffy/server/security"
)

type PageConfig struct {
	ID               string
	ExtraHTMLClasses []string
}

func (p PageConfig) HTMLClasses() string {
	return "page-" + p.ID + " " + strings.Join(p.ExtraHTMLClasses, " ")
}

type Meta struct {
	Conf     *config.Config
	PageConf PageConfig
	Nonce    string
}

func NewMeta(ctx context.Context, conf *config.Config, pc PageConfig) (*Meta, error) {
	nonce, err := security.CSPNonce(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting CSP nonce: %w", err)
	}
	return &Meta{
		Conf:     conf,
		PageConf: pc,
		Nonce:    nonce,
	}, nil
}

func (m Meta) InlineJS(path string) template.HTML {
	src, err := assets.AssetAsString(m.Conf, path)
	if err != nil {
		panic("loading asset: " + err.Error())
	}
	var buf bytes.Buffer
	data := struct {
		Meta
		Content template.JS
	}{
		Meta:    m,
		Content: template.JS(src),
	}
	tmpl := m.Conf.Templates.Must("include/inline-js.html")
	if err := tmpl.ExecuteTemplate(&buf, "inline-js.html", data); err != nil {
		panic("executing template: " + err.Error())
	} else {
		return template.HTML(buf.String())
	}
}

func (m Meta) AssetURL(path string) string {
	url, err := assets.AssetURL(m.Conf, path)
	if err != nil {
		panic("loading asset: " + err.Error())
	}
	return url
}

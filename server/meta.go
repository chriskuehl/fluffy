package server

import (
	"bytes"
	"context"
	"html/template"

	"github.com/chriskuehl/fluffy/server/assets"
	"github.com/chriskuehl/fluffy/server/config"
)

type meta struct {
	Config     *config.Config
	PageConfig pageConfig
	Nonce      string
}

func NewMeta(ctx context.Context, config *config.Config, pc pageConfig) meta {
	nonce, ok := ctx.Value(cspNonceKey{}).(string)
	if !ok {
		panic("no nonce in context")
	}
	return meta{
		Config:     config,
		PageConfig: pc,
		Nonce:      nonce,
	}
}

var inlineJSTemplate = pageTemplate("include/inline-js.html")

func (m meta) InlineJS(path string) template.HTML {
	src, err := assets.AssetAsString(path)
	if err != nil {
		panic("loading asset: " + err.Error())
	}
	var buf bytes.Buffer
	data := struct {
		meta
		Content template.JS
	}{
		meta:    m,
		Content: template.JS(src),
	}
	if err := inlineJSTemplate.ExecuteTemplate(&buf, "inline-js.html", data); err != nil {
		panic("executing template: " + err.Error())
	} else {
		return template.HTML(buf.String())
	}
}

func (m meta) AssetURL(path string) string {
	url, err := assets.AssetURL(m.Config, path)
	if err != nil {
		panic("loading asset: " + err.Error())
	}
	return url
}

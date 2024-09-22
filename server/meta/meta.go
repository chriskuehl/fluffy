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
	"github.com/chriskuehl/fluffy/server/utils"
)

type PageConfig struct {
	ID               string
	ExtraHTMLClasses []string
	IsStatic         bool
}

func (p PageConfig) HTMLClasses() string {
	return "page-" + p.ID + " " + strings.Join(p.ExtraHTMLClasses, " ")
}

type Meta struct {
	Conf     *config.Config
	PageConf PageConfig
	nonce    string
}

func (m Meta) Nonce() string {
	if m.nonce == "" {
		panic("no nonce set; this should only be called for dynamic pages")
	}
	return m.nonce
}

func NewMeta(ctx context.Context, conf *config.Config, pc PageConfig) (*Meta, error) {
	nonce := ""
	if !pc.IsStatic {
		n, err := security.CSPNonce(ctx)
		if err != nil {
			return nil, fmt.Errorf("getting CSP nonce: %w", err)
		}
		nonce = n
	}
	return &Meta{
		Conf:     conf,
		PageConf: pc,
		nonce:    nonce,
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

func (m Meta) MIMEIcon(filename string) string {
	// Try "file.tar.gz" => "tar.gz" => "gz" => "unknown".
	parts := strings.Split(filename, ".")
	for i := 0; i < len(parts); i++ {
		ext := strings.Join(parts[i:], ".")
		if _, ok := m.Conf.Assets.MIMEExtensions[ext]; ok {
			return ext
		}
	}
	return "unknown"
}

func (m Meta) MIMEIconSmallURL(iconName string) string {
	return m.AssetURL("img/mime/small/" + iconName + ".png")
}

func (m Meta) FormatBytes(bytes int64) string {
	return utils.FormatBytes(bytes)
}

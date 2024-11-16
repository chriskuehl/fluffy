package highlighting

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

type MarkdownHighlighter struct{}

func (m *MarkdownHighlighter) Name() string {
	return "Rendered Markdown"
}

func (m *MarkdownHighlighter) RenderAsDiff() bool {
	return false
}

func (m *MarkdownHighlighter) RenderAsRichText() bool {
	return true
}

func (p *MarkdownHighlighter) ExtraHTMLClasses() []string {
	return []string{"markdown"}
}

func (p *MarkdownHighlighter) GenerateTexts(text string) []*Text {
	return []*Text{simpleText(text)}
}

func (m *MarkdownHighlighter) Highlight(text *Text) (template.HTML, error) {
	md := goldmark.New(
		// TODO: add syntax highlighting
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	var html strings.Builder
	if err := md.Convert([]byte(text.Text), &html); err != nil {
		return "", fmt.Errorf("rendering Markdown: %w", err)
	}

	return template.HTML(html.String()), nil
}

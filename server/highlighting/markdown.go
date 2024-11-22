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

func (h *MarkdownHighlighter) Name() string {
	return "Rendered Markdown"
}

func (h *MarkdownHighlighter) RenderAsDiff() bool {
	return false
}

func (h *MarkdownHighlighter) RenderAsRichText() bool {
	return true
}

func (h *MarkdownHighlighter) RenderAsTerminal() bool {
	return false
}

func (h *MarkdownHighlighter) ExtraHTMLClasses() []string {
	return []string{"markdown"}
}

func (h *MarkdownHighlighter) GenerateTexts(text string) []*Text {
	return []*Text{simpleText(text)}
}

func (h *MarkdownHighlighter) Highlight(text *Text) (template.HTML, error) {
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

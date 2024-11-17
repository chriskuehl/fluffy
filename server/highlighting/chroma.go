package highlighting

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
)

func guessLexer(text, language, filename string) chroma.Lexer {
	var lexer chroma.Lexer
	if language != "" && language != "autodetect" {
		lexer = lexers.Get(language)
		fmt.Printf("lexer from language: %v\n", lexer)
	}
	if lexer == nil && filename != "" {
		lexer = lexers.Match(filename)
		fmt.Printf("lexer from filename: %v\n", lexer)
	}
	if lexer == nil {
		lexer = lexers.Analyse(text)
		fmt.Printf("lexer from text: %v\n", lexer)
	}
	if lexer == nil {
		lexer = lexers.Fallback
		fmt.Printf("lexer from fallback: %v\n", lexer)
	}
	return lexer
}

var Formatter = html.New(html.WithClasses(true))

type ChromaHighlighter struct {
	lexer chroma.Lexer
}

func NewChromaHighlighter(lexer chroma.Lexer) *ChromaHighlighter {
	return &ChromaHighlighter{lexer: chroma.Coalesce(lexer)}
}

func (h *ChromaHighlighter) Name() string {
	name := h.lexer.Config().Name
	if name == "plaintext" {
		return "Plain Text"
	}
	return name
}

func (h *ChromaHighlighter) RenderAsDiff() bool {
	return false
}

func (h *ChromaHighlighter) RenderAsRichText() bool {
	return false
}

func (h *ChromaHighlighter) ExtraHTMLClasses() []string {
	return nil
}

func (h *ChromaHighlighter) GenerateTexts(text string) []*Text {
	return []*Text{simpleText(text)}
}

func (h *ChromaHighlighter) Highlight(text *Text) (template.HTML, error) {
	iterator, err := h.lexer.Tokenise(nil, text.Text)
	if err != nil {
		return "", fmt.Errorf("tokenizing: %w", err)
	}

	var html strings.Builder
	if err := Formatter.Format(&html, DefaultStyle.ChromaStyle, iterator); err != nil {
		return "", fmt.Errorf("formatting: %w", err)
	}
	return template.HTML(html.String()), nil
}

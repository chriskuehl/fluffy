package highlighting

import (
	"html/template"
	"strings"
)

var UILanguagesMap = map[string]string{
	"bash":        "Bash / Shell",
	"c":           "C",
	"c++":         "C++",
	"cheetah":     "Cheetah",
	"diff":        "Diff",
	"groovy":      "Groovy",
	"haskell":     "Haskell",
	"html":        "HTML",
	"java":        "Java",
	"javascript":  "JavaScript",
	"json":        "JSON",
	"kotlin":      "Kotlin",
	"lua":         "Lua",
	"makefile":    "Makefile",
	"objective-c": "Objective-C",
	"php":         "PHP",
	"python3":     "Python",
	"ruby":        "Ruby",
	"rust":        "Rust",
	"scala":       "Scala",
	"sql":         "SQL",
	"swift":       "Swift",
	"yaml":        "YAML",
}

type StyleCategory struct {
	Name   string
	Styles []Style
}

type Style struct {
	Name string
}

var Styles = []StyleCategory{
	{
		Name: "Light",
		Styles: []Style{
			{Name: "default"},
			{Name: "pastie"},
		},
	},
	{
		Name: "Dark",
		Styles: []Style{
			{Name: "monokai"},
			{Name: "solarized-dark"},
		},
	},
}

// Text represents a single piece of text.
//
// A single paste usually consists of a single Text, but in the case of a diff, it may contain
// multiple Texts.
type Text struct {
	Text string
	// Array index corresponds to zero-indexed line number in this text,
	// and the value is the array of zero-indexed line numbers that line
	// corresponds to in the original text.
	LineNumberMapping [][]int
}

func simpleText(text string) *Text {
	// This is a little tricky because whether lines end in newlines depends on the source of the
	// text. Our approach is to not count a final \n if it's present since the user probably
	// doesn't intend to create an empty line at the end of the text. We do count multiple newlines
	// though since that is more clear.
	// "a" => 1 line
	// "a\n" => 1 line
	// "a\nb" => 2 lines
	// "a\nb\n" => 2 lines
	// "a\nb\n\n" => 3 lines
	lineCount := strings.Count(text, "\n") + 1
	if strings.HasSuffix(text, "\n") {
		lineCount -= 1
	}
	mapping := make([][]int, lineCount)
	for i := range mapping {
		mapping[i] = []int{i}
	}
	return &Text{
		Text:              text,
		LineNumberMapping: mapping,
	}
}

// Highlighter is an interface for syntax highlighting.
//
// A Highlighter is responsible for taking a piece of text and returning an HTML representation.
type Highlighter interface {
	// Name returns the name of the highlighter (e.g. "Python").
	Name() string
	// RenderAsDiff returns true if the paste should be rendered as a diff.
	RenderAsDiff() bool
	// RenderAsRichText returns true if the paste should be rendered as rich text, without line
	// numbers or other plaintext formatting. Useful for e.g. Markdown.
	RenderAsRichText() bool
	// ExtraHTMLClasses returns a extra CSS classes that should be added.
	ExtraHTMLClasses() []string
	// GenerateTexts takes a piece of text and returns a slice of Texts.
	// Generally this will return a single Text, but in the case of a diff, it may return multiple.
	GenerateTexts(text string) []*Text
	// Highlight takes a piece of text and returns an HTML representation.
	Highlight(text *Text) (template.HTML, error)
}

func GuessHighlighterForPaste(text string, language string) Highlighter {
	if language == "rendered-markdown" {
		return &MarkdownHighlighter{}
	}

	// TODO: implement
	return &PlainTextHighlighter{}
}

type PlainTextHighlighter struct{}

func (p *PlainTextHighlighter) Name() string {
	return "Plain Text"
}

func (p *PlainTextHighlighter) RenderAsDiff() bool {
	return false
}

func (p *PlainTextHighlighter) RenderAsRichText() bool {
	return false
}

func (p *PlainTextHighlighter) ExtraHTMLClasses() []string {
	return nil
}

func (p *PlainTextHighlighter) GenerateTexts(text string) []*Text {
	return []*Text{simpleText(text)}
}

func (p *PlainTextHighlighter) Highlight(text *Text) (template.HTML, error) {
	var html strings.Builder
	for _, line := range strings.Split(text.Text, "\n") {
		html.WriteString(template.HTMLEscapeString(line))
		html.WriteString("<br />")
	}
	return template.HTML(html.String()), nil
}

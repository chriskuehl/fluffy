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

type Highlighter interface {
	Name() string
	IsDiff() bool
	Highlight(text *Text) (template.HTML, error)
}

type PlainTextHighlighter struct{}

func (p *PlainTextHighlighter) Name() string {
	return "Plain Text"
}

func (p *PlainTextHighlighter) IsDiff() bool {
	return false
}

func (p *PlainTextHighlighter) Highlight(text *Text) (template.HTML, error) {
	var html strings.Builder
	for _, line := range strings.Split(text.Text, "\n") {
		html.WriteString(template.HTMLEscapeString(line))
		html.WriteString("<br />")
	}
	return template.HTML(html.String()), nil
}

type Text struct {
	Text string
	// Array index corresponds to zero-indexed line number in this text,
	// and the value is the array of zero-indexed line numbers that line
	// corresponds to in the original text.
	LineNumberMapping [][]int
}

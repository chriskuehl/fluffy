package highlighting

import (
	"html/template"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
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

var diffLinePrefixes = []string{
	"diff --git",
	"--- ",
	"+++ ",
	"index ",
	"@@ ",
	"Author:",
	"AuthorDate:",
	"Commit:",
	"CommitDate:",
	"commit ",
}

type StyleCategory struct {
	Name   string
	Styles []Style
}

type Color string

type FluffyColorSet struct {
	ToolbarForeground                Color
	ToolbarBackground                Color
	Border                           Color
	LineNumbersForeground            Color
	LineNumbersBackground            Color
	LineNumbersHoverBackground       Color
	LineNumbersSelectedBackground    Color
	SelectedLineBackground           Color
	DiffAddLineBackground            Color
	DiffAddSelectedLineBackground    Color
	DiffRemoveLineBackground         Color
	DiffRemoveSelectedLineBackground Color
}

var lightFluffyColors = &FluffyColorSet{
	ToolbarForeground:                "#333",
	ToolbarBackground:                "#e0e0e0",
	Border:                           "#eee",
	LineNumbersForeground:            "#222",
	LineNumbersBackground:            "#fafafa",
	LineNumbersHoverBackground:       "#ffeaaf",
	LineNumbersSelectedBackground:    "#ffe18e",
	SelectedLineBackground:           "#fff3d3",
	DiffAddLineBackground:            "#e2ffe2",
	DiffAddSelectedLineBackground:    "#e8ffbc",
	DiffRemoveLineBackground:         "#ffe5e5",
	DiffRemoveSelectedLineBackground: "#ffdfbf",
}

type Style struct {
	Name         string
	ChromaStyle  *chroma.Style
	ANSIColors   *ANSIColorSet
	FluffyColors *FluffyColorSet
}

var DefaultStyle = Style{
	Name:         "default",
	ChromaStyle:  styles.Get("xcode"),
	ANSIColors:   ansiColorsLight,
	FluffyColors: lightFluffyColors,
}

var DefaultDarkStyle = Style{
	Name:        "monokai",
	ChromaStyle: styles.Get("monokai"),
	ANSIColors:  ansiColorsDark,
	FluffyColors: &FluffyColorSet{
		ToolbarForeground:                "#333",
		ToolbarBackground:                "#e0e0e0",
		Border:                           "#454545",
		LineNumbersForeground:            "#999",
		LineNumbersBackground:            "#272822",
		LineNumbersHoverBackground:       "#8D8D8D",
		LineNumbersSelectedBackground:    "#5F5F5F",
		SelectedLineBackground:           "#545454",
		DiffAddLineBackground:            "#3d523d",
		DiffAddSelectedLineBackground:    "#607b60",
		DiffRemoveLineBackground:         "#632727",
		DiffRemoveSelectedLineBackground: "#9e4848",
	},
}

var Styles = []StyleCategory{
	{
		Name: "Light",
		Styles: []Style{
			DefaultStyle,
			{
				Name:         "pastie",
				ChromaStyle:  styles.Get("pastie"),
				ANSIColors:   ansiColorsLight,
				FluffyColors: lightFluffyColors,
			},
		},
	},
	{
		Name: "Dark",
		Styles: []Style{
			DefaultDarkStyle,
			{
				Name:        "catppuccin-frappe",
				ChromaStyle: styles.Get("catppuccin-frappe"),
				ANSIColors:  ansiColorsDark,
				FluffyColors: &FluffyColorSet{
					ToolbarForeground:                "#333",
					ToolbarBackground:                "#e0e0e0",
					Border:                           "#454545",
					LineNumbersForeground:            "#656565",
					LineNumbersBackground:            "#002b36",
					LineNumbersHoverBackground:       "#00596f",
					LineNumbersSelectedBackground:    "#004252",
					SelectedLineBackground:           "#004252",
					DiffAddLineBackground:            "#0e400e",
					DiffAddSelectedLineBackground:    "#176117",
					DiffRemoveLineBackground:         "#632727",
					DiffRemoveSelectedLineBackground: "#9e4848",
				},
			},
			{
				Name:        "solarized-dark",
				ChromaStyle: styles.Get("solarized-dark"),
				ANSIColors:  ansiColorsDark,
				FluffyColors: &FluffyColorSet{
					ToolbarForeground:                "#333",
					ToolbarBackground:                "#e0e0e0",
					Border:                           "#454545",
					LineNumbersForeground:            "#656565",
					LineNumbersBackground:            "#002b36",
					LineNumbersHoverBackground:       "#00596f",
					LineNumbersSelectedBackground:    "#004252",
					SelectedLineBackground:           "#004252",
					DiffAddLineBackground:            "#0e400e",
					DiffAddSelectedLineBackground:    "#176117",
					DiffRemoveLineBackground:         "#632727",
					DiffRemoveSelectedLineBackground: "#9e4848",
				},
			},
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
	// RenderAsTerminal returns true if the paste should be rendered as terminal output.
	RenderAsTerminal() bool
	// ExtraHTMLClasses returns a extra CSS classes that should be added.
	ExtraHTMLClasses() []string
	// GenerateTexts takes a piece of text and returns a slice of Texts.
	//   * Generally this will return a single Text, but in the case of a diff, it may return multiple.
	//	 * It will always return at least one Text.
	//   * The first returned Text is the primary Text.
	//   * Line mappings in the returned Texts are in reference to the primary Text.
	GenerateTexts(text string) []*Text
	// Highlight takes a piece of text and returns an HTML representation.
	//
	// For rich-text highlighters, the returned HTML may be anything.
	//
	// For plain-text highlighters, the returned HTML should contain:
	//   * A <pre> with class "chroma" wrapping the highlighted text.
	//   * Each line should be its own element (e.g. <span> or <div>) with classes "line" and
	//     "line-NUMBER" where NUMBER is the 1-indexed line number.
	Highlight(text *Text) (template.HTML, error)
}

func looksLikeDiff(text string) bool {
	// TODO: improve this
	return strings.HasPrefix(text, "diff --git") || strings.Contains(text, "\ndiff --git")
}

func looksLikeAnsiColor(text string) bool {
	return strings.Contains(text, "\x1b[")
}

// stripDiffThings removes things from text that make it look like a diff.
//
// The purpose of this is so we can run guessLexer over the source text. If
// we have a diff of Python, Chroma might tell us the language is "Diff".
// Really, we want it to highlight it like it's Python, and then we'll apply
// the diff formatting on top.
func stripDiffThings(text string) string {
	var s strings.Builder

outer:
	for _, line := range strings.Split(text, "\n") {
		for _, prefix := range diffLinePrefixes {
			if strings.HasPrefix(line, prefix) {
				continue outer
			}
		}

		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") {
			line = line[1:]
		}

		s.WriteString(line + "\n")
	}

	return s.String()
}

func GuessHighlighterForPaste(text, language, filename string) Highlighter {
	if language == "rendered-markdown" {
		return &MarkdownHighlighter{}
	}

	if (language == "" || language == "autodetect") && looksLikeAnsiColor(text) {
		return &ANSIHighlighter{}
	}

	diffRequested := (language == "diff" ||
		strings.HasSuffix(filename, ".diff") ||
		strings.HasSuffix(filename, ".patch"))
	diffRequestedLanguage := ""

	if strings.HasPrefix(language, "diff-") {
		diffRequested = true
		diffRequestedLanguage = strings.TrimPrefix(language, "diff-")
	}

	lexer := guessLexer(text, language, filename)
	lexerName := strings.ToLower(lexer.Config().Name)
	if lexerName != strings.ToLower(language) && (diffRequested || lexerName == "diff" || looksLikeDiff(text)) {
		// It wasn't a perfect match and it looks like a diff, so apply diff formatting rather than
		// regular Chroma formatting.
		lexer = guessLexer(stripDiffThings(text), diffRequestedLanguage, filename)
		// TODO: implement
		return &PlainTextHighlighter{}
	}

	return &ChromaHighlighter{lexer: lexer}
}

// TODO: probably get rid of this
type PlainTextHighlighter struct{}

func (h *PlainTextHighlighter) Name() string {
	return "Plain Text"
}

func (h *PlainTextHighlighter) RenderAsDiff() bool {
	return false
}

func (h *PlainTextHighlighter) RenderAsRichText() bool {
	return false
}

func (h *PlainTextHighlighter) RenderAsTerminal() bool {
	return false
}

func (h *PlainTextHighlighter) ExtraHTMLClasses() []string {
	return nil
}

func (h *PlainTextHighlighter) GenerateTexts(text string) []*Text {
	return []*Text{simpleText(text)}
}

func (h *PlainTextHighlighter) Highlight(text *Text) (template.HTML, error) {
	var html strings.Builder
	for _, line := range strings.Split(text.Text, "\n") {
		html.WriteString(template.HTMLEscapeString(line))
		html.WriteString("<br />")
	}
	return template.HTML(html.String()), nil
}

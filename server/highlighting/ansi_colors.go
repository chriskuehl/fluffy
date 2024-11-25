package highlighting

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

type ANSIColors struct {
	Black   Color
	Red     Color
	Green   Color
	Yellow  Color
	Blue    Color
	Magenta Color
	Cyan    Color
	White   Color
}

type ANSIColorSet struct {
	Foreground      ANSIColors
	ForegroundFaint ANSIColors
	Background      ANSIColors
}

var ansiColorsLight = &ANSIColorSet{
	Foreground: ANSIColors{
		Black:   "#000000",
		Red:     "#EF2929",
		Green:   "#62CA00",
		Yellow:  "#DAC200",
		Blue:    "#3465A4",
		Magenta: "#CE42BE",
		Cyan:    "#34E2E2",
		White:   "#FFFFFF",
	},
	// TODO: double-check these colors
	ForegroundFaint: ANSIColors{
		Black:   "#676767",
		Red:     "#ff6d67",
		Green:   "#5ff967",
		Yellow:  "#fefb67",
		Blue:    "#6871ff",
		Magenta: "#ff76ff",
		Cyan:    "#5ffdff",
		White:   "#feffff",
	},
	Background: ANSIColors{
		Black:   "#000000",
		Red:     "#EF2929",
		Green:   "#8AE234",
		Yellow:  "#FCE94F",
		Blue:    "#3465A4",
		Magenta: "#C509C5",
		Cyan:    "#34E2E2",
		White:   "#FFFFFF",
	},
}

var ansiColorsDark = &ANSIColorSet{
	Foreground: ANSIColors{
		Black:   "#555753",
		Red:     "#FF5C5C",
		Green:   "#8AE234",
		Yellow:  "#FCE94F",
		Blue:    "#8FB6E1",
		Magenta: "#FF80F1",
		Cyan:    "#34E2E2",
		White:   "#EEEEEC",
	},
	// TODO: double-check these colors
	ForegroundFaint: ANSIColors{
		Black:   "#676767",
		Red:     "#ff6d67",
		Green:   "#5ff967",
		Yellow:  "#fefb67",
		Blue:    "#6871ff",
		Magenta: "#ff76ff",
		Cyan:    "#5ffdff",
		White:   "#feffff",
	},
	Background: ANSIColors{
		Black:   "#555753",
		Red:     "#F03D3D",
		Green:   "#6ABC1B",
		Yellow:  "#CEB917",
		Blue:    "#6392C6",
		Magenta: "#FF80F1",
		Cyan:    "#2FC0C0",
		White:   "#BFBFBF",
	},
}

type ANSIHighlighter struct{}

func (h *ANSIHighlighter) Name() string {
	return "ANSI Color"
}

func (h *ANSIHighlighter) RenderAsDiff() bool {
	return false
}

func (h *ANSIHighlighter) RenderAsRichText() bool {
	return false
}

func (h *ANSIHighlighter) RenderAsTerminal() bool {
	return true
}

func (h *ANSIHighlighter) ExtraHTMLClasses() []string {
	return nil
}

func (h *ANSIHighlighter) GenerateTexts(text string) []*Text {
	return []*Text{simpleText(text)}
}

func (h *ANSIHighlighter) Highlight(text *Text) (template.HTML, error) {
	var sb strings.Builder

	sb.WriteString(`<pre class="chroma"><code>`)

	lines := strings.Split(strings.TrimSuffix(text.Text, "\n"), "\n")
	state := ansiState{}

	for i, line := range lines {
		fmt.Fprintf(&sb, `<span class="line line-%d">`, i+1)

		for _, parsed := range parseANSI(line, state) {
			state = parsed.state // For next iteration.
			classes := state.cssClasses()
			styles := state.cssStyles()
			if len(classes) > 0 || len(styles) > 0 {
				sb.WriteString("<span")

				if len(classes) > 0 {
					fmt.Fprintf(&sb, ` class="%s"`, strings.Join(classes, " "))
				}

				if len(styles) > 0 {
					sb.WriteString(` style="`)
					first := true
					for key, value := range styles {
						if !first {
							sb.WriteString(";")
						}
						fmt.Fprintf(&sb, "%s: %s", key, value)
						first = false
					}
					sb.WriteString(`"`)
				}
				sb.WriteString(">")

				sb.WriteString(html.EscapeString(parsed.text))
				sb.WriteString("</span>")
			} else {
				sb.WriteString(html.EscapeString(parsed.text))
			}

		}

		sb.WriteString("\n") // Newline at the end ensures that empty lines are still rendered.
		sb.WriteString(`</span>`)
	}

	sb.WriteString(`</code></pre>`)

	return template.HTML(sb.String()), nil
}

type rgb struct {
	r uint8
	g uint8
	b uint8
}

type ansiColor struct {
	index uint8
	// nil indicates standard color and index is set
	rgb *rgb
}

type ansiState struct {
	bold          bool
	faint         bool
	italic        bool
	underline     bool
	strikethrough bool
	// nil indicates no color is set
	foreground *ansiColor
	background *ansiColor
}

func (s ansiState) cssStyles() map[string]string {
	ret := map[string]string{}
	if s.bold {
		ret["font-weight"] = "bold"
	}
	if s.italic {
		ret["font-style"] = "italic"
	}
	if s.underline {
		ret["text-decoration"] = "underline"
	}
	if s.strikethrough {
		ret["text-decoration"] = "line-through"
	}
	if s.foreground != nil && s.foreground.rgb != nil {
		r := s.foreground.rgb.r
		g := s.foreground.rgb.g
		b := s.foreground.rgb.b
		if s.faint {
			// TODO: no idea if this is correct
			r = r / 2
			g = g / 2
			b = b / 2
		}
		ret["color"] = fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)
	}
	if s.background != nil && s.background.rgb != nil {
		ret["background-color"] = fmt.Sprintf("rgb(%d, %d, %d)", s.background.rgb.r, s.background.rgb.g, s.background.rgb.b)
	}
	return ret
}

func (s ansiState) cssClasses() []string {
	// Index colors are done through classes so that they can be customized by themes.
	ret := []string{}
	if s.foreground != nil && s.foreground.rgb == nil {
		class := fmt.Sprintf("fg-%d", s.foreground.index)
		if s.faint {
			class += "-faint"
		}
		ret = append(ret, class)
	}
	if s.background != nil && s.background.rgb == nil {
		ret = append(ret, fmt.Sprintf("bg-%d", s.background.index))
	}
	return ret
}

func (s ansiState) update(command string) ansiState {
	switch command {
	case "0":
		return ansiState{}
	case "1":
		s.bold = true
	case "2":
		s.faint = true
	case "3":
		s.italic = true
	case "4":
		s.underline = true
	case "9":
		s.strikethrough = true
	case "22":
		s.bold = false
		s.faint = false
	case "23":
		s.italic = false
	case "24":
		s.underline = false
	case "29":
		s.strikethrough = false
	case "30":
		s.foreground = &ansiColor{index: 0}
	case "31":
		s.foreground = &ansiColor{index: 1}
	case "32":
		s.foreground = &ansiColor{index: 2}
	case "33":
		s.foreground = &ansiColor{index: 3}
	case "34":
		s.foreground = &ansiColor{index: 4}
	case "35":
		s.foreground = &ansiColor{index: 5}
	case "36":
		s.foreground = &ansiColor{index: 6}
	case "37":
		s.foreground = &ansiColor{index: 7}
	case "39":
		s.foreground = nil
	case "40":
		s.background = &ansiColor{index: 0}
	case "41":
		s.background = &ansiColor{index: 1}
	case "42":
		s.background = &ansiColor{index: 2}
	case "43":
		s.background = &ansiColor{index: 3}
	case "44":
		s.background = &ansiColor{index: 4}
	case "45":
		s.background = &ansiColor{index: 5}
	case "46":
		s.background = &ansiColor{index: 6}
	case "47":
		s.background = &ansiColor{index: 7}
	case "49":
		s.background = nil

	}
	return s
}

type parsedANSI struct {
	text  string
	state ansiState
}

// parse parses the given text and returns the parsed text, the new state, and the remaining text.
func parseANSI(text string, state ansiState) []parsedANSI {
	ret := []parsedANSI{}
	cur := strings.Builder{}
	escape := false

	output := func(force bool) {
		if cur.Len() > 0 || force {
			ret = append(ret, parsedANSI{
				text:  cur.String(),
				state: state,
			})
			cur.Reset()
		}
	}

	for i := 0; i < len(text); i++ {
		if text[i] == '\x1b' {
			if len(text) > i+1 && text[i+1] == '[' {
				i++
				escape = true
				output(false)
				continue
			}
		}
		if escape {
			if text[i] >= 0x40 && text[i] <= 0x7e {
				escape = false
				commands := cur.String()
				cur.Reset()
				if text[i] == 'm' {
					for _, command := range strings.Split(commands, ";") {
						state = state.update(command)
					}
				}
				continue
			}
		}
		cur.WriteByte(text[i])
	}
	output(true)
	return ret
}

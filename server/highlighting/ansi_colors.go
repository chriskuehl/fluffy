package highlighting

import (
	"fmt"
	"html"
	"html/template"
	"strings"
)

type ANSIColors struct {
	Black         Color
	Red           Color
	Green         Color
	Yellow        Color
	Blue          Color
	Magenta       Color
	Cyan          Color
	White         Color
	BrightBlack   Color
	BrightRed     Color
	BrightGreen   Color
	BrightYellow  Color
	BrightBlue    Color
	BrightMagenta Color
	BrightCyan    Color
	BrightWhite   Color
}

type ANSIColorSet struct {
	Foreground ANSIColors
	Background ANSIColors
}

var ansiColorsLight = &ANSIColorSet{
	Foreground: ANSIColors{
		Black:         "#000000",
		Red:           "#EF2929",
		Green:         "#62CA00",
		Yellow:        "#DAC200",
		Blue:          "#3465A4",
		Magenta:       "#CE42BE",
		Cyan:          "#34E2E2",
		White:         "#FFFFFF",
		BrightBlack:   "#676767",
		BrightRed:     "#ff6d67",
		BrightGreen:   "#5ff967",
		BrightYellow:  "#fefb67",
		BrightBlue:    "#6871ff",
		BrightMagenta: "#ff76ff",
		BrightCyan:    "#5ffdff",
		BrightWhite:   "#feffff",
	},
	Background: ANSIColors{
		Black:         "#000000",
		Red:           "#EF2929",
		Green:         "#8AE234",
		Yellow:        "#FCE94F",
		Blue:          "#3465A4",
		Magenta:       "#C509C5",
		Cyan:          "#34E2E2",
		White:         "#FFFFFF",
		BrightBlack:   "#676767",
		BrightRed:     "#ff6d67",
		BrightGreen:   "#5ff967",
		BrightYellow:  "#fefb67",
		BrightBlue:    "#6871ff",
		BrightMagenta: "#ff76ff",
		BrightCyan:    "#5ffdff",
		BrightWhite:   "#feffff",
	},
}

var ansiColorsDark = &ANSIColorSet{
	Foreground: ANSIColors{
		Black:         "#555753",
		Red:           "#FF5C5C",
		Green:         "#8AE234",
		Yellow:        "#FCE94F",
		Blue:          "#8FB6E1",
		Magenta:       "#FF80F1",
		Cyan:          "#34E2E2",
		White:         "#EEEEEC",
		BrightBlack:   "#676767",
		BrightRed:     "#ff6d67",
		BrightGreen:   "#5ff967",
		BrightYellow:  "#fefb67",
		BrightBlue:    "#6871ff",
		BrightMagenta: "#ff76ff",
		BrightCyan:    "#5ffdff",
		BrightWhite:   "#feffff",
	},
	Background: ANSIColors{
		Black:         "#555753",
		Red:           "#F03D3D",
		Green:         "#6ABC1B",
		Yellow:        "#CEB917",
		Blue:          "#6392C6",
		Magenta:       "#FF80F1",
		Cyan:          "#2FC0C0",
		White:         "#BFBFBF",
		BrightBlack:   "#676767",
		BrightRed:     "#ff6d67",
		BrightGreen:   "#5ff967",
		BrightYellow:  "#fefb67",
		BrightBlue:    "#6871ff",
		BrightMagenta: "#ff76ff",
		BrightCyan:    "#5ffdff",
		BrightWhite:   "#feffff",
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
	if s.faint {
		ret["opacity"] = "0.5"
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
		ret["color"] = fmt.Sprintf(
			"rgb(%d, %d, %d)",
			s.foreground.rgb.r,
			s.foreground.rgb.g,
			s.foreground.rgb.b,
		)
	}
	if s.background != nil && s.background.rgb != nil {
		ret["background-color"] = fmt.Sprintf(
			"rgb(%d, %d, %d)",
			s.background.rgb.r,
			s.background.rgb.g,
			s.background.rgb.b,
		)
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

func ansi256Color(index uint8) *ansiColor {
	switch {
	case index < 16: // Standard colors
		return &ansiColor{index: index}
	case index >= 16 && index < 232: // 6x6x6 cube (216 colors)
		index -= 16
		r := index / 36
		g := (index % 36) / 6
		b := index % 6
		return &ansiColor{rgb: &rgb{
			r: uint8(r*40 + 55),
			g: uint8(g*40 + 55),
			b: uint8(b*40 + 55),
		}}
	case index >= 232 && index <= 255: // Grayscale ramp (24 colors)
		value := uint8((index-232)*10 + 8)
		return &ansiColor{rgb: &rgb{r: value, g: value, b: value}}
	}
	panic("unreachable")
}

func (s ansiState) update(commands []string) ansiState {
outer:
	for len(commands) > 0 {
		command := commands[0]
		commands = commands[1:]
		switch command {
		case "0":
			s = ansiState{}
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
		case "30", "31", "32", "33", "34", "35", "36", "37":
			s.foreground = &ansiColor{index: uint8(command[1] - '0')}
		case "39":
			s.foreground = nil
		case "40", "41", "42", "43", "44", "45", "46", "47":
			s.background = &ansiColor{index: uint8(command[1] - '0')}
		case "49":
			s.background = nil
		case "90", "91", "92", "93", "94", "95", "96", "97":
			s.foreground = &ansiColor{index: uint8(command[1] - '0' + 8)}
		case "100", "101", "102", "103", "104", "105", "106", "107":
			s.background = &ansiColor{index: uint8(command[2] - '0' + 8)}
		case "38", "48":
			if len(commands) < 1 {
				break outer
			}
			typeCode := commands[0]
			commands = commands[1:]

			if typeCode == "5" {
				if len(commands) < 1 {
					break outer
				}
				var index uint8
				fmt.Sscanf(commands[0], "%d", &index)
				commands = commands[1:]
				if command == "38" {
					s.foreground = ansi256Color(index)
				} else {
					s.background = ansi256Color(index)
				}
			} else if typeCode == "2" {
				if len(commands) < 3 {
					break outer
				}
				r := commands[0]
				g := commands[1]
				b := commands[2]
				commands = commands[3:]
				rgb := rgb{}
				fmt.Sscanf(r, "%d", &rgb.r)
				fmt.Sscanf(g, "%d", &rgb.g)
				fmt.Sscanf(b, "%d", &rgb.b)
				if command == "38" {
					s.foreground = &ansiColor{rgb: &rgb}
				} else {
					s.background = &ansiColor{rgb: &rgb}
				}
			}
		}
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
					state = state.update(strings.Split(commands, ";"))
				}
				continue
			}
		}
		cur.WriteByte(text[i])
	}
	output(true)
	return ret
}

package main

import (
	"fmt"
	"strings"

	"github.com/chriskuehl/fluffy/server/highlighting"
)

func main() {
	for _, cat := range highlighting.Styles {
		for _, style := range cat.Styles {
			var chromaCSS strings.Builder
			highlighting.Formatter.WriteCSS(&chromaCSS, style.ChromaStyle)

			fmt.Printf(`
				.style-%s {
					/* Chroma */
					%s

					/* Fluffy UI colors */
					.line-numbers {
						background-color: %s;
						border-color: %s;
					}
					.text {
						background-color: %s;
					}
					.line-numbers a {
						color: %s;
					}
					.line-numbers a:hover {
						background-color: %s !important;
					}
					.line-numbers a.selected {
						background-color: %s;
					}
					.paste-toolbar {
						background-color: %s;
						color: %s;
					}
					.text .chroma .line.selected {
						background-color: %s;
					}
					.text .chroma .line.diff-add {
						background-color: %s;
					}
					.text .chroma .line.diff-add.selected {
						background-color: %s;
					}
					.text .chroma .line.diff-remove {
						background-color: %s;
					}
					.text .chroma .line.diff-remove.selected {
						background-color: %s;
					}

					/* ANSI colors */
					.text .chroma .fg-0 { color: %s; }
					.text .chroma .fg-1 { color: %s; }
					.text .chroma .fg-2 { color: %s; }
					.text .chroma .fg-3 { color: %s; }
					.text .chroma .fg-4 { color: %s; }
					.text .chroma .fg-5 { color: %s; }
					.text .chroma .fg-6 { color: %s; }
					.text .chroma .fg-7 { color: %s; }
					.text .chroma .fg-8 { color: %s; }
					.text .chroma .fg-9 { color: %s; }
					.text .chroma .fg-10 { color: %s; }
					.text .chroma .fg-11 { color: %s; }
					.text .chroma .fg-12 { color: %s; }
					.text .chroma .fg-13 { color: %s; }
					.text .chroma .fg-14 { color: %s; }
					.text .chroma .fg-15 { color: %s; }
					.text .chroma .bg-0 { background-color: %s; }
					.text .chroma .bg-1 { background-color: %s; }
					.text .chroma .bg-2 { background-color: %s; }
					.text .chroma .bg-3 { background-color: %s; }
					.text .chroma .bg-4 { background-color: %s; }
					.text .chroma .bg-5 { background-color: %s; }
					.text .chroma .bg-6 { background-color: %s; }
					.text .chroma .bg-7 { background-color: %s; }
					.text .chroma .bg-8 { background-color: %s; }
					.text .chroma .bg-9 { background-color: %s; }
					.text .chroma .bg-10 { background-color: %s; }
					.text .chroma .bg-11 { background-color: %s; }
					.text .chroma .bg-12 { background-color: %s; }
					.text .chroma .bg-13 { background-color: %s; }
					.text .chroma .bg-14 { background-color: %s; }
					.text .chroma .bg-15 { background-color: %s; }

				}
				`,
				style.Name,
				chromaCSS.String(),
				style.FluffyColors.LineNumbersBackground,
				style.FluffyColors.Border,
				style.FluffyColors.Border,
				style.FluffyColors.LineNumbersForeground,
				style.FluffyColors.LineNumbersHoverBackground,
				style.FluffyColors.LineNumbersSelectedBackground,
				style.FluffyColors.ToolbarBackground,
				style.FluffyColors.ToolbarForeground,
				style.FluffyColors.SelectedLineBackground,
				style.FluffyColors.DiffAddLineBackground,
				style.FluffyColors.DiffAddSelectedLineBackground,
				style.FluffyColors.DiffRemoveLineBackground,
				style.FluffyColors.DiffRemoveSelectedLineBackground,
				style.ANSIColors.Foreground.Black,
				style.ANSIColors.Foreground.Red,
				style.ANSIColors.Foreground.Green,
				style.ANSIColors.Foreground.Yellow,
				style.ANSIColors.Foreground.Blue,
				style.ANSIColors.Foreground.Magenta,
				style.ANSIColors.Foreground.Cyan,
				style.ANSIColors.Foreground.White,
				style.ANSIColors.Background.BrightBlack,
				style.ANSIColors.Background.BrightRed,
				style.ANSIColors.Background.BrightGreen,
				style.ANSIColors.Background.BrightYellow,
				style.ANSIColors.Background.BrightBlue,
				style.ANSIColors.Background.BrightMagenta,
				style.ANSIColors.Background.BrightCyan,
				style.ANSIColors.Background.BrightWhite,
				style.ANSIColors.Background.Black,
				style.ANSIColors.Background.Red,
				style.ANSIColors.Background.Green,
				style.ANSIColors.Background.Yellow,
				style.ANSIColors.Background.Blue,
				style.ANSIColors.Background.Magenta,
				style.ANSIColors.Background.Cyan,
				style.ANSIColors.Background.White,
				style.ANSIColors.Background.BrightBlack,
				style.ANSIColors.Background.BrightRed,
				style.ANSIColors.Background.BrightGreen,
				style.ANSIColors.Background.BrightYellow,
				style.ANSIColors.Background.BrightBlue,
				style.ANSIColors.Background.BrightMagenta,
				style.ANSIColors.Background.BrightCyan,
				style.ANSIColors.Background.BrightWhite,
			)
		}
	}
}

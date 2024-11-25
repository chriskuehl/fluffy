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
					.text .chroma .fg-0 {
						color: %s;
					}
					.text .chroma .fg-0-faint {
						color: %s;
					}
					.text .chroma .fg-1 {
						color: %s;
					}
					.text .chroma .fg-1-faint {
						color: %s;
					}
					.text .chroma .fg-2 {
						color: %s;
					}
					.text .chroma .fg-2-faint {
						color: %s;
					}
					.text .chroma .fg-3 {
						color: %s;
					}
					.text .chroma .fg-3-faint {
						color: %s;
					}
					.text .chroma .fg-4 {
						color: %s;
					}
					.text .chroma .fg-4-faint {
						color: %s;
					}
					.text .chroma .fg-5 {
						color: %s;
					}
					.text .chroma .fg-5-faint {
						color: %s;
					}
					.text .chroma .fg-6 {
						color: %s;
					}
					.text .chroma .fg-6-faint {
						color: %s;
					}
					.text .chroma .fg-7 {
						color: %s;
					}
					.text .chroma .fg-7-faint {
						color: %s;
					}
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
				style.ANSIColors.ForegroundFaint.Black,
				style.ANSIColors.Foreground.Red,
				style.ANSIColors.ForegroundFaint.Red,
				style.ANSIColors.Foreground.Green,
				style.ANSIColors.ForegroundFaint.Green,
				style.ANSIColors.Foreground.Yellow,
				style.ANSIColors.ForegroundFaint.Yellow,
				style.ANSIColors.Foreground.Blue,
				style.ANSIColors.ForegroundFaint.Blue,
				style.ANSIColors.Foreground.Magenta,
				style.ANSIColors.ForegroundFaint.Magenta,
				style.ANSIColors.Foreground.Cyan,
				style.ANSIColors.ForegroundFaint.Cyan,
				style.ANSIColors.Foreground.White,
				style.ANSIColors.ForegroundFaint.White,
			)
		}
	}
}

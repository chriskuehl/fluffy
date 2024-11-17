package main

import (
	"fmt"
	"os"

	"github.com/chriskuehl/fluffy/server/highlighting"
)

func main() {
	for _, cat := range highlighting.Styles {
		for _, style := range cat.Styles {
			fmt.Printf(".style-%s {\n", style.ChromaStyle.Name)
			highlighting.Formatter.WriteCSS(os.Stdout, style.ChromaStyle)
			fmt.Printf(".line-numbers {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.LineNumbersBackground)
			fmt.Printf("  border-color: %s;\n", style.FluffyColors.Border)
			fmt.Printf("}\n")
			fmt.Printf(".text {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.Border)
			fmt.Printf("}\n")
			fmt.Printf(".line-numbers a {\n")
			fmt.Printf("  color: %s;\n", style.FluffyColors.LineNumbersForeground)
			fmt.Printf("}\n")
			fmt.Printf(".line-numbers a:hover {\n")
			fmt.Printf("  background-color: %s !important;\n", style.FluffyColors.LineNumbersHoverBackground)
			fmt.Printf("}\n")
			fmt.Printf(".line-numbers a.selected {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.LineNumbersSelectedBackground)
			fmt.Printf("}\n")
			fmt.Printf(".paste-toolbar {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.ToolbarBackground)
			fmt.Printf("  color: %s;\n", style.FluffyColors.ToolbarForeground)
			fmt.Printf("}\n")
			fmt.Printf(".text .highlight > pre > span.selected {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.SelectedLineBackground)
			fmt.Printf("}\n")
			fmt.Printf(".text .highlight > pre > span.diff-add {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.DiffAddLineBackground)
			fmt.Printf("}\n")
			fmt.Printf(".text .highlight > pre > span.diff-add.selected {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.DiffAddSelectedLineBackground)
			fmt.Printf("}\n")
			fmt.Printf(".text .highlight > pre > span.diff-remove {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.DiffRemoveLineBackground)
			fmt.Printf("}\n")
			fmt.Printf(".text .highlight > pre > span.diff-remove.selected {\n")
			fmt.Printf("  background-color: %s;\n", style.FluffyColors.DiffRemoveSelectedLineBackground)
			fmt.Printf("}\n")
			fmt.Printf("}\n")
		}
	}
}

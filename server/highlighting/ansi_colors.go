package highlighting

type ansiColors struct {
	black   Color
	red     Color
	green   Color
	yellow  Color
	blue    Color
	magenta Color
	cyan    Color
	white   Color
}

type ansiColorSet struct {
	foreground ansiColors
	background ansiColors
}

var ansiColorsLight = &ansiColorSet{
	foreground: ansiColors{
		black:   "#000000",
		red:     "#EF2929",
		green:   "#62CA00",
		yellow:  "#DAC200",
		blue:    "#3465A4",
		magenta: "#CE42BE",
		cyan:    "#34E2E2",
		white:   "#FFFFFF",
	},
	background: ansiColors{
		black:   "#000000",
		red:     "#EF2929",
		green:   "#8AE234",
		yellow:  "#FCE94F",
		blue:    "#3465A4",
		magenta: "#C509C5",
		cyan:    "#34E2E2",
		white:   "#FFFFFF",
	},
}

var ansiColorsDark = &ansiColorSet{
	foreground: ansiColors{
		black:   "#555753",
		red:     "#FF5C5C",
		green:   "#8AE234",
		yellow:  "#FCE94F",
		blue:    "#8FB6E1",
		magenta: "#FF80F1",
		cyan:    "#34E2E2",
		white:   "#EEEEEC",
	},
	background: ansiColors{
		black:   "#555753",
		red:     "#F03D3D",
		green:   "#6ABC1B",
		yellow:  "#CEB917",
		blue:    "#6392C6",
		magenta: "#FF80F1",
		cyan:    "#2FC0C0",
		white:   "#BFBFBF",
	},
}

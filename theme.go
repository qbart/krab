package main

import (
	"github.com/gdamore/tcell"
)

// Theme keeps editor colors.
type Theme struct {
	cursorColor      tcell.Color
	bgColor          tcell.Color
	fgColor          tcell.Color
	highlightFgColor tcell.Color
	highlightBgColor tcell.Color
	footerBgColor    tcell.Color
	zebraRowColor    tcell.Color
	zebraRowAltColor tcell.Color
}

// NewTheme returns new theme with default style.
func NewTheme() *Theme {
	return &Theme{
		cursorColor:      tcell.ColorRed,
		bgColor:          tcell.NewRGBColor(38, 39, 47),
		fgColor:          tcell.NewRGBColor(70, 73, 90),
		highlightFgColor: tcell.NewRGBColor(255, 255, 255),
		highlightBgColor: tcell.NewRGBColor(70, 73, 90),
		footerBgColor:    tcell.NewRGBColor(70, 73, 90), //;tcell.NewRGBColor(92, 139, 154)
		zebraRowColor:    tcell.NewRGBColor(147, 112, 219),
		zebraRowAltColor: tcell.NewRGBColor(32, 178, 170),
	}
}

func (theme *Theme) RunnableRegionColorByIndex(index int) tcell.Color {
	if index == 0 {
		return theme.fgColor
	}

	switch index % 2 {
	case 0:
		return theme.zebraRowColor
	case 1:
		return theme.zebraRowAltColor
	}

	return theme.fgColor
}

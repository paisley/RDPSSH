package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type DarkGreenTheme struct{}

var _ fyne.Theme = (*DarkGreenTheme)(nil)

func (m DarkGreenTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameSuccess {
		// Dark Green #006400
		return color.RGBA{R: 0, G: 100, B: 0, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m DarkGreenTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m DarkGreenTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m DarkGreenTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

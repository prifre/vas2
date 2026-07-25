package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomTheme struct {
	fyne.Theme
}

func (m CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// 1. Bakgrundsfärg på knappen
	if name == theme.ColorNameButton {
		return color.RGBA{R: 50, G: 55, B: 65, A: 255}
	}

	// 2. Textfärg på knapparna (Tydlig vit text)
	if name == theme.ColorNameForeground {
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	return m.Theme.Color(name, variant)
}

// NewCustomTheme skapar och returnerar ditt anpassade tema
func NewCustomTheme() fyne.Theme {
	return &CustomTheme{Theme: theme.DefaultTheme()}
}

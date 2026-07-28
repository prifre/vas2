package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type CustomTheme struct {
	fyne.Theme
}

func NewCustomTheme() fyne.Theme {
	return &CustomTheme{Theme: theme.DefaultTheme()}
}

// Tvinga Fyne att behandla detta som ett mörkt tema i Windows:
func (m CustomTheme) Variant() fyne.ThemeVariant {
	return theme.VariantDark
}

func (m CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {

	// 1. All text i appen och i inmatningsfälten (Vit)
	case theme.ColorNameForeground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}

	// 2. Bakgrunden INUTI inmatningsfälten (Mörkgrå så vit text syns tydligt!)
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 45, G: 50, B: 58, A: 255}

	// 3. Hjälptext/Placeholder i fält när de är tomma
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 150, G: 150, B: 150, A: 255}

	// 4. Bakgrund för toppmeny, popup-dialoger och fönster
	case theme.ColorNameBackground, theme.ColorNameMenuBackground, theme.ColorNameOverlayBackground, theme.ColorNameHeaderBackground:
		return color.RGBA{R: 30, G: 35, B: 40, A: 255}

	// 5. Knappar
	case theme.ColorNameButton:
		return color.RGBA{R: 20, G: 25, B: 30, A: 255}

	// 6. Markerad text / Hover
	case theme.ColorNameHover, theme.ColorNameSelection:
		return color.RGBA{R: 50, G: 65, B: 80, A: 255}

	// Fallback
	default:
		return m.Theme.Color(name, variant)
	}
}

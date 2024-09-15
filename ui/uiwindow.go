package ui

import "fyne.io/fyne/v2"

func WindowSetup(w fyne.Window) {
	w1 := fyne.CurrentApp().Preferences().FloatWithFallback("winWidth", 1024)
	h1 := fyne.CurrentApp().Preferences().FloatWithFallback("winHeight", 768)
	if w1 < 800 {
		w1 = 800
	}
	if h1 < 600 {
		h1 = 600
	}
	w.Resize(fyne.NewSize(float32(w1), float32(h1)))
}

func (g *game) windowResize() {
	// if g.window.Canvas().Size().Width < 5 || g.window.Canvas().Size().Height < 5 {
	// 	return
	// } else {
	// 	g.winSize.Width = g.window.Canvas().Size().Width
	// 	g.winSize.Height = g.window.Canvas().Size().Height
	// }
	// switch g.countrunning() {
	// case 0, 1:
	// 	g.chartSize.Width = g.winSize.Width
	// 	g.chartSize.Height = g.winSize.Height
	// case 2:
	// 	g.chartSize = g.winSize
	// 	g.chartSize.Width = g.winSize.Width / 2
	// case 6:
	// 	g.chartSize.Width = g.winSize.Width / 3
	// 	g.chartSize.Height = g.winSize.Height / 2
	// case 7, 8:
	// 	g.chartSize.Width = g.winSize.Width / 4
	// 	g.chartSize.Height = g.winSize.Height / 2
	// }
}

//Clears current settings and then searches for instruments and sets Prefernces to found ports

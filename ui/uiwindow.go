package ui

func WindowSetup() w *fyne.Window {
w1 := fyne.CurrentApp().Preferences().FloatWithFallback("winWidth", 1024)
h1 := fyne.CurrentApp().Preferences().FloatWithFallback("winHeight", 768)
if w1 < 800 {
	w1 = 800
}
if h1 < 600 {
	h1 = 600
}
w.Resize(fyne.NewSize(float32(w1), float32(h1)))
return w
}
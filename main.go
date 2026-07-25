package main

import (
	"vas2/ui"

	"fyne.io/fyne/v2/app"
)

func main() {
	a := app.NewWithID("vas2")
	a.Settings().SetTheme(ui.NewCustomTheme())

	w := a.NewWindow("vas2")

	// Bygg UI och tillstånd
	ui.AppState(w)

	// Visa fönstret och starta huvudloopen
	w.ShowAndRun()
}

package general

import (
	"fmt"
	"os"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

func Getmemoryinfo() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var r string
	r += fmt.Sprintf("Memory Usage = %v MB\n", (m.Alloc / 1024 / 1024))
	r += fmt.Sprintf("Application Memory = %v MB", (m.Sys / 1024 / 1024))
	return r
}

func Getappinfo() string {
	var m string

	app := fyne.CurrentApp()

	// 1. Säker koll av Storage Root (Kraschskydd om den är nil)
	if rootURI := app.Storage().RootURI(); rootURI != nil {
		m += "\nStorage Root Name\t" + rootURI.Name()
		m += "\nStorage Root Path\t" + rootURI.Path()
	} else {
		m += "\nStorage Root\t<Not Set>"
	}

	// 2. Metadata
	meta := app.Metadata()
	m += "\nApp Name\t" + meta.Name
	m += "\nApp ID\t" + meta.ID
	m += "\nVersion\t" + meta.Version
	m += fmt.Sprintf("\nBuild\t%v", meta.Build)
	m += fmt.Sprintf("\nRelease\t%v", meta.Release)

	// 3. Mappar
	hd, _ := os.Getwd()
	m += "\nWorking Dir\t" + hd
	m += "\nGetHomeDir()\t" + GetHomeDir()
	m += "\nMemory\t\n" + Getmemoryinfo()

	// 4. Hämta ändringsdatum på den KÖRBARA BINÄREN
	if exePath, err := os.Executable(); err == nil {
		if fi, err := os.Stat(exePath); err == nil {
			m += "\nCreated date\t" + fi.ModTime().Format("2006-01-02 15:04:05")
		}
	}

	// 5. Språk
	var l fyne.Locale = lang.SystemLocale()
	m += "\nLanguageString\t" + l.LanguageString()

	return m
}

func Showlogo2() fyne.CanvasObject {
	img := canvas.NewImageFromResource(resourceLAB360LogoLightBluePng)
	img.FillMode = canvas.ImageFillContain
	return img
}

func ShowAbout(window fyne.Window) {
	// 1. Skapa textkomponenten
	t := "\nProgram for measuring with TSI\nAeroTrak, DustTrak and PTrak\n\nby Peter Freund\n\n"
	t += fmt.Sprintf("Version %v", fyne.CurrentApp().Metadata().Version)

	lbl := widget.NewLabel(t)
	lbl.Alignment = fyne.TextAlignCenter // Centrera texten under loggan

	// 2. Hämta loggan och ge den en bra storlek
	logo := Showlogo2()
	logoContainer := container.NewGridWrap(fyne.NewSize(200, 100), logo)

	// 3. Lägg ihop loggan och texten vertikalt
	content := container.NewVBox(
		logoContainer,
		lbl,
	)

	// 4. Visa anpassad dialogruta med en "OK"-knapp
	dialog.ShowCustom("About...", "OK", content, window)
}

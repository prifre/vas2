package general

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

func SetupEnvironment(window fyne.Window) {
	// Sätt CloseIntercept för snygg nedstängning
	window.SetCloseIntercept(func() {
		Closeapp(window)
	})

	SetupFiles()
	SetupLogging()
	SetupWindow(window)

	log.Println("SetupEnvironment finished.")
}

func SetupLogging() {
	// Hämtar konfigurerad documentpath (eller fallback till home dir)
	documentPath := fyne.CurrentApp().Preferences().StringWithFallback("documentpath", GetHomeDir())
	logFileName := fmt.Sprintf("%s.log", fyne.CurrentApp().UniqueID())
	lfn := filepath.Join(documentPath, logFileName)

	fyne.CurrentApp().Preferences().SetString("vaslog", lfn)

	logfile, err := os.OpenFile(lfn, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("#2 SetupLogging: OpenFile error:", err)
		return
	}

	// Om debug är påslaget, skicka loggen till både fil och konsol (Stdout)
	if fyne.CurrentApp().Preferences().Bool("mydebug") {
		multi := io.MultiWriter(logfile, os.Stdout)
		log.SetOutput(multi)
	} else {
		log.SetOutput(logfile)
	}

	log.Println("Logging started:", lfn)
}

func SetupWindow(w fyne.Window) {
	w1 := fyne.CurrentApp().Preferences().FloatWithFallback("winWidth", 1024)
	h1 := fyne.CurrentApp().Preferences().FloatWithFallback("winHeight", 768)

	if w1 < 800 {
		w1 = 800
	}
	if h1 < 600 {
		h1 = 600
	}

	w.Resize(fyne.NewSize(float32(w1), float32(h1)))
	w.SetMaster()
	w.CenterOnScreen()
}

func SetupFiles() {
	documentPath := fyne.CurrentApp().Preferences().StringWithFallback("documentpath", GetHomeDir())
	if documentPath == "" {
		documentPath = GetHomeDir()
	}

	homeDir := GetHomeDir()

	// Spara inställningar och sökvägar med filepath.Join
	prefs := fyne.CurrentApp().Preferences()
	prefs.SetString("documentpath", documentPath)
	prefs.SetString("homedir", homeDir)
	prefs.SetString("dbfilename", filepath.Join(documentPath, "vasdatabase.db"))
	prefs.SetString("excelfile", filepath.Join(documentPath, "vas.xlsx"))
}

package general

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
)

// GetHomeDir sätter upp och returnerar sökvägen till datamappen "vasdata".
func GetHomeDir() string {
	var err error
	var path string

	if fyne.CurrentApp().Preferences().Bool("mydebug") {
		path, err = os.Getwd()
		if err != nil {
			log.Println("#1 GetHomeDir Failed to get Getwd:", err)
			path = "."
		}
	} else {
		path, err = os.UserHomeDir()
		if err != nil {
			log.Println("#2 GetHomeDir Failed to get UserHomeDir:", err)
			path = "."
		}
	}

	if IsTesting() {
		exePath, err := os.Executable()
		if err == nil {
			dir := filepath.Dir(exePath)
			// Klättra uppåt i katalogträdet tills vi hittar "vas2"
			for {
				parent := filepath.Dir(dir)
				if filepath.Base(dir) == "vas2" || parent == dir {
					break
				}
				dir = parent
			}
			path = dir
		}
	}

	// Använd filepath.Join för säkra sökvägar på både Windows och Linux
	path = filepath.Join(path, "vasdata")

	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, 0755) // MkdirAll skapar hela sökvägen om undermappar saknas
			if err != nil {
				log.Println("#3 GetHomeDir Could not make dir:", path, err)
			}
		} else {
			log.Println("#4 GetHomeDir Stat error:", path, err)
		}
	}
	return path
}
func Openstoragelocation() {
	documentpath := fyne.CurrentApp().Preferences().StringWithFallback("documentpath", GetHomeDir())
	if documentpath == "" {
		documentpath = GetHomeDir()
	}
	log.Println("Opening storage location:", documentpath)

	// Använd url.Parse för att skapa en giltig file://-URL
	u, err := url.Parse("file://" + documentpath)
	if err != nil {
		log.Println("Error parsing URL for path "+documentpath+":", err)
		return
	}

	err = fyne.CurrentApp().OpenURL(u)
	if err != nil {
		log.Println("Error opening location "+documentpath+":", err)
	}
}
func ShowMyDebug(s string) {
	if fyne.CurrentApp().Preferences().Bool("mydebug") {
		log.Printf("FUNCTION: %v\n%v\n", Printfunc(), s)
	}
}

func Printfunc() string {
	pc := make([]uintptr, 1)
	n := runtime.Callers(2, pc)
	if n == 0 {
		return "unknown"
	}
	f := runtime.FuncForPC(pc[0])
	if f == nil {
		return "unknown"
	}
	return f.Name()
}

func IsTesting() bool {
	if strings.HasSuffix(os.Args[0], ".test") {
		return true
	}
	if flag.Lookup("test.v") != nil {
		return true
	}
	return false
}

func Openurl(uu string) error {
	// Om det inte är en http/https-länk, lägg till file:// så url.Parse förstår sökvägen
	if !strings.HasPrefix(uu, "http://") && !strings.HasPrefix(uu, "https://") && !strings.HasPrefix(uu, "file://") {
		uu = "file://" + uu
	}

	u, err := url.Parse(uu)
	if err != nil {
		log.Println("#1 Openurl error parsing URI:", err)
		return err
	}
	return fyne.CurrentApp().OpenURL(u)
}
func GetcurrentIP() []string {
	host, err := os.Hostname()
	if err != nil {
		log.Println("No network/hostname error:", err)
		return nil
	}

	a1, err := net.LookupHost(host)
	if err != nil {
		log.Println("Host lookup failed:", err)
	}

	var theip []string
	for _, ipStr := range a1 {
		if a2 := net.ParseIP(ipStr); a2 != nil {
			theip = append(theip, a2.String())
		}
	}
	return theip
}

func Getint64(s string) int64 {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		log.Println("int64 conversion error:", err)
	}
	return v
}

func Closeapp(w fyne.Window) {
	if w != nil && w.Canvas() != nil {
		height := w.Canvas().Size().Height
		width := w.Canvas().Size().Width
		fmt.Println(height, width)
		fyne.CurrentApp().Preferences().SetFloat("winWidth", float64(width))
		fyne.CurrentApp().Preferences().SetFloat("winHeight", float64(height))
	}
	if fyne.CurrentApp().Preferences().Bool("mydebug") {
		log.Println("Quit application")
	}
	fyne.CurrentApp().Quit()
}
func GetWindowSize() fyne.Size {
	// Hämta sparade värden som float64 med en standardstorlek som fallback (t.ex. 1024x764)
	w := fyne.CurrentApp().Preferences().FloatWithFallback("winWidth", 1024)
	h := fyne.CurrentApp().Preferences().FloatWithFallback("winHeight", 764)

	// Konvertera till float32 för fyne.NewSize
	width := float32(w)
	height := float32(h)

	if width > 100 && height > 100 {
		return fyne.NewSize(width, height)
	}

	return fyne.NewSize(1024, 764)
}

package general

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

func Doscreenshot(currentwindow fyne.Window) string {
	var err error
	img := currentwindow.Canvas().Capture()
	time.Sleep(time.Second)
	t := time.Now().Format(time.RFC3339)
	t = strings.Replace(t, "-", "", 99)
	t = strings.Replace(t, ":", "", 99)
	fn := t[:13] + ".png"
	m := 28
	upLeft := image.Point{0, 0}
	lowRight := image.Point{img.Bounds().Max.X, img.Bounds().Max.Y - m}
	img2 := image.NewRGBA(image.Rectangle{upLeft, lowRight})
	if m > 0 {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			for x := 0; x < img.Bounds().Max.X; x++ {
				if y > m {
					img2.Set(x, y-m-1, img.At(x, y))
				}
			}
		}
	}
	err = saveimg(img2, filepath.Join(g.getdocumentpath().Path(), fn))
	if err != nil {
		return ""
	}
	return fn
}

func saveimg(img image.Image, fn string) error {
	var err error
	var f2 *os.File
	f2, err = os.Create(fn)
	if err != nil {
		return err
	}
	err = png.Encode(f2, img)
	if err != nil {
		return err
	}
	time.Sleep(time.Second)
	f2.Close()
	return nil
}

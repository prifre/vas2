package vascharts

import (
	"fmt"
	"image/color"
	"testing"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func TestFixMax(t *testing.T) {
	v := []float32{5, 12934, 11, 899, 4000, 40000000}
	a := []float32{10, 25000, 25, 1000, 5000, 50000000}
	for i := 0; i < len(v); i++ {
		x := fixmax(v[i])
		fmt.Println(v[i], " -> ", x)
		if x != a[i] {
			t.Fatalf(fmt.Sprintf("INCORRECT (%v != %v)", v[i], a[i]))
		}
	}
}

func TestDrawBox(t *testing.T) {
	app := app.NewWithID("x")
	w := app.NewWindow("Hello")
	var r *canvas.Rectangle = canvas.NewRectangle(color.White)
	fmt.Println(r)
	c := *container.NewGridWithRows(2)
	c.Add(r)
	w.SetContent(&c)
	w.ShowAndRun()
}

func TestMain(t *testing.T) {
	app := app.NewWithID("x")
	w := app.NewWindow("Hello")
	var r *canvas.Rectangle = canvas.NewRectangle(color.White)
	fmt.Println(r)
	c := *container.NewGridWithRows(2)
	c.Add(r)
	w.SetContent(&c)
	w.ShowAndRun()
}

func Test_setyaxisscale(t *testing.T) {
	var oldvt string
	var vt []string
	var tvalue float64
	tvalue = 1000000
	vt = setyaxisscale(float64(tvalue))
	for i := 0; i < len(vt); i++ {
		if vt[i] == oldvt {
			t.Fatal("error: ", vt[i], "Scale: ", vt)
		} else {
			oldvt = vt[i]
		}
	}
	t.Error(vt)
}

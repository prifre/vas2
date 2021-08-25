package main

import (
	"fmt"
	"testing"
)

func (lc *LineChart) TestFixMax(t *testing.T) {
	v := []int{5, 12934, 11, 899, 4000, 40000000}
	a := []int{10, 25000, 25, 1000, 5000, 50000000}
	for i := 0; i < len(v); i++ {
		x := lc.FixMax(v[i])
		//		fmt.Println(v[i], " -> ", x)
		if x != a[i] {
			t.Fatalf(fmt.Sprintf("INCORRECT (%v != %v)", v[i], a[i]))
		}
	}
}

func TestDrawBox(t *testing.T) {
	// app := app.NewWithID("x")
	// w := app.NewWindow("Hello")
	// var r *canvas.Rectangle = canvas.NewRectangle(color.White)
	// fmt.Println(r)
	// c := *container.NewGridWithRows(2)
	// c.Add(r)
	// w.SetContent(&c)
	// w.ShowAndRun()
}

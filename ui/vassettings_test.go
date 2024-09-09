package ui

import (
	"fmt"
	"image/color"
	"testing"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/storage/repository"
)

func TestMain3(T *testing.T) {
	t := "C:/Users"
	fmt.Println(repository.ForScheme(t))
	u, err := storage.ParseURI(t)
	fmt.Println("#####", u, err.Error())
}
func TestMain2(T *testing.T) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Border Layout")

	top := canvas.NewText("top bar", color.White)
	left := canvas.NewText("left", color.White)
	middle := canvas.NewText("content", color.White)
	content := container.New(layout.NewBorderLayout(top, nil, left, nil),
		top, left, middle)
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}

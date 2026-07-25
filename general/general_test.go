package general

import (
	"fmt"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func TestGetHomedir(t *testing.T) {
	//  Doscreenshot()
	app.NewWithID("vas2")
	fyne.CurrentApp().Preferences().SetBool("mydebug", true)
	mydebug := fyne.CurrentApp().Preferences().Bool("mydebug")
	if mydebug {
		fmt.Println("DEBUG ON!!!")
	} else {
		fmt.Println("DEBUG OFF")
	}
	hd := GetHomeDir()
	if strings.Contains(hd, "ERROR") {
		t.Fail()
	}
	fmt.Println("HomeDir set to: ", hd)
}
func TestGetappinfo(t *testing.T) {
	app.NewWithID("vas2")
	fmt.Println(Getappinfo())
	fmt.Println("RUNNING FUNCTION:\t", Printfunc())
}
func TestSetupEnvironment(t *testing.T) {
	a := app.NewWithID("vas2")
	w := a.NewWindow("vas2")
	SetupEnvironment(w)
	fmt.Println(Getappinfo())
}

package ui

/*

2021-04-03 23:23
TCP routines to handle TSI AeroTrak
2021-04-10 got it to work.

*/

import (
	"fmt"
	"log"
	"vas/vasinstruments"

	"fyne.io/fyne/v2"
)
func DoSetupInstruments() {
	var err error
	var s string
	log.Println("Searching for TSI network instruments DustTrak & AeroTrak")
	log.Println("Searching for TSI serial instrument PTrak")
	instsettings := fmt.Sprintf("AeroTrak: %v, DustTrak: %v, PTrak: %v",
		fyne.CurrentApp().Preferences().String("AeroTrak"), fyne.CurrentApp().Preferences().String("DustTrak"), fyne.CurrentApp().Preferences().String("PTrak"))
	if instsettings != "AeroTrak: , DustTrak: , PTrak: " {
		log.Println("Clearing old instruments settings!\n (" + instsettings + ")")
		fyne.CurrentApp().Preferences().SetString("AeroTrak", "")
		fyne.CurrentApp().Preferences().SetString("DustTrak", "")
		fyne.CurrentApp().Preferences().SetString("PTrak", "")
	}
	_, err = new(vasinstruments.PTraktype).FindPTrak()
	if err != nil {
		log.Println("#1 SetupInstruments:FindPTrak:\n", err.Error())
	}
	// if g.MyTSIscanner() != nil {
	// 	log.Println("#2 SetupInstruments:MyTSIscanner:\n", err.Error())
	// }
	if fyne.CurrentApp().Preferences().String("AeroTrak") > "" {
		s += "\nAeroTrak: " + fyne.CurrentApp().Preferences().String("AeroTrak")
	}
	if fyne.CurrentApp().Preferences().String("DustTrak") > "" {
		s += "\nDustTrak: " + fyne.CurrentApp().Preferences().String("DustTrak")
	}
	if fyne.CurrentApp().Preferences().String("PTrak") > "" {
		s += "\nPTrak: " + fyne.CurrentApp().Preferences().String("PTrak")
	}
	// if s == "" {
	// 	dialog.ShowInformation("", "No instruments found.", g.window)
	// } else {
	// 	log.Println("Finished autosetup of instruments: " + s)
	// 	dialog.ShowInformation("Found instruments: ", strings.Replace(s, ", ", "\n", -1), g.window)
	// }
}

package vasinstruments

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	bugserial "go.bug.st/serial"
)

func DoSetupInstruments() string {
	var err error
	var s string

	log.Println("Searching for TSI serial instrument PTrak...")

	// Hämta befintliga inställningar
	atPort := fyne.CurrentApp().Preferences().String("AeroTrak")
	dtPort := fyne.CurrentApp().Preferences().String("DustTrak")
	ptPort := fyne.CurrentApp().Preferences().String("PTrak")

	instsettings := fmt.Sprintf("AeroTrak: %v, DustTrak: %v, PTrak: %v", atPort, dtPort, ptPort)
	log.Println("Current instrument settings:", instsettings)

	// Sök efter PTrak över serieport
	pt := new(PTraktype)
	if _, err = pt.FindPTrak(); err != nil {
		log.Println("#1 SetupInstruments:FindPTrak failed:", err.Error())
	}

	// Om du vill köra nätverksscannern för TSI-instrument:
	if err := MyTSIscanner(); err != nil {
		log.Println("#2 SetupInstruments:MyTSIscanner failed:", err.Error())
	}

	// Läs om uppdaterade preferences efter sökningen
	atPort = fyne.CurrentApp().Preferences().String("AeroTrak")
	dtPort = fyne.CurrentApp().Preferences().String("DustTrak")
	ptPort = fyne.CurrentApp().Preferences().String("PTrak")

	if atPort != "" {
		s += "\nAeroTrak: " + atPort
	}
	if dtPort != "" {
		s += "\nDustTrak: " + dtPort
	}
	if ptPort != "" {
		s += "\nPTrak: " + ptPort
	}

	if s == "" {
		s = "No instruments found."
	}

	log.Println("SetupInstruments complete. Found:" + s)

	return s
}
func GetallSerialPorts() string {
	// 1. Hämta alla serieportar i datorn
	ports, err := bugserial.GetPortsList()
	var msg string
	if err != nil || len(ports) == 0 {
		msg = ""
	} else {
		msg = ""
		for _, p := range ports {
			msg += fmt.Sprintf("%s\n", p)
		}
	}
	return msg
}

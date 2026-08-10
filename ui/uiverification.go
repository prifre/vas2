package ui

import (
	"fmt"
	"io"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var logfiledata string = ""

func ShowInstumentVerification(parentWindow fyne.Window) error {
	var manualDialog dialog.Dialog
	prefs := fyne.CurrentApp().Preferences()
	logfiledata = getLatestLogEntry()

	atport := prefs.StringWithFallback("AeroTrak", "")
	allResult := widget.NewMultiLineEntry()
	allResult.SetMinRowsVisible(10) // Sätt lagom rader för att undvika layout-krockar

	// Helt trådsäker uppdatering av textfönstret
	appendResult := func(text string) {
		logText := getLatestLogEntry()
		fyne.Do(func() {
			current := allResult.Text
			if current != "" && current[len(current)-1] != '\n' {
				current += "\n"
			}
			allResult.SetText(current + text + logText)
		})
	}

	// --- AeroTrak ---
	atLabel := widget.NewLabel("AeroTrak")
	atSetup := widget.NewButton("Setup", func() {
		go func() {
			r := ""
			err := ActiveMeasurement.AT.AeroTrakSetup()
			if err != nil {
				r = " Could not connect to AeroTrak at " + atport + "\n"
			}
			v, dataErr := ActiveMeasurement.AT.GetAeroTrakdata()
			if dataErr != nil {
				r += " Error getting AeroTrak data: " + dataErr.Error() + "\n"
			} else {
				r += fmt.Sprintf(" AeroTrak data: %v\n", v)
			}
			appendResult(fmt.Sprintf("AeroTrak Setup:\n%s", r))
		}()
	})

	atStart := widget.NewButton("Start", func() {
		go func() {
			r := ""
			err := ActiveMeasurement.AT.AeroTrakStart()
			if err != nil {
				r = " Could not connect to AeroTrak at " + atport + "\n"
			} else {
				r = " AeroTrak started at " + atport + "\n"
			}
			v, dataErr := ActiveMeasurement.AT.GetAeroTrakdata()
			if dataErr != nil {
				r += " Error getting AeroTrak data: " + dataErr.Error() + "\n"
			} else {
				r += fmt.Sprintf(" AeroTrak data: %v\n", v)
			}
			appendResult(fmt.Sprintf("AeroTrak Start:\n%s", r))
		}()
	})

	atGetData := widget.NewButton("Get Data", func() {
		go func() {
			r := ""
			v, dataErr := ActiveMeasurement.AT.GetAeroTrakdata()
			if dataErr != nil {
				r += " Error getting AeroTrak data: " + dataErr.Error() + "\n"
			} else {
				r += fmt.Sprintf(" AeroTrak data: %v\n", v)
			}
			appendResult(fmt.Sprintf("AeroTrak Get Data:\n%s", r))
		}()
	})

	atStop := widget.NewButton("Stop", func() {
		go func() {
			err := ActiveMeasurement.AT.AeroTrakStop()
			out := ""
			if err != nil {
				out = fmt.Sprintf(" AeroTrak Stop Error: %s\n", err.Error())
			} else {
				out = " AeroTrak Stopped\n"
			}
			appendResult("AeroTrak Stop:\n" + out)
		}()
	})

	// --- DustTrak ---
	dtport := prefs.StringWithFallback("DustTrak", "")
	dtLabel := widget.NewLabel("DustTrak")

	dtStart := widget.NewButton("Start", func() {
		go func() {
			r := ""
			err := ActiveMeasurement.DT.DustTrakStart()
			if err != nil {
				r = " Could not connect to DustTrak at " + dtport + "\n"
			} else {
				r = " DustTrak started at " + dtport + "\n"
			}
			v, dataErr := ActiveMeasurement.DT.GetDustTrakdata()
			if dataErr != nil {
				r += " Error getting DustTrak data: " + dataErr.Error() + "\n"
			} else {
				r += fmt.Sprintf(" DustTrak data: %v\n", v)
			}
			appendResult(fmt.Sprintf("DustTrak Start:\n%s", r))
		}()
	})

	dtGetData := widget.NewButton("Get Data", func() {
		go func() {
			r := ""
			v, dataErr := ActiveMeasurement.DT.GetDustTrakdata()
			if dataErr != nil {
				r += " Error getting DustTrak data: " + dataErr.Error() + "\n"
			} else {
				r += fmt.Sprintf(" DustTrak data: %v\n", v)
			}
			appendResult(fmt.Sprintf("DustTrak Get Data:\n%s", r))
		}()
	})

	dtStop := widget.NewButton("Stop", func() {
		go func() {
			err := ActiveMeasurement.DT.DustTrakstop()
			out := ""
			if err != nil {
				out = fmt.Sprintf(" DustTrak Stop Error: %s\n", err.Error())
			} else {
				out = " DustTrak Stopped\n"
			}
			appendResult("DustTrak Stop:\n" + out)
		}()
	})

	// --- PTrak ---
	ptport := prefs.StringWithFallback("PTrak", "")
	ptLabel := widget.NewLabel("PTrak")

	ptStart := widget.NewButton("Start", func() {
		go func() {
			r := ""
			err := ActiveMeasurement.PT.PTrakStart()
			if err != nil {
				r = " Could not connect to PTrak at " + ptport + "\n"
			} else {
				r = " PTrak started at " + ptport + "\n"
			}
			v, dataErr := ActiveMeasurement.PT.GetPTrakdata()
			if dataErr != nil {
				r += " Error getting PTrak data: " + dataErr.Error() + "\n"
			} else {
				r += fmt.Sprintf(" PTrak data: %v\n", v)
			}
			appendResult(fmt.Sprintf("PTrak Start:\n%s", r))
		}()
	})

	ptGetData := widget.NewButton("Get Data", func() {
		go func() {
			r := ""
			v, dataErr := ActiveMeasurement.PT.GetPTrakdata()
			if dataErr != nil {
				r += " Error getting PTrak data: " + dataErr.Error() + "\n"
			} else {
				r += fmt.Sprintf(" PTrak data: %v\n", v)
			}
			appendResult(fmt.Sprintf("PTrak Get Data:\n%s", r))
		}()
	})

	ptStop := widget.NewButton("Stop", func() {
		go func() {
			err := ActiveMeasurement.PT.PTrakStop()
			out := ""
			if err != nil {
				out = fmt.Sprintf(" PTrak Stop Error: %s\n", err.Error())
			} else {
				out = " PTrak Stopped\n"
			}
			appendResult("PTrak Stop:\n" + out)
		}()
	})

	okbutton := widget.NewButton("OK", func() {
		if manualDialog != nil {
			manualDialog.Hide()
		}
	})

	atGroup := container.NewGridWithColumns(5, atLabel, atSetup, atStart, atGetData, atStop)
	dtGroup := container.NewGridWithColumns(4, dtLabel, dtStart, dtGetData, dtStop)
	ptGroup := container.NewGridWithColumns(4, ptLabel, ptStart, ptGetData, ptStop)

	// Använd Border-layout för att hålla loggfönstret flexibelt och OK-knappen fixerad i botten
	scrollResult := container.NewScroll(allResult)
	controls := container.NewVBox(atGroup, dtGroup, ptGroup)
	content := container.NewBorder(controls, okbutton, nil, nil, scrollResult)

	manualDialog = dialog.NewCustomWithoutButtons("Instrument Verification", content, parentWindow)
	manualDialog.Resize(fyne.NewSize(750, 500))
	manualDialog.Show()

	return nil
}
func getLatestLogEntry() string {
	// 1. Slå av loggen under tiden vi läser
	log.SetOutput(io.Discard)

	// 2. Garantera att loggningen slås på igen när funktionen är KLAR
	// (Körs sist av allt, oavsett om vi gör en tidig 'return')
	defer restoreLogging()

	prefs := fyne.CurrentApp().Preferences()
	logFileName := prefs.StringWithFallback("logfile", "")

	// Läs filen
	data, err := os.ReadFile(logFileName)
	if err != nil {
		return ""
	}

	newdata := string(data)

	// Om inga ändringar har skett
	if newdata == logfiledata {
		return ""
	}

	// Om filen har blivit mindre (t.ex. rensats/roterats)
	if len(newdata) <= len(logfiledata) {
		logfiledata = newdata
		return logfiledata
	}

	// Plocka ut det nya
	addedText := newdata[len(logfiledata):]
	logfiledata = newdata

	return addedText
}

// Hjälpfunktion för att återställa loggningen på ett säkert sätt
func restoreLogging() {
	prefs := fyne.CurrentApp().Preferences()
	logFileName := prefs.StringWithFallback("logfile", "")

	logfile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// Om det misslyckas, logga till konsolen i alla fall så det inte blir helt tyst
		log.SetOutput(os.Stdout)
		return
	}

	if prefs.Bool("mydebug") {
		multi := io.MultiWriter(logfile, os.Stdout)
		log.SetOutput(multi)
	} else {
		log.SetOutput(logfile)
	}
}

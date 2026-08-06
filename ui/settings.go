package ui

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"
	"vas2/general"
	"vas2/vascharts"
	"vas2/vasinstruments"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"
)

func createColorIcon(c color.Color) fyne.Resource {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for x := 0; x < 16; x++ {
		for y := 0; y < 16; y++ {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)

	return fyne.NewStaticResource("color_icon.png", buf.Bytes())
}
func SetStrokeColor(c int, cc color.Color) {
	r, g, b, a := cc.RGBA()
	prefs := fyne.CurrentApp().Preferences()
	// Skifta ned från 16-bit till 8-bit (0-255)
	prefs.SetInt(fmt.Sprintf("%vStrokeColorR", c), int(r>>8))
	prefs.SetInt(fmt.Sprintf("%vStrokeColorG", c), int(g>>8))
	prefs.SetInt(fmt.Sprintf("%vStrokeColorB", c), int(b>>8))
	prefs.SetInt(fmt.Sprintf("%vStrokeColorA", c), int(a>>8))
}
func GetStrokeColor(c int) color.Color {
	var ChartColors = [8]color.Color{
		colornames.Green, colornames.Yellow, colornames.Red, colornames.Red,
		colornames.Red, colornames.Red, colornames.Red, colornames.Red,
	}
	r2, g2, b2, a2 := ChartColors[c].RGBA()
	prefs := fyne.CurrentApp().Preferences()

	r1 := uint8(prefs.IntWithFallback(fmt.Sprintf("%vStrokeColorR", c), int(r2>>8)))
	g1 := uint8(prefs.IntWithFallback(fmt.Sprintf("%vStrokeColorG", c), int(g2>>8)))
	b1 := uint8(prefs.IntWithFallback(fmt.Sprintf("%vStrokeColorB", c), int(b2>>8)))
	a1 := uint8(prefs.IntWithFallback(fmt.Sprintf("%vStrokeColorA", c), int(a2>>8)))

	return color.RGBA{R: r1, G: g1, B: b1, A: a1}
}
func DoSettings(parentWindow fyne.Window) error {
	var settingsDialog dialog.Dialog
	var pruningoptions = []string{"Save 12/min", "Save 6/min", "Save 1/min"}
	var countunitsoptions = []string{"Δ #", "Δ #/m³", "Δ #/ft³"}
	prefs := fyne.CurrentApp().Preferences()

	// Instrument-etiketter
	pruninglabel := widget.NewLabel("Pruning:")
	pruningpopup := widget.NewSelect(pruningoptions, func(value string) {})
	pruningpopup.SetSelectedIndex(prefs.IntWithFallback("savefrequency", 0))

	countunitslabel := widget.NewLabel("Count units:")
	countunitspopup := widget.NewSelect(countunitsoptions, func(value string) {})
	switch prefs.IntWithFallback("countunits", 1) {
	case 1:
		countunitspopup.SetSelectedIndex(0)
	case 21201:
		countunitspopup.SetSelectedIndex(1)
	case 600:
		countunitspopup.SetSelectedIndex(2)
	default:
		countunitspopup.SetSelectedIndex(0)
	}

	chkSimulateAeroTrak := widget.NewCheck("AeroTrak", func(value bool) {})
	chkSimulateAeroTrak.SetChecked(prefs.BoolWithFallback("SimulateAeroTrak", false))

	chkSimulateDustTrak := widget.NewCheck("DustTrak", func(value bool) {})
	chkSimulateDustTrak.SetChecked(prefs.BoolWithFallback("SimulateDustTrak", false))

	chkSimulatePTrak := widget.NewCheck("PTrak", func(value bool) {})
	chkSimulatePTrak.SetChecked(prefs.BoolWithFallback("SimulatePTrak", false))

	chksync := widget.NewCheck("Synchronized measuring", func(value bool) {})
	chksync.SetChecked(prefs.Bool("synchronizedmeasuring"))

	chkmydebug := widget.NewCheck("Debugging", func(value bool) {})
	chkmydebug.SetChecked(prefs.Bool("mydebug"))

	chkautostart := widget.NewCheck("Autostart measuring", func(value bool) {})
	chkautostart.SetChecked(prefs.BoolWithFallback("autostartmeasuring", false))

	// Skapa knappen med ikon direkt
	// DATAPOINTSMAX
	optionsshow := []string{"2", "5", "10", "20", "50", "100"}
	sparatshow := prefs.IntWithFallback("datapoints", 100)

	selectshow := widget.NewSelect(optionsshow, func(selected string) {
		valInt, err := strconv.Atoi(selected)
		if err == nil {
			prefs.SetInt("datapoints", valInt)
			vascharts.SetMaxDatapoints(valInt)
		}
	})
	selectshow.SetSelected(strconv.Itoa(sparatshow))
	if selectshow.Selected == "" {
		selectshow.SetSelected("10")
	}
	labelshow := widget.NewLabel("Datapoints: ")

	// AVERAGEPOINTS
	optionsav := []string{"1", "2", "5", "10", "20", "50", "100"}
	sparatav := prefs.IntWithFallback("averagepoints", 10)
	selectav := widget.NewSelect(optionsav, func(selected string) {
		valInt, err := strconv.Atoi(selected)
		if err == nil {
			prefs.SetInt("averagepoints", valInt)
		}
	})
	selectav.SetSelected(strconv.Itoa(sparatav))
	labelav := widget.NewLabel("Avg.datapoints: ")

	// SAMPLE INTERVAL
	optionssa := []string{"1ms", "10ms", "50ms", "100ms", "500ms", "1000ms", "2000ms", "5000ms"}
	selectsa := widget.NewSelect(optionssa, func(selected string) {
		var valInt int
		fmt.Sscanf(selected, "%dms", &valInt)

		if valInt > 0 {
			prefs.SetInt("sampleinterval", valInt)
			nyttIntervall := time.Duration(valInt) * time.Millisecond
			if ActiveMeasurement != nil && ActiveMeasurement.IntervalChan != nil {
				select {
				case ActiveMeasurement.IntervalChan <- nyttIntervall:
				default:
				}
			}
		}
	})
	sparatVal := prefs.IntWithFallback("sampleinterval", 10)
	selectsa.SetSelected(fmt.Sprintf("%dms", sparatVal))
	if selectsa.Selected == "" {
		selectsa.SetSelected("1000ms")
	}
	labelsa := widget.NewLabel("Interval: ")
	containersa := container.NewHBox(labelsa, selectsa)

	documentpath := widget.NewEntry()
	documentpath.SetText(prefs.StringWithFallback("documentpath", ""))
	doclabel := widget.NewLabel("Path:")

	setdocumentpathbutton := widget.NewButton("Set Default Path", func() {
	})
	documentpathbutton := widget.NewButton("Choose Path", func() {
		dialog.ShowFolderOpen(func(fo fyne.ListableURI, err error) {
			if err != nil {
				fyne.LogError("#1 Error on selecting folder", err)
				dialog.ShowError(err, parentWindow)
				return
			} else if fo == nil {
				return
			}
			documentpath.SetText(fo.Path())
		}, parentWindow)
	})
	docinfo := widget.NewButton("Document Path Info", func() {
		h, _ := os.Hostname()
		username := "Unknown"
		if u, err := user.Current(); err == nil && u != nil {
			username = u.Username
		}

		msg := "Computer Name: " + h + "\n"
		msg += "User Name:     " + username + "\n"
		msg += "RootURI:       " + fyne.CurrentApp().Storage().RootURI().Path() + "\n"
		msg += "HomeDir:       " + general.GetHomeDir() + "\n"
		msg += "Preferences:   " + prefs.StringWithFallback("documentpath", "") + "\n"

		// Skapa label med Monospace
		msgLabel := widget.NewLabel(msg)
		msgLabel.TextStyle = fyne.TextStyle{Monospace: true}

		infoDialog := dialog.NewCustom("Defaults:", "OK", msgLabel, parentWindow)
		infoDialog.Show()
	})

	// GEMENSAM SPARRUTIN
	saveAll := func() {
		prefs.SetBool("mydebug", chkmydebug.Checked)
		prefs.SetBool("SimulateAeroTrak", chkSimulateAeroTrak.Checked)
		prefs.SetBool("SimulateDustTrak", chkSimulateDustTrak.Checked)
		prefs.SetBool("SimulatePTrak", chkSimulatePTrak.Checked)
		prefs.SetBool("autostartmeasuring", chkautostart.Checked)
		prefs.SetBool("synchronizedmeasuring", chksync.Checked)

		switch pruningpopup.Selected {
		case pruningoptions[0]:
			prefs.SetInt("savefrequency", 0)
		case pruningoptions[1]:
			prefs.SetInt("savefrequency", 1)
		case pruningoptions[2]:
			prefs.SetInt("savefrequency", 2)
		}

		switch countunitspopup.Selected {
		case countunitsoptions[0]:
			prefs.SetInt("countunits", 1)
		case countunitsoptions[1]:
			prefs.SetInt("countunits", 21201)
		case countunitsoptions[2]:
			prefs.SetInt("countunits", 600)
		}

		prefs.SetString("documentpath", documentpath.Text)

		if settingsDialog != nil {
			settingsDialog.Hide()
		}
	}

	okbutton := widget.NewButton("OK", func() {
		saveAll()
	})
	cancelbutton := widget.NewButton("Cancel", func() {
		if settingsDialog != nil {
			settingsDialog.Hide()
		}
	})

	simGroup := container.NewGridWithColumns(4, widget.NewLabel("Simulate:"), chkSimulateAeroTrak, chkSimulateDustTrak, chkSimulatePTrak)
	//	colgroup := container.NewGridWithColumns(2, b1, r1, b2, r2, b3, r3)
	//	colgroup := widget.NewCard("", "", container.NewGridWithColumns(2, ATColButton, DTColButton, PTColButton))
	chkgroup := container.NewVBox(chksync, chkautostart, chkmydebug)
	topgroup := container.NewGridWithColumns(1, simGroup)

	containershow := container.NewBorder(nil, nil, labelshow, nil, selectshow)
	containerav := container.NewBorder(nil, nil, labelav, nil, selectav)
	pruninggroup := container.NewBorder(nil, nil, pruninglabel, nil, pruningpopup)
	cugroup := container.NewBorder(nil, nil, countunitslabel, nil, countunitspopup)
	popupgroup := container.NewVBox(containershow, containerav, containersa, pruninggroup, cugroup)

	group2 := container.NewHBox(popupgroup, chkgroup)
	pathgroup := container.NewBorder(nil, nil, doclabel, nil, documentpath)
	pathgroup2 := container.NewGridWithColumns(3, setdocumentpathbutton, documentpathbutton, docinfo)

	c := container.NewVBox(topgroup, group2, pathgroup, pathgroup2, container.NewGridWithColumns(2, cancelbutton, okbutton))

	// Skapa en modal dialog över huvudfönstret!
	settingsDialog = dialog.NewCustomWithoutButtons("Settings", c, parentWindow)
	settingsDialog.Show()

	return nil
}
func DoAeroTrakSettings(parentWindow fyne.Window) error {
	var manualDialog dialog.Dialog
	var currentColor2 color.Color = GetStrokeColor(2) // Startfärg för AeroTrak
	var ATColButton *widget.Button
	ATColButton = widget.NewButtonWithIcon("AeroTrak", createColorIcon(currentColor2), func() {
		picker := dialog.NewColorPicker("AeroTrak", "Line Color", func(c color.Color) {
			// 1. Spara/använd den nya färgen i din applikation
			currentColor2 = c

			// 2. Uppdatera ikonens färg på knappen
			ATColButton.SetIcon(createColorIcon(c))
		}, parentWindow)
		picker.Advanced = true
		picker.Show()
	})
	//	ATSettings := canvas.NewText(prefs.StringWithFallback("AeroTrak", ""), colornames.White)
	prefs := fyne.CurrentApp().Preferences()

	f1ecom := widget.NewEntry()
	f1ecom.SetText(prefs.String("AeroTrak"))
	f1ecmd := widget.NewMultiLineEntry()
	f1ecmd.SetText(prefs.String("aerotrakcmd"))
	f1ecmd.SetMinRowsVisible(30)
	f1ecmd.TextStyle = fyne.TextStyle{Monospace: true} // Gör tecknen fasta och mer kompakta
	f1default := widget.NewButton("Set Defaults", func() {
		f1ecom.SetText("192.168.0.130:502")
		f1ecmd.SetText(new(vasinstruments.AeroTraktype).Setupaerotrakcode())
	})
	f1info := widget.NewButton("Info", func() {
		port := prefs.StringWithFallback("AeroTrak", "")
		go func() {
			r, err := vasinstruments.ModbusAeroTrakgetinfo(port)
			if err != nil || r == "" {
				r = "Could not connect to AeroTrak at " + port
			}

			// Kör UI-anropet säkert på huvudtråden!
			fyne.Do(func() {
				dialog.ShowInformation("AeroTrak info", r, parentWindow)
			})
		}()
	})
	f1help := widget.NewButton("Help", func() {
		r := "AeroTrak command program help:\n\n" +
			"setup: - commands to run when setting up the AeroTrak\n" +
			"beforeread: - commands to run before reading data\n" +
			"afterread: - commands to run after reading data\n" +
			"stop: - commands to run when stopping the AeroTrak\n" +
			"restart: - commands to run when restarting the AeroTrak\n\n" +
			"You can use comments starting with // or #\n" +
			"Each command should be on a new line.\n"

		fyne.Do(func() {
			dialog.ShowInformation("AeroTrak info", r, parentWindow)
		})
	})

	clearall := widget.NewButton("Clear", func() {
		f1ecom.SetText("")
		f1ecmd.SetText("")
	})

	okbutton := widget.NewButton("OK", func() {
		prefs.SetString("AeroTrak", f1ecom.Text)
		prefs.SetString("aerotrakcmd", f1ecmd.Text)
		SetStrokeColor(2, currentColor2)
		SetStrokeColor(3, currentColor2)
		SetStrokeColor(4, currentColor2)
		SetStrokeColor(5, currentColor2)
		SetStrokeColor(6, currentColor2)
		SetStrokeColor(7, currentColor2)
		if manualDialog != nil {
			manualDialog.Hide()
		}
	})

	cancelbutton := widget.NewButton("Cancel", func() {
		if manualDialog != nil {
			manualDialog.Hide()
		}
	})

	// 🟢 Använd widget.Form med HBox för knapparna – då linjerar allting knivskarpt!
	form := widget.NewForm(
		widget.NewFormItem("AeroTrak setting:", f1ecom),
		widget.NewFormItem("AeroTrak Setup:", f1ecmd))

	btnInfo := container.NewGridWithColumns(4, f1default, clearall, f1info, f1help)
	btnGroup := container.NewGridWithColumns(2, cancelbutton, okbutton)
	content := container.NewVBox(form, ATColButton, btnInfo, btnGroup)

	manualDialog = dialog.NewCustomWithoutButtons("AeroTrak Settings", content, parentWindow)
	manualDialog.Resize(fyne.NewSize(1200, 756))
	manualDialog.Show()

	return nil

}
func DoDustTrakSettings(parentWindow fyne.Window) error {
	var manualDialog dialog.Dialog
	var currentColor1 color.Color = GetStrokeColor(1) // Startfärg för DustTrak
	var DTColButton *widget.Button
	DTColButton = widget.NewButtonWithIcon("DustTrak", createColorIcon(currentColor1), func() {
		picker := dialog.NewColorPicker("DustTrak", "Line Color", func(c color.Color) {
			// 1. Spara/använd den nya färgen i din applikation
			currentColor1 = c

			// 2. Uppdatera ikonens färg på knappen
			DTColButton.SetIcon(createColorIcon(c))
		}, parentWindow)
		picker.Advanced = true
		picker.Show()
	})
	prefs := fyne.CurrentApp().Preferences()

	f1com := widget.NewEntry()
	f1com.SetText(prefs.String("DustTrak"))
	f1default := widget.NewButton("Set Defaults", func() {
		f1com.SetText("192.168.0.131:3602")
	})

	f1info := widget.NewButton("Info", func() {
		port := prefs.StringWithFallback("DustTrak", "")
		go func() {
			info := new(vasinstruments.DustTraktype).GetDustTrakinfo(port)
			if info == "" {
				info = "Could not connect to DustTrak at " + port
			}
			fyne.Do(func() {
				dialog.ShowInformation("DustTrak info", info, parentWindow)
			})
		}()
	})
	f1help := widget.NewButton("Help", func() {
		info := "DustTrak help:\n\n" +
			"To get DustTrak info, click the Info button.\n" +
			"To set the DustTrak port, enter it in the text field.\n" +
			"DustTrak standard commands for start and stop are MSTART and MSTOP.\n" +
			"The default port is 192.168.0.131:3602\n"
		dialog.ShowInformation("DustTrak help", info, parentWindow)
	})

	clearall := widget.NewButton("Clear", func() {
		f1com.SetText("")
	})

	okbutton := widget.NewButton("OK", func() {
		prefs.SetString("DustTrak", f1com.Text)
		SetStrokeColor(1, currentColor1)
		if manualDialog != nil {
			manualDialog.Hide()
		}
	})

	cancelbutton := widget.NewButton("Cancel", func() {
		if manualDialog != nil {
			manualDialog.Hide()
		}
	})

	// 🟢 Använd widget.Form med HBox för knapparna – då linjerar allting knivskarpt!
	form := widget.NewForm(
		widget.NewFormItem("DustTrak setting:", f1com))

	btnInfo := container.NewGridWithColumns(4, f1default, clearall, f1info, f1help)
	btnGroup := container.NewGridWithColumns(2, cancelbutton, okbutton)
	content := container.NewVBox(form, DTColButton, btnInfo, layout.NewSpacer(), btnGroup)

	manualDialog = dialog.NewCustomWithoutButtons("DustTrak Settings", content, parentWindow)
	//	manualDialog.Resize(fyne.NewSize(600, 240))
	manualDialog.Show()

	return nil
}
func DoPTrakSettings(parentWindow fyne.Window) error {
	var PTColButton *widget.Button
	prefs := fyne.CurrentApp().Preferences()
	var currentColor0 color.Color = GetStrokeColor(0) // Startfärg för PTrak
	PTColButton = widget.NewButtonWithIcon("PTrak   ", createColorIcon(currentColor0), func() {
		picker := dialog.NewColorPicker("PTrak   ", "Line Color", func(c color.Color) {
			// 1. Spara/använd den nya färgen i din applikation
			currentColor0 = c
			// 2. Uppdatera ikonens färg på knappen
			PTColButton.SetIcon(createColorIcon(c))
		}, parentWindow)
		picker.Advanced = true
		picker.Show()
	})

	f1ecom := widget.NewEntry()
	f1ecom.SetText(prefs.String("PTrak"))
	f1bdefault := widget.NewButton("Set Default", func() {
		p := vasinstruments.GetallSerialPorts()
		if len(p) == 0 {
			dialog.ShowInformation("No serial ports found", "No serial ports found on this computer.", parentWindow)
			return
		}
		p1 := strings.Split(p, "\n")[0]
		f1ecom.SetText(p1)
	})

	var manualDialog dialog.Dialog

	f2binfo := widget.NewButton("Info", func() {
		// 1. Hämta alla serieportar i datorn
		msg := vasinstruments.GetallSerialPorts()
		if msg == "" {
			msg = "No serial ports found on this computer."
		}
		// 2. Visa den dynamiska listan i dialogrutan
		dialog.ShowInformation("Available Serial Ports", msg, parentWindow)
	})
	clearall := widget.NewButton("Clear", func() {
		f1ecom.SetText("")
	})

	okbutton := widget.NewButton("OK", func() {
		prefs.SetString("PTrak", f1ecom.Text)
		SetStrokeColor(0, currentColor0)
		if manualDialog != nil {
			manualDialog.Hide()
		}
	})
	cancelbutton := widget.NewButton("Cancel", func() {
		if manualDialog != nil {
			manualDialog.Hide()
		}
	})

	// 🟢 Använd widget.Form med HBox för knapparna – då linjerar allting knivskarpt!
	form := widget.NewForm(
		widget.NewFormItem("PTrak:", f1ecom))

	btnGroup1 := container.NewGridWithColumns(3, clearall, f1bdefault, f2binfo)
	btnGroup2 := container.NewGridWithColumns(2, cancelbutton, okbutton)
	content := container.NewVBox(form, PTColButton, layout.NewSpacer(), btnGroup1, btnGroup2)

	manualDialog = dialog.NewCustomWithoutButtons("PTrak Settings", content, parentWindow)
	//	manualDialog.Resize(fyne.NewSize(600, 240))
	manualDialog.Show()

	return nil
}

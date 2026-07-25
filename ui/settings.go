package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"time"
	"vas2/vascharts"
	"vas2/vasinstruments"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"
)

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
	var pruningoptions = []string{"Save every 5 seconds", "Save every 10 seconds", "Save every minute"}
	var countunitsoptions = []string{"Δ #", "Δ #/m³", "Δ #/ft³"}
	prefs := fyne.CurrentApp().Preferences()

	// Instrument-etiketter
	ins1 := canvas.NewText("AeroTrak: "+prefs.StringWithFallback("AeroTrak", ""), colornames.White)
	ins2 := canvas.NewText("DustTrak: "+prefs.StringWithFallback("DustTrak", ""), colornames.White)
	ins3 := canvas.NewText("PTrak: "+prefs.StringWithFallback("PTrak", ""), colornames.White)

	pruninglabel := canvas.NewText("Pruning:", colornames.White)
	pruningpopup := widget.NewSelect(pruningoptions, func(value string) {})
	pruningpopup.SetSelectedIndex(prefs.IntWithFallback("savefrequency", 0))

	countunitslabel := canvas.NewText("Count units:", colornames.White)
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

	chkSimulateAeroTrak := widget.NewCheck("Simulate AeroTrak", func(value bool) {})
	chkSimulateAeroTrak.SetChecked(prefs.BoolWithFallback("SimulateAeroTrak", false))

	chkSimulateDustTrak := widget.NewCheck("Simulate DustTrak", func(value bool) {})
	chkSimulateDustTrak.SetChecked(prefs.BoolWithFallback("SimulateDustTrak", false))

	chkSimulatePTrak := widget.NewCheck("Simulate PTrak", func(value bool) {})
	chkSimulatePTrak.SetChecked(prefs.BoolWithFallback("SimulatePTrak", false))

	chksync := widget.NewCheck("Synchronized measuring", func(value bool) {})
	chksync.SetChecked(prefs.Bool("synchronizedmeasuring"))

	chkmydebug := widget.NewCheck("Debugging", func(value bool) {})
	chkmydebug.SetChecked(prefs.Bool("mydebug"))

	chkautostart := widget.NewCheck("Autostart measuring", func(value bool) {})
	chkautostart.SetChecked(prefs.BoolWithFallback("autostartmeasuring", true))

	// Färgväljare (Använder nu parentWindow)
	r1 := canvas.NewRectangle(GetStrokeColor(2))
	b1 := widget.NewButton("AeroTrak line color:", func() {
		picker := dialog.NewColorPicker("AeroTrak", "Line Color", func(c color.Color) {
			r1.FillColor = c
			r1.Refresh()
		}, parentWindow)
		picker.Advanced = true
		picker.Show()
	})

	r2 := canvas.NewRectangle(GetStrokeColor(1))
	b2 := widget.NewButton("DustTrak line color:", func() {
		picker := dialog.NewColorPicker("DustTrak", "Line Color", func(c color.Color) {
			r2.FillColor = c
			r2.Refresh()
		}, parentWindow)
		picker.Advanced = true
		picker.Show()
	})

	r3 := canvas.NewRectangle(GetStrokeColor(0))
	b3 := widget.NewButton("PTrak line color:", func() {
		picker := dialog.NewColorPicker("PTrak", "Line Color", func(c color.Color) {
			r3.FillColor = c
			r3.Refresh()
		}, parentWindow)
		picker.Advanced = true
		picker.Show()
	})

	// DATAPOINTSMAX
	optionsshow := []string{"2", "5", "10", "20", "50", "100"}
	sparatshow := prefs.IntWithFallback("datapoints", 10)

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
	labelshow := widget.NewLabel("Datapoints to show: ")

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
	labelav := widget.NewLabel("Averaged datapoints: ")

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
	sparatVal := prefs.IntWithFallback("sampleinterval", 1000)
	selectsa.SetSelected(fmt.Sprintf("%dms", sparatVal))
	if selectsa.Selected == "" {
		selectsa.SetSelected("1000ms")
	}
	labelsa := widget.NewLabel("Sample Interval: ")
	containersa := container.NewHBox(labelsa, selectsa)

	documentpath := widget.NewEntry()
	documentpath.SetText(prefs.StringWithFallback("documentpath", ""))

	documentpathbutton := widget.NewButton("Set Path", func() {
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

		SetStrokeColor(0, r3.FillColor)
		SetStrokeColor(1, r2.FillColor)
		SetStrokeColor(2, r1.FillColor)
		SetStrokeColor(3, r1.FillColor)
		SetStrokeColor(4, r1.FillColor)
		SetStrokeColor(5, r1.FillColor)
		SetStrokeColor(6, r1.FillColor)
		SetStrokeColor(7, r1.FillColor)

		if settingsDialog != nil {
			settingsDialog.Hide()
		}
	}

	okbutton := widget.NewButton("OK", func() {
		saveAll()
	})

	instgroup := container.NewGridWithColumns(2, ins1, chkSimulateAeroTrak, ins2, chkSimulateDustTrak, ins3, chkSimulatePTrak)
	colgroup := container.NewGridWithColumns(2, b1, r1, b2, r2, b3, r3)
	topgroup := container.NewGridWithColumns(2, instgroup, colgroup)

	containershow := container.NewBorder(nil, nil, labelshow, nil, selectshow)
	containerav := container.NewBorder(nil, nil, labelav, nil, selectav)
	pruninggroup := container.NewBorder(nil, nil, pruninglabel, nil, pruningpopup)
	cugroup := container.NewBorder(nil, nil, countunitslabel, nil, countunitspopup)
	popupgroup := container.NewVBox(containershow, containerav, containersa, pruninggroup)

	chkgroup := container.NewVBox(chksync, chkautostart, chkmydebug, cugroup)
	group2 := container.NewHBox(popupgroup, chkgroup)
	pathgroup := container.NewGridWithColumns(2, documentpath, documentpathbutton)

	c := container.NewVBox(topgroup, group2, pathgroup, layout.NewSpacer(), okbutton)

	// Skapa en modal dialog över huvudfönstret!
	settingsDialog = dialog.NewCustomWithoutButtons("Settings", c, parentWindow)
	settingsDialog.Resize(fyne.NewSize(650, 500))
	settingsDialog.Show()

	return nil
}
func DoManualSettings(parentWindow fyne.Window) error {
	prefs := fyne.CurrentApp().Preferences()

	f1e1 := widget.NewEntry()
	f1e2 := widget.NewEntry()
	f1e3 := widget.NewEntry()

	f1e1.SetText(prefs.String("AeroTrak"))
	f1e2.SetText(prefs.String("DustTrak"))
	f1e3.SetText(prefs.String("PTrak"))

	f1b1 := widget.NewButton("Set Default", func() { f1e1.SetText("192.168.0.130:502") })
	f1b2 := widget.NewButton("Set Default", func() { f1e2.SetText("192.168.0.131:3602") })
	f1b3 := widget.NewButton("Set Default", func() { f1e3.SetText("COM3") })

	var manualDialog dialog.Dialog

	f2b1 := widget.NewButton("Info", func() {
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
	f2b2 := widget.NewButton("Info", func() {
		port := prefs.StringWithFallback("DustTrak", "")

		go func() {
			info := new(vasinstruments.DustTraktype).GetDustTrakinfo(port)
			if info == "" {
				info = "Could not connect to DustTrak at " + port
			}

			dialog.ShowInformation("DustTrak info", info, parentWindow)
		}()
	})

	f2b3 := widget.NewButton("Info", func() {
		// 1. Hämta alla serieportar i datorn
		msg := vasinstruments.GetallSerialPorts()
		// 2. Visa den dynamiska listan i dialogrutan
		dialog.ShowInformation("Available Serial Ports", msg, parentWindow)
	})
	clearall := widget.NewButton("Clear", func() {
		f1e1.SetText("")
		f1e2.SetText("")
		f1e3.SetText("")
	})

	okbutton := widget.NewButton("OK", func() {
		prefs.SetString("AeroTrak", f1e1.Text)
		prefs.SetString("DustTrak", f1e2.Text)
		prefs.SetString("PTrak", f1e3.Text)

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
		widget.NewFormItem("AeroTrak:", container.NewBorder(nil, nil, nil, container.NewHBox(f1b1, f2b1), f1e1)),
		widget.NewFormItem("DustTrak:", container.NewBorder(nil, nil, nil, container.NewHBox(f1b2, f2b2), f1e2)),
		widget.NewFormItem("PTrak:", container.NewBorder(nil, nil, nil, container.NewHBox(f1b3, f2b3), f1e3)),
	)

	btnGroup := container.NewGridWithColumns(3, clearall, cancelbutton, okbutton)
	content := container.NewVBox(form, layout.NewSpacer(), btnGroup)

	manualDialog = dialog.NewCustomWithoutButtons("Manual Settings", content, parentWindow)
	manualDialog.Resize(fyne.NewSize(600, 240))
	manualDialog.Show()

	return nil
}
func DoAeroTrakSettings(parentWindow fyne.Window) error {
	prefs := fyne.CurrentApp().Preferences()

	f1e1 := widget.NewMultiLineEntry()
	f1e1.SetText(prefs.String("aerotrakcmd"))

	// 1. Deklarera dialogvariabeln först
	var aeroDialog dialog.Dialog

	f1b1 := widget.NewButton("Set Defaults", func() {
		// Logik för standardkommandon (i gorrutin om det görs tunga anrop)
		go func() {
			f1e1.SetText("DEFAULT_COMMANDS_HERE")
		}()
	})

	okbutton := widget.NewButton("OK", func() {
		prefs.SetString("aerotrakcmd", f1e1.Text)

		if aeroDialog != nil {
			aeroDialog.Hide()
		}
	})

	cancelbutton := widget.NewButton("Cancel", func() {
		if aeroDialog != nil {
			aeroDialog.Hide()
		}
	})

	c0 := container.NewGridWithColumns(3, f1b1, cancelbutton, okbutton)
	c1 := container.NewBorder(nil, c0, nil, nil, f1e1)

	// 2. Skapa den modala dialogen över parentWindow
	aeroDialog = dialog.NewCustomWithoutButtons("Special AeroTrak Settings", c1, parentWindow)
	aeroDialog.Resize(fyne.NewSize(400, 300))
	aeroDialog.Show()

	return nil
}

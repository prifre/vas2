package main

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"
)

func (g *game) setStrokeColor(c int, cc color.Color) {
	r1, g1, b1, a1 := cc.RGBA()
	g.app.Preferences().SetInt(fmt.Sprintf("%vStrokeColorR", c), int(r1))
	g.app.Preferences().SetInt(fmt.Sprintf("%vStrokeColorG", c), int(g1))
	g.app.Preferences().SetInt(fmt.Sprintf("%vStrokeColorB", c), int(b1))
	g.app.Preferences().SetInt(fmt.Sprintf("%vStrokeColorA", c), int(a1))
}
func (g *game) getStrokeColor(c int) color.Color {
	var cc color.Color
	r2, g2, b2, a2 := ChartColors[c].RGBA()
	r1 := uint8(g.app.Preferences().IntWithFallback(fmt.Sprintf("%vStrokeColorR", c), int(r2)))
	g1 := uint8(g.app.Preferences().IntWithFallback(fmt.Sprintf("%vStrokeColorG", c), int(g2)))
	b1 := uint8(g.app.Preferences().IntWithFallback(fmt.Sprintf("%vStrokeColorB", c), int(b2)))
	a1 := uint8(g.app.Preferences().IntWithFallback(fmt.Sprintf("%vStrokeColorA", c), int(a2)))
	cc = color.RGBA{r1, g1, b1, a1}
	return cc
}
func DoSettings(g *game) error {
	settings := g.app.NewWindow("Settings")
	settings.SetCloseIntercept(func() {
	})
	b0 := canvas.NewText("                        ", colornames.Black)
	b0.TextSize = 40
	ins1 := canvas.NewText("AeroTrak: "+g.app.Preferences().StringWithFallback("AeroTrak", ""), colornames.White)
	ins2 := canvas.NewText("DustTrak: "+g.app.Preferences().StringWithFallback("DustTrak", ""), colornames.White)
	ins3 := canvas.NewText("PTrak: "+g.app.Preferences().StringWithFallback("PTrak", ""), colornames.White)
	b0.TextSize = 14

	pruninglabel := canvas.NewText("Pruning of data:", colornames.White)
	pruningpopup := widget.NewSelect([]string{"Save every 5 seconds", "Save every 10 seconds", "Save every minute"}, func(value string) {
	})
	pruningpopup.SetSelectedIndex(g.app.Preferences().IntWithFallback("savefrequency", 0))

	countunitslabel := canvas.NewText("Count units::", colornames.White)
	countunitspopup := widget.NewSelect([]string{"Δ #", "Δ #/m³", "Δ #/ft³"}, func(value string) {
	})
	countunitspopup.SetSelectedIndex(g.app.Preferences().IntWithFallback("countunits", 0))

	chksimulateAeroTrak := widget.NewCheck("Simulate AeroTrak", func(value bool) {})
	chksimulateAeroTrak.SetChecked(g.app.Preferences().BoolWithFallback("SimulateAeroTrak", false))
	if g.app.Preferences().String("AeroTrak") != "" {
		chksimulateAeroTrak.SetChecked(false)
		chksimulateAeroTrak.Disable()
	}

	chksimulateDustTrak := widget.NewCheck("Simulate DustTrak", func(value bool) {})
	chksimulateDustTrak.SetChecked(g.app.Preferences().BoolWithFallback("SimulateDustTrak", false))
	if g.app.Preferences().String("DustTrak") != "" {
		chksimulateDustTrak.SetChecked(false)
		chksimulateDustTrak.Disable()
	}

	chksimulatePTrak := widget.NewCheck("Simulate PTrak", func(value bool) {})
	chksimulatePTrak.SetChecked(g.app.Preferences().BoolWithFallback("SimulatePTrak", false))
	if g.app.Preferences().String("PTrak") != "" {
		chksimulatePTrak.SetChecked(false)
		chksimulatePTrak.Disable()
	}

	chksync := widget.NewCheck("Synchronized measuring", func(value bool) {})
	chksync.SetChecked(g.app.Preferences().Bool("synchronizedmeasuring"))

	chkmydebug := widget.NewCheck("Debugging", func(value bool) {})
	chkmydebug.SetChecked(g.app.Preferences().Bool("mydebug"))

	chkautostart := widget.NewCheck("Autostart measuring", func(value bool) {})
	chkautostart.SetChecked(g.app.Preferences().BoolWithFallback("autostartmeasuring", true))

	r1 := canvas.NewRectangle(colornames.Green)
	r1.FillColor = g.getStrokeColor(2)
	b1 := widget.NewButton("AeroTrak line color:", func() {
		picker := dialog.NewColorPicker("AeroTrak", "Line Color", func(c color.Color) {
			r1.FillColor = c
			r1.Refresh()
		}, settings)
		picker.Advanced = true
		picker.Show()
	})
	r2 := canvas.NewRectangle(colornames.White)
	r2.FillColor = g.getStrokeColor(1)
	b2 := widget.NewButton("DustTrak line color:", func() {
		picker := dialog.NewColorPicker("DustTrak", "Line Color", func(c color.Color) {
			r2.FillColor = c
			r2.Refresh()
		}, settings)
		picker.Advanced = true
		picker.Show()
	})

	r3 := canvas.NewRectangle(g.getStrokeColor(0))
	b3 := widget.NewButton("PTrak line color:", func() {
		picker := *dialog.NewColorPicker("PTrak", "Line Color", func(c color.Color) {
			r3.FillColor = c
			r3.Refresh()
		}, settings)
		picker.Advanced = true
		picker.Show()
	})

	f1 := float64(g.app.Preferences().IntWithFallback("datapoints", 10))
	data := binding.BindFloat(&f1)
	slide := widget.NewSliderWithData(2, 100, data)
	slide.Step = 1
	entry := widget.NewLabelWithData(binding.FloatToStringWithFormat(data, "Number of datapoints to show: %v"))
	slide.Refresh()

	f1av := float64(g.app.Preferences().IntWithFallback("datapoints", 10))
	dataav := binding.BindFloat(&f1av)
	slideav := widget.NewSliderWithData(1, 100, dataav)
	slideav.Step = 1
	entryav := widget.NewLabelWithData(binding.FloatToStringWithFormat(dataav, "Number of averaged datapoints: %v"))
	slideav.Refresh()

	documentpath := widget.NewEntry()
	documentpath.SetText(g.app.Preferences().StringWithFallback("documentpath", g.getdocumentpath().Path()))
	documentpath.MultiLine = true

	documentpathbutton := widget.NewButton("Set Path", func() {
		dialog.ShowFolderOpen(func(fo fyne.ListableURI, err error) {
			if err != nil {
				fyne.LogError("#1 Error on selecting folder", err)
				dialog.ShowError(err, settings)
				return
			} else if fo == nil {
				return
			}
			documentpath.Text = fo.Path()
			documentpath.Refresh()
		}, settings)
	})

	okbutton := widget.NewButton("OK", func() {
		g.app.Preferences().SetInt("datapoints", int(f1))
		g.datapoints = int(f1)
		g.app.Preferences().SetInt("averagepoints", int(f1av))
		g.averagepoints = int(f1av)
		if g.app.Preferences().Bool("mydebug") != chkmydebug.Checked {
			g.app.Preferences().SetBool("mydebug", chkmydebug.Checked)
			g.setuplogging() // since myDebug may have changed.
		}
		g.app.Preferences().SetBool("SimulateAeroTrak", chksimulateAeroTrak.Checked)
		g.app.Preferences().SetBool("SimulateDustTrak", chksimulateDustTrak.Checked)
		g.app.Preferences().SetBool("SimulatePTrak", chksimulatePTrak.Checked)
		g.app.Preferences().SetBool("autostartmeasuring", chkautostart.Checked)
		g.app.Preferences().SetBool("synchronizedmeasuring", chksync.Checked)
		switch pruningpopup.Selected {
		case "Save every 5 seconds":
			g.app.Preferences().SetInt("savefrequency", 0)
		case "Save every 10 seconds":
			g.app.Preferences().SetInt("savefrequency", 1)
		case "Save every minute":
			g.app.Preferences().SetInt("savefrequency", 2)
		}
		switch countunitspopup.Selected {
		case "Δ #":
			g.app.Preferences().SetInt("countunits", 0)
			g.countunits = 1
		case "Δ #/m³":
			g.app.Preferences().SetInt("countunits", 1)
			g.countunits = 21201
		case "Δ #/ft³":
			g.app.Preferences().SetInt("countunits", 2)
			g.countunits = 600
		}
		g.app.Preferences().SetString("documentpath", documentpath.Text)
		g.setStrokeColor(0, r3.FillColor)
		g.setStrokeColor(1, r2.FillColor)
		g.setStrokeColor(2, r1.FillColor)
		g.setStrokeColor(3, r1.FillColor)
		g.setStrokeColor(4, r1.FillColor)
		g.setStrokeColor(5, r1.FillColor)
		g.setStrokeColor(6, r1.FillColor)
		g.setStrokeColor(7, r1.FillColor)
		settings.Close()
	})
	cancelbutton := widget.NewButton("Cancel", func() {
		settings.Close()
	})
	instgroup := container.NewGridWithColumns(2, ins1, chksimulateAeroTrak, ins2, chksimulateDustTrak, ins3, chksimulatePTrak)
	colgroup := container.NewGridWithColumns(2, b1, r1, b2, r2, b3, r3)
	instcol := container.NewGridWithColumns(2, instgroup, colgroup)

	chkgroup := container.New(layout.NewVBoxLayout(), chksync, chkautostart, chkmydebug)
	pruninggroup := container.NewGridWithRows(2, pruninglabel, pruningpopup)
	cugroup := container.NewGridWithRows(2, countunitslabel, countunitspopup)
	group1 := container.NewGridWithColumns(2, pruninggroup, cugroup)

	slidegroup := container.NewGridWithRows(2, slide, entry)
	slidegroup2 := container.NewGridWithRows(2, slideav, entryav)
	slidegroup3 := container.NewGridWithRows(2, slidegroup, slidegroup2)
	pathgroup := container.NewGridWithColumns(2, documentpath, documentpathbutton)
	group2 := container.NewGridWithColumns(2, slidegroup3, chkgroup)
	c6 := container.NewGridWithColumns(2, cancelbutton, okbutton)
	c := container.New(layout.NewVBoxLayout(), instcol, group2, group1, pathgroup, layout.NewSpacer(), c6)
	settings.SetContent(c)
	settings.CenterOnScreen()
	settings.Show()
	return nil
}

func DoFTPSettings(g *game) error {
	ftpsettings := g.app.NewWindow("FTP Settings")
	g.f1.ftpserver = g.app.Preferences().String("ftpserver")
	g.f1.ftpusername = g.app.Preferences().String("ftpusername")
	g.f1.ftppassword = g.app.Preferences().String("ftppassword")
	g.f1.ftpdir = g.app.Preferences().String("ftpdir")
	f1e1 := widget.NewEntry()
	f1e2 := widget.NewEntry()
	f1e3 := widget.NewEntry()
	f1e4 := widget.NewEntry()
	sp1 := widget.NewLabel("Please note that the connection information is not stored in a secure way!")
	spacing := widget.NewLabel(strings.Repeat(string(rune(32)), 30))
	f1e1.SetText(g.f1.ftpserver)
	f1e2.SetText(g.f1.ftpusername)
	f1e3.SetText(g.f1.ftppassword)
	f1e4.SetText(g.f1.ftpdir)
	f1e1.MultiLine = false
	f1e2.MultiLine = false
	f1e3.MultiLine = false
	f1e4.MultiLine = false
	f1t1 := widget.NewLabel("FTP-server: ")
	f1t2 := widget.NewLabel("FTP-username: ")
	f1t3 := widget.NewLabel("FTP-password: ")
	f1t4 := widget.NewLabel("FTP-path: ")
	okbutton := widget.NewButton("OK", func() {
		g.app.Preferences().SetString("ftpserver", f1e1.Text)
		g.app.Preferences().SetString("ftpusername", f1e2.Text)
		g.app.Preferences().SetString("ftppassword", f1e3.Text)
		g.app.Preferences().SetString("ftpdir", f1e4.Text)
		ftpsettings.Close()
	})
	cancelbutton := widget.NewButton("Cancel", func() {
		ftpsettings.Close()
	})
	ftpsettings.SetCloseIntercept(func() {
		ftpsettings.Close()
	})
	c1 := container.NewGridWithColumns(2, sp1, spacing, f1t1, f1e1, f1t2, f1e2, f1t3, f1e3, f1t4, f1e4, cancelbutton, okbutton)
	c1.Resize(fyne.NewSize(c1.Size().Width*1.5, c1.Size().Height*1.5))
	ftpsettings.SetContent(c1)
	ftpsettings.CenterOnScreen()
	ftpsettings.Show()
	return nil
}

func DoManualSettings(g *game) error {
	msettings := g.app.NewWindow("Manual Settings")
	spacing := widget.NewLabel(strings.Repeat(string(rune(32)), 35))
	temp_AeroTrak := g.app.Preferences().String("AeroTrak")
	temp_DustTrak := g.app.Preferences().String("DustTrak")
	temp_PTrak := g.app.Preferences().String("PTrak")
	f1e1 := widget.NewEntry()
	f1e2 := widget.NewEntry()
	f1e3 := widget.NewEntry()
	f1e1.SetText(temp_AeroTrak)
	f1e2.SetText(temp_DustTrak)
	f1e3.SetText(temp_PTrak)
	f1e1.MultiLine = false
	f1e2.MultiLine = false
	f1e3.MultiLine = false
	f1t1 := widget.NewLabel("AeroTrak: ")
	f1t2 := widget.NewLabel("DustTrak: ")
	f1t3 := widget.NewLabel("PTrak: ")
	f1b1 := widget.NewButton("Set", func() {
		f1e1.SetText("192.168.0.130:502")
	})
	f1b2 := widget.NewButton("Set", func() {
		f1e2.SetText("192.168.0.131:3602")
	})
	f1b3 := widget.NewButton("Set", func() {
		f1e3.SetText("COM3")
	})
	clearall := widget.NewButton("Clear", func() {
		f1e1.SetText("")
		f1e2.SetText("")
		f1e3.SetText("")
	})
	f2b1 := widget.NewButton("Info", func() {
		g.at.AeroTrakport = g.app.Preferences().StringWithFallback("AeroTrak", "")
		r, _ := g.at.ModbusAeroTrakgetinfo()
		dialog.ShowInformation("AeroTrak info", r, msettings)
	})
	f2b2 := widget.NewButton("Info", func() {
		g.dt.DustTrakport = g.app.Preferences().StringWithFallback("DustTrak", "")
		dialog.ShowInformation("DustTrak info", g.dt.GetDustTrakinfo(), msettings)
	})
	f2b3 := widget.NewButton("Info", func() {
		g.pt.PTrakport = g.app.Preferences().StringWithFallback("PTrak", "")
		dialog.ShowInformation("PTrak info", "It's a PTrak, ok?", msettings)
	})
	okbutton := widget.NewButton("OK", func() {
		g.app.Preferences().SetString("AeroTrak", f1e1.Text)
		g.app.Preferences().SetString("DustTrak", f1e2.Text)
		g.app.Preferences().SetString("PTrak", f1e3.Text)
		msettings.Close()
	})
	cancelbutton := widget.NewButton("Cancel", func() {
		msettings.Close()
	})
	c1 := container.NewGridWithColumns(4, f1t1, f1e1, f1b1, f2b1, f1t2, f1e2, f1b2, f2b2, f1t3, f1e3, f1b3, f2b3, clearall, cancelbutton, okbutton, spacing)
	c1.Resize(fyne.NewSize(c1.Size().Width*1.5, c1.Size().Height*1.5))
	msettings.SetCloseIntercept(func() {
		msettings.Close()
	})
	msettings.SetContent(c1)
	msettings.CenterOnScreen()
	msettings.Show()
	return nil
}
func DoAeroTrakSettings(g *game) error {
	msettings := g.app.NewWindow("Special AeroTrak Settings")
	// spacing := widget.NewLabel(strings.Repeat(string(rune(32)), 35))
	temp_ATcmd := g.app.Preferences().String("aerotrakcmd")
	f1e1 := widget.NewEntry()
	f1e1.Resize(fyne.NewSize(300, 400))
	f1e1.SetText(temp_ATcmd)
	f1e1.MultiLine = true
	f1b1 := widget.NewButton("Set", func() {
		f1e1.SetText(g.at.setupaerotrakcode())
	})
	okbutton := widget.NewButton("OK", func() {
		g.app.Preferences().SetString("aerotrakcmd", f1e1.Text)
		msettings.Close()
	})
	cancelbutton := widget.NewButton("Cancel", func() {
		msettings.Close()
	})
	c0 := container.NewGridWithColumns(3, f1b1, cancelbutton, okbutton)
	mx := container.NewMax(f1e1)
	c1 := container.NewBorder(nil, c0, nil, nil, mx)
	//	c1.Resize(fyne.NewSize(c1.Size().Width*10, c1.Size().Height))
	msettings.SetCloseIntercept(func() {
		msettings.Close()
	})
	msettings.SetContent(c1)
	msettings.CenterOnScreen()
	msettings.Resize(fyne.NewSize(g.winSize.Width-100, g.winSize.Height))
	msettings.Show()
	return nil
}

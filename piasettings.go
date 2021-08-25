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

	pruninglabel := canvas.NewText("Measurements to database", colornames.White)
	pruningpopup := widget.NewSelect([]string{"Save every 5 seconds", "Save every 10 seconds", "Save every minute"}, func(value string) {
	})
	pruningpopup.SetSelectedIndex(g.app.Preferences().IntWithFallback("savefrequency", 0))

	chksimulateAeroTrak := widget.NewCheck("Simulate AeroTrak", func(value bool) {})
	chksimulateAeroTrak.SetChecked(g.app.Preferences().Bool("SimulateAeroTrak"))

	chksimulateDustTrak := widget.NewCheck("Simulate DustTrak", func(value bool) {})
	chksimulateDustTrak.SetChecked(g.app.Preferences().Bool("SimulateDustTrak"))

	chksimulatePTrak := widget.NewCheck("Simulate PTrak", func(value bool) {})
	chksimulatePTrak.SetChecked(g.app.Preferences().BoolWithFallback("SimulatePTrak", false))

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
	entry := widget.NewEntryWithData(binding.FloatToStringWithFormat(data, "Number of datapoints to show: %v"))
	slide.Refresh()

	okbutton := widget.NewButton("OK", func() {
		g.app.Preferences().SetInt("datapoints", int(f1))
		g.datapoints = int(f1)
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
	c1 := container.NewGridWithColumns(2, ins1, chksimulateAeroTrak, ins2, chksimulateDustTrak, ins3, chksimulatePTrak, b0)
	c2 := container.NewGridWithColumns(2, pruninglabel, chksync, pruningpopup, chkautostart, b0, chkmydebug)
	c3 := container.NewGridWithColumns(2, b1, r1, b2, r2, b3, r3, b0, b0)
	c4 := container.NewGridWithRows(2, slide, entry)
	c := container.NewGridWithColumns(2, c1, c2, c3, c4, cancelbutton, okbutton)
	settings.SetContent(c)
	settings.CenterOnScreen()
	settings.Show()
	return nil
}

func DoFTPSettings(g *game) error {
	ftpsettings := g.app.NewWindow("FTP Settings")
	ftpsettings.SetCloseIntercept(func() {
	})
	g.f1.ftpserver = g.app.Preferences().String("ftpserver")
	g.f1.ftpusername = g.app.Preferences().String("ftpusername")
	g.f1.ftppassword = g.app.Preferences().String("ftppassword")
	g.f1.ftpdir = g.app.Preferences().String("ftpdir")
	f1e1 := widget.NewEntry()
	f1e2 := widget.NewEntry()
	f1e3 := widget.NewEntry()
	f1e4 := widget.NewEntry()
	sp1 := widget.NewLabel("Please note that the connection information is not stored in a secure way!")
	sp2 := widget.NewLabel(strings.Repeat(string(rune(32)), 30))
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
	c1 := container.NewGridWithColumns(2, sp1, sp2, f1t1, f1e1, f1t2, f1e2, f1t3, f1e3, f1t4, f1e4, cancelbutton, okbutton)
	c1.Resize(fyne.NewSize(c1.Size().Width*1.5, c1.Size().Height*1.5))
	ftpsettings.SetContent(c1)
	ftpsettings.CenterOnScreen()
	ftpsettings.Show()
	return nil
}

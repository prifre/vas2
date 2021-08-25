package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"
)

const (
	version       = 0.29
	chartnum      = 8
	datapointsmax = 100
)

type game struct {
	MyDebug      bool
	Logpath      fyne.URI
	Settingspath fyne.URI
	app          fyne.App
	window       fyne.Window
	winSize      fyne.Size
	start        time.Time
	widget.BaseWidget
	datapoints            int
	d                     *dbtype
	f1                    *ftptype
	at                    AeroTraktype
	dt                    DustTraktype
	pt                    PTraktype
	synchronizedmeasuring bool
	autostartmeasuring    bool
	recordcount           int
	laps                  int
	paused                bool
	endmeasuring          bool
	cont                  fyne.Container
	chart                 *[8]LineChart
	chartSize             fyne.Size
}

var instruments = [3]string{"PTrak", "DustTrak", "AeroTrak"}
var tbl []string = []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}
var g game

func main() {
	//	os.Setenv("FYNE_SCALE", "1.0")
	//	var g game = *newGame(&[8]LineChart{})
	g = *new(game)
	g.chart = new([8]LineChart)
	g.app = app.NewWithID("particles.in.air")
	g.window = g.app.NewWindow("Particles in Air")
	g.window.SetMainMenu(g.buildMenu())
	g.window.Canvas().SetOnTypedRune(g.typedRune)
	g.d = new(dbtype)
	g.f1 = new(ftptype)
	g.MyDebug = g.app.Preferences().BoolWithFallback("mydebug", true)
	w := g.app.Preferences().FloatWithFallback("winWidth", 1024)
	h := g.app.Preferences().FloatWithFallback("winHeight", 768)
	g.winSize = fyne.NewSize(float32(w), float32(h))
	g.datapoints = g.app.Preferences().IntWithFallback("datapoints", 10)
	g.synchronizedmeasuring = g.app.Preferences().BoolWithFallback("synchronizedmeasuring", true)
	g.autostartmeasuring = g.app.Preferences().BoolWithFallback("autostartmeasuring", true)
	g.d.nanostamp = Getint64(g.app.Preferences().StringWithFallback("nanostamp", "0"))
	g.setuplogging()
	g.start = time.Now()
	log.Printf("Program started %v", g.start)
	g.window.SetCloseIntercept(func() {
		g.closeapp()
	})
	g.window.Resize(g.winSize)
	g.window.CenterOnScreen()
	g.setupcharts()
	g.StopMeasurement()
	if g.autostartmeasuring {
		log.Printf("Autostart active, starting %v", g.d.nanostamp)
		g.StartMeasurement()
	} else {
		g.StopMeasurement()
	}
	g.window.ShowAndRun()
	g.closeapp()
}

func (g *game) Measure() {
	var msg string
	var err error
	var fn string
	var oldtstamp string
	for {
		if g.endmeasuring {
			return
		}
		if g.winSize.Width != g.window.Canvas().Size().Width ||
			g.winSize.Height != g.window.Canvas().Size().Height ||
			g.window.Canvas().Size().Height > 0 && g.window.Canvas().Size().Width > 0 {
			g.drawCharts()
			g.windowResize()
		}
		if g.getData() && !g.paused && oldtstamp != g.d.tstamp {
			g.recordcount = g.recordcount + 1
			oldtstamp = g.d.tstamp
			r := runtime.NumGoroutine()
			g.d.AddMeasurement()
			g.updateCharts()
			for r < runtime.NumGoroutine() {
				time.Sleep(time.Millisecond)
			}
			g.laps = g.laps + 1
			if time.Now().String()[14:19] == "00:00" {
				runtime.GC()
				g.d.Pruning(g)
				msg, err = g.d.Exporttotext(g)
				if err != nil {
					log.Println("Error exporting: ", msg, g.window)
				} else {
					for i := 0; i < len(tbl); i++ {
						fn = tbl[i] + ".txt"
						g.Doftp(fn)
					}
				}
				fn = g.Doscreenshot()
				if fn > "" {
					g.Doftp(fn)
				}
			}
			if g.recordcount%100 == 0 && g.MyDebug {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Printf("%v  Go: %v  Time: %vs  Rec: %v  Speed: %.2f Rec/s  Laps: %v  Stack: %v \n",
					time.Now().String()[11:19],
					runtime.NumGoroutine(), strings.Split(time.Since(g.start).String(), ".")[0],
					g.recordcount, float64(g.recordcount)/(time.Since(g.d.dbstart).Seconds()), g.laps, m.StackInuse)
			}
		}
	}
}

// application menu
func (g *game) buildMenu() *fyne.MainMenu {
	// var err error
	// a quit item will be appended to our first menu
	var err error
	var msg string
	mFile := fyne.NewMenu("File",
		fyne.NewMenuItem("Setup instruments...", func() {
			g.StopMeasurement()
			// cnf := dialog.NewConfirm("Setup of instruments", "Search for instruments?", func(reply bool) {
			// 	if reply {
			g.SetupInstruments()
			// 	}
			// }, g.window)
			// cnf.SetDismissText("No")
			// cnf.SetConfirmText("Yes")
			// cnf.Show()
		}),
		fyne.NewMenuItem("FTP settings...", func() {
			g.StopMeasurement()
			DoFTPSettings(g)
		}),
		fyne.NewMenuItem("Settings...", func() {
			g.StopMeasurement()
			DoSettings(g)
		}))
	mEdit := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Cut", func() { dialog.ShowInformation("Cut? ", "Not implemented, sorry!", g.window) }),
		fyne.NewMenuItem("Copy", func() { dialog.ShowInformation("Copy? ", "Not implemented, sorry!", g.window) }),
		fyne.NewMenuItem("Paste", func() { dialog.ShowInformation("Paste? ", "Not implemented, sorry!", g.window) }))
	mMeasurements := fyne.NewMenu("Measurements",
		fyne.NewMenuItem("Start Measuring", func() {
			g.StartMeasurement()
		}),
		fyne.NewMenuItem("End Measuring", func() {
			g.StopMeasurement()
			g.d.CloseMeasurement()
			log.Printf("Measurement '%v' (%v) stopped", g.d.mname, g.d.nanostamp)
			g.app.Preferences().SetString("nanostamp", "0")
			g.d.mname = ""
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Measurements maintainance", func() {
			g.StopMeasurement()
			g.d.DoMeasurements(g)
		}),
		fyne.NewMenuItem("Save screen", func() {
			g.Doscreenshot()
		}),
		fyne.NewMenuItem("Upload to FTP-server", func() {
			var fn string
			for i := 0; i < len(tbl); i++ {
				fn = tbl[i] + ".txt"
				_, err = os.Stat(filepath.Join(filepath.Dir(g.d.Databasepath), fn))
				if err != nil {
					log.Print("File ", fn, " not found: ", err.Error())
				} else {
					g.Doftp(fn)
				}
			}
			fn = g.Doscreenshot()
			if fn > "" {
				g.Doftp(fn)
			}
		}),
		fyne.NewMenuItem("Export all measurements to textfile", func() {
			g.StopMeasurement()
			msg, err = g.d.Exporttotext(g)
			if err != nil {
				dialog.ShowInformation("Error exporting: ", msg, g.window)
			} else {
				dialog.ShowInformation("Export report", msg, g.window)
			}
		}),
		fyne.NewMenuItem("Export current measurement to textfile", func() {
			g.StopMeasurement()
			msg, err = g.d.Exportonetotext(g)
			if err != nil {
				dialog.ShowInformation("Error exporting: ", msg, g.window)
			} else {
				dialog.ShowInformation("Export report", msg, g.window)
			}
		}),
		// fyne.NewMenuItem("Repair database", func() {
		// 	DorepairDatabase(g.d, g.app)
		// }),
		fyne.NewMenuItem("Remove redundant measurements", func() {
			g.StopMeasurement()
			g.d.Pruning(g)
		}),
		fyne.NewMenuItem("Open storage location", func() {
			openstoragelocation(g)
		}))
	mHelp := fyne.NewMenu("Help",
		fyne.NewMenuItem("About...", func() {
			t := "\nProgram for measuring with TSI\nAeroTrak, DustTrak and PTrak\n\nby Peter Freund\n\n"
			t = t + fmt.Sprintf("Version %v", version)
			dialog.ShowInformation("About...", t, g.window)
		}),
		fyne.NewMenuItem("Copyright info...", func() {
			t := "All packages used are below:\n"
			t += "Fyne - fyne.io (fyne.io/fyne/v2)\n"
			t += "Sqlite3 - (github.com/mattn/go-sqlite3)\n"
			t += "FTP - (github.com/jlaffaye/ftp)\n"
			t += "modbus - (github.com/goburrow/modbus)\n"
			t += "Excelexport - (github.com/360EntSecGroup-Skylar/excelize)\n"
			t += "Serial - (github.com/jacobsa/go-serial/serial)\n"

			dialog.ShowInformation("About...", t, g.window)
		}),
		fyne.NewMenuItem("Check for update", func() {
			g.Checkforupdate()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Open the PARTICLES IN AIR webpage!", func() {
			g.openurl("http://www.prifre.com/pia")
		}),
	)
	m3 := fyne.NewMainMenu(mFile, mEdit, mMeasurements, mHelp)
	return m3
}
func (g *game) countrunning() int {
	var c1 int = 0
	if g.at.AeroTrakrunning {
		c1 = c1 + 6
	}
	if g.dt.DustTrakrunning {
		c1 = c1 + 1
	}
	if g.pt.PTrakrunning {
		c1 = c1 + 1
	}
	return c1
}
func (g *game) windowResize() {
	if g.window.Canvas().Size().Width < 5 || g.window.Canvas().Size().Height < 5 {
		return
	} else {
		g.winSize.Width = g.window.Canvas().Size().Width
		g.winSize.Height = g.window.Canvas().Size().Height
	}
	switch g.countrunning() {
	case 0, 1:
		g.chartSize.Width = g.winSize.Width
		g.chartSize.Height = g.winSize.Height
	case 2:
		g.chartSize = g.winSize
		g.chartSize.Width = g.winSize.Width / 2
	case 6:
		g.chartSize.Width = g.winSize.Width / 3
		g.chartSize.Height = g.winSize.Height / 2
	case 7, 8:
		g.chartSize.Width = g.winSize.Width / 4
		g.chartSize.Height = g.winSize.Height / 2
	}
}
func (g *game) getoldmeasurement() error {
	var temp []string
	var err error
	if !g.app.Preferences().BoolWithFallback("autostartmeasuring", false) {
		return fmt.Errorf("no autostart ")
	}
	if Getint64(g.app.Preferences().StringWithFallback("nanostamp", "0")) == 0 {
		return fmt.Errorf("no nanostamp")
	}
	g.d.nanostamp = Getint64(g.app.Preferences().StringWithFallback("nanostamp", "0"))
	temp, err = g.d.Getsql(fmt.Sprintf("SELECT mname FROM tblMain WHERE nanostamp=%v", g.d.nanostamp))
	if err != nil {
		return fmt.Errorf("#1 Getting mname Error1: %v", err.Error())
	}
	if temp == nil {
		return fmt.Errorf("#1 Getting mname Error2: ''")
	}
	if len(temp) < 1 {
		return fmt.Errorf("#1 Getting mname Error3: %v", err.Error())
	}
	g.d.mname = fmt.Sprintf("%v", temp[0])
	temp, err = g.d.Getsql(fmt.Sprintf("SELECT tstamp FROM tblMain WHERE nanostamp = %v", g.d.nanostamp))
	if err != nil {
		return fmt.Errorf("#1 Getting tstamp Error: %v", err.Error())
	}
	if temp == nil {
		return fmt.Errorf("#1 Getting tstamp Error2: nil")
	}
	if len(temp) < 1 {
		return fmt.Errorf("#1 Getting tstamp Error3: %v", err.Error())
	}
	g.d.tstamp = fmt.Sprintf("%v", temp[0])
	return nil
}
func (g *game) StartMeasurement() {
	var err error
	g.d.Setupdb(g)
	g.synchronizedmeasuring = g.app.Preferences().BoolWithFallback("synchronizedmeasuring", true)
	g.autostartmeasuring = g.app.Preferences().BoolWithFallback("autostartmeasuring", true)
	g.checkinstruments()
	if !g.at.AeroTrakrunning && !g.dt.DustTrakrunning && !g.pt.PTrakrunning {
		return
	}
	err = g.getoldmeasurement()
	g.d.dbstart = time.Now()
	if err != nil {
		log.Println("restart was not possible due to: ", err.Error())
		g.d.nanostamp = time.Now().UnixNano()
		g.app.Preferences().SetString("nanostamp", fmt.Sprintf("%v", g.d.nanostamp))
		g.d.tstamp = time.Now().Format(time.RFC3339)
		g.d.mname = "Measurement" + g.d.tstamp
		g.d.mname = strings.Replace(g.d.mname, "-", "", 99)
		g.d.mname = strings.Replace(g.d.mname, "/", "", 99)
		g.d.mname = strings.Replace(g.d.mname, ":", "", 99)
		g.d.mname = strings.Replace(g.d.mname, ".", "", 99)
		g.d.mname = g.d.mname[:26]
		if g.at.simulateAeroTrak || g.dt.simulateDustTrak || g.pt.simulatePTrak {
			g.d.note = "Simulated measurement "
			if g.at.simulateAeroTrak {
				g.d.note = g.d.note + " AeroTrak"
			}
			if g.dt.simulateDustTrak {
				g.d.note = g.d.note + " DustTrak"
			}
			if g.pt.simulatePTrak {
				g.d.note = g.d.note + " PTrak"
			}
		}
		g.d.mdata = [8]int32{-1, -1, -1, -1, -1, -1, -1, -1}
		err = g.d.AddMeasurement()
		if err != nil {
			log.Println("#1 StartMeasurement - Problems adding Main measurement ", err.Error())
		}
	}
	g.window.Resize(g.winSize)
	g.setupcharts()
	g.at.AeroTrakport = g.app.Preferences().StringWithFallback("AeroTrak", "")
	g.dt.DustTrakport = g.app.Preferences().StringWithFallback("DustTrak", "")
	g.pt.PTrakport = g.app.Preferences().StringWithFallback("PTrak", "")
	log.Printf("Measurement '%v' (%v) started at %v.", g.d.mname, g.d.nanostamp, g.d.tstamp)
	g.recordcount = 0
	g.paused = false
	g.endmeasuring = false
	go g.Measure()
	// g.d.Pruning(g)
	// os.Exit(0)
}

func (g *game) StopMeasurement() {
	g.endmeasuring = true
	var err error
	if g.at.AeroTrakrunning && !g.at.simulateAeroTrak {
		time.Sleep(time.Second)
		err = g.at.AeroTrakstop()
		if err != nil {
			log.Println("#1 StopMeasurement - Problems AeroTrak stop", err.Error())
		}
	}
	if g.dt.DustTrakrunning && !g.dt.simulateDustTrak {
		time.Sleep(time.Second)
		g.dt.DustTrakstop()
		err = g.dt.DustTrakstop()
		if err != nil {
			log.Println("#2 StopMeasurement - Problems DustTrak stop", err.Error())
		}
	}
	g.at.AeroTrakrunning = false
	g.dt.DustTrakrunning = false
	g.pt.PTrakrunning = false
	for c := 0; c < chartnum; c++ {
		for i := 0; i < g.datapoints; i++ {
			g.chart[c].mLine[i].Position1.X = 0
			g.chart[c].mLine[i].Position2.X = 0
		}
		g.chart[c].mChart.Hidden = true
		g.chart[c].gotpoints = 0
	}
}

// adds data and converts to range 0.0-1.0 for mX and 0.0-1.0 for mY
func (g *game) getData() bool {
	var atdata, ptdata, dtdata bool
	if g.pt.PTrakrunning {
		if g.pt.simulatePTrak {
			g.d.mdata[0] = SimulatedPTrak()
		} else {
			g.d.mdata[0] = g.pt.GetPTrakdata()
		}
		ptdata = g.d.mdata[0] > 0
	}
	if g.dt.DustTrakrunning {
		if g.dt.simulateDustTrak {
			g.d.mdata[1] = SimulatedDustTrak()
		} else {
			g.d.mdata[1] = g.dt.GetDustTrakdata()
		}
		dtdata = g.d.mdata[0] > 0
	}
	if g.at.AeroTrakrunning {
		if g.at.simulateAeroTrak {
			g.d.mdata[2] = SimulatedAeroTrak()[0]
			g.d.mdata[3] = SimulatedAeroTrak()[1]
			g.d.mdata[4] = SimulatedAeroTrak()[2]
			g.d.mdata[5] = SimulatedAeroTrak()[3]
			g.d.mdata[6] = SimulatedAeroTrak()[4]
			g.d.mdata[7] = SimulatedAeroTrak()[5]
		} else {
			m := g.at.GetAeroTrakdata()
			if m[2] > 0 {
				g.d.mdata[2] = m[0]
				g.d.mdata[3] = m[1]
				g.d.mdata[4] = m[2]
				g.d.mdata[5] = m[3]
				g.d.mdata[6] = m[4]
				g.d.mdata[7] = m[5]
			}
		}
		atdata = g.d.mdata[2] > 0
	}
	g.d.tstamp = time.Now().Format(time.RFC3339)
	if !g.synchronizedmeasuring {
		return atdata || dtdata || ptdata
	} else {
		if g.at.AeroTrakrunning && !atdata {
			return false
		}
		if g.dt.DustTrakrunning && !dtdata {
			return false
		}
		if g.pt.PTrakrunning && !ptdata {
			return false
		}
		return true
	}
}

func (g *game) updateCharts() {
	for c := 0; c < chartnum; c++ {
		if g.d.mdata[c] > -1 {
			g.chart[c].gotpoints++
			if g.chart[c].gotpoints > g.datapoints {
				g.chart[c].gotpoints = g.datapoints
			}
			g.chart[c].mY[g.chart[c].gotpoints-1] = float32(g.d.mdata[c])
			g.chart[c].mTime[g.chart[c].gotpoints-1] = time.Now().Format(time.RFC3339)[14:19]

			if g.chart[c].mY[g.chart[c].gotpoints-1] > g.chart[c].mMax {
				newmax := g.chart[c].FixMax(int(g.chart[c].mY[g.chart[c].gotpoints-1]))
				// recalculate all values based on the change to max!!!!
				for i := 0; i < g.chart[c].gotpoints-1; i++ {
					g.chart[c].mY[i] = g.chart[c].mY[i] * g.chart[c].mMax / float32(newmax+1)
				}
				// new max? set max into struct
				g.chart[c].mMax = float32(newmax)
				g.chart[c].Refreshchart(c)
			}
			g.chart[c].mX[g.chart[c].gotpoints-1] = 1
			g.chart[c].mY[g.chart[c].gotpoints-1] = g.chart[c].mY[g.chart[c].gotpoints-1] / (g.chart[c].mMax + 1)
			for i := 1; i < g.chart[c].gotpoints; i++ {
				g.chart[c].mX[i-1] = float32(float32(i) / float32(g.chart[c].gotpoints))
				if g.chart[c].gotpoints == g.datapoints {
					g.chart[c].mY[i-1] = g.chart[c].mY[i]
					g.chart[c].mTime[i-1] = g.chart[c].mTime[i]
				}
			}
		}
	}
}

//Clears current settings and then searches for instruments and sets Prefernces to found ports
func (g *game) SetupInstruments() {
	var err error
	var s string
	log.Println("Searching for TSI  network instruments DustTrak & AeroTrak")
	log.Println("Searching for TSI serial instrument PTrak")
	instsettings := fmt.Sprintf("AeroTrak: %v\nDustTrak: %v\nPTrak: %v",
		g.app.Preferences().String("AeroTrak"), g.app.Preferences().String("DustTrak"), g.app.Preferences().String("PTrak"))

	log.Println("Clearing old instruments settings! (" + instsettings)
	g.app.Preferences().SetString("AeroTrak", "")
	g.app.Preferences().SetString("DustTrak", "")
	g.app.Preferences().SetString("PTrak", "")
	_, err = g.MyTSIscanner()
	if err != nil {
		log.Println("#1 SetupInstruments - Problems finding network instruments.\n", err.Error())
	}
	_, err = g.pt.FindPTrak()
	if err != nil {
		log.Println("#2 SetupInstruments - Problems finding serial instrument.", err.Error())
	}
	if g.app.Preferences().String("AeroTrak") > "" {
		s += "\nAeroTrak: " + g.app.Preferences().String("AeroTrak")
	}
	if g.app.Preferences().String("DustTrak") > "" {
		s += "\nDustTrak: " + g.app.Preferences().String("DustTrak")
	}
	if g.app.Preferences().String("PTrak") > "" {
		s += "PTrak: " + g.app.Preferences().String("PTrak")
	}
	log.Println("Finished autosetup of instruments: " + s)
	dialog.ShowInformation("Found instruments: ", strings.Replace(s, ", ", "\n", -1), g.window)
}

// sets instruments to running depending of simulated settings and Preferences ports
func (g *game) checkinstruments() {
	g.at.AeroTrakport = g.app.Preferences().StringWithFallback("AeroTrak", "")
	g.dt.DustTrakport = g.app.Preferences().StringWithFallback("DustTrak", "")
	g.pt.PTrakport = g.app.Preferences().StringWithFallback("PTrak", "")
	g.at.simulateAeroTrak = g.app.Preferences().BoolWithFallback("SimulateAeroTrak", true) && (g.at.AeroTrakport == "")
	g.dt.simulateDustTrak = g.app.Preferences().BoolWithFallback("SimulateDustTrak", true) && (g.dt.DustTrakport == "")
	g.pt.simulatePTrak = g.app.Preferences().BoolWithFallback("SimulatePTrak", true) && (g.pt.PTrakport == "")
	g.at.AeroTrakrunning = (g.at.AeroTrakport > "" || g.at.simulateAeroTrak)
	g.dt.DustTrakrunning = (g.dt.DustTrakport > "" || g.dt.simulateDustTrak)
	g.pt.PTrakrunning = (g.pt.PTrakport > "" || g.pt.simulatePTrak)
	g.app.Preferences().SetBool("SimulateAeroTrak", g.at.simulateAeroTrak)
	g.app.Preferences().SetBool("SimulateDustTrak", g.dt.simulateDustTrak)
	g.app.Preferences().SetBool("SimulatePTrak", g.pt.simulatePTrak)
}

func (g *game) closeapp() {
	g.StopMeasurement()
	g.app.Preferences().SetFloat("winWidth", float64(g.window.Canvas().Size().Width))
	g.app.Preferences().SetFloat("winHeight", float64(g.window.Canvas().Size().Height))
	time.Sleep(2 * time.Second)
	g.app.Quit()
}

// Charts!
var ChartColors = [8]color.Color{colornames.Green, colornames.Yellow, colornames.Red, colornames.Red,
	colornames.Red, colornames.Red, colornames.Red, colornames.Red}
var ChartTitles = [8]string{"Nano Particles 0.02-1.0µm", "Particles µg", "Fine Particles 0.3-0.5µm", "Fine Particles 0.5-1.0µm",
	"Fine Particles 1.0-3.0µm", "Coarse Particles 3.0-5.0µm", "Coarse Particles 5.0-10.0µm", "Coarse Particles >10.0µm"}
var YaxisTitles = [8]string{"Nano Particles 0.02-1.0µm", "Particles µg", "Fine Particles 0.3-0.5µm", "Fine Particles 0.5-1.0µm",
	"Fine Particles 1.0-3.0µm", "Coarse Particles 3.0-5.0µm", "Coarse Particles 5.0-10.0µm", "Coarse Particles >10.0µm"}

func (g *game) setupcharts() {
	g.chart[0].mChart.Hidden = !g.pt.PTrakrunning
	g.chart[1].mChart.Hidden = !g.dt.DustTrakrunning
	g.chart[2].mChart.Hidden = !g.at.AeroTrakrunning
	g.chart[3].mChart.Hidden = !g.at.AeroTrakrunning
	g.chart[4].mChart.Hidden = !g.at.AeroTrakrunning
	g.chart[5].mChart.Hidden = !g.at.AeroTrakrunning
	g.chart[6].mChart.Hidden = !g.at.AeroTrakrunning
	g.chart[7].mChart.Hidden = !g.at.AeroTrakrunning
	for c := 0; c < chartnum; c++ {
		g.chart[c].gotpoints = 0
		for i := 0; i < datapointsmax; i++ {
			g.chart[c].mLine[i].Position1.X = 0
			g.chart[c].mLine[i].Position2.X = 0
			g.chart[c].mLine[i].Position1.Y = 0
			g.chart[c].mLine[i].Position2.Y = 0
			if i < g.datapoints {
				g.chart[c].mLine[i].StrokeColor = g.getStrokeColor(c)
			}
			g.chart[c].mLine[i].StrokeWidth = 2
		}
		g.chart[c].gotpoints = 0
	}
	if len(g.cont.Objects) == 0 {
		if g.countrunning() == 1 {
			g.cont = *container.NewMax()
		} else {
			if g.countrunning() == 2 {
				g.cont = *container.NewGridWithColumns(2)
			} else {
				g.cont = *container.NewGridWithRows(2)
			}
		}
		for c := 0; c < chartnum; c++ {
			g.chart[c].DrawChart(g.chartSize)
			g.cont.Add(&g.chart[c].mChart)
			g.chart[c].Refreshchart(c)
		}
		g.window.SetContent(&g.cont)
	}
	g.cont.Refresh()
}

func (g *game) drawCharts() {
	for c := 0; c < chartnum; c++ {
		if g.chart[c].gotpoints > 1 {
			if g.chart[c].gotpoints > g.datapoints {
				g.chart[c].gotpoints = g.datapoints
			}
			g.chart[c].drawChartLines(g.chart[c].gotpoints)
			g.chart[c].drawXaxistexts(g.chart[c].mTime)
			g.chart[c].Refreshchart(c)
		}
	}
	g.cont.Refresh()
}

func (g *game) typedRune(r rune) {
	if r == ' ' {
		g.paused = !g.paused
	}
}
func (g *game) Tapped(ev *fyne.PointEvent) {
	g.paused = !g.paused
	g.Refresh()
}

func (g *game) TappedSecondary(ev *fyne.PointEvent) {
	g.paused = !g.paused
	g.Refresh()
}

func openstoragelocation(g *game) {
	var err error
	var fn string = g.app.Storage().RootURI().Path()
	if err != nil {
		log.Println("Error getting path to "+fn+".", err)
	}
	log.Println("Opening " + fn)
	err = g.openurl(fn)
	if err != nil {
		log.Println("Error opening "+fn+".", err)
	}
}

func (g *game) ShowMyDebug(s string) {
	if g.MyDebug {
		log.Printf("FUNCTION: %v\n%v\n", printfunc(), s)
	}
}
func printfunc() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

func (g *game) openurl(uu string) error {
	var u *url.URL
	var err error
	u, err = url.Parse(uu)
	if err != nil {
		log.Println("#1 openurl error", err.Error())
	}
	err = fyne.App.OpenURL(g.app, u)
	return err
}

// setups logging,  handles debugging, based on myDebug variable. if myDebug is on, it also prints all logging to console
func (g *game) setuplogging() {
	var logfile *os.File
	var err error
	g.Logpath, err = storage.Child(g.app.Storage().RootURI(), g.app.UniqueID()+".log")
	if err != nil {
		log.Fatal("setuplogging storage.Child error", err.Error())
	}

	logfile, err = os.OpenFile(g.Logpath.Path(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("err_setup#2", err.Error())
	}
	if g.MyDebug {
		multi := io.MultiWriter(logfile, os.Stdout)
		log.SetOutput(multi)
	} else {
		log.SetOutput(os.Stdout)
	}
	//	log.Printf("Log =\t%v", g.Logpath.Path())
	//	log.Printf("Settings =\t%v", strings.Replace(g.Logpath.Path(), g.app.UniqueID()+".log", "preferences.json", 1))
}

func GetcurrentIP() []string {
	// get current computers IP-address
	var err error
	var host string
	var a2 net.IP
	var a1, theip []string
	host, err = os.Hostname()
	if err != nil {
		log.Println("No network???!")
		return nil
	}

	a1, err = net.LookupHost(host)
	if err != nil {
		log.Println("Host lookup failed")
	}
	for i := 0; i < len(a1); i++ {
		a2 = net.ParseIP(a1[i])
		if a2 != nil {
			theip = append(theip, a2.String())
		}
	}
	return theip
}

func Getint64(s string) int64 {
	var err error
	var v int64
	v, err = strconv.ParseInt(s, 0, 64)
	if err != nil {
		log.Println("in64 conversion error", err.Error())
	}
	//	fmt.Printf("Getint64 conversion from %v -> %v \n",s,v)
	return v
}

// Finds IP-instruments AeroTrak and DustTrak
// with 502 and 3602 ports and returns []string with "DustTrak=192.168.0.121:3602" and "AeroTrak=192.168.0.171:502"
// Setini("AeroTrak") and Setini("DustTrak") is set
func (g *game) MyTSIscanner() ([]string, error) {
	var tcpinstruments []string
	var adr string
	var err error
	var myIP []string = GetcurrentIP()
	if myIP != nil {
		return nil, nil
	}
	for i := 0; i < len(myIP); i++ {
		for myIP[i][len(myIP[i])-1] != '.' {
			myIP[i] = myIP[i][:len(myIP[i])-1]
		}
	}
	timeout := time.Duration(time.Second / 100)
	for j := 0; j < len(myIP); j++ {
		for i := 0; i < 256; i++ {
			adr = fmt.Sprintf("%v%v%v", myIP[j], i, ":502")
			_, err = net.DialTimeout("tcp", adr, timeout)
			if err == nil {
				tcpinstruments = append(tcpinstruments, "AeroTrak="+adr)
				g.app.Preferences().SetString("AeroTrak", adr)
				log.Println("Found AeroTrak at " + adr)
				continue
			}
			adr = fmt.Sprintf("%v%v%v", myIP[j], i, ":3602")
			_, err = net.DialTimeout("tcp", adr, timeout)
			if err == nil {
				tcpinstruments = append(tcpinstruments, "DustTrak="+adr)
				g.app.Preferences().SetString("DustTrak", adr)
				log.Println("Found DustTrak at " + adr)
			}
		}
	}
	return tcpinstruments, err
}

func (g *game) Doftp(fn string) {
	var err error
	var content []byte
	var buf *bytes.Buffer
	g.f1.ftpserver = g.app.Preferences().StringWithFallback("ftpserver", "")
	g.f1.ftpusername = g.app.Preferences().StringWithFallback("ftpusername", "")
	g.f1.ftppassword = g.app.Preferences().StringWithFallback("ftppassword", "")
	g.f1.ftpdir = g.app.Preferences().StringWithFallback("ftpdir", "")
	if len(g.f1.ftpserver) == 0 || len(g.f1.ftpusername) == 0 || len(g.f1.ftpserver) == 0 {
		return
	}
	err = g.f1.ftplogin()
	if err != nil {
		log.Println("#1 DoUpload Login failed", err.Error())
		return
	}
	content, err = ioutil.ReadFile(filepath.Join(filepath.Dir(g.d.Databasepath), fn))
	buf = bytes.NewBuffer(content)
	if err != nil {
		log.Println("#2 Doftp reading ", filepath.Join(filepath.Dir(g.d.Databasepath), fn), " failed:\n ", err.Error())
		return
	}
	err = g.f1.ftpupload(g.f1.ftpdir+fn, *buf)
	if err != nil {
		log.Println("#3 Doftp upload failed ", err.Error())
		return
	}
	g.f1.ftplogout()
}

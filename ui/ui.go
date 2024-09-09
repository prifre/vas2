package ui

// To create app, use "fyne package -os windows -icon resources/vas.png"
// to generate "loggor.go" I used; "fyne bundle -o loggor.go resources"

import (
	"bytes"
	"errors"
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
	"vas/general"
	"vas/vasdatabase"

	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	//	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/colornames"
)

const (
	version       = 0.47
	chartnum      = 8
	datapointsmax = 100
)

type game struct {
	MyDebug      bool
	Logpath      fyne.URI
	logo         fyne.Container
	Settingspath fyne.URI
	app          fyne.App
	window       fyne.Window
	winSize      fyne.Size
	start        time.Time
	widget.BaseWidget
	datapoints            int
	synchronizedmeasuring bool
	autostartmeasuring    bool
	recordcount           int
	laps                  int
	paused                bool
	endmeasuring          bool
	cont                  fyne.Container
	chart                 *[]fyne.Container
	chartSize             fyne.Size
	countunits            int32
	averagepoints         int
	retries               int	
}

var tbl []string = []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}
var g game

// Create will stitch together all ui components
func Create(window fyne.Window) *fyne.Window {
	Setupfiles()
	//	os.Setenv("FYNE_SCALE", "1.0")
	//	var g game = *newGame(&[8]LineChart{})
	g.window.SetMaster()
	g.window.SetMainMenu(g.buildMenu())
	g.addkeyshortcuts()
	g.window.Canvas().SetOnTypedRune(g.typedRune)
	g.MyDebug = fyne.CurrentApp().Preferences().BoolWithFallback("mydebug", true)
	w := fyne.CurrentApp().Preferences().FloatWithFallback("winWidth", 1024)
	h := fyne.CurrentApp().Preferences().FloatWithFallback("winHeight", 768)
	if w < 800 {
		w = 800
	}
	if h < 600 {
		h = 600
	}
	g.winSize = fyne.NewSize(float32(w), float32(h))
	g.window.Resize(g.winSize)
	g.datapoints = fyne.CurrentApp().Preferences().IntWithFallback("datapoints", 10)
	cu := fyne.CurrentApp().Preferences().IntWithFallback("countunits", 0)
	g.countunits = []int32{1, 21201, 600}[cu]
	g.averagepoints = fyne.CurrentApp().Preferences().IntWithFallback("averagepoints", 30)
	g.synchronizedmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("synchronizedmeasuring", true)
	g.autostartmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", true)
	g.setuplogging()
	new(vasdatabase.DBtype).Setupdb()
	g.start = time.Now()
	log.Printf("Program started %v", g.start)
	g.window.SetCloseIntercept(func() {
		g.closeapp()
	})
	g.window.CenterOnScreen()
	g.showlogo()
	if g.autostartmeasuring {
		log.Printf("Autostart active, starting %v",time.Now().UnixNano())
		g.StartMeasurement()
	} else {
		g.StopMeasurement()
	}
	return &g.window
}

func (g *game) Measure() {
	var msg string
	var err error
	var fn string
	for {
		if g.endmeasuring {
			break
		}
		if g.getData() && !g.paused {
			g.recordcount = g.recordcount + 1
			r := runtime.NumGoroutine()
			err = new(vasdatabase.DBtype).Addmeasurement()
			if err != nil {
				log.Println("#1 database error: ", err.Error())
			}
			err = Updatecharts()
			if err != nil {
				log.Println("#2 Updatecharts error: ", err.Error())
			}
			for r < runtime.NumGoroutine() {
				time.Sleep(time.Millisecond)
			}
			g.laps = g.laps + 1
			if time.Now().String()[14:19] == "00:00" {
				pauserestore := g.paused
				g.paused = true
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
				fn = general.Doscreenshot(g.window)
				if fn > "" {
					g.Doftp(fn)
				}
				g.paused = pauserestore
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

func (g *game) addkeyshortcuts() {
	key_s := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_s, func(shortcut fyne.Shortcut) {
		g.StartMeasurement()
	})
	key_e := desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_e, func(shortcut fyne.Shortcut) {
		g.StopMeasurement()
	})
	key_q := desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_q, func(shortcut fyne.Shortcut) {
		g.closeapp()
	})
	key_u := desktop.CustomShortcut{KeyName: fyne.KeyU, Modifier: desktop.ControlModifier}
	g.window.Canvas().AddShortcut(&key_u, func(shortcut fyne.Shortcut) {
		g.Checkforupdate()
	})
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
		}),
		fyne.NewMenuItem("Manual Settings...", func() {
			g.StopMeasurement()
			DoManualSettings(g)
		}),
		fyne.NewMenuItem("Special Aerotrak Settings...", func() {
			g.StopMeasurement()
			DoAeroTrakSettings(g)
		}))
	// mEdit := fyne.NewMenu("Edit",
	// 	fyne.NewMenuItem("Cut", func() { dialog.ShowInformation("Cut? ", "Not implemented, sorry!", g.window) }),
	// 	fyne.NewMenuItem("Copy", func() { dialog.ShowInformation("Copy? ", "Not implemented, sorry!", g.window) }),
	// 	fyne.NewMenuItem("Paste", func() { dialog.ShowInformation("Paste? ", "Not implemented, sorry!", g.window) }))
	mMeasurements := fyne.NewMenu("Measurements",
		fyne.NewMenuItem("Start Measuring", func() {
			g.StartMeasurement()
		}),
		fyne.NewMenuItem("End Measuring", func() {
			g.StopMeasurement()
			g.d.Closemeasurement()
			log.Printf("Measurement '%v' (%v) stopped", g.d.mname, g.d.nanostamp)
			fyne.CurrentApp().Preferences().SetString("nanostamp", "0")
			g.d.mname = ""
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Measurements maintainance", func() {
			g.StopMeasurement()
			new(vasdatabase.DBtype).DoMeasurements(g)
		}),
		fyne.NewMenuItem("Save screen", func() {
			time.Sleep(time.Second)
			new(general).Doscreenshot()
		}),
		fyne.NewMenuItem("Upload to FTP-server", func() {
			var fn string
			for i := 0; i < len(tbl); i++ {
				fn = tbl[i] + ".txt"
				hd:=fyne.CurrentApp().Preferences().String("homedir")
				_, err = os.Stat(filepath.Join(hd, fn))
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
			g.showlogo2()
			dialog.ShowInformation("About...", t, g.window)
			g.showlogo2()
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
		fyne.NewMenuItem("Open the VISIBLE AIR SYSTEM webpage!", func() {
			g.openurl("http://www.prifre.com/vas")
		}),
	)
	return fyne.NewMainMenu(mFile, mMeasurements, mHelp)
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
	// if g.window.Canvas().Size().Width < 5 || g.window.Canvas().Size().Height < 5 {
	// 	return
	// } else {
	// 	g.winSize.Width = g.window.Canvas().Size().Width
	// 	g.winSize.Height = g.window.Canvas().Size().Height
	// }
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
	if g.d.nanostamp == 0 {
		if !fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", false) {
			return fmt.Errorf("no autostart")
		}
		if Getint64(fyne.CurrentApp().Preferences().StringWithFallback("nanostamp", "0")) == 0 {
			return fmt.Errorf("no nanostamp")
		}
		g.d.nanostamp = Getint64(fyne.CurrentApp().Preferences().StringWithFallback("nanostamp", "0"))
	}
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
	g.synchronizedmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("synchronizedmeasuring", true)
	g.autostartmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", true)
	g.checkinstruments()
	g.at.AeroTrakport = fyne.CurrentApp().Preferences().StringWithFallback("AeroTrak", "")
	g.dt.DustTrakport = fyne.CurrentApp().Preferences().StringWithFallback("DustTrak", "")
	g.pt.PTrakport = fyne.CurrentApp().Preferences().StringWithFallback("PTrak", "")
	if !g.at.AeroTrakrunning && !g.dt.DustTrakrunning && !g.pt.PTrakrunning {
		return
	}
	err = g.getoldmeasurement()
	g.d.dbstart = time.Now()
	if err != nil {
		if err.Error() != "no autostart" {
			log.Println("Restart not possible due to: ", err.Error())
		}
		log.Println("Autocreating new measurement!")
		g.d.nanostamp = time.Now().UnixNano()
		fyne.CurrentApp().Preferences().SetString("nanostamp", fmt.Sprintf("%v", g.d.nanostamp))
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
		err = g.d.Addmeasurement()
		if err != nil {
			log.Println("#1 StartMeasurement - Problems adding Main measurement ", err.Error())
		}
	}
	g.window.Resize(g.winSize)
	g.windowResize()
	g.Setupcharts()
	g.window.SetContent(&g.cont)
	log.Printf("Measurement '%v' (%v) started at %v.", g.d.mname, g.d.nanostamp, g.d.tstamp)
	g.recordcount = 0
	g.paused = false
	g.endmeasuring = false
	go g.Measure()
	// g.d.Pruning(g)
	// os.Exit(0)
}

func (g *game) StopMeasurement() {
	g.showlogo()
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
	// for c := 0; c < chartnum; c++ {
	// 	for i := 0; i < g.datapoints; i++ {
	// 		g.chart[c].mLine[i].Position1.X = 0
	// 		g.chart[c].mLine[i].Position2.X = 0
	// 	}
	// 	g.chart[c].mChart.Hidden = true
	// 	g.chart[c].gotpoints = 0
	// }
	g.d.nanostamp = 0
}

// adds data and converts to range 0.0-1.0 for mX and 0.0-1.0 for mY
func (g *game) getData() bool {
	var atdata, ptdata, dtdata bool
	g.d.mdata[0] = -1
	g.d.mdata[1] = -1
	g.d.mdata[2] = -1
	if g.pt.PTrakrunning {
		if g.pt.simulatePTrak {
			g.d.mdata[0] = new(vasinstruments).SimulatedPTrak()
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
			g.d.mdata[2] = SimulatedAeroTrak()[0] * g.countunits
			g.d.mdata[3] = SimulatedAeroTrak()[1] * g.countunits
			g.d.mdata[4] = SimulatedAeroTrak()[2] * g.countunits
			g.d.mdata[5] = SimulatedAeroTrak()[3] * g.countunits
			g.d.mdata[6] = SimulatedAeroTrak()[4] * g.countunits
			g.d.mdata[7] = SimulatedAeroTrak()[5] * g.countunits
		} else {
			m, err := g.at.GetAeroTrakdata()
			if err != nil {
				g.retries++
				if g.retries > 5 {
					dialog.ShowInformation("Problems!", "Problem communicating with AeroTrak.", g.window)
					g.retries = 0
					g.StopMeasurement()
				}
			} else {
				g.retries = 0
			}
			if m[0] > 0 {
				g.d.mdata[2] = m[0] * g.countunits
				g.d.mdata[3] = m[1] * g.countunits
				g.d.mdata[4] = m[2] * g.countunits
				g.d.mdata[5] = m[3] * g.countunits
				g.d.mdata[6] = m[4] * g.countunits
				g.d.mdata[7] = m[5] * g.countunits
				if g.at.showdata {
					fmt.Printf("Data (%v): %v %v\n", g.recordcount, g.d.tstamp[11:19], g.d.mdata)
				}
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

//Clears current settings and then searches for instruments and sets Prefernces to found ports
func (g *game) SetupInstruments() {
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
	_, err = g.pt.FindPTrak()
	if err != nil {
		log.Println("#1 SetupInstruments:FindPTrak:\n", err.Error())
	}
	if g.MyTSIscanner() != nil {
		log.Println("#2 SetupInstruments:MyTSIscanner:\n", err.Error())
	}
	if fyne.CurrentApp().Preferences().String("AeroTrak") > "" {
		s += "\nAeroTrak: " + fyne.CurrentApp().Preferences().String("AeroTrak")
	}
	if fyne.CurrentApp().Preferences().String("DustTrak") > "" {
		s += "\nDustTrak: " + fyne.CurrentApp().Preferences().String("DustTrak")
	}
	if fyne.CurrentApp().Preferences().String("PTrak") > "" {
		s += "\nPTrak: " + fyne.CurrentApp().Preferences().String("PTrak")
	}
	if s == "" {
		dialog.ShowInformation("", "No instruments found.", g.window)
	} else {
		log.Println("Finished autosetup of instruments: " + s)
		dialog.ShowInformation("Found instruments: ", strings.Replace(s, ", ", "\n", -1), g.window)
	}
}

// sets instruments to running depending of simulated settings and Preferences ports
func (g *game) checkinstruments() {
	g.at.AeroTrakport = fyne.CurrentApp().Preferences().StringWithFallback("AeroTrak", "")
	g.dt.DustTrakport = fyne.CurrentApp().Preferences().StringWithFallback("DustTrak", "")
	g.pt.PTrakport = fyne.CurrentApp().Preferences().StringWithFallback("PTrak", "")
	g.at.simulateAeroTrak = fyne.CurrentApp().Preferences().BoolWithFallback("SimulateAeroTrak", true) && (g.at.AeroTrakport == "")
	g.dt.simulateDustTrak = fyne.CurrentApp().Preferences().BoolWithFallback("SimulateDustTrak", true) && (g.dt.DustTrakport == "")
	g.pt.simulatePTrak = fyne.CurrentApp().Preferences().BoolWithFallback("SimulatePTrak", true) && (g.pt.PTrakport == "")
	g.at.AeroTrakrunning = (g.at.AeroTrakport > "" || g.at.simulateAeroTrak)
	g.dt.DustTrakrunning = (g.dt.DustTrakport > "" || g.dt.simulateDustTrak)
	g.pt.PTrakrunning = (g.pt.PTrakport > "" || g.pt.simulatePTrak)
	fyne.CurrentApp().Preferences().SetBool("SimulateAeroTrak", g.at.simulateAeroTrak)
	fyne.CurrentApp().Preferences().SetBool("SimulateDustTrak", g.dt.simulateDustTrak)
	fyne.CurrentApp().Preferences().SetBool("SimulatePTrak", g.pt.simulatePTrak)
	i64, _ := strconv.ParseInt(fyne.CurrentApp().Preferences().StringWithFallback("ATdelay", "200"), 10, 64)
	g.at.defdelay = i64
}

func (g *game) closeapp() {
	g.StopMeasurement()
	fyne.CurrentApp().Preferences().SetFloat("winWidth", float64(g.window.Canvas().Size().Width))
	fyne.CurrentApp().Preferences().SetFloat("winHeight", float64(g.window.Canvas().Size().Height))
	g.app.Quit()
}

// Charts!
var ChartColors = [8]color.Color{colornames.Green, colornames.Yellow, colornames.Red, colornames.Red,
	colornames.Red, colornames.Red, colornames.Red, colornames.Red}
var ChartTitles = [8]string{
	"   Nano Particles\t0.02-1.0µm (#/cm³)",
	"Weight Particles\t          µg/m³",
	"Fine Particles\t   0.3-0.5µm",
	"Fine Particles\t   0.5-1.0µm",
	"Fine Particles\t   1.0-3.0µm",
	"Coarse Particles\t   3.0-5.0µm",
	"Coarse Particles\t   5.0-10.0µm",
	"Coarse Particles\t     >10.0µm"}

//var YaxisTitles = [8]string{"Nano Particles 0.02-1.0µm", "Particles µg", "Fine Particles 0.3-0.5µm", "Fine Particles 0.5-1.0µm",
//	"Fine Particles 1.0-3.0µm", "Coarse Particles 3.0-5.0µm", "Coarse Particles 5.0-10.0µm", "Coarse Particles >10.0µm"}

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
	var fn string = g.getdocumentpath().Path()
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
	g.Logpath, err = storage.Child(g.getdocumentpath(), g.app.UniqueID()+".log")
	fmt.Println()
	if err != nil {
		log.Println("#1 setuplogging:storage.Child error: ", err.Error())
		return
	}

	logfile, err = os.OpenFile(g.Logpath.Path(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("#2 setuplogging:OpenFile  error: ", err.Error())
		return
	}
	if g.MyDebug {
		multi := io.MultiWriter(logfile, os.Stdout)
		log.SetOutput(multi)
	} else {
		log.SetOutput(logfile)
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
func (g *game) MyTSIscanner() error {
	var tcpinstruments []string
	var adr string
	var err error
	var myIP []string = GetcurrentIP()
	if myIP == nil {
		return errors.New("#1 MyTSIscanner:no current IP address exists")
	}
	for i := 0; i < len(myIP); i++ {
		fmt.Println("#", myIP[i], "#")
		if strings.Contains(myIP[i], ".") {
			for myIP[i][len(myIP[i])-1] != '.' {
				myIP[i] = myIP[i][:len(myIP[i])-1]
			}
		}
	}
	timeout := time.Duration(time.Second / 200)
	for j := 0; j < len(myIP); j++ {
		if strings.Contains(myIP[j], ".") {
			log.Println("Searching IP: ", myIP[j])
			for i := 0; i < 256; i++ {
				time.Sleep(time.Duration(10))
				adr = fmt.Sprintf("%v%v%v", myIP[j], i, ":502")
				_, err = net.DialTimeout("tcp", adr, timeout)
				if err == nil {
					tcpinstruments = append(tcpinstruments, "AeroTrak="+adr)
					fyne.CurrentApp().Preferences().SetString("AeroTrak", adr)
					log.Println("Found AeroTrak at " + adr)
					continue
				}
				adr = fmt.Sprintf("%v%v%v", myIP[j], i, ":3602")
				_, err = net.DialTimeout("tcp", adr, timeout)
				if err == nil {
					tcpinstruments = append(tcpinstruments, "DustTrak="+adr)
					fyne.CurrentApp().Preferences().SetString("DustTrak", adr)
					log.Println("Found DustTrak at " + adr)
				}
			}
		}
	}
	if len(tcpinstruments) == 0 {
		return errors.New("no instruments found")
	}
	return nil
}

func (g *game) Doftp(fn string) {
	var err error
	var content []byte
	var buf *bytes.Buffer
	f1:=new(vasftp.ftptype)
	f1.ftpserver = fyne.CurrentApp().Preferences().StringWithFallback("ftpserver", "")
	f1.ftpusername = fyne.CurrentApp().Preferences().StringWithFallback("ftpusername", "")
	f1.ftppassword = fyne.CurrentApp().Preferences().StringWithFallback("ftppassword", "")
	g.f1.ftpdir = fyne.CurrentApp().Preferences().StringWithFallback("ftpdir", "")
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
		log.Println("#2 Doftp:ReadFile: ", filepath.Join(filepath.Dir(g.d.Databasepath), fn), " failed:\n ", err.Error())
		return
	}
	err = g.f1.ftpupload(g.f1.ftpdir+fn, *buf)
	if err != nil {
		log.Println("#3 Doftp:ftpupload: ", err.Error())
		return
	}
	g.f1.ftplogout()
}

func (g *game) getdocumentpath() fyne.URI {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	u := storage.NewFileURI(fyne.CurrentApp().Preferences().StringWithFallback("documentpath", path))
	if _, err := os.Stat("/path/to/whatever"); os.IsNotExist(err) {
		u = storage.NewFileURI(path)
	}
	return u
}

func (g *game) showlogo() {
	g.logo = *container.NewMax(canvas.NewImageFromResource(resource4SafeAirLogga2Jpg))
	// g.cont.Add(&g.logo)
	// g.cont.Refresh()
	if len(g.logo.Objects) == 0 {
		g.logo.Add(&g.logo)
	}
	g.window.SetContent(&g.logo)
	g.logo.Refresh()
}
func (g *game) showlogo2() {
	g.logo = *container.NewMax(canvas.NewImageFromResource(resourceLAB360LogoLightBluePng))
	// g.cont.Add(&g.logo)
	// g.cont.Refresh()
	if len(g.logo.Objects) == 0 {
		g.logo.Add(&g.logo)
	}
	g.window.SetContent(&g.logo)
	g.logo.Refresh()
}

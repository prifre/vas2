package ui

// To create app, use "fyne package -os windows -icon resources/vas.png"
// to generate "loggor.go" I used; "fyne bundle -o loggor.go resources"

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"vas/vasftp"
	"vas/vasmeasure"

	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	//	"fyne.io/fyne/v2/container"
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
	measure 	 *vasmeasure.Measuretype
}

var tbl []string = []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}

// Create will stitch together all ui components
func Create(window fyne.Window) *fyne.Window {
	
	// g.MyDebug = fyne.CurrentApp().Preferences().BoolWithFallback("mydebug", true)
	Setupfiles()
	//	os.Setenv("FYNE_SCALE", "1.0")
	//	var g game = *newGame(&[8]LineChart{})
	window.SetMaster()
	window.SetMainMenu(BuildMenu())
	// Addkeyshortcuts()
	// g.window.Canvas().SetOnTypedRune(g.TypedRune)
	// WindowSetup()
	// g.measure =vasmeasure.MeasurementSetup()
	// g.setuplogging()
	// new(vasdatabase.DBtype).Setupdb()
	// g.measure.Start = time.Now()
	// log.Printf("Program started %v", g.measure.Start)
	// g.window.SetCloseIntercept(func() {
	// 	g.Closeapp()
	// })
	window.CenterOnScreen()
	// g.showlogo()
	// if g.measure.autostartmeasuring {
	// 	log.Printf("Autostart active, starting %v",time.Now().UnixNano())
	// 	g.measure.StartMeasurement()
	// } else {
	// 	g.measure.StopMeasurement()
	// }
	return &window
}


// sets instruments to running depending of Simulated settings and Preferences ports
func (g *game) Closeapp() {
	// g.StopMeasurement()
	fyne.CurrentApp().Preferences().SetFloat("winWidth", float64(g.window.Canvas().Size().Width))
	fyne.CurrentApp().Preferences().SetFloat("winHeight", float64(g.window.Canvas().Size().Height))
	g.app.Quit()
}


//var YaxisTitles = [8]string{"Nano Particles 0.02-1.0µm", "Particles µg", "Fine Particles 0.3-0.5µm", "Fine Particles 0.5-1.0µm",
//	"Fine Particles 1.0-3.0µm", "Coarse Particles 3.0-5.0µm", "Coarse Particles 5.0-10.0µm", "Coarse Particles >10.0µm"}

// func (g *vasmeasure.Measuretype) TypedRune(r rune) {
// 	if r == ' ' {
// 		g.Paused = !g.Paused
// 	}
// }
// func (g *vasmeasure.Measuretype) Tapped(ev *fyne.PointEvent) {
// 	g.Paused = !g.Paused
// }

// func (g *vasmeasure.Measuretype) TappedSecondary(ev *fyne.PointEvent) {
// 	g.Paused = !g.Paused
// }

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

func Doftp(fn string) {
	var err error
	var content []byte
	var buf *bytes.Buffer
	f1:=new(vasftp.Ftptype)
	f1.Ftpserver = fyne.CurrentApp().Preferences().StringWithFallback("ftpserver", "")
	f1.Ftpusername = fyne.CurrentApp().Preferences().StringWithFallback("ftpusername", "")
	f1.Ftppassword = fyne.CurrentApp().Preferences().StringWithFallback("ftppassword", "")
	f1.Ftpdir = fyne.CurrentApp().Preferences().StringWithFallback("ftpdir", "")
	if len(f1.Ftpserver) == 0 || len(f1.Ftpusername) == 0 || len(f1.Ftpserver) == 0 {
		return
	}
	err = f1.Ftplogin()
	if err != nil {
		log.Println("#1 DoUpload Login failed", err.Error())
		return
	}
	content, err = os.ReadFile(filepath.Join(filepath.Dir(GetHomeDir()), fn))
	buf = bytes.NewBuffer(content)
	hd:=fyne.CurrentApp().Preferences().String("homedir")
	if err != nil {
		log.Println("#2 Doftp:ReadFile: ", filepath.Join(hd, fn), " failed:\n ", err.Error())
		return
	}
	err = f1.Ftpupload(filepath.Join(hd, fn), *buf)
	if err != nil {
		log.Println("#3 Doftp:ftpupload: ", err.Error())
		return
	}
	f1.Ftplogout()
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

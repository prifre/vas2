package ui

// To create app, use "fyne package -os windows -icon resources/vas.png"
// to generate "loggor.go" I used; "fyne bundle -o loggor.go resources"

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"vas/general"
	"vas/vascharts"
	"vas/vasdatabase"
	"vas/vasftp"

	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	//	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
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
	datapoints            int
	synchronizedmeasuring bool
	autostartmeasuring    bool
	recordcount           int
	laps                  int
	paused                bool
	endmeasuring          bool
	cont                  fyne.Container
	retries               int	
}

var tbl []string = []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}
var g game

// Create will stitch together all ui components
func Create(window fyne.Window) *fyne.Window {
	g.MyDebug = fyne.CurrentApp().Preferences().BoolWithFallback("mydebug", true)
	Setupfiles()
	//	os.Setenv("FYNE_SCALE", "1.0")
	//	var g game = *newGame(&[8]LineChart{})
	g.window.SetMaster()
	g.window.SetMainMenu(g.BuildMenu())
	g.Addkeyshortcuts()
	g.window.Canvas().SetOnTypedRune(g.typedRune)
	WindowSetup(g.window)
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
//Clears current settings and then searches for instruments and sets Prefernces to found ports

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

func Doftp(fn string) {
	var err error
	var content []byte
	var buf *bytes.Buffer
	f1:=new(vasftp.Ftptype)
	f1.ftpserver = fyne.CurrentApp().Preferences().StringWithFallback("ftpserver", "")
	f1.ftpusername = fyne.CurrentApp().Preferences().StringWithFallback("ftpusername", "")
	f1.ftppassword = fyne.CurrentApp().Preferences().StringWithFallback("ftppassword", "")
	g.f1.ftpdir = fyne.CurrentApp().Preferences().StringWithFallback("ftpdir", "")
	if len(g.f1.ftpserver) == 0 || len(g.f1.ftpusername) == 0 || len(g.f1.ftpserver) == 0 {
		return
	}
	err = f1.Ftplogin()
	if err != nil {
		log.Println("#1 DoUpload Login failed", err.Error())
		return
	}
	content, err = os.ReadFile(filepath.Join(filepath.Dir(g.d.Databasepath), fn))
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

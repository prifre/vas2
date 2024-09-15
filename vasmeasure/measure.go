package vasmeasure

// To create app, use "fyne package -os windows -icon resources/vas.png"
// to generate "loggor.go" I used; "fyne bundle -o loggor.go resources"

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"vas/vascharts"
	"vas/vasdatabase"
	"vas/vasinstruments"

	"time"

	"fyne.io/fyne/v2"
	//	"fyne.io/fyne/v2/container"
)
type Measuretype struct {
	chart                 *[](vascharts.LineChart)
	chartSize             fyne.Size
	countunits            int32
	averagepoints         int
	synchronizedmeasuring bool
	autostartmeasuring    bool
	recordcount           int
	Start                 time.Time
	laps                  int
	retries               int
 	Paused                bool
	endmeasuring          bool
	MyDebug               bool
	Datapoints            int
	at vasinstruments.AeroTraktype
	dt vasinstruments.DustTraktype
	pt vasinstruments.PTraktype
	d *vasdatabase.DBtype
}
func MeasurementSetup() *Measuretype {
	g:=new(Measuretype)
	g.Datapoints = fyne.CurrentApp().Preferences().IntWithFallback("datapoints", 10)
	cu := fyne.CurrentApp().Preferences().IntWithFallback("countunits", 0)
	g.countunits = []int32{1, 21201, 600}[cu]
	g.averagepoints = fyne.CurrentApp().Preferences().IntWithFallback("averagepoints", 30)
	g.synchronizedmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("synchronizedmeasuring", true)
	g.autostartmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", true)
	return g
}
func (g *Measuretype) Measure() {
	var err error
	// var msg string
	// var fn string
	for {
		if g.endmeasuring {
			break
		}
		if g.GetData() && !g.Paused {
			g.recordcount = g.recordcount + 1
			r := runtime.NumGoroutine()
			err = new(vasdatabase.DBtype).Addmeasurement()
			if err != nil {
				log.Println("#1 database error: ", err.Error())
			}
			// err = vascharts.Updatecharts(nil,g.d.Mdata)
			if err != nil {
				log.Println("#2 Updatecharts error: ", err.Error())
			}
			for r < runtime.NumGoroutine() {
				time.Sleep(time.Millisecond)
			}
			g.laps = g.laps + 1
			if time.Now().String()[14:19] == "00:00" {
				pauserestore := g.Paused
				g.Paused = true
				runtime.GC()
				new(vasdatabase.DBtype).Pruning()
				// msg, err = new(vasdatabase.DBtype).Exporttotext()
				// if err != nil {
				// 	log.Println("Error exporting: ", msg)
				// } else {
				// 	for i := 0; i < len(tbl); i++ {
				// 		fn = tbl[i] + ".txt"
				// 		Doftp(fn)
				// 	}
				// }
				// fn = general.Doscreenshot(game.window)
				// if fn > "" {
				// 	Doftp(fn)
				// }
				g.Paused = pauserestore
			}
			// if g.recordcount%100 == 0 && g.MyDebug {
			// 	var m1 runtime.MemStats
			// 	runtime.ReadMemStats(&m1)
			// 	fmt.Printf("%v  Go: %v  Time: %vs  Rec: %v  Speed: %.2f Rec/s  Laps: %v  Stack: %v \n",
			// 		time.Now().String()[11:19],
			// 		runtime.NumGoroutine(), strings.Split(time.Since(g.Start).String(), ".")[0],
			// 		g.recordcount, 0, g.laps, m1.StackInuse)
			// 		// g.recordcount, float64(g.recordcount)/(time.Since(g.d.dbstart).Seconds()), g.laps, m.StackInuse)
			// }
		}
	}
}
func (g *Measuretype) Getoldmeasurement() error {
	var temp []string
	var err error
	if g.d.Nanostamp == 0 {
		if !fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", false) {
			return fmt.Errorf("no autostart")
		}
		if fyne.CurrentApp().Preferences().IntWithFallback("nanostamp", 0) == 0 {
			return fmt.Errorf("no nanostamp")
		}
		g.d.Nanostamp = int64(fyne.CurrentApp().Preferences().IntWithFallback("nanostamp",0))
	}
	temp, err = g.d.Getsql(fmt.Sprintf("SELECT mname FROM tblMain WHERE nanostamp=%v", g.d.Nanostamp))
	if err != nil {
		return fmt.Errorf("#1 Getting mname Error1: %v", err.Error())
	}
	if temp == nil {
		return fmt.Errorf("#1 Getting mname Error2: ''")
	}
	if len(temp) < 1 {
		return fmt.Errorf("#1 Getting mname Error3: %v", err.Error())
	}
	g.d.Mname = fmt.Sprintf("%v", temp[0])
	temp, err = g.d.Getsql(fmt.Sprintf("SELECT tstamp FROM tblMain WHERE nanostamp = %v", g.d.Nanostamp))
	if err != nil {
		return fmt.Errorf("#1 Getting tstamp Error: %v", err.Error())
	}
	if temp == nil {
		return fmt.Errorf("#1 Getting tstamp Error2: nil")
	}
	if len(temp) < 1 {
		return fmt.Errorf("#1 Getting tstamp Error3: %v", err.Error())
	}
	g.d.Tstamp = fmt.Sprintf("%v", temp[0])
	return nil
}
func (g *Measuretype) StartMeasurement() {
	var err error
	g.synchronizedmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("synchronizedmeasuring", true)
	g.autostartmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", true)
	// Checkinstruments()
	g.at.AeroTrakport = fyne.CurrentApp().Preferences().StringWithFallback("AeroTrak", "")
	g.dt.DustTrakport = fyne.CurrentApp().Preferences().StringWithFallback("DustTrak", "")
	g.pt.PTrakport = fyne.CurrentApp().Preferences().StringWithFallback("PTrak", "")
	if !g.at.AeroTrakrunning && !g.dt.DustTrakrunning && !g.pt.PTrakrunning {
		return
	}
	err = g.Getoldmeasurement()
	if err != nil {
		if err.Error() != "no autostart" {
			log.Println("Restart not possible due to: ", err.Error())
		}
		log.Println("Autocreating new measurement!")
		g.d.Nanostamp = time.Now().UnixNano()
		fyne.CurrentApp().Preferences().SetString("nanostamp", fmt.Sprintf("%v", g.d.Nanostamp))
		g.d.Tstamp = time.Now().Format(time.RFC3339)
		g.d.Mname = "Measurement" + g.d.Tstamp
		g.d.Mname = strings.Replace(g.d.Mname, "-", "", 99)
		g.d.Mname = strings.Replace(g.d.Mname, "/", "", 99)
		g.d.Mname = strings.Replace(g.d.Mname, ":", "", 99)
		g.d.Mname = strings.Replace(g.d.Mname, ".", "", 99)
		g.d.Mname = g.d.Mname[:26]
		if g.at.SimulateAeroTrak || g.dt.SimulateDustTrak || g.pt.SimulatePTrak {
			g.d.Note = "Simulated measurement "
			if g.at.SimulateAeroTrak {
				g.d.Note = g.d.Note + " AeroTrak"
			}
			if g.dt.SimulateDustTrak {
				g.d.Note = g.d.Note + " DustTrak"
			}
			if g.pt.SimulatePTrak {
				g.d.Note = g.d.Note + " PTrak"
			}
		}
		g.d.Mdata = [8]int32{-1, -1, -1, -1, -1, -1, -1, -1}
		err = g.d.Addmeasurement()
		if err != nil {
			log.Println("#1 StartMeasurement - Problems adding Main measurement ", err.Error())
		}
	}
	vascharts.Setupcharts()
	// g.window.SetContent(&g.cont)
	log.Printf("Measurement '%v' (%v) started at %v.", g.d.Mname, g.d.Nanostamp, g.d.Tstamp)
	g.recordcount = 0
	// g.paused = false
	g.endmeasuring = false
	go g.Measure()
	// g.d.Pruning(g)
	// os.Exit(0)
}

func (g *Measuretype) StopMeasurement() {
	// g.showlogo()
	g.endmeasuring = true
	var err error
	if g.at.AeroTrakrunning && !g.at.SimulateAeroTrak {
		time.Sleep(time.Second)
		err = g.at.AeroTrakstop()
		if err != nil {
			log.Println("#1 StopMeasurement - Problems AeroTrak stop", err.Error())
		}
	}
	if g.dt.DustTrakrunning && !g.dt.SimulateDustTrak {
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
	g.d.Nanostamp = 0
}

// adds data and converts to range 0.0-1.0 for mX and 0.0-1.0 for mY
func (g *Measuretype) etData() bool {
	var atdata, ptdata, dtdata bool
	g.d.Mdata[0] = -1
	g.d.Mdata[1] = -1
	g.d.Mdata[2] = -1
	if g.pt.PTrakrunning {
		if g.pt.SimulatePTrak {
			g.d.Mdata[0] = vasinstruments.SimulatedPTrak()
		} else {
			g.d.Mdata[0] = g.pt.GetPTrakdata()
		}
		ptdata = g.d.Mdata[0] > 0
	}
	if g.dt.DustTrakrunning {
		if g.dt.SimulateDustTrak {
			g.d.Mdata[1] = vasinstruments.SimulatedDustTrak()
		} else {
			g.d.Mdata[1] = g.dt.GetDustTrakdata()
		}
		dtdata = g.d.Mdata[0] > 0
	}
	if g.at.AeroTrakrunning {
		if g.at.SimulateAeroTrak {
			g.d.Mdata[2] = vasinstruments.SimulatedAeroTrak()[0] * g.countunits
			g.d.Mdata[3] = vasinstruments.SimulatedAeroTrak()[1] * g.countunits
			g.d.Mdata[4] = vasinstruments.SimulatedAeroTrak()[2] * g.countunits
			g.d.Mdata[5] = vasinstruments.SimulatedAeroTrak()[3] * g.countunits
			g.d.Mdata[6] = vasinstruments.SimulatedAeroTrak()[4] * g.countunits
			g.d.Mdata[7] = vasinstruments.SimulatedAeroTrak()[5] * g.countunits
		} else {
			m, err := g.at.GetAeroTrakdata()
			if err != nil {
				g.retries++
				if g.retries > 5 {
					// dialog.ShowInformation("Problems!", "Problem communicating with AeroTrak.", g.window)
					g.retries = 0
					g.StopMeasurement()
				}
			} else {
				g.retries = 0
			}
			if m[0] > 0 {
				g.d.Mdata[2] = m[0] * g.countunits
				g.d.Mdata[3] = m[1] * g.countunits
				g.d.Mdata[4] = m[2] * g.countunits
				g.d.Mdata[5] = m[3] * g.countunits
				g.d.Mdata[6] = m[4] * g.countunits
				g.d.Mdata[7] = m[5] * g.countunits
				if g.at.Showdata {
					fmt.Printf("Data (%v): %v %v\n", g.recordcount, g.d.Tstamp[11:19], g.d.Mdata)
				}
			}
		}
		atdata = g.d.Mdata[2] > 0
	}
	g.d.Tstamp = time.Now().Format(time.RFC3339)
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

func (g *Measuretype) Checkinstruments() {
	g.at.AeroTrakport = fyne.CurrentApp().Preferences().StringWithFallback("AeroTrak", "")
	g.dt.DustTrakport = fyne.CurrentApp().Preferences().StringWithFallback("DustTrak", "")
	g.pt.PTrakport = fyne.CurrentApp().Preferences().StringWithFallback("PTrak", "")
	// g.at.SimulateAeroTrak = fyne.CurrentApp().Preferences().BoolWithFallback("SimulateAeroTrak", true) && (g.at.AeroTrakport == "")
	// g.dt.SimulateDustTrak = fyne.CurrentApp().Preferences().BoolWithFallback("SimulateDustTrak", true) && (g.dt.DustTrakport == "")
	// g.pt.SimulatePTrak = fyne.CurrentApp().Preferences().BoolWithFallback("SimulatePTrak", true) && (g.pt.PTrakport == "")
	// g.at.AeroTrakrunning = (g.at.AeroTrakport > "" || g.at.SimulateAeroTrak)
	// g.dt.DustTrakrunning = (g.dt.DustTrakport > "" || g.dt.SimulateDustTrak)
	// g.pt.PTrakrunning = (g.pt.PTrakport > "" || g.pt.SimulatePTrak)
	// fyne.CurrentApp().Preferences().SetBool("SimulateAeroTrak", g.at.SimulateAeroTrak)
	// fyne.CurrentApp().Preferences().SetBool("SimulateDustTrak", g.dt.SimulateDustTrak)
	// fyne.CurrentApp().Preferences().SetBool("SimulatePTrak", g.pt.SimulatePTrak)
	i64, _ := strconv.ParseInt(fyne.CurrentApp().Preferences().StringWithFallback("ATdelay", "200"), 10, 64)
	g.at.Defdelay = i64
}

// func (g *game) countrunning() int {
// 	var c1 int = 0
// 	if g.measure.at.AeroTrakrunning {
// 		c1 = c1 + 6
// 	}
// 	if g.measure.dt.DustTrakrunning {
// 		c1 = c1 + 1
// 	}
// 	if g.measure.pt.PTrakrunning {
// 		c1 = c1 + 1
// 	}
// 	return c1
// }
// adds data and converts to range 0.0-1.0 for mX and 0.0-1.0 for mY
func (g *Measuretype) GetData() bool {
	var atdata, ptdata, dtdata bool
	g.d.Mdata[0] = -1
	g.d.Mdata[1] = -1
	g.d.Mdata[2] = -1
	if g.pt.PTrakrunning {
		if g.pt.SimulatePTrak {
			g.d.Mdata[0] = vasinstruments.SimulatedPTrak()
		} else {
			g.d.Mdata[0] = g.pt.GetPTrakdata()
		}
		ptdata = g.d.Mdata[0] > 0
	}
	if g.dt.DustTrakrunning {
		if g.dt.SimulateDustTrak {
			g.d.Mdata[1] = vasinstruments.SimulatedDustTrak()
		} else {
			g.d.Mdata[1] = g.dt.GetDustTrakdata()
		}
		dtdata = g.d.Mdata[0] > 0
	}
	if g.at.AeroTrakrunning {
		if g.at.SimulateAeroTrak {
			g.d.Mdata[2] = vasinstruments.SimulatedAeroTrak()[0] * g.countunits
			g.d.Mdata[3] = vasinstruments.SimulatedAeroTrak()[1] * g.countunits
			g.d.Mdata[4] = vasinstruments.SimulatedAeroTrak()[2] * g.countunits
			g.d.Mdata[5] = vasinstruments.SimulatedAeroTrak()[3] * g.countunits
			g.d.Mdata[6] = vasinstruments.SimulatedAeroTrak()[4] * g.countunits
			g.d.Mdata[7] = vasinstruments.SimulatedAeroTrak()[5] * g.countunits
		} else {
			m, err := g.at.GetAeroTrakdata()
			if err != nil {
				g.retries++
				if g.retries > 5 {
					// dialog.ShowInformation("Problems!", "Problem communicating with AeroTrak.", g.window)
					g.retries = 0
					g.StopMeasurement()
				}
			} else {
				g.retries = 0
			}
			if m[0] > 0 {
				g.d.Mdata[2] = m[0] * g.countunits
				g.d.Mdata[3] = m[1] * g.countunits
				g.d.Mdata[4] = m[2] * g.countunits
				g.d.Mdata[5] = m[3] * g.countunits
				g.d.Mdata[6] = m[4] * g.countunits
				g.d.Mdata[7] = m[5] * g.countunits
				if g.at.Showdata {
					fmt.Printf("Data (%v): %v %v\n", g.recordcount, g.d.Tstamp[11:19], g.d.Mdata)
				}
			}
		}
		atdata = g.d.Mdata[2] > 0
	}
	g.d.Tstamp = time.Now().Format(time.RFC3339)
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


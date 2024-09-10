package ui

// To create app, use "fyne package -os windows -icon resources/vas.png"
// to generate "loggor.go" I used; "fyne bundle -o loggor.go resources"

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"vas/general"
	"vas/vascharts"
	"vas/vasdatabase"
	"vas/vasinstruments"

	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	//	"fyne.io/fyne/v2/container"
)
type Measuretype struct {
	chart                 *[](vascharts.LineChart)
	chartSize             fyne.Size
	countunits            int32
	averagepoints         int

}
func (g *Measuretype) Measure() {
	var msg string
	var err error
	var fn string
	for {
		if g.Endmeasuring {
			break
		}
		if g.getData() && !g.paused {
			g.recordcount = g.recordcount + 1
			r := runtime.NumGoroutine()
			err = new(vasdatabase.DBtype).Addmeasurement()
			if err != nil {
				log.Println("#1 database error: ", err.Error())
			}
			err = vascharts.Updatecharts(nil,g.getData())
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
				new(vasdatabase.DBtype).Pruning()
				msg, err = new(vasdatabase.DBtype).Exporttotext()
				if err != nil {
					log.Println("Error exporting: ", msg, g.window)
				} else {
					for i := 0; i < len(tbl); i++ {
						fn = tbl[i] + ".txt"
						Doftp(fn)
					}
				}
				fn = general.Doscreenshot(g.window)
				if fn > "" {
					Doftp(fn)
				}
				g.paused = pauserestore
			}
			if g.recordcount%100 == 0 && g.MyDebug {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				fmt.Printf("%v  Go: %v  Time: %vs  Rec: %v  Speed: %.2f Rec/s  Laps: %v  Stack: %v \n",
					time.Now().String()[11:19],
					runtime.NumGoroutine(), strings.Split(time.Since(g.start).String(), ".")[0],
					g.recordcount, 0, g.laps, m.StackInuse)
					// g.recordcount, float64(g.recordcount)/(time.Since(g.d.dbstart).Seconds()), g.laps, m.StackInuse)
			}
		}
	}
}

func (g *Measuretype) getoldmeasurement() error {
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
func (g *Measuretype) StartMeasurement() {
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
		err = g.Addmeasurement()
		if err != nil {
			log.Println("#1 StartMeasurement - Problems adding Main measurement ", err.Error())
		}
	}
	g.window.Resize(g.winSize)
	g.windowResize()
	vascharts.Setupcharts()
	g.window.SetContent(&g.cont)
	log.Printf("Measurement '%v' (%v) started at %v.", g.d.mname, g.d.nanostamp, g.d.tstamp)
	g.recordcount = 0
	g.paused = false
	g.endmeasuring = false
	go g.Measure()
	// g.d.Pruning(g)
	// os.Exit(0)
}

func (g *Measuretype) StopMeasurement() {
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
func (g *Measuretype) getData() bool {
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
			g.d.mdata[2] = vasinstruments.SimulatedAeroTrak()[0] * g.countunits
			g.d.mdata[3] = vasinstruments.SimulatedAeroTrak()[1] * g.countunits
			g.d.mdata[4] = vasinstruments.SimulatedAeroTrak()[2] * g.countunits
			g.d.mdata[5] = vasinstruments.SimulatedAeroTrak()[3] * g.countunits
			g.d.mdata[6] = vasinstruments.SimulatedAeroTrak()[4] * g.countunits
			g.d.mdata[7] = vasinstruments.SimulatedAeroTrak()[5] * g.countunits
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


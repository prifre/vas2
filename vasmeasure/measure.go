package vasmeasure

// To create app, use "fyne package -os windows -icon resources/vas.png"
// to generate "loggor.go" I used; "fyne bundle -o loggor.go resources"

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"vas2/vasdatabase"
	"vas2/vasinstruments"

	"time"

	"fyne.io/fyne/v2"
	//	"fyne.io/fyne/v2/container"
)

type Measuretype struct {
	countunits            int32
	averagepoints         int
	synchronizedmeasuring bool
	Autostartmeasuring    bool
	recordcount           int
	Start                 time.Time
	laps                  int
	retries               int
	Paused                bool
	Endmeasuring          bool
	IsRunning             bool
	IntervalChan          chan time.Duration
	SampleInterval        time.Duration
	AT                    *vasinstruments.AeroTraktype
	DT                    *vasinstruments.DustTraktype
	PT                    *vasinstruments.PTraktype
	D                     *vasdatabase.DBtype
}

const Datapointsmax = 5
const Chartnum = 8

var gotdata int

var measuredata []int32

func SetupMeasurements(d *vasdatabase.DBtype) *Measuretype {
	g := &Measuretype{
		AT:           new(vasinstruments.AeroTraktype),
		DT:           new(vasinstruments.DustTraktype),
		PT:           new(vasinstruments.PTraktype),
		D:            d,
		IntervalChan: make(chan time.Duration, 1), // 👈 Lägg till denna!
	}
	i64, _ := strconv.ParseInt(fyne.CurrentApp().Preferences().StringWithFallback("ATdelay", "200"), 10, 64)
	g.AT.Defdelay = i64
	g.countunits = int32(fyne.CurrentApp().Preferences().IntWithFallback("countunits", 1))
	g.averagepoints = fyne.CurrentApp().Preferences().IntWithFallback("averagepoints", 30)
	g.SampleInterval = time.Duration(fyne.CurrentApp().Preferences().IntWithFallback("sampleinterval", 10)) * time.Millisecond
	g.synchronizedmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("synchronizedmeasuring", true)
	g.Autostartmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", false)
	return g
}
func (g *Measuretype) Getoldmeasurement() error {
	if !fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", false) {
		return fmt.Errorf("no autostart")
	}

	if g.D.Nanostamp == 0 {
		nanoStr := fyne.CurrentApp().Preferences().StringWithFallback("nanostamp", "")
		if nanoStr == "" {
			return fmt.Errorf("no nanostamp")
		}

		nanostamp, err := strconv.ParseInt(nanoStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid nanostamp: %v", err)
		}
		g.D.Nanostamp = nanostamp
	}

	// Hämta mname
	q1 := fmt.Sprintf("SELECT mname FROM tblMain WHERE nanostamp=%d", g.D.Nanostamp)
	mnameRes, err := g.D.Getsql(q1)
	if err != nil || len(mnameRes) == 0 {
		return fmt.Errorf("#1 Getting mname error: %v", err)
	}

	// Hämta tstamp
	q2 := fmt.Sprintf("SELECT tstamp FROM tblMain WHERE nanostamp=%d", g.D.Nanostamp)
	tstampRes, err := g.D.Getsql(q2)
	if err != nil || len(tstampRes) == 0 {
		return fmt.Errorf("#2 Getting tstamp error: %v", err)
	}

	g.D.Mname = mnameRes[0]
	g.D.Tstamp = tstampRes[0]

	return nil
}
func (g *Measuretype) StartMeasurement() {
	var err error

	g.Autostartmeasuring = fyne.CurrentApp().Preferences().BoolWithFallback("autostartmeasuring", true)
	ms := fyne.CurrentApp().Preferences().IntWithFallback("sampleinterval", 1000)

	// 2. Multiplicera heltalet med time.Millisecond för att få en time.Duration
	g.SampleInterval = time.Duration(ms) * time.Millisecond
	if g.D == nil {
		log.Println("Error: Database object (g.D) is nil!")
		return
	}

	err = g.Getoldmeasurement()
	if err != nil {
		if err.Error() != "no autostart" {
			log.Println("Restart not possible due to: ", err.Error())
		}

		log.Println("Autocreating new measurement!")
		g.D.Nanostamp = time.Now().UnixNano()

		// SPARA SOM STRING HÄR
		fyne.CurrentApp().Preferences().SetString("nanostamp", strconv.FormatInt(g.D.Nanostamp, 10))

		g.D.Tstamp = time.Now().Format(time.RFC3339)

		// Rensa tecken i namnet
		cleanTstamp := g.D.Tstamp
		for _, char := range []string{"-", "/", ":", "."} {
			cleanTstamp = strings.ReplaceAll(cleanTstamp, char, "")
		}

		g.D.Mname = "Measurement" + cleanTstamp
		if len(g.D.Mname) > 26 {
			g.D.Mname = g.D.Mname[:26]
		}
		g.D.Mdata = []int32{-1, -1, -1, -1, -1, -1, -1, -1}
		err = g.D.Addmeasurement()
		if err != nil {
			log.Println("#1 StartMeasurement - Problems adding Main measurement: ", err.Error())
		}
	}

	g.AT.SimulateAeroTrak = fyne.CurrentApp().Preferences().Bool("SimulateAeroTrak")
	g.AT.AeroTrakport = fyne.CurrentApp().Preferences().StringWithFallback("AeroTrak", "")
	g.AT.AeroTrakrunning = (g.AT.SimulateAeroTrak || (g.AT.AeroTrakport > ""))

	g.DT.SimulateDustTrak = fyne.CurrentApp().Preferences().Bool("SimulateDustTrak")
	g.DT.DustTrakport = fyne.CurrentApp().Preferences().StringWithFallback("DustTrak", "")
	g.DT.DustTrakrunning = (g.DT.SimulateDustTrak || (g.DT.DustTrakport > ""))

	g.PT.SimulatePTrak = fyne.CurrentApp().Preferences().Bool("SimulatePTrak")
	g.PT.PTrakport = fyne.CurrentApp().Preferences().StringWithFallback("PTrak", "")
	g.PT.PTrakrunning = (g.PT.SimulatePTrak || (g.PT.PTrakport > ""))

	if g.AT.SimulateAeroTrak || g.DT.SimulateDustTrak || g.PT.SimulatePTrak {
		g.D.Note = "Simulated measurement"
		if g.AT.SimulateAeroTrak {
			g.D.Note += " AeroTrak"
		}
		if g.DT.SimulateDustTrak {
			g.D.Note += " DustTrak"
		}
		if g.PT.SimulatePTrak {
			g.D.Note += " PTrak"
		}
	}
	if g.AT.AeroTrakport > "" && !g.AT.SimulateAeroTrak {
		g.AT.AeroTrakStart()
	}
	if g.DT.DustTrakport > "" && !g.DT.SimulateDustTrak {
		g.DT.DustTrakStart()
	}
	if g.PT.PTrakport > "" && !g.PT.SimulatePTrak {
		g.PT.PTrakStart()
	}

	log.Printf("Measurement '%v' (%v) started at %v.", g.D.Mname, g.D.Nanostamp, g.D.Tstamp)
	g.recordcount = 0
	g.Endmeasuring = false
	g.Paused = false
	g.IsRunning = true
}
func (g *Measuretype) StopMeasurement() {
	if g == nil {
		return
	}

	g.Paused = false
	g.Endmeasuring = true

	var err error

	// Stop AeroTrak
	if g.AT != nil && g.AT.AeroTrakrunning && !g.AT.SimulateAeroTrak {
		time.Sleep(500 * time.Millisecond) // Kanske räcker 500ms?
		if err = g.AT.AeroTrakStop(); err != nil {
			log.Println("#1 StopMeasurement - Problems AeroTrak stop:", err.Error())
		}
	}

	// Stop DustTrak
	if g.DT != nil && g.DT.DustTrakrunning && !g.DT.SimulateDustTrak {
		time.Sleep(500 * time.Millisecond)
		if err = g.DT.DustTrakstop(); err != nil { // KORRIGERAT: Endast ett anrop här!
			log.Println("#2 StopMeasurement - Problems DustTrak stop:", err.Error())
		}
	}

	// Återställ tillstånd
	if g.AT != nil {
		g.AT.AeroTrakrunning = false
	}
	if g.DT != nil {
		g.DT.DustTrakrunning = false
	}
	if g.PT != nil {
		g.PT.PTrakrunning = false
	}
}
func (g *Measuretype) GetData() bool {
	var err error
	if len(g.D.Mdata) != Chartnum {
		g.D.Mdata = make([]int32, Chartnum)
	}

	// Återställ alla värden till -1 (ingen mätning)
	for i := range g.D.Mdata {
		g.D.Mdata[i] = -1
	}

	// 1. PTrak
	if g.PT.PTrakrunning {
		if g.PT.SimulatePTrak {
			g.D.Mdata[0] = vasinstruments.SimulatedPTrak()
		} else {
			g.D.Mdata[0], err = g.PT.GetPTrakdata()
			if err != nil {
				log.Println("Error getting PTrak data:", err.Error())
			}
		}
	}

	// 2. DustTrak
	if g.DT.DustTrakrunning {
		if g.DT.SimulateDustTrak {
			g.D.Mdata[1] = vasinstruments.SimulatedDustTrak()
		} else {
			g.D.Mdata[1], err = g.DT.GetDustTrakdata()
			if err != nil {
				log.Println("Error getting DustTrak data:", err.Error())
			}
		}
	}

	// 3. AeroTrak
	if g.AT.AeroTrakrunning {
		if g.AT.SimulateAeroTrak {
			g.countunits = 1
			// 🟢 ANROP EN GÅNG: Spara i en variabel först!
			simData := vasinstruments.SimulatedAeroTrak()
			if len(simData) >= 6 {
				for i := 0; i < 6; i++ {
					g.D.Mdata[2+i] = simData[i] * g.countunits
				}
			}
		} else {
			m, err := g.AT.GetAeroTrakdata()
			if err != nil {
				g.retries++
				if g.retries > 5 {
					g.retries = 0
					g.StopMeasurement()
				}
			} else {
				g.retries = 0
				if len(m) >= 6 && m[0] > 0 {
					for i := 0; i < 6; i++ {
						g.D.Mdata[2+i] = m[i] * g.countunits
					}
					if g.AT.Showdata {
						fmt.Printf("Data (%v): %v %v\n", g.recordcount, g.D.Tstamp[11:19], g.D.Mdata)
					}
				}
			}
		}
	}

	g.D.Tstamp = time.Now().Format(time.RFC3339)
	gotdata++

	return true
}

func (g *Measuretype) Measure(chartUpdateChan chan []int32, IntervalChan chan time.Duration) {
	var err error

	interval := g.SampleInterval
	if interval <= 0 {
		interval = 1 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if g.Endmeasuring {
			break
		}

		select {
		// A) Nytt intervall från UI
		case newInterval, ok := <-IntervalChan:
			if !ok {
				// Kanalen har stängts utifrån – sätt till nil så select ignorerar den
				IntervalChan = nil
				break
			}
			if newInterval > 0 {
				g.SampleInterval = newInterval
				ticker.Reset(newInterval)
				log.Println("Samplingsintervall ändrat till:", newInterval)
			}

		// B) Timern slog till
		case <-ticker.C:
			// Dubbelkolla om vi stängt under tiden timern väntade
			if g.Endmeasuring {
				break
			}

			// Snabbkoll: Hoppa över direkt om pausad
			if g.Paused {
				continue
			}

			// Läs av instrumenten
			if g.GetData() {
				g.recordcount++

				err = g.D.Addmeasurement()
				if err != nil {
					log.Println("#1 database error: ", err.Error())
				}

				// Skicka data till UI om vi inte stoppar
				if !g.Endmeasuring && !g.Paused && chartUpdateChan != nil {
					select {
					case chartUpdateChan <- g.D.Mdata:
						// Datan skickades ok
					default:
						// Undvik låsning om kanalen är full
					}
				}

				g.laps++

				// Pruning vid varje heltimme
				now := time.Now()
				if now.Minute() == 0 && now.Second() == 0 {
					pauserestore := g.Paused
					g.Paused = true
					runtime.GC()
					g.D.Pruning()
					g.Paused = pauserestore
				}
			}
		}
	}

	// Stäng kanalen säkert när loopen bryts
	if chartUpdateChan != nil {
		close(chartUpdateChan)
	}
}

package vascharts

// external functions:
// Setupcharts - destroys old chars & setups new...
// Updatecharts - updates current charts
// Redrawcharts - when window size has changed...

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"golang.org/x/image/colornames"
)

type LineChart struct {
	TheChart *fyne.Container
	mBox     canvas.Rectangle
	mTitle   canvas.Text
	mTitle2  canvas.Text
	//	mYaxisTitle canvas.Image
	mXaxis      [axisticks]canvas.Line
	mYaxis      [axisticks]canvas.Line
	mXaxisline  canvas.Line
	mYaxisline  canvas.Line
	mLine       []*canvas.Line
	mValue      []float32
	mTime       []string
	mCounter    int
	mXaxistext  [axisticks]canvas.Text
	mYaxistext  [axisticks]canvas.Text
	mMax        float32
	StrokeColor color.Color
	StrokeWidth float32
	mdata       []int32
}

var ChartTitles = [8]string{
	"Nano Particles\t0.02-1.0µm (#/cm³)",
	"Weight Particles\t          µg/m³",
	"Fine Particles\t   0.3-0.5µm",
	"Fine Particles\t   0.5-1.0µm",
	"Fine Particles\t   1.0-3.0µm",
	"Coarse Particles\t   3.0-5.0µm",
	"Coarse Particles\t   5.0-10.0µm",
	"Coarse Particles\t     >10.0µm"}

const (
	linemarginsY    = .72
	linemarginsX    = .8
	titledown       = 0.03 // Lite marginal i överkant för titeln
	axismarg        = 0.08 // SÄTT DENNA > 0! Ger luft för X/Y-texten vid kanterna
	axisticks       = 5
	ticksize        = 3.5
	chartnum        = 8
	textsizedivider = 60 // Gör texten något mindre på små skärmar
)

var Datapointsmax int = 80
var charts []*LineChart

// main Line chart drawing routine
// takes lc.mX[] and lc.mY[] and draws
// where values are already adjusted to 0...1

// Return a slice of pointers. This prevents massive struct copying in memory.
func NewCharts() []*LineChart {
	// This initializes your 8 charts automatically at startup
	// and keeps their data in memory.
	datapoints := fyne.CurrentApp().Preferences().IntWithFallback("datapoints", 10)

	chartnum := 8
	charts = make([]*LineChart, chartnum)

	sz := fyne.NewSize(1600, 1600)
	for c := 0; c < chartnum; c++ {
		charts[c] = new(LineChart)
		if strings.Contains(ChartTitles[c], "\t") {
			charts[c].mTitle2.Text = strings.Split(ChartTitles[c], "\t")[1]
			charts[c].mTitle.Text = strings.Split(ChartTitles[c], "\t")[0]
		}
		charts[c].TheChart = container.NewWithoutLayout()
		charts[c].TheChart.Resize(sz)
		lc := charts[c]
		lc.mBox.StrokeColor = colornames.Blue
		lc.mBox.FillColor = colornames.Black
		lc.mBox.StrokeWidth = 1
		lc.mBox.Resize(lc.TheChart.Size()) // resize box to chart size
		lc.TheChart.Add(&lc.mBox)          // object 0
		lc.TheChart.Add(&lc.mYaxisline)    // object 5 2*axisticks+2
		lc.TheChart.Add(&lc.mXaxisline)
		lc.TheChart.Add(&lc.mTitle) // object 1
		lc.TheChart.Add(&lc.mTitle2)
		for i := 0; i < axisticks; i++ {
			lc.TheChart.Add(&lc.mXaxistext[i])
			lc.TheChart.Add(&lc.mYaxistext[i])
			lc.TheChart.Add(&lc.mXaxis[i])
			lc.TheChart.Add(&lc.mYaxis[i])
		}
		lc.Refreshonechart()

		charts[c].mLine = make([]*canvas.Line, 0, Datapointsmax)
		charts[c].mValue = make([]float32, Datapointsmax)
		charts[c].mTime = make([]string, Datapointsmax)

		for i := 0; i < Datapointsmax; i++ {
			var sc color.Color
			r := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorR", c))
			g := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorG", c))
			b := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorB", c))
			a := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorA", c))
			sc = color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(a),
			}
			newLine := canvas.NewLine(sc)
			newLine.StrokeWidth = 2
			newLine.Position1 = fyne.NewPos(0, 0)
			newLine.Position2 = fyne.NewPos(0, 0)
			newLine.Hidden = false

			// 1. Lägg till linjen i Fynes container (UI)
			charts[c].TheChart.Add(newLine)

			// 2. Spara EXAKT samma linje i din struct-slice!
			charts[c].mLine = append(charts[c].mLine, newLine)

			charts[c].mValue[i] = 0
			charts[c].mTime[i] = ""
		}
		lc.TheChart.Refresh()
	}
	SetMaxDatapoints(datapoints)
	return charts
}

func UpdateChart(lc []*LineChart, mdata []int32) error {
	for c := 0; c < chartnum; c++ {
		newdata := float32(mdata[c])
		if newdata < 0 {
			continue // Ignorera ogiltiga värden
		}

		currMaxPoints := len(lc[c].mValue)

		// 1. Skifta datan i bufferten (Sliding Window)
		for i := 1; i < currMaxPoints; i++ {
			lc[c].mValue[i-1] = lc[c].mValue[i]
			lc[c].mTime[i-1] = lc[c].mTime[i]
		}

		// 2. Lägg till sista punkten
		lc[c].mValue[currMaxPoints-1] = newdata
		lc[c].mTime[currMaxPoints-1] = time.Now().Format("15:04")
		lc[c].mCounter++

		// 3. Räkna ut dynamiskt Max
		var currentMax float32 = 0
		for i := 0; i < currMaxPoints; i++ {
			if lc[c].mValue[i] > currentMax {
				currentMax = lc[c].mValue[i]
			}
		}

		newMaxScaled := Fixmax(currentMax)
		if newMaxScaled != lc[c].mMax {
			lc[c].mMax = newMaxScaled
			lc[c].Drawoneyaxistexts()
		}

		// 4. Uppdatera X-axeln & Rita linjerna
		lc[c].Drawonexaxistexts(lc[c].mTime)
		lc[c].Drawonechartlines()
		lc[c].Refreshonechart()
	}
	return nil
}

func (lc *LineChart) Drawonechartlines() {
	totalPoints := len(lc.mValue)
	if totalPoints < 2 {
		return
	}

	chartSize := lc.TheChart.Size()
	height := chartSize.Height
	width := chartSize.Width

	// Antal segments/linjer som ska ritas är alltid antal punkter - 1
	maxSegments := float32(totalPoints - 1)

	for i := 0; i < totalPoints-1; i++ {
		// Om i hamnar utanför mLine-slicen skyddar vi oss mot panic
		if i >= len(lc.mLine) {
			break
		}

		// 1. Beräkna X-procent spikrakt utifrån index 'i' dividerat med max punkter
		pctX1 := float32(i) / maxSegments
		pctX2 := float32(i+1) / maxSegments

		x1 := pctX1*width*linemarginsX + (width * 0.5 * (1 - linemarginsX))
		x2 := pctX2*width*linemarginsX + (width * 0.5 * (1 - linemarginsX))

		// 2. Beräkna Y-procent utifrån mValue och mMax
		divMax := lc.mMax
		if divMax == 0 {
			divMax = 1 // Förhindra division med 0
		}

		pctY1 := lc.mValue[i] / divMax
		pctY2 := lc.mValue[i+1] / divMax

		y1 := height*linemarginsY - pctY1*height*linemarginsY + (height * 0.5 * (1 - linemarginsY))
		y2 := height*linemarginsY - pctY2*height*linemarginsY + (height * 0.5 * (1 - linemarginsY))

		// 3. Uppdatera positionerna på Fyne-linjen
		lc.mLine[i].Position1 = fyne.NewPos(x1, y1)
		lc.mLine[i].Position2 = fyne.NewPos(x2, y2)
		lc.mLine[i].Show() // Se till att linjen syns
	}

	// 4. Göm alla eventuella överblivna linjer som ligger i mLine-slicen
	for i := totalPoints - 1; i < len(lc.mLine); i++ {
		lc.mLine[i].Hide()
	}

	lc.TheChart.Refresh()
}
func (lc *LineChart) Refreshonechart() {
	lc.Drawonetitle()
	lc.Drawoneyaxistexts()
	lc.DrawXaxisline()
	lc.DrawYaxisline()
	lc.DrawYaxis()
	lc.DrawXaxis()
}

// SetMaxDatapoints uppdaterar Datapointsmax för alla diagram och anpassar UI-objekten
func SetMaxDatapoints(newMax int) {
	if newMax < 2 {
		newMax = 2
	}
	Datapointsmax = newMax

	for c, lc := range charts { // 👈 c är diagrammets index (0-7)
		if lc == nil {
			continue
		}

		oldLen := len(lc.mValue)
		if newMax > oldLen {
			for i := oldLen; i < newMax; i++ {
				// ANVÄND 'c' ISTÄLLET FÖR '0':
				r := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorR", c))
				g := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorG", c))
				b := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorB", c))
				a := fyne.CurrentApp().Preferences().Int(fmt.Sprintf("%vStrokeColorA", c))

				// Sätt en fallback om inställningen vore 0/tom
				if a == 0 {
					a = 255
				}

				newLine := canvas.NewLine(color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)})
				newLine.StrokeWidth = 2
				newLine.Hide()

				lc.TheChart.Add(newLine)
				lc.mLine = append(lc.mLine, newLine)
				lc.mValue = append(lc.mValue, 0)
				lc.mTime = append(lc.mTime, "")
			}
		} else if newMax < oldLen {
			for i := newMax; i < oldLen; i++ {
				if i < len(lc.mLine) {
					lc.TheChart.Remove(lc.mLine[i])
				}
			}
			lc.mLine = lc.mLine[:newMax]
			lc.mValue = lc.mValue[:newMax]
			lc.mTime = lc.mTime[:newMax]
		}

		lc.Drawonechartlines()
		lc.Refreshonechart()
	}
}

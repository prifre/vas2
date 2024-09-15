package vascharts

// external functions:
// Setupcharts - destroys old chars & setups new...
// Updatecharts - updates current charts
// Redrawcharts - when window size has changed...

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"golang.org/x/image/colornames"
)

const datapointsmax = 80
const (
	linemarginsY    = .72
	linemarginsX    = .68
	titledown       = .001
	axismarg        = 0.12
	textmargin      = 1.3 //position of axis - (size of textwidth/2) * textmargin
	axisticks       = 5
	ticksize        = 3.5
	textsizedivider = 40
)
type LineChart struct {
	mChart  fyne.Container
	mBox    canvas.Rectangle
	mTitle  canvas.Text
	mTitle2 canvas.Text
	//	mYaxisTitle canvas.Image
	mXaxis      [axisticks]canvas.Line
	mYaxis      [axisticks]canvas.Line
	mXaxisline  canvas.Line
	mYaxisline  canvas.Line
	mLine       []canvas.Line
	mX          []float32
	mY          []float32
	mValue      []float32
	mTime       []string
	mXaxistext  [axisticks]canvas.Text
	mYaxistext  [axisticks]canvas.Text
	gotpoints   int
	mMax        float32
	StrokeColor color.Color
	StrokeWidth float32
}
// Charts!
var ChartTitles = [8]string{
	"   Nano Particles\t0.02-1.0µm (#/cm³)",
	"Weight Particles\t          µg/m³",
	"Fine Particles\t   0.3-0.5µm",
	"Fine Particles\t   0.5-1.0µm",
	"Fine Particles\t   1.0-3.0µm",
	"Coarse Particles\t   3.0-5.0µm",
	"Coarse Particles\t   5.0-10.0µm",
	"Coarse Particles\t     >10.0µm"}


// main Line chart drawing routine
// takes lc.mX[] and lc.mY[] and draws
// where values are already adjusted to 0...1

func Setupcharts() []LineChart {
	chartnum :=fyne.CurrentApp().Preferences().Int("chartcount")
	chartsizex :=float32(fyne.CurrentApp().Preferences().Float("chartsizex"))
	chartsizey :=float32(fyne.CurrentApp().Preferences().Float("chartsizey"))
	var ch []LineChart
	for c := 0; c < chartnum; c++ {
		lc := new(LineChart)
		ch=append(ch,*lc)
		for i := 0; i < datapointsmax; i++ {
			ch[c].drawonechart(fyne.NewSize(chartsizex,chartsizey))
			ch[c].mLine[i].Position1.X = 0
			ch[c].mLine[i].Position2.X = 0
			ch[c].mLine[i].Position1.Y = 0
			ch[c].mLine[i].Position2.Y = 0
			ch[c].mLine[i].StrokeWidth = 2
			ch[c].mLine[i].StrokeColor = color.Black
		}
		ch[c].gotpoints = 0
	}
	return ch
}

func Updatecharts(chart []LineChart,mdata []int) error {
	if chart == nil {
		chart:=Setupcharts()
		Redrawcharts(chart)
	}
	datapoints:=len(mdata)
	avsum:=0
	for a1 := 0; a1 < len(mdata); a1++ {
		avsum += mdata[a1]
	}
	averagepoints:= avsum / len(mdata)
	for c := 0; c < len(chart); c++ {
		newdata := float32(mdata[c])
		if newdata > -1 {
			chart[c].gotpoints++
			if chart[c].gotpoints > datapoints {
				chart[c].gotpoints = datapoints
			}
			// chart[c].mY[chart[c].gotpoints-1] = float32(g.d.Mdata[c])
			chart[c].mValue[chart[c].gotpoints-1] = newdata
			if averagepoints > 1 {
				avpoints := averagepoints
				if avpoints < 1 {
					avpoints = 1
				}
				if chart[c].gotpoints < avpoints {
					avpoints = chart[c].gotpoints
				}
				var avsum float32 = 0
				for a1 := 0; a1 < avpoints; a1++ {
					avsum += chart[c].mValue[a1]
				}
				newdata = avsum / float32(avpoints)
			}
			if newdata > chart[c].mMax {
				chart[c].mMax = fixmax(newdata)
				chart[c].drawoneyaxistexts()
				// recalculate all T-axis values based on the change to max!!!!
				for i := 0; i < chart[c].gotpoints; i++ {
					chart[c].mY[i] = chart[c].mValue[i] / (chart[c].mMax + 1)
				}
			} else {
				// adjust for Y-size 0 ... 1 just the new Y-value
				chart[c].mY[chart[c].gotpoints-1] = newdata / (chart[c].mMax + 1)
			}
			// find new max when oldMax is dissapearing...
			if fixmax(chart[c].mValue[0]) == chart[c].mMax {
				var newmax float32 = 0
				for i := 1; i < chart[c].gotpoints; i++ {
					if newmax < chart[c].mValue[i] {
						newmax = chart[c].mValue[i]
					}
				}
				chart[c].mMax = fixmax(newmax)
				chart[c].drawoneyaxistexts()
				for i := 1; i < chart[c].gotpoints; i++ {
					chart[c].mY[i] = chart[c].mValue[i] / (chart[c].mMax + 1)
				}
			}
			// time...
			chart[c].mTime[chart[c].gotpoints-1] = time.Now().Format(time.RFC3339)[11:16]
			// move that line...
			chart[c].mX[chart[c].gotpoints-1] = 1
			for i := 1; i < chart[c].gotpoints; i++ {
				chart[c].mX[i-1] = float32(float32(i) / float32(chart[c].gotpoints))
				if chart[c].gotpoints == datapoints {
					chart[c].mY[i-1] = chart[c].mY[i]
					chart[c].mValue[i-1] = chart[c].mValue[i]
					chart[c].mTime[i-1] = chart[c].mTime[i]
				}
			}
		}
	}
	return Redrawcharts(chart)
}
func Redrawcharts(chart []LineChart) error {
	if chart == nil {
		Setupcharts()
	}
	for c := 0; c < len(chart); c++ {
		if chart[c].gotpoints > 1 {
			if chart[c].gotpoints > datapointsmax {
				chart[c].gotpoints = datapointsmax
			}
			chart[c].drawonechartlines(chart[c].gotpoints)
			chart[c].drawonexaxistexts(chart[c].mTime)
			chart[c].refreshonechart(c)
		}
	}
	return nil
}
func (lc *LineChart) drawonebox() {
	lc.mBox.StrokeColor = colornames.Blue
	lc.mBox.FillColor = colornames.Black
	lc.mBox.StrokeWidth = 1
	lc.mBox.Resize(lc.mChart.Size())
}
func (lc *LineChart) drawonetitle(title string) {
	var title2 string
	if strings.Contains(title, "\t") {
		title2 = strings.Split(title, "\t")[1]
		title = strings.Split(title, "\t")[0]
	}
	lc.mTitle.Text = title
	lc.mTitle.Color = colornames.White
	lc.mTitle.TextStyle.Bold = true
	lc.mTitle.TextSize = lc.mChart.Size().Width/textsizedivider + 4 //g.contoldW/150 + 2
	lc.mTitle.Move(fyne.NewPos(lc.mChart.Size().Width/2-lc.mTitle.MinSize().Width/2, lc.mChart.Size().Height*titledown))
	if len(title2) > 0 {
		lc.mTitle2.Text = title2
		lc.mTitle2.Color = colornames.White
		lc.mTitle2.TextStyle.Bold = true
		lc.mTitle2.TextSize = lc.mChart.Size().Width/textsizedivider + 4 //g.contoldW/150 + 2
		ypos2 := lc.mChart.Size().Height * titledown
		ypos2 = ypos2 + lc.mTitle2.TextSize
		lc.mTitle2.Move(fyne.NewPos(lc.mChart.Size().Width/2-lc.mTitle.MinSize().Width/2, ypos2))
	}
}
func (lc *LineChart) drawonexaxistexts(t []string) {
	if lc.gotpoints > axisticks {
		j := 0
		gp := len(t)
		if gp > lc.gotpoints {
			gp = lc.gotpoints
		}
		st := gp / axisticks
		for i := 1; i < gp; i = i + st {
			if j < axisticks {
				t[j] = t[i]
				j++
			}
		}
		t[j] = t[gp-1]
	}
	for i := 0; i < axisticks; i++ {
		lc.mXaxistext[i].Color = color.White
		lc.mXaxistext[i].Text = t[axisticks-i-1]
		lc.mXaxistext[i].TextSize = lc.mChart.Size().Width/textsizedivider + 3
		lc.mXaxistext[i].TextStyle.Bold = true
		lc.mXaxistext[i].Move(fyne.NewPos(lc.mChart.Size().Width-lc.mChart.Size().Width*axismarg-
			(lc.mChart.Size().Width-(2*axismarg*lc.mChart.Size().Width))*float32(i)/axisticks-
			lc.mXaxistext[i].MinSize().Width/2,
			lc.mChart.Size().Height-lc.mChart.Size().Height*axismarg))
	}
}
func setyaxisscale(max float64) []string {
	var vt []string
	for i := 0; i < axisticks; i++ {
		m1 := math.Round(((float64(axisticks - i)) * max))
		v := float64(m1 / axisticks)
		switch {
		case v < 1000:
			vt = append(vt, fmt.Sprintf("%v", v))
		case v < 1000000:
			vt = append(vt, fmt.Sprintf("%vk", math.Round(v/1000)))
		default:
			vt = append(vt, fmt.Sprintf("%vM", math.Round(v/1000000)))
		}
	}
	return vt
}
func (lc *LineChart) drawoneyaxistexts() {
	vt := setyaxisscale(float64(lc.mMax))
	for i := 0; i < len(vt); i++ {
		lc.mYaxistext[i].Color = color.White
		lc.mYaxistext[i].Text = vt[i]
		lc.mYaxistext[i].TextSize = lc.mChart.Size().Width/textsizedivider + 3
		lc.mYaxistext[i].TextStyle.Bold = true
		lc.mYaxistext[i].Move(fyne.NewPos(axismarg*lc.mChart.Size().Width-
			lc.mYaxistext[i].MinSize().Width-ticksize,
			(lc.mChart.Size().Height-
				(2*axismarg*lc.mChart.Size().Height))*float32(i)/(axisticks)+lc.mChart.Size().Height*axismarg-
				float32(lc.mYaxistext[i].MinSize().Height)/2))
	}
}
func fixmax(x float32) float32 {
	if x < 10 {
		return 10
	}
	y := math.Log10(float64(x))
	y = float64(int(y)) + 1
	y = math.Pow(10, y)
	for y/2 > float64(x) && y/2 == float64(int(y/2)) && fmt.Sprintf("%v", int(y/2))[:2] != "12" {
		y = y / 2
	}
	return float32(y)
}
func (lc *LineChart) drawXaxisline() {
	//X-axis:
	lc.mXaxisline.StrokeColor = color.White
	lc.mXaxisline.StrokeWidth = 2
	lc.mXaxisline.Position1.X = lc.mChart.Size().Width * axismarg
	lc.mXaxisline.Position2.X = lc.mChart.Size().Width - lc.mChart.Size().Width*axismarg
	lc.mXaxisline.Position1.Y = lc.mChart.Size().Height - lc.mChart.Size().Height*axismarg
	lc.mXaxisline.Position2.Y = lc.mXaxisline.Position1.Y
}
func (lc *LineChart) drawYaxisline() {
	//Y-axis
	lc.mYaxisline.StrokeColor = color.White
	lc.mYaxisline.StrokeWidth = 2
	lc.mYaxisline.Position1.X = lc.mChart.Size().Width * axismarg
	lc.mYaxisline.Position2.X = lc.mYaxisline.Position1.X
	lc.mYaxisline.Position1.Y = lc.mChart.Size().Height * axismarg
	lc.mYaxisline.Position2.Y = lc.mChart.Size().Height - lc.mChart.Size().Height*axismarg
}
func (lc *LineChart) drawXaxis() {
	//X-axis ticks:
	for i := 0; i < axisticks; i++ {
		lc.mXaxis[i].StrokeColor = color.White
		lc.mXaxis[i].StrokeWidth = 2
		lc.mXaxis[i].Position1.X = lc.mChart.Size().Width - lc.mChart.Size().Width*axismarg - (lc.mChart.Size().Width-(2*axismarg*lc.mChart.Size().Width))*float32(i)/axisticks
		lc.mXaxis[i].Position2.X = lc.mChart.Size().Width - lc.mChart.Size().Width*axismarg - (lc.mChart.Size().Width-(2*axismarg*lc.mChart.Size().Width))*float32(i)/axisticks
		lc.mXaxis[i].Position1.Y = lc.mChart.Size().Height - lc.mChart.Size().Height*axismarg - ticksize
		lc.mXaxis[i].Position2.Y = lc.mChart.Size().Height - lc.mChart.Size().Height*axismarg + ticksize
	}
}
func (lc *LineChart) drawYaxis() {
	//Y-axis ticks
	for i := 0; i < axisticks; i++ {
		lc.mYaxis[i].StrokeColor = color.White
		lc.mYaxis[i].StrokeWidth = 2
		lc.mYaxis[i].Position1.X = lc.mChart.Size().Width*axismarg - ticksize
		lc.mYaxis[i].Position2.X = lc.mChart.Size().Width*axismarg + ticksize
		lc.mYaxis[i].Position1.Y = (lc.mChart.Size().Height-(2*axismarg*lc.mChart.Size().Height))*float32(i)/(axisticks) + lc.mChart.Size().Height*axismarg
		lc.mYaxis[i].Position2.Y = (lc.mChart.Size().Height-(2*axismarg*lc.mChart.Size().Height))*float32(i)/(axisticks) + lc.mChart.Size().Height*axismarg
	}
}
func (lc *LineChart) drawonechart(sz fyne.Size) {
	//	lc.mChart.Resize(sz)
	if len(lc.mChart.Objects) == 0 {
		lc.mChart = *container.NewWithoutLayout()
		lc.mChart.Resize(sz)
		lc.mChart.Add(&lc.mBox)
		lc.mChart.Add(&lc.mTitle)
		lc.mChart.Add(&lc.mTitle2)
		for i := 0; i < axisticks; i++ {
			lc.mChart.Add(&lc.mXaxistext[i])
			lc.mChart.Add(&lc.mYaxistext[i])
			lc.mChart.Add(&lc.mXaxis[i])
			lc.mChart.Add(&lc.mYaxis[i])
		}
		lc.mChart.Add(&lc.mYaxisline)
		lc.mChart.Add(&lc.mXaxisline)
		for i := 0; i < datapointsmax; i++ {
			lc.mChart.Add(&lc.mLine[i])
		}
		lc.mChart.Refresh()
	}
}
func (lc *LineChart) drawonechartlines(j int) {
	for i := 1; i < j; i++ {
		y1 := lc.mChart.Size().Height*linemarginsY - lc.mY[i-1]*lc.mChart.Size().Height*linemarginsY + (lc.mChart.Size().Height * .5 * (1 - linemarginsY))
		x1 := lc.mX[i-1]*lc.mChart.Size().Width*linemarginsX + (lc.mChart.Size().Width * .5 * (1 - linemarginsX))
		y2 := lc.mChart.Size().Height*linemarginsY - lc.mY[i]*lc.mChart.Size().Height*linemarginsY + (lc.mChart.Size().Height * .5 * (1 - linemarginsY))
		x2 := lc.mX[i]*lc.mChart.Size().Width*linemarginsX + (lc.mChart.Size().Width * .5 * (1 - linemarginsX))
		lc.mLine[i-1].Position1.X = x1
		lc.mLine[i-1].Position2.X = x2
		lc.mLine[i-1].Position1.Y = y1
		lc.mLine[i-1].Position2.Y = y2
	}
}
func (lc *LineChart) refreshonechart(c int) {
	lc.drawonebox()
	ct:=fyne.CurrentApp().Preferences().String("charttitles")
	ChartTitles:=strings.Split(ct,",")
	lc.drawonetitle(ChartTitles[c])
	lc.drawoneyaxistexts()
	lc.drawXaxisline()
	lc.drawYaxisline()
	lc.drawXaxis()
	lc.drawYaxis()
}

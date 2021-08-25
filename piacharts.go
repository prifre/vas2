package main

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"golang.org/x/image/colornames"
)

type LineChart struct {
	mChart fyne.Container
	mBox   canvas.Rectangle
	mTitle canvas.Text
	//	mYaxisTitle canvas.Image
	mXaxis      [axisticks]canvas.Line
	mYaxis      [axisticks]canvas.Line
	mXaxisline  canvas.Line
	mYaxisline  canvas.Line
	mLine       [datapointsmax]canvas.Line
	mX          [datapointsmax]float32
	mY          [datapointsmax]float32
	mTime       [datapointsmax]string
	mXaxistext  [axisticks]canvas.Text
	mYaxistext  [axisticks]canvas.Text
	gotpoints   int
	mMax        float32
	StrokeColor color.Color
	StrokeWidth float32
}

const (
	linemarginsY = .72
	linemarginsX = .68
	titledown    = .001
	axismarg     = 0.12
	textmargin   = 1.3 //position of axis - (size of textwidth/2) * textmargin
	axisticks    = 5
	ticksize     = 3.5
)

// main Line chart drawing routine
// takes lc.mX[] and lc.mY[] and draws
// where values are already adjusted to 0...1
func (lc *LineChart) Refreshchart(c int) {
	lc.drawBox()
	lc.drawTitle(ChartTitles[c])
	lc.drawYaxistexts()
	lc.drawXaxisline()
	lc.drawYaxisline()
	lc.drawXaxis()
	lc.drawYaxis()
}
func (lc *LineChart) DrawChart(sz fyne.Size) {
	lc.mChart.Resize(sz)
	if len(lc.mChart.Objects) == 0 {
		lc.mChart = *container.NewWithoutLayout()
		lc.mChart.Resize(sz)
		lc.mChart.Add(&lc.mBox)
		lc.mChart.Add(&lc.mTitle)
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
	}
	lc.mChart.Refresh()
}
func (lc *LineChart) drawChartLines(j int) {
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
func (lc *LineChart) drawBox() {
	lc.mBox.StrokeColor = colornames.Blue
	lc.mBox.FillColor = colornames.Black
	lc.mBox.StrokeWidth = 1
	lc.mBox.Resize(lc.mChart.Size())
}
func (lc *LineChart) drawTitle(title string) {
	lc.mTitle.Text = title
	lc.mTitle.Color = colornames.White
	lc.mTitle.TextStyle.Bold = true
	lc.mTitle.TextSize = lc.mChart.Size().Width/30 + 4 //g.contoldW/150 + 2
	lc.mTitle.Move(fyne.NewPos(lc.mChart.Size().Width/2-lc.mTitle.MinSize().Width/2, lc.mChart.Size().Height*titledown))
}

func (lc *LineChart) drawXaxistexts(t [datapointsmax]string) {
	// if len(t) > 10 {
	// 	for i := 0; i < len(t); i++ {
	// 		fmt.Println("#"+t[i])
	// 	}
	// 	os.Exit(1)
	// }
	for i := 0; i < axisticks; i++ {
		lc.mXaxistext[i].Color = color.White
		lc.mXaxistext[i].Text = t[i]
		lc.mXaxistext[i].TextSize = lc.mChart.Size().Width/30 + 3
		lc.mXaxistext[i].TextStyle.Bold = true
		lc.mXaxistext[i].Move(fyne.NewPos(lc.mChart.Size().Width-lc.mChart.Size().Width*axismarg-
			(lc.mChart.Size().Width-(2*axismarg*lc.mChart.Size().Width))*float32(i)/axisticks-
			lc.mXaxistext[i].MinSize().Width/2,
			lc.mChart.Size().Height-lc.mChart.Size().Height*axismarg))
	}
}
func (lc *LineChart) drawYaxistexts() {
	for i := 0; i < axisticks; i++ {
		lc.mYaxistext[i].Color = color.White
		lc.mYaxistext[i].Text = fmt.Sprintf("%v", math.Round(((float64(axisticks-i))*float64(lc.mMax))/(axisticks)))
		lc.mYaxistext[i].TextSize = lc.mChart.Size().Width/30 + 3
		lc.mYaxistext[i].TextStyle.Bold = true
		lc.mYaxistext[i].Move(fyne.NewPos(axismarg*lc.mChart.Size().Width-
			lc.mYaxistext[i].MinSize().Width-ticksize,
			(lc.mChart.Size().Height-
				(2*axismarg*lc.mChart.Size().Height))*float32(i)/(axisticks)+lc.mChart.Size().Height*axismarg-
				float32(lc.mYaxistext[i].MinSize().Height)/2))
	}
}

func (lc *LineChart) FixMax(x int) int {
	if x < 10 {
		return 10
	}
	y := math.Log10(float64(x))
	y = float64(int(y)) + 1
	y = math.Pow(10, y)
	for y/2 > float64(x) && y/2 == float64(int(y/2)) && fmt.Sprintf("%v", int(y/2))[:2] != "12" {
		y = y / 2
	}
	return int(y)
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

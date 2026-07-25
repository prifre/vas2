package vascharts

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"golang.org/x/image/colornames"
)

func (lc *LineChart) DrawXaxisline() {
	w := lc.TheChart.Size().Width
	h := lc.TheChart.Size().Height
	bottomMarg := float32(0.15) // Matchar textens nya bottenmarginal

	lc.mXaxisline.StrokeColor = color.White
	lc.mXaxisline.StrokeWidth = 2
	lc.mXaxisline.Position1 = fyne.NewPos(w*axismarg, h-h*bottomMarg)
	lc.mXaxisline.Position2 = fyne.NewPos(w-w*axismarg, lc.mXaxisline.Position1.Y)
}

func (lc *LineChart) DrawXaxis() {
	w := lc.TheChart.Size().Width
	h := lc.TheChart.Size().Height
	bottomMarg := float32(0.15)

	for i := 0; i < axisticks; i++ {
		lc.mXaxis[i].StrokeColor = color.White
		lc.mXaxis[i].StrokeWidth = 2
		x := w - w*axismarg - (w-(axismarg*w))*float32(i)/axisticks
		lc.mXaxis[i].Position1 = fyne.NewPos(x, h-h*bottomMarg-ticksize)
		lc.mXaxis[i].Position2 = fyne.NewPos(x, h-h*bottomMarg+ticksize)
	}
}

func (lc *LineChart) Drawonexaxistexts(t []string) {
	if (len(lc.mValue) > axisticks) || (len(lc.mTime) > axisticks) {
		j := 0
		gp := len(t)
		if gp > len(lc.mValue) {
			gp = len(lc.mValue)
		}
		st := gp / axisticks
		if st > 0 {
			for i := 1; i < gp; i += st {
				if j < axisticks {
					t[j] = t[i]
					j++
				}
			}
			if j < axisticks {
				t[j] = t[gp-1]
			}
		}
	}

	w := lc.TheChart.Size().Width
	h := lc.TheChart.Size().Height

	// Beräkna dynamisk textstorlek
	fontSize := w/textsizedivider + 3
	if fontSize > 20 {
		fontSize = 20
	} else if fontSize < 4 {
		fontSize = 4
	}

	bottomMarg := float32(0.15)

	for i := 0; i < axisticks; i++ {
		lc.mXaxistext[i].Color = color.White

		// Säker avläsning av index från t
		textVal := ""
		idx := axisticks - 1 - i
		if idx >= 0 && idx < len(t) {
			textVal = t[idx]
		}

		lc.mXaxistext[i].Text = textVal
		lc.mXaxistext[i].TextSize = fontSize
		lc.mXaxistext[i].TextStyle.Bold = true

		x := w - w*axismarg - (w-(axismarg*w))*float32(i)/axisticks - lc.mXaxistext[i].MinSize().Width/2
		y := h - h*bottomMarg + ticksize + 2

		lc.mXaxistext[i].Move(fyne.NewPos(x, y))
	}
}

func (lc *LineChart) Drawonetitle() {
	lc.mTitle.Color = colornames.White
	lc.mTitle.TextStyle.Bold = true
	lc.mTitle.TextSize = lc.TheChart.Size().Width/textsizedivider + 4
	lc.mTitle.Move(fyne.NewPos(lc.TheChart.Size().Width/2-lc.mTitle.MinSize().Width/2, lc.TheChart.Size().Height*titledown))

	if len(lc.mTitle2.Text) > 0 {
		lc.mTitle2.Color = colornames.White
		lc.mTitle2.TextStyle.Bold = true
		lc.mTitle2.TextSize = lc.TheChart.Size().Width/textsizedivider + 4
		ypos2 := lc.TheChart.Size().Height*titledown + lc.mTitle2.TextSize

		// FIX: Centrera baserat på mTitle2:s egen bredd
		lc.mTitle2.Move(fyne.NewPos(lc.TheChart.Size().Width/2-lc.mTitle2.MinSize().Width/2, ypos2))
	}
}
func (lc *LineChart) Drawoneyaxistexts() {
	vt := Setyaxisscale(float64(lc.mMax))
	w := lc.TheChart.Size().Width
	h := lc.TheChart.Size().Height
	bottomMarg := float32(0.15)

	fontSize := w/textsizedivider + 3
	if fontSize > 20 {
		fontSize = 20
	} else if fontSize < 4 {
		fontSize = 4
	}

	drawableHeight := h - (axismarg+bottomMarg)*h

	for i := 0; i < len(vt); i++ {
		lc.mYaxistext[i].Color = color.White
		lc.mYaxistext[i].Text = vt[i]
		lc.mYaxistext[i].TextSize = fontSize
		lc.mYaxistext[i].TextStyle.Bold = true

		// Y-position för tick-linjen med korrekt höjdfördelning
		yTick := drawableHeight*float32(i)/float32(axisticks) + h*axismarg

		yPos := yTick - float32(lc.mYaxistext[i].MinSize().Height)/2

		minTopMargin := fontSize/2 + 1
		if yPos < minTopMargin {
			yPos = minTopMargin
		}

		xPos := axismarg*w - lc.mYaxistext[i].MinSize().Width - ticksize - 2

		lc.mYaxistext[i].Move(fyne.NewPos(xPos, yPos))
	}
}
func Setyaxisscale(max float64) []string {
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
func (lc *LineChart) DrawYaxisline() {
	w := lc.TheChart.Size().Width
	h := lc.TheChart.Size().Height
	bottomMarg := float32(0.15)

	// Y-axis
	lc.mYaxisline.StrokeColor = color.White
	lc.mYaxisline.StrokeWidth = 2
	// Startar vid topp-marginalen (axismarg)
	lc.mYaxisline.Position1 = fyne.NewPos(w*axismarg, h*axismarg)
	// Slutar exakt där X-axellinjen korsar (bottomMarg)
	lc.mYaxisline.Position2 = fyne.NewPos(lc.mYaxisline.Position1.X, h-h*bottomMarg)
}
func (lc *LineChart) DrawYaxis() {
	w := lc.TheChart.Size().Width
	h := lc.TheChart.Size().Height
	bottomMarg := float32(0.15)

	// Den faktiska ritbara höjden mellan toppen och X-axeln
	drawableHeight := h - (axismarg+bottomMarg)*h

	for i := 0; i < axisticks; i++ {
		lc.mYaxis[i].StrokeColor = color.White
		lc.mYaxis[i].StrokeWidth = 2

		// Beräkna Y-positionen baserat på den utökade bottenmarginalen
		yPos := drawableHeight*float32(i)/float32(axisticks) + h*axismarg

		lc.mYaxis[i].Position1 = fyne.NewPos(w*axismarg-ticksize, yPos)
		lc.mYaxis[i].Position2 = fyne.NewPos(w*axismarg+ticksize, yPos)
	}
}
func Fixmax(x float32) float32 {
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

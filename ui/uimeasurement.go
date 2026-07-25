package ui

import (
	"log"

	"fyne.io/fyne/v2"

	"vas2/vascharts"
	"vas2/vasdatabase"
	"vas2/vasmeasure"
)

func UIstartmeasurement() {
	// 1. Om mätningen var pausad, återuppta den
	if ActiveMeasurement != nil && ActiveMeasurement.Paused {
		ActiveMeasurement.Paused = false
		return
	}

	// 2. Om mätningen redan snurrar, gör ingenting
	if ActiveMeasurement != nil && ActiveMeasurement.IsRunning {
		return
	}

	logoContainer.Hide()
	chartsContainer.Show()
	chartsContainer.Refresh()

	// 3. Initiera databas
	if ActiveDatabase == nil {
		ActiveDatabase = new(vasdatabase.DBtype)
		if err := ActiveDatabase.SetupDatabase(); err != nil {
			log.Printf("SetupDatabase error: %v\n", err)
			return
		}
	}

	// 4. Skapa ett nytt mätobjekt
	ActiveMeasurement = vasmeasure.SetupMeasurements(ActiveDatabase)
	ActiveMeasurement.StartMeasurement()
	ActiveMeasurement.IsRunning = true // 👈 Markera som igång!

	// Sätt fönstertiteln säkert
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		windows[0].SetTitle("VAS Analyzer - " + ActiveMeasurement.D.Mname)
	}

	// 5. Skapa en ny buffrad kanal för UI-uppdateringar
	ChartUpdateChan = make(chan []int32, 100)

	// 6. Starta bakgrundsmätningen
	go ActiveMeasurement.Measure(ChartUpdateChan, ActiveMeasurement.IntervalChan)

	// 7. Lyssna på data och uppdatera UI
	go func() {
		for mdata := range ChartUpdateChan {
			currentData := mdata

			fyne.Do(func() {
				vascharts.UpdateChart(ActiveCharts, currentData)
				chartsContainer.Refresh()
			})
		}
	}()
}

func UIstopmeasurement() {
	if ActiveMeasurement != nil {
		ActiveMeasurement.StopMeasurement()
		ActiveMeasurement.IsRunning = false // 👈 Markera som stoppad
		// Låt ActiveMeasurement ligga kvar i minnet eller nollställ vid nästa start
	}

	chartsContainer.Hide()
	logoContainer.Show()
	logoContainer.Refresh()

	// Återställ diagram-objekten
	ActiveCharts = vascharts.NewCharts()
	chartObjects = nil

	for _, chart := range ActiveCharts {
		if chart.TheChart != nil {
			chartObjects = append(chartObjects, chart.TheChart)
		}
	}

	// Uppdatera innehållet i gridet direkt
	if c, ok := chartsContainer.(*fyne.Container); ok {
		c.Objects = chartObjects
		c.Refresh()
	}

	// Sätt tillbaka originaltiteln
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		windows[0].SetTitle("VAS Analyzer")
	}
}

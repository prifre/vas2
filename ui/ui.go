package ui

// To create app, use "fyne package -os windows -icon resources/vas.png"
// to generate "loggor.go" I used; "fyne bundle -o loggor.go resources"

import (
	"log"
	"vas2/general"
	"vas2/vascharts"
	"vas2/vasdatabase"
	"vas2/vasmeasure"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

const (
	chartnum = 8
)

// Create will stitch together all ui components
var (
	logoContainer     fyne.CanvasObject
	chartsContainer   fyne.CanvasObject
	chartObjects      []fyne.CanvasObject
	ActiveCharts      []*vascharts.LineChart
	ActiveMeasurement *vasmeasure.Measuretype
	ActiveDatabase    *vasdatabase.DBtype
	ChartUpdateChan   = make(chan []int32, vasmeasure.Datapointsmax)
)

func AppState(myWindow fyne.Window) {
	general.SetupEnvironment(myWindow)
	SetupMenus(myWindow)

	// 1. Bygg logotyp-container
	logoContainer = general.Showlogo()

	// 2. Initiera databas och diagram
	if ActiveDatabase == nil {
		ActiveDatabase = new(vasdatabase.DBtype)
		if ActiveDatabase.SetupDatabase() != nil {
			log.Print("SetupDatabase failed")
		}
	}

	ActiveCharts = vascharts.NewCharts()
	chartObjects = nil // Återställ slicen säkert

	for _, chart := range ActiveCharts {
		if chart.TheChart != nil {
			chartObjects = append(chartObjects, chart.TheChart)
		}
	}

	// 3. Bygg diagram-grid (gömd från start)
	chartsContainer = container.NewGridWithColumns(2, chartObjects...)
	chartsContainer.Hide()
	logoContainer.Show()

	// 4. Lägg ihop dina vyer i en Stack
	mainContent := container.NewStack(logoContainer, chartsContainer)

	// 5. Använd GridWrap + Stack för att sätta minsta fönsterstorlek (380x680)
	minSizeWrapper := container.NewStack(
		container.NewGridWrap(fyne.NewSize(680, 380)),
		mainContent,
	)

	myWindow.Resize(general.GetWindowSize())
	myWindow.SetContent(minSizeWrapper)

	// 6. Kontrollera om Autostart är aktiverat vid programstart
	// Instead of new(), use a constructor if you have one
	ActiveMeasurement = vasmeasure.SetupMeasurements(ActiveDatabase)
	if ActiveMeasurement.Autostartmeasuring {
		log.Println("Autostart är aktiverat – startar mätning...")

		// Starta inställningar och ladda gammal mätning
		ActiveMeasurement.StartMeasurement()
		// Sätt titeln säkert utan risk för index panic
		myWindow.SetTitle("VAS Analyzer - " + ActiveMeasurement.D.Mname)

		// Visa diagram-griden istället för logotypen
		logoContainer.Hide()
		chartsContainer.Show()

		// Starta mätningsloopen i en egen gorutin (bakgrundstråd)
		go ActiveMeasurement.Measure(ChartUpdateChan, ActiveMeasurement.IntervalChan)

		// Starta en lyssnare som tar emot data från mätningen och uppdaterar UI/diagrammen
		go func() {
			for mdata := range ChartUpdateChan {
				// Använd fyne.Do eller liknande trådsäker uppdatering för UI
				dataCopy := make([]int32, len(mdata))
				copy(dataCopy, mdata)

				fyne.CurrentApp().Driver().DoFromGoroutine(func() {
					vascharts.UpdateChart(ActiveCharts, dataCopy)
				}, false)
			}
		}()
	}
}

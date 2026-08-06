package ui

import (
	"log"
	"os/user"
	"vas2/general"
	"vas2/vascharts"
	"vas2/vasftp"
	"vas2/vasinstruments"
	"vas2/vasmeasure"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
)

var mm fyne.Window

// SetupMenus bygger och sätter huvudmenyn på fönstret
func SetupMenus(window fyne.Window) {
	var nmm *fyne.MainMenu
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("Kunde inte hämta användare: %v", err)
	}
	if currentUser.Name == "prifre" {
		nmm = fyne.NewMainMenu(
			BuildFileMenu(window),
			BuildMeasurementMenu(window),
			BuildHelpMenu(window),
			BuildTestMenu(window),
		)
	} else {
		nmm = fyne.NewMainMenu(
			BuildFileMenu(window),
			BuildMeasurementMenu(window),
			BuildHelpMenu(window),
		)
	}
	window.SetMainMenu(nmm)
}

func BuildFileMenu(window fyne.Window) *fyne.Menu {
	// Skapa Quit-objektet
	itemQuit := fyne.NewMenuItem("Quit", func() {
		general.Closeapp(window)
	})
	// Koppla Ctrl+Q som genväg
	itemQuit.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyQ, Modifier: fyne.KeyModifierControl}
	itemQuit.IsQuit = true

	return fyne.NewMenu("File",
		fyne.NewMenuItem("Setup instruments...", func() {
			result := vasinstruments.DoSetupInstruments()
			dialog.ShowInformation("Setup Instruments result", result, window)
		}),
		fyne.NewMenuItem("FTP settings...", func() {
			vasftp.DoFTPSettings(window)
		}),
		fyne.NewMenuItem("Settings...", func() { DoSettings(window) }),
		fyne.NewMenuItem("Aerotrak Settings...", func() { DoAeroTrakSettings(window) }),
		fyne.NewMenuItem("DustTrak Settings...", func() { DoDustTrakSettings(window) }),
		fyne.NewMenuItem("PTrak Settings...", func() { DoPTrakSettings(window) }),
		fyne.NewMenuItemSeparator(),
		itemQuit, // Lägg till objektet med genvägen
	)
}

func BuildEditMenu(window fyne.Window) *fyne.Menu {
	return fyne.NewMenu("Edit",
		fyne.NewMenuItem("Cut", func() { dialog.ShowInformation("Cut", "Not implemented, sorry!", window) }),
		fyne.NewMenuItem("Copy", func() { dialog.ShowInformation("Copy", "Not implemented, sorry!", window) }),
		fyne.NewMenuItem("Paste", func() { dialog.ShowInformation("Paste", "Not implemented, sorry!", window) }),
	)
}

func BuildMeasurementMenu(window fyne.Window) *fyne.Menu {
	itemNewStart := fyne.NewMenuItem("Start New Measurement", func() {
		// 1. Stoppa pågående mätning
		UIstopmeasurement()

		// 2. Tvinga nytt nanostamp i Preferences & instans
		fyne.CurrentApp().Preferences().SetString("nanostamp", "0")

		// 3. Skapa ny mätinstans
		ActiveMeasurement = vasmeasure.SetupMeasurements(ActiveDatabase)
		ActiveMeasurement.D.Nanostamp = 0
		ActiveMeasurement.D.Mname = ""

		// 4. Starta mätningen (sätter upp nya instrument & SQLite-rad)
		UIstartmeasurement()

		// 5. VIKTIGT: Skapa en ny kanal och starta om bakgrundsloopen!
		ChartUpdateChan = make(chan []int32, vasmeasure.Datapointsmax)

		// Starta mätningsloopen
		go ActiveMeasurement.Measure(ChartUpdateChan, ActiveMeasurement.IntervalChan)

		// Starta UI-lyssnaren
		go func() {
			for mdata := range ChartUpdateChan {
				dataCopy := make([]int32, len(mdata))
				copy(dataCopy, mdata)

				fyne.CurrentApp().Driver().DoFromGoroutine(func() {
					vascharts.UpdateChart(ActiveCharts, dataCopy)
				}, false)
			}
		}()
	})
	itemNewStart.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyN, Modifier: fyne.KeyModifierControl}

	itemStart := fyne.NewMenuItem("Start Measurement", func() {
		UIstartmeasurement()
	})
	itemStart.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}

	itemEnd := fyne.NewMenuItem("End Measurement", func() {
		UIstopmeasurement()
	})
	itemEnd.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyE, Modifier: fyne.KeyModifierControl}

	itemPause := fyne.NewMenuItem("Pause Measurement", func() {
		if ActiveMeasurement != nil {
			ActiveMeasurement.Paused = true
		}
	})
	itemMaintainance := fyne.NewMenuItem("Measurements maintenance", func() {
		// Snygg hantering av underfönstret för att undvika nollpekarkrascher
		if mm == nil {
			mm = fyne.CurrentApp().NewWindow("Measurements Maintenance")
			mm.SetOnClosed(func() {
				mm = nil
			})
			ActiveDatabase.MeasurementsMaintainance(mm)
			mm.Show()
		} else {
			mm.RequestFocus()
		}
	})
	itemMaintainance.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl}
	return fyne.NewMenu("Measurements",
		itemNewStart,
		itemStart,
		itemPause,
		itemEnd,
		fyne.NewMenuItemSeparator(),
		itemMaintainance,
		fyne.NewMenuItem("Export all measurements to textfiles", func() {
			err := ActiveDatabase.Exporttotext()
			if err != nil {
				dialog.ShowInformation("Measurements export failed ", err.Error(), window)
			} else {
				dialog.ShowInformation("All measurements exported ", "ok", window)
			}
		}),
		fyne.NewMenuItem("Upload textfiles to FTP-server", func() {
			go func() {
				err := vasftp.UploadtoFTPserver()
				// dialog.ShowInformation går utmärkt att anropa från gorutiner i Fyne
				if err != nil {
					dialog.ShowInformation("FTP Upload", err.Error(), window)
				} else {
					dialog.ShowInformation("FTP Upload", "Upload complete!", window)
				}
			}()
		}),
		fyne.NewMenuItem("Remove redundant measurements", func() {
			ActiveDatabase.Pruning()
		}),
		fyne.NewMenuItem("Open storage location", func() {
			general.Openstoragelocation()
		}),
	)
}

func BuildHelpMenu(window fyne.Window) *fyne.Menu {
	return fyne.NewMenu("Help",
		fyne.NewMenuItem("About...", func() {
			general.ShowAbout(window)
		}),
		fyne.NewMenuItem("Copyright info...", func() {
			t := "All packages used are below:\n" +
				"Fyne - fyne.io (fyne.io/fyne/v2)\n" +
				"Sqlite3 - (github.com/mattn/go-sqlite3)\n" +
				"FTP - (github.com/jlaffaye/ftp)\n" +
				"modbus - (github.com/goburrow/modbus)\n" +
				"Excelexport - (github.com/360EntSecGroup-Skylar/excelize)\n" +
				"Serial - (github.com/jacobsa/go-serial/serial)\n"

			dialog.ShowInformation("Copyright", t, window)
		}),
		fyne.NewMenuItem("Check for update", func() {
			msg := general.Checkforupdate()
			dialog.ShowInformation("Update information...", msg, window)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Open the VISIBLE AIR SYSTEM webpage!", func() {
			general.Openurl("http://www.prifre.com/vas")
		}),
	)
}

func BuildTestMenu(window fyne.Window) *fyne.Menu {
	return fyne.NewMenu("Tests",
		fyne.NewMenuItem("Test Appinfo()", func() {
			dialog.ShowInformation("Appinfo", general.Getappinfo(), window)
		}),
		fyne.NewMenuItem("Test FillDatabase()", func() {
			ActiveDatabase.FillDatabase()
		}),
		fyne.NewMenuItem("Clear all prefs and data", func() {
			// CLEAR ALL PREFS
			general.ClearPrefsFromJSONKeys("com.prifre.vas")
			ClearAllVasPreferences()
		}))
}

func ClearAllVasPreferences() {
	p := fyne.CurrentApp().Preferences()

	// Lista på alla nycklar som appen sparar
	keys := []string{
		"dbfilename",
		"documentpath",
		"excelfile",
		"homedir",
		"vaslog",
		"winHeight",
		"winWidth",
		"datapoints",
		"countunits",
		"averagepoints",
		"sampleinterval",
		"synchronizedmeasuring",
		"autostartmeasuring",
	}

	// Radera alla nycklar ur Fynes minnes-cache
	for _, k := range keys {
		p.RemoveValue(k)
	}
}

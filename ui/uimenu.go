package ui

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"vas/general"
	"vas/vascharts"
	"vas/vasdatabase"
	"vas/vasftp"

	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"

	//	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"golang.org/x/image/colornames"
)

// application menu
func (g *game) BuildMenu() *fyne.MainMenu {
	// var err error
	// a quit item will be appended to our first menu
	var err error
	var msg string
	mFile := fyne.NewMenu("File",
		fyne.NewMenuItem("Setup instruments...", func() {
			g.StopMeasurement()
			// cnf := dialog.NewConfirm("Setup of instruments", "Search for instruments?", func(reply bool) {
			// 	if reply {
			g.SetupInstruments()
			// 	}
			// }, g.window)
			// cnf.SetDismissText("No")
			// cnf.SetConfirmText("Yes")
			// cnf.Show()
		}),
		fyne.NewMenuItem("FTP settings...", func() {
			g.StopMeasurement()
			DoFTPSettings(g)
		}),
		fyne.NewMenuItem("Settings...", func() {
			g.StopMeasurement()
			DoSettings(g)
		}),
		fyne.NewMenuItem("Manual Settings...", func() {
			g.StopMeasurement()
			DoManualSettings(g)
		}),
		fyne.NewMenuItem("Special Aerotrak Settings...", func() {
			g.StopMeasurement()
			DoAeroTrakSettings(g)
		}))
	// mEdit := fyne.NewMenu("Edit",
	// 	fyne.NewMenuItem("Cut", func() { dialog.ShowInformation("Cut? ", "Not implemented, sorry!", g.window) }),
	// 	fyne.NewMenuItem("Copy", func() { dialog.ShowInformation("Copy? ", "Not implemented, sorry!", g.window) }),
	// 	fyne.NewMenuItem("Paste", func() { dialog.ShowInformation("Paste? ", "Not implemented, sorry!", g.window) }))
	mMeasurements := fyne.NewMenu("Measurements",
		fyne.NewMenuItem("Start Measuring", func() {
			g.StartMeasurement()
		}),
		fyne.NewMenuItem("End Measuring", func() {
			g.StopMeasurement()
			new(vasdatabase.DBtype).Closemeasurement()
			log.Printf("Measurement '%v' (%v) stopped", g.d.mname, g.d.nanostamp)
			fyne.CurrentApp().Preferences().SetString("nanostamp", "0")
			g.d.mname = ""
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Measurements maintainance", func() {
			g.StopMeasurement()
			DoMeasurements()
		}),
		fyne.NewMenuItem("Save screen", func() {
			time.Sleep(time.Second)
			general.Doscreenshot(g.window)
		}),
		fyne.NewMenuItem("Upload to FTP-server", func() {
			var fn string
			for i := 0; i < len(tbl); i++ {
				fn = tbl[i] + ".txt"
				hd:=fyne.CurrentApp().Preferences().String("homedir")
				_, err = os.Stat(filepath.Join(hd, fn))
				if err != nil {
					log.Print("File ", fn, " not found: ", err.Error())
				} else {
					Doftp(fn)
				}
			}
			fn = general.Doscreenshot()
			if fn > "" {
				Doftp(fn)
			}
		}),
		fyne.NewMenuItem("Export all measurements to textfile", func() {
			g.StopMeasurement()
			msg, err = g.d.Exporttotext(g)
			if err != nil {
				dialog.ShowInformation("Error exporting: ", msg, g.window)
			} else {
				dialog.ShowInformation("Export report", msg, g.window)
			}
		}),
		fyne.NewMenuItem("Export current measurement to textfile", func() {
			g.StopMeasurement()
			msg, err = g.d.Exportonetotext(g)
			if err != nil {
				dialog.ShowInformation("Error exporting: ", msg, g.window)
			} else {
				dialog.ShowInformation("Export report", msg, g.window)
			}
		}),
		// fyne.NewMenuItem("Repair database", func() {
		// 	DorepairDatabase(g.d, g.app)
		// }),
		fyne.NewMenuItem("Remove redundant measurements", func() {
			g.StopMeasurement()
			g.d.Pruning(g)
		}),
		fyne.NewMenuItem("Open storage location", func() {
			openstoragelocation(g)
		}))
	mHelp := fyne.NewMenu("Help",
		fyne.NewMenuItem("About...", func() {
			t := "\nProgram for measuring with TSI\nAeroTrak, DustTrak and PTrak\n\nby Peter Freund\n\n"
			t = t + fmt.Sprintf("Version %v", version)
			g.showlogo2()
			dialog.ShowInformation("About...", t, g.window)
			g.showlogo2()
		}),
		fyne.NewMenuItem("Copyright info...", func() {
			t := "All packages used are below:\n"
			t += "Fyne - fyne.io (fyne.io/fyne/v2)\n"
			t += "Sqlite3 - (github.com/mattn/go-sqlite3)\n"
			t += "FTP - (github.com/jlaffaye/ftp)\n"
			t += "modbus - (github.com/goburrow/modbus)\n"
			t += "Excelexport - (github.com/360EntSecGroup-Skylar/excelize)\n"
			t += "Serial - (github.com/jacobsa/go-serial/serial)\n"

			dialog.ShowInformation("About...", t, g.window)
		}),
		fyne.NewMenuItem("Check for update", func() {
			msg:=general.Checkforupdate()
			dialog.ShowInformation("Update information...",msg, g.window)

		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Open the VISIBLE AIR SYSTEM webpage!", func() {
			g.openurl("http://www.prifre.com/vas")
		}),
	)
	return fyne.NewMainMenu(mFile, mMeasurements, mHelp)
}


package vasdatabase

import (
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Maintainance for measurements
func (db *DBtype) MeasurementsMaintainance(mm fyne.Window) string {
	var d1, nanostring, n1 []string
	var err error
	var srclist binding.ExternalStringList
	var myID int = -1
	currentWindow := fyne.CurrentApp().Driver().AllWindows()[0]
	orgsize := currentWindow.Canvas().Size()
	orgsize.Height = orgsize.Height - 100

	// w := fyne.CurrentApp().NewWindow("Measurements")
	// w.SetCloseIntercept(func() {
	// 	w.Close()
	// 	w = nil
	// })

	d1, err = db.Getsql("SELECT mname FROM tblMain ORDER BY nanostamp ASC")
	if err != nil {
		log.Println("#1 DoMeasurements SELECT ", err.Error())
	}
	if len(d1) == 0 {
		log.Println("#2 len(d1)=0")
		srclist = binding.BindStringList(&[]string{""})
		srclist.Append("empty")
	} else {
		srclist = binding.BindStringList(&d1)
	}
	details := widget.NewLabel("")
	nanostring, err = db.Getsql("SELECT nanostamp FROM tblMain ORDER BY nanostamp ASC")
	if err != nil {
		log.Println("#2 DoMeasurements SELECT ", err.Error())
	}
	list := widget.NewListWithData(srclist, func() fyne.CanvasObject {
		co99 := widget.NewLabel("Placeholder")
		return co99
	}, func(item binding.DataItem, obj fyne.CanvasObject) {
		t1 := obj.(*widget.Label)
		t1.Bind(item.(binding.String))
	})
	list.OnSelected = func(id int) {
		myID = id
		if len(nanostring) > 0 {
			details.Text = db.Updatedetails(nanostring[myID])
			details.Refresh()
		}
	}
	wt := widget.NewToolbar(
		//Edit Selected Measurement
		widget.NewToolbarAction(theme.DocumentCreateIcon(),
			func() {
				if myID > -1 && d1 != nil && nanostring != nil {
					d := widget.NewEntry()
					d.SetText(d1[myID])
					d.MultiLine = false
					note := widget.NewEntry()
					note.MultiLine = true
					// GET NOTE FROM DB
					n1, err = db.Getsql("SELECT note FROM tblMain WHERE nanostamp=" + nanostring[myID])
					if err != nil {
						log.Print("#3 DoMeasurements Getting note failed", err.Error())
						return
					}
					note.SetText(n1[0])
					dialog.ShowForm("Measurement information"+strings.Repeat(" ", 100), "Enter", "Cancel", []*widget.FormItem{
						widget.NewFormItem("Measurement name:", d),
						widget.NewFormItem("Note: ", note)}, func(bool) {
						d1[myID] = d.Text
						// update measurementname based on nanostampID
						db.UpdateMeasurementNameNote(nanostring[myID], d.Text, note.Text)
						srclist.Reload()
						details.Text = db.Updatedetails(nanostring[myID])
						details.Refresh()
					}, mm)
					mm.Close()
				}
			}),
		widget.NewToolbarAction(theme.DeleteIcon(),
			// Remove selected measurement!
			func() {
				// Remove the element at index i from a.
				if myID > -1 && len(d1) > 0 && nanostring != nil {
					// if nanostring[myID] == fmt.Sprintf("%v", g.d.Nanostamp) && g.d.Nanostamp > 0 {
					// 	dialog.ShowInformation("Warning", "Active measurement cannot be removed!", w)
					// 	return
					// }
					db.Deleteall(nanostring[myID])
					if myID < len(d1) {
						copy(d1[myID:], d1[myID+1:])                 // Shift a[i+1:] left one index.
						copy(nanostring[myID:], nanostring[myID+1:]) // Shift a[i+1:] left one index.
					}
					d1[len(d1)-1] = ""  // Erase last element (write zero value).
					d1 = d1[:len(d1)-1] // Truncate slice.
					if myID == len(d1) {
						myID--
					}
					srclist.Reload()
					nanostring[len(nanostring)-1] = ""          // Erase last element (write zero value).
					nanostring = nanostring[:len(nanostring)-1] // Truncate slice.
					// details.Text = db.updatedetails(nanostring[myID], d1[myID])
					// details.Refresh()
					list.Unselect(myID)
					list.Refresh()
				}
			}),
		widget.NewToolbarAction(theme.FileTextIcon(),
			func() {
				var n int64
				if myID > -1 && nanostring != nil {
					n, err = strconv.ParseInt(nanostring[myID], 10, 64)
					if err != nil {
						log.Printf("Error converting string to int64: %v\n", err)
						dialog.ShowInformation("One Measurement export failed ", err.Error(), mm)
					}
					db.Nanostamp = n
					err := db.Exportonetotext()
					if err != nil {
						dialog.ShowInformation("One Measurement export failed ", err.Error(), mm)
					} else {
						dialog.ShowInformation("One Measurement exported ", "Textexport finished ok.", mm)
					}
				}
			}),

		widget.NewToolbarAction(theme.DocumentSaveIcon(),
			// Export selected measurement to to Excel!!!
			func() {
				if myID > -1 && nanostring != nil {
					fn := fyne.CurrentApp().Preferences().String("excelfile")
					db.Exporttoexcel(nanostring[myID], fn)
					log.Println("Exported to Excel: " + fn)
					dialog.ShowInformation("Export to Excel!", "Export finished ok.", mm)
				} else {
					dialog.ShowInformation("Export to Excel!", "Select a measurement first!", mm)
				}
			}),
		//Pruning
		widget.NewToolbarAction(theme.SearchReplaceIcon(),
			func() {
				db.Pruning()
				m1 := "Pruning done for selected record.\n"
				m1 = m1 + "Did remove records based on setting 'Save every' so only averages are saved.\n"
				m1 = m1 + "Normally vas saves data every second if possible. This creates too much data\n"
				m1 = m1 + "when exported automatically and data might become unmanagable.\n"
				m1 = m1 + "Pruning automatically or here manually tries to solve this problem."
				dialog.ShowInformation(m1, "Select a measurement first!", mm)
			}),
		//Start Measurement
		widget.NewToolbarAction(theme.MediaPlayIcon(),
			func() {
				if len(nanostring) > 0 {
					if nanostring[myID] > "" {
						fyne.CurrentApp().Preferences().SetBool("autostartmeasuring", true)
						fyne.CurrentApp().Preferences().SetString("nanostamp", nanostring[myID])
						mm.Close()
					}
				}
			}),
		//		widget.NewToolbarSeparator(),
		//		widget.NewToolbarSpacer(),
		//		widget.NewToolbarAction(theme.HomeIcon(),
		//			func() {}),
		//		widget.NewToolbarAction(theme.HelpIcon(),
		//			func() {}),
	)
	c00 := container.NewGridWithColumns(2, list, details)
	mainLayout := container.NewBorder(
		wt,  // Top
		nil, // Bottom
		nil, // Left
		nil, // Right
		c00, // Remaining center space
	)
	//	c1.Resize(fyne.NewSize(c1.Size().Width*10, c1.Size().Height))
	mm.SetCloseIntercept(func() {
		mm.Close()
	})
	mm.SetContent(mainLayout)
	mm.Resize(orgsize)

	mm.CenterOnScreen()
	mm.Show()
	return ""
}

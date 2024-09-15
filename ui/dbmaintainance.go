package ui

import (
	"log"
	"strings"
	"vas/vasdatabase"
	"vas/vasmeasure"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Maintainance for measurements
func DoMeasurements() error {
	var d1, nanostring, n1 []string
	var c0, c2 fyne.Container
	var err error
	var myID int = -1
	w := fyne.CurrentApp().NewWindow("Measurements")
	w.SetCloseIntercept(func() {
		w.Close()
		w = nil
	})
	d1, err = new(vasdatabase.DBtype).Getsql("SELECT mname FROM tblMain ORDER BY nanostamp ASC")
	if err != nil {
		log.Println("#1 handlemeasurements SELECT ", err.Error())
	}
	if len(d1) == 0 {
		return err
	}
	srclist := binding.BindStringList(&d1)
	details := widget.NewLabel("")
	nanostring, err = new(vasdatabase.DBtype).Getsql("SELECT nanostamp FROM tblMain ORDER BY nanostamp ASC")
	if err != nil {
		log.Println("#1 handlemeasurements SELECT ", err.Error())
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
		details.Text = new(vasdatabase.DBtype).Updatedetails(nanostring[myID], d1[myID])
		details.Refresh()
	}
	wt := widget.NewToolbar(
		//Edit Selected Measurement
		widget.NewToolbarAction(theme.DocumentCreateIcon(),
			func() {
				if myID > -1 {
					d := widget.NewEntry()
					d.SetText(d1[myID])
					d.MultiLine = false
					note := widget.NewEntry()
					note.MultiLine = true
					// GET NOTE FROM DB
					n1, err = new(vasdatabase.DBtype).Getsql("SELECT note FROM tblMain WHERE nanostamp=" + nanostring[myID])
					if err != nil {
						log.Print("Getting note failed", err.Error())
						return
					}
					note.SetText(n1[0])
					dialog.ShowForm("Measurement information"+strings.Repeat(" ", 100), "Enter", "Cancel", []*widget.FormItem{
						widget.NewFormItem("Measurement name:", d),
						widget.NewFormItem("Note: ", note)}, func(bool) {
						d1[myID] = d.Text
						// update measurementname based on nanostampID
						new(vasdatabase.DBtype).UpdateMeasurementNameNote(nanostring[myID],d.Text,note.Text)
						srclist.Reload()
						details.Text =new(vasdatabase.DBtype).Updatedetails(nanostring[myID], d1[myID])
						details.Refresh()
					}, w)
				}
			}),
		widget.NewToolbarAction(theme.DeleteIcon(),
			// Remove selected measurement!
			func() {
				// Remove the element at index i from a.
				if myID > -1 && len(d1) > 0 {
					// if nanostring[myID] == fmt.Sprintf("%v", g.d.Nanostamp) && g.d.Nanostamp > 0 {
					// 	dialog.ShowInformation("Warning", "Active measurement cannot be removed!", w)
					// 	return
					// }
					new(vasdatabase.DBtype).Deleteall(nanostring[myID])
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
		widget.NewToolbarAction(theme.DocumentSaveIcon(),
			// Export selected measurement to to Excel!!!
			func() {
				if myID > -1 {
					fn := fyne.CurrentApp().Preferences().String("excelfile")
					new(vasdatabase.DBtype).Exporttoexcel(nanostring[myID], fn)
					log.Println("Exported to Excel: " + fn)
					dialog.ShowInformation("Export to Excel!", "Export finished ok.", w)
				} else {
					dialog.ShowInformation("Export to Excel!", "Select a measurement first!", w)
				}
			}),
		//Start Measurement
		widget.NewToolbarAction(theme.SearchReplaceIcon(),
			func() {
				new(vasdatabase.DBtype).Pruning()
				m1 := "Pruning done for selected record.\n"
				m1 = m1 + "Did remove records based on setting 'Save every' so only averages are saved.\n"
				m1 = m1 + "Normally vas saves data every second if possible. This creates too much data\n"
				m1 = m1 + "when exported automatically and data might become unmanagable.\n"
				m1 = m1 + "Pruning automatically or here manually tries to solve this problem."
				dialog.ShowInformation(m1, "Select a measurement first!", w)
			}),
		//Start Measurement
		widget.NewToolbarAction(theme.MediaPlayIcon(),
			func() {
				if nanostring[myID] > "" {
					fyne.CurrentApp().Preferences().SetBool("autostart", true)
					new(vasmeasure.Measuretype).StartMeasurement()
					w.Close()
				}
			}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HomeIcon(),
			func() {}),
		widget.NewToolbarAction(theme.HelpIcon(),
			func() {}),
	)

	c2 = *container.New(layout.NewGridLayoutWithColumns(2), list, details)
	c0 = *container.New(layout.NewBorderLayout(wt, nil, nil, nil), wt, &c2)
	w.SetContent(&c0)
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()
	w.Show()
	return err
}

package vasftp

import (
	"vas2/general"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

func DoFTPSettings(parentWindow fyne.Window) error {
	prefs := fyne.CurrentApp().Preferences()

	f1e1 := widget.NewEntry()
	f1e2 := widget.NewEntry()
	f1e3 := widget.NewPasswordEntry() // Maskerar lösenordet
	f1e4 := widget.NewEntry()

	f1e1.SetText(prefs.String("ftpserver"))
	f1e2.SetText(prefs.String("ftpusername"))
	encPass := fyne.CurrentApp().Preferences().StringWithFallback("ftppassword", "")
	decrypt, _ := general.DecryptSecret(encPass)
	f1e3.SetText(decrypt)
	f1e4.SetText(prefs.String("ftpdir"))

	sp1 := widget.NewLabel("Please note that connection details are stored in plain text.")

	f1t1 := widget.NewLabel("FTP-server:")
	f1t2 := widget.NewLabel("FTP-username:")
	f1t3 := widget.NewLabel("FTP-password:")
	f1t4 := widget.NewLabel("FTP-path:")

	// 1. Deklarera dialogvariabeln så knappar kan dölja den
	var ftpDialog dialog.Dialog

	okbutton := widget.NewButton("OK", func() {
		prefs.SetString("ftpserver", f1e1.Text)
		prefs.SetString("ftpusername", f1e2.Text)
		// Spara krypterat
		encPass, _ := general.EncryptSecret(f1e3.Text)
		prefs.SetString("ftppassword", encPass)
		prefs.SetString("ftpdir", f1e4.Text)

		if ftpDialog != nil {
			ftpDialog.Hide()
		}
	})

	cancelbutton := widget.NewButton("Cancel", func() {
		if ftpDialog != nil {
			ftpDialog.Hide()
		}
	})

	form := container.NewGridWithColumns(2,
		f1t1, f1e1,
		f1t2, f1e2,
		f1t3, f1e3,
		f1t4, f1e4,
	)

	btnGroup := container.NewGridWithColumns(2, cancelbutton, okbutton)
	content := container.NewVBox(sp1, form, layout.NewSpacer(), btnGroup)

	// 2. Skapa den modala dialogen över parentWindow
	ftpDialog = dialog.NewCustomWithoutButtons("FTP Settings", content, parentWindow)
	ftpDialog.Resize(fyne.NewSize(450, 250))
	ftpDialog.Show()

	return nil
}

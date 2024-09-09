package ui

import (
	"fmt"
	"io"
	"log"
	"os"

	"fyne.io/fyne"
)

// contains preferences for setting up all paths
func GetHomeDir() string {
	var err error
	var mydebug bool
	var path string
	mydebug = fyne.CurrentApp().Preferences().Bool("debug")
	if mydebug {
		path, err = os.Getwd()
		if err != nil {
			log.Fatal("#1 GetHomeDir Failed to get Getwd!!!(" + path + ") ", err.Error())
		}
	} else {
		path, err = os.UserHomeDir()
		if err != nil {
			log.Println("#2 GetHomeDir Failed to get UserHomeDir!!! (" + path + ")", err.Error())
		}
	}
	if fyne.CurrentApp().UniqueID()=="testApp" {
		path="C:\\dev\\go\\src\\vasdata"
	}
	path = fmt.Sprintf("%s%c%s", path, os.PathSeparator, "vasdata")
	if _, err = os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, 0755)
			if err!=nil {
				log.Fatal("#3 GetHomeDir Could not make " + path + " " + err.Error())
			}
		} else {
			log.Fatal("#4 GetHomeDir Could not make " + path + " " + err.Error())
			// other error
		}
	}
	return path
}
func Setupfiles() {
	var wrt io.Writer
	var err error
	fyne.CurrentApp().Preferences().SetString("homedir", fmt.Sprintf("%s", GetHomeDir()))
	fyne.CurrentApp().Preferences().SetString("dbfilename", fmt.Sprintf("%s%c%s", GetHomeDir(), os.PathSeparator, "vasdatabase.db"))
	fyne.CurrentApp().Preferences().SetString("groupsfile", fmt.Sprintf("%s%c%s", GetHomeDir(), os.PathSeparator, "groups.txt"))
	fyne.CurrentApp().Preferences().SetString("historyfile", fmt.Sprintf("%s%c%s", GetHomeDir(), os.PathSeparator, "history.txt"))
	fyne.CurrentApp().Preferences().SetString("pfsmsdb", fmt.Sprintf("%s%c%s", GetHomeDir(), os.PathSeparator, "pfsms.db"))
	fyne.CurrentApp().Preferences().SetString("pfsmslog", fmt.Sprintf("%s%c%s", GetHomeDir(), os.PathSeparator, "pfsms.log"))
	f, err := os.OpenFile(fyne.CurrentApp().Preferences().StringWithFallback("pfsmslog", ""), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic("Error opening pfsms.log:" + err.Error())
	}
	//	defer f.Close()
	wrt = io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)
}
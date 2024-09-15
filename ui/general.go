package ui

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"runtime"
	"strconv"

	"fyne.io/fyne/v2"
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
	fyne.CurrentApp().Preferences().SetString("homedir",  GetHomeDir())
	fyne.CurrentApp().Preferences().SetString("dbfilename", fmt.Sprintf("%s%c%s", GetHomeDir(), os.PathSeparator, "vasdatabase.db"))
	fyne.CurrentApp().Preferences().SetString("excelfile", fmt.Sprintf("%s%c%s", GetHomeDir(), os.PathSeparator, "vas.xlsx"))
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

func Openstoragelocation(g *game) {
	var err error
	var fn string = GetHomeDir()
	if err != nil {
		log.Println("Error getting path to "+fn+".", err)
	}
	log.Println("Opening " + fn)
	err = g.openurl(fn)
	if err != nil {
		log.Println("Error opening "+fn+".", err)
	}
}

func (g *game) ShowMyDebug(s string) {
	if g.MyDebug {
		log.Printf("FUNCTION: %v\n%v\n", printfunc(), s)
	}
}
func printfunc() string {
	pc := make([]uintptr, 1)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	return f.Name()
}

func (g *game) openurl(uu string) error {
	var u *url.URL
	var err error
	u, err = url.Parse(uu)
	if err != nil {
		log.Println("#1 openurl error", err.Error())
	}
	err = fyne.App.OpenURL(g.app, u)
	return err
}

// setups logging,  handles debugging, based on myDebug variable. if myDebug is on, it also prints all logging to console
func (g *game) setuplogging() {
	var err error
	lfn := fmt.Sprintf("%s%d%s%s",GetHomeDir() , os.PathSeparator , g.app.UniqueID(),".log")
	fmt.Println()
	if err != nil {
		log.Println("#1 setuplogging: "+lfn, err.Error())
		return
	}

	logfile, err := os.OpenFile(lfn, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("#2 setuplogging:OpenFile  error: ", err.Error())
		return
	}
	if g.MyDebug {
		multi := io.MultiWriter(logfile, os.Stdout)
		log.SetOutput(multi)
	} else {
		log.SetOutput(logfile)
	}
	//	log.Printf("Log =\t%v", g.Logpath.Path())
	//	log.Printf("Settings =\t%v", strings.Replace(g.Logpath.Path(), g.app.UniqueID()+".log", "preferences.json", 1))
}

func GetcurrentIP() []string {
	// get current computers IP-address
	var err error
	var host string
	var a2 net.IP
	var a1, theip []string
	host, err = os.Hostname()
	if err != nil {
		log.Println("No network???!")
		return nil
	}

	a1, err = net.LookupHost(host)
	if err != nil {
		log.Println("Host lookup failed")
	}
	for i := 0; i < len(a1); i++ {
		a2 = net.ParseIP(a1[i])
		if a2 != nil {
			theip = append(theip, a2.String())
		}
	}
	return theip
}

func Getint64(s string) int64 {
	var err error
	var v int64
	v, err = strconv.ParseInt(s, 0, 64)
	if err != nil {
		log.Println("in64 conversion error", err.Error())
	}
	//	fmt.Printf("Getint64 conversion from %v -> %v \n",s,v)
	return v
}


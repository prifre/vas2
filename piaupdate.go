package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"fyne.io/fyne/v2/dialog"
)

func (g *game) Checkforupdate() {
	var txt string
	var html []byte
	var webversion, created string
	var yourversion string
	url := "http://prifre.com/pia/pia.txt"
	resp, err := http.Get(url)
	// handle the error if there is one
	if err != nil {
		log.Println("#1 getUpdate", err.Error())
	}
	defer resp.Body.Close()
	html, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("#2 getUpdate ioread failed", err.Error())
	}
	// show the HTML code as a string %s
	txt = fmt.Sprintf("%v", html)
	for _, s := range strings.Split(txt, "\n") {
		if len(s) > 9 {
			if s[:8] == "version=" {
				webversion = s[8:]
			}
			if s[:8] == "created=" {
				created = s[8:]
			}
		}
	}
	yourversion = fmt.Sprintf("%v", version)
	yourversion = strings.Trim(yourversion, " \n")
	if webversion <= yourversion {
		dialog.ShowInformation("Update information...",
			strings.Repeat("_", 70)+"\n\n\n"+
				"You have the latest version!\n"+
				webversion+"\nYour version: "+yourversion+"\nCreated: "+created, g.window)
	} else {
		dialog.ShowInformation("Update information...",
			strings.Repeat("_", 70)+"\n\n\n"+
				"There is a newer version!\n\n"+
				"Your version: "+fmt.Sprintf("%v", version)+"\n"+
				"Latest version: "+webversion+"\n\nVersion was updated: "+created+
				"\n\nDownload it from: http://www.prifre.com/pia", g.window)
	}
}

package general

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func Checkforupdate() {
	var txt string
	var html []byte
	var webversion, created string
	var yourversion string
	version:=fyne.CurrentApp().Preferences().String("version")
	url := "https://prifre.com/vas/vas.htm"

	//	resp, err := http.Get(url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("#0 Checkforupdate", err.Error())
	}
	req.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36`)
	resp, err := client.Do(req)

	// handle the error if there is one
	if err != nil {
		log.Println("#1 getUpdate", err.Error())
	}
	defer resp.Body.Close()
	html, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Println("#2 getUpdate ioread failed", err.Error())
	}
	// show the HTML code as a string %s
	txt = string(html)
	fmt.Println(txt)
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
				"\n\nDownload it from: http://www.prifre.com/vas", g.window)
	}
}

package general

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

func Checkforupdate() string {
	var webversion, created string
	url := "https://prifre.com/vas/vas.htm"

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("#0 Checkforupdate NewRequest failed: %v", err)
		return "Could not check for updates: invalid request."
	}
	req.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36`)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("#1 Checkforupdate HTTP request failed: %v", err)
		return "Could not check for updates.\nPlease check your internet connection."
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("#2 Checkforupdate unexpected status code: %d", resp.StatusCode)
		return fmt.Sprintf("Could not check for updates (Server returned HTTP %d).", resp.StatusCode)
	}

	// Läs svaret rad för rad och gör det oberoende av versaler/gemener
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lowerLine := strings.ToLower(line)

		if strings.HasPrefix(lowerLine, "version=") {
			webversion = strings.TrimSpace(line[8:])
		} else if strings.HasPrefix(lowerLine, "created=") {
			created = strings.TrimSpace(line[8:])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("#3 Checkforupdate scan failed: %v", err)
		return "Failed to parse update information from server."
	}

	if webversion == "" {
		return "Could not retrieve version information from server."
	}

	yourversion := fyne.CurrentApp().Metadata().Version

	// Om versionerna är identiska
	if webversion == yourversion {
		return fmt.Sprintf("%s\n\n\nYou have the latest version!\n\n"+
			"Latest public version: %s\n"+
			"Latest public version created: %s\n\n"+
			"Your version: %s\n"+
			"Your version was created: %s",
			strings.Repeat("_", 50), webversion, created, yourversion, AppCreatedTime())
	}
	if webversion < yourversion {
		return fmt.Sprintf("%s\n\n\nYou have a later version than the public version!\n\n"+
			"Latest public version: %s\n"+
			"Latest public version created: %s\n\n"+
			"Your version: %s\n"+
			"Your version was created: %s",
			strings.Repeat("_", 50), webversion, created, yourversion, AppCreatedTime())
	}
	// Om versionerna skiljer sig
	return fmt.Sprintf("%s\n\n\nThere is a newer version available!\n\n"+
		"Your version: %s\n"+
		"Latest version: %s\n\n"+
		"Version was updated: %s\n\n"+
		"Download it from: https://www.prifre.com/vas",
		strings.Repeat("_", 50), yourversion, webversion, created)
}

func AppCreatedTime() string {
	path, err := os.Executable()
	if err != nil {
		log.Printf("AppCreatedTime error: %v", err)
		return "Unknown"
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Printf("AppCreatedTime stat error: %v", err)
		return "Unknown"
	}

	return fileInfo.ModTime().Format("2006-01-02 15:04:05")
}

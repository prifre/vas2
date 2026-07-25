package vasftp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	"vas2/general"

	"fyne.io/fyne/v2"
	"github.com/jlaffaye/ftp"
)

type Ftptype struct {
	client      *ftp.ServerConn
	Ftpserver   string
	Ftpusername string
	Ftppassword string
	Ftpdir      string
}

func UploadtoFTPserver() error {
	prefs := fyne.CurrentApp().Preferences()
	ftpServer := prefs.StringWithFallback("ftpserver", "")
	ftpUsername := prefs.StringWithFallback("ftpusername", "")
	encPass := fyne.CurrentApp().Preferences().StringWithFallback("ftppassword", "")
	ftpPassword, _ := general.DecryptSecret(encPass)
	ftpDir := prefs.StringWithFallback("ftpdir", "")

	if ftpServer == "" || ftpUsername == "" || ftpPassword == "" || ftpDir == "" {
		return errors.New("server, username, password or ftpdir missing")
	}

	tables := []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}
	homeDir := general.GetHomeDir()

	for _, table := range tables {
		fn := table + ".txt"
		fullPath := filepath.Join(homeDir, fn)

		if _, err := os.Stat(fullPath); err != nil {
			log.Printf("File %s not found: %v", fn, err)
			return fmt.Errorf("upload of %s failed: file does not exist", fn)
		}

		if err := Doftp(fn, ftpServer, ftpUsername, ftpPassword, ftpDir); err != nil {
			return fmt.Errorf("failed to upload %s: %w", fn, err)
		}
	}
	return nil
}

func (f1 *Ftptype) Ftplogin() error {
	if f1.Ftpserver == "" || f1.Ftpusername == "" {
		return errors.New("login details not set")
	}

	// Om det saknas port, lägg till standardport :21
	if !strings.Contains(f1.Ftpserver, ":") {
		f1.Ftpserver = f1.Ftpserver + ":21"
		log.Print("Adding port info to FTP-server: ", f1.Ftpserver)
	}

	// 🟢 Koppla upp med 5 sekunders timeout så att UI inte fryser vid felaktig adress/server down
	client, err := ftp.Dial(f1.Ftpserver, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Printf("#Login ftp.Dial failed for %s: %v", f1.Ftpserver, err)
		return fmt.Errorf("ftp dial failed: %w", err)
	}
	f1.client = client

	if err := f1.client.Login(f1.Ftpusername, f1.Ftppassword); err != nil {
		log.Printf("#Login username/password problem for %s: %v", f1.Ftpusername, err)
		return fmt.Errorf("ftp login failed: %w", err)
	}

	return nil
}

func (f1 *Ftptype) Ftplogout() error {
	if f1.client == nil {
		log.Println("Ftplogout: not logged in or already closed")
		return nil
	}

	err := f1.client.Quit()
	if err != nil {
		log.Printf("err_ftplogout#1: %v", err)
	}

	f1.client = nil
	return err
}

func (f1 *Ftptype) Ftpupload(remoteName string, r io.Reader) error {
	if f1.client == nil {
		log.Println("err_ftpupload#1: no f1.client")
		return errors.New("no ftp client connection")
	}

	err := f1.client.Stor(remoteName, r)
	if err != nil {
		log.Printf("err_ftpupload#13: failed to upload %s: %v", remoteName, err)
		return fmt.Errorf("client.Stor failed for %s: %w", remoteName, err)
	}

	log.Println("Upload to", remoteName, "ok.")
	return nil
}

func Doftp(fn, ftpserver, ftpuser, ftppassword, ftpdir string) error {
	if ftpserver == "" || ftpuser == "" {
		return errors.New("no ftp server or username specified")
	}

	f1 := &Ftptype{
		Ftpserver:   ftpserver,
		Ftpusername: ftpuser,
		Ftppassword: ftppassword,
		Ftpdir:      ftpdir,
	}

	// 1. Logga in
	if err := f1.Ftplogin(); err != nil {
		log.Printf("#1 DoUpload Login failed: %v", err)
		return fmt.Errorf("FTP login failed: %w", err)
	}
	defer f1.Ftplogout()

	// 2. Byt mapp på FTP-servern
	if f1.Ftpdir != "" {
		if err := f1.client.ChangeDir(f1.Ftpdir); err != nil {
			log.Printf("Could not change FTP directory to %s: %v", f1.Ftpdir, err)
			return fmt.Errorf("failed to change FTP dir to %s: %w", f1.Ftpdir, err)
		}
	}

	// 3. Öppna den lokala filen
	fullPath := filepath.Join(general.GetHomeDir(), fn)
	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("#2 Doftp Open file failed for %s: %v", fullPath, err)
		return fmt.Errorf("read file failed (%s): %w", fn, err)
	}
	defer file.Close()

	// 4. Ladda upp filen
	if err := f1.Ftpupload(fn, file); err != nil {
		return err
	}

	return nil
}

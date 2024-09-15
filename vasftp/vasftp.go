package vasftp

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/jlaffaye/ftp"
)

type Ftptype struct {
	client      *ftp.ServerConn
	Ftpserver   string
	Ftpusername string
	Ftppassword string
	Ftpdir      string
}

func (f1 *Ftptype) Ftplogin() error {
	var err error
	if f1.Ftpserver == "" || f1.Ftpusername == "" {
		return fmt.Errorf("login details not set")
	}
	if f1.Ftpserver[len(f1.Ftpserver)-3:] != ":21" && !strings.Contains(f1.Ftpserver, ":") {
		f1.Ftpserver = f1.Ftpserver + ":21"
		log.Print("Adding port info to FTP-server:", f1.Ftpserver)
	}
	f1.client, err = ftp.Dial(f1.Ftpserver)
	if err != nil {
		log.Print("#Login ftp.Dial failed:", err.Error())
		return err
	}
	err = f1.client.Login(f1.Ftpusername, f1.Ftppassword)
	if err != nil {
		log.Print("#Login username/password problem:", err.Error())
		return err
	}
	return nil
}
func (f1 *Ftptype) Ftplogout() error {
	var err error
	if f1.client == nil {
		log.Println("not logged in")
		return nil
	}
	if err = f1.client.Quit(); err != nil {
		log.Print("err_ftplogout#1", err.Error())
		return err
	}
	return nil
}
func (f1 *Ftptype) Ftpupload(remote_name string, buf bytes.Buffer) error {
	// gets ftp-file from server and saves locally
	var err error
	if f1.client == nil {
		log.Println("err_ftpupload#1: no f1.client")
		return err
	}
	if err != nil {
		log.Println("err_ftpupload#2:", err.Error())
		return err
	}
	err = f1.client.Stor(remote_name, &buf)
	if err != nil {
		log.Println("err_ftpupload#13", err.Error())
		return err
	}
	log.Println("Upload to ", remote_name, " ok.")
	return nil
}

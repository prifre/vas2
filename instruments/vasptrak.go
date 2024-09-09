/*

module for reading PTrak data

*/

package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"golang.org/x/sys/windows/registry"
)

type PTraktype struct {
	simulatePTrak   bool
	PTrakrunning    bool
	PTrakconnection io.ReadWriteCloser
	PTrakolddata    int
	PTrakport       string
	mdata           int64
}

func (pt *PTraktype) openPTrak() io.ReadWriteCloser {
	var port io.ReadWriteCloser
	var err error
	// Opens Windows Registry
	if pt.PTrakport == "" {
		log.Println("#1 openPTrakport no port in settings")
		port = nil
		return nil
	}
	options := serial.OpenOptions{PortName: pt.PTrakport, BaudRate: 9600, DataBits: 8, StopBits: 1, MinimumReadSize: 4}
	// Open the port.
	port, err = serial.Open(options)
	if err != nil {
		log.Println("#1 openPtrakport serialOpen: ", err.Error())
		port = nil
		return nil
	}
	return port
}
func (pt *PTraktype) GetPTrakdata() int32 {
	var v int = -1
	var oldv int
	var r int
	if pt.PTrakconnection == nil {
		pt.PTrakconnection = pt.openPTrak()
		if pt.PTrakconnection == nil {
			log.Println("#2 getPTrakdata could not open PTrak")
		}
	}
	if pt.PTrakconnection != nil {
		for r < 1 {
			b := []byte("QLCONC\r\n")
			n2, err := pt.PTrakconnection.Write(b)
			if err != nil {
				log.Println("#2 getPTrakdata portWrite: ", err, n2)
			}
			lpData := make([]byte, 30)
			n1, err := pt.PTrakconnection.Read(lpData)
			if err != nil {
				log.Println("#3 getPTrakdata portRead: ", err, n1)
			}
			reply := strings.Split(string(lpData), "\r")[0]
			if len(reply) == 12 {
				if reply[:6] == "QLCONC" {
					v, err = strconv.Atoi(reply[7:])
					if err != nil {
						log.Println("#4 getPTrakdataerr conversion", err)
					} else {
						if v != oldv {
							r = r + 1
						}
					}
				}
			}
		}
	}
	if v == pt.PTrakolddata {
		v = -1
	} else {
		pt.PTrakolddata = v
	}
	return int32(v)
}
func (pt *PTraktype) PTrakstop() {
	if pt.PTrakconnection == nil {
		return
	}
	err := pt.PTrakconnection.Close()
	if err != nil {
		log.Println("#1 PTrakstop.Close error")
	}
}
func SimulatedPTrak() int32 {
	return rand.Int31n(1000)
}

func (pt *PTraktype) TestGetPTrakdata(t *testing.T) {
	var d int32
	for i := 0; i < 100; i++ {
		for pt.mdata < 0 {
			time.Sleep(time.Second / 10)
			d = pt.GetPTrakdata()
		}
		fmt.Println(d)
	}
}

// Finds PTrak by searching serial ports and returns  "PTrak:COM4",nil if found or "",err if error or not found
func (pt *PTraktype) FindPTrak() (string, error) {
	var port io.ReadWriteCloser
	var err error
	var theport string = ""
	var k registry.Key
	k, err = registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\\DEVICEMAP\\SERIALCOMM`, registry.QUERY_VALUE)
	if err != nil {
		return "", errors.New("#1 Findports - COM-ports missing")
	}
	defer k.Close()
	ki, err := k.Stat()
	if err != nil {
		return "", errors.New(fmt.Sprint("#2 Findports - Stat: ", err.Error()))
	}
	s, err := k.ReadValueNames(int(ki.ValueCount))
	if err != nil {
		return "", errors.New(fmt.Sprint("#3 Findports - ReadValueNames: ", err.Error()))
	}
	comports := make([]string, ki.ValueCount)
	for i, test := range s {
		q, _, err := k.GetStringValue(test)
		if err != nil {
			log.Println("#4  Findports - GetStringValue: ", err)
		}
		comports[i] = q
	}
	log.Println("Found COM-ports: ", comports)
	// Checks ports to find PTrak...
	for i := 0; i < len(comports); i++ {
		options := serial.OpenOptions{PortName: comports[i], BaudRate: 9600, DataBits: 8, StopBits: 1, MinimumReadSize: 4}
		// Open the port.
		//		fmt.Println("Checking port", comports[i])
		port, err = serial.Open(options)
		if err != nil {
			return "", errors.New(fmt.Sprint("#5 Findports:serial.Open: ", err.Error()))
		}
		// Make sure to close it later.
		s1 := "QLMODL\r\n"
		n2, err := port.Write([]byte(s1))
		if err != nil {
			log.Println("#6 Findports - portWrite: ", err, n2)
		}
		lpData := make([]byte, 30)
		n1, err := port.Read(lpData)
		if err != nil {
			log.Println("#7 Findports - portRead: ", err, n1)
		}
		reply := strings.TrimSpace(strings.Split(string(lpData), "\n")[0])
		if reply == "QLMODL8525" {
			theport = comports[i]
			log.Println("Found PTrak at port: ", theport)
			g.app.Preferences().SetString("PTrak", theport)
		}
		port.Close()
	}
	if theport > "" {
		pt.PTrakport = theport
		return "PTrak:" + theport, nil
	}
	return "", err
}

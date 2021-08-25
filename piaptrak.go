/*

module for reading PTrak data

*/

package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jacobsa/go-serial/serial"
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

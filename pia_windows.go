/*

module for reading PTrak data
part that just works on Windows...
*/

package main

import (
	"io"
	"log"
	"strings"

	"github.com/jacobsa/go-serial/serial"
	"golang.org/x/sys/windows/registry"
)

// Finds PTrak by searching serial ports and returns  "PTrak:COM4",nil if found or "",err if error or not found
func (pt *PTraktype) FindPTrak() (string, error) {
	var port io.ReadWriteCloser
	var err error
	var theport string = ""
	var k registry.Key
	k, err = registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\\DEVICEMAP\\SERIALCOMM`, registry.QUERY_VALUE)
	if err != nil {
		log.Println("#1 Findports - FindPtrak: ", err.Error())
		return "", err
	}
	defer k.Close()
	ki, err := k.Stat()
	if err != nil {
		log.Println("#2 Findports - Registry stat problem: ", err.Error())
		return "", err
	}
	s, err := k.ReadValueNames(int(ki.ValueCount))
	if err != nil {
		log.Println("#3 Findports - ReadValueNames error: ", err.Error())
		return "", err
	}
	comports := make([]string, ki.ValueCount)
	for i, test := range s {
		q, _, err := k.GetStringValue(test)
		if err != nil {
			log.Println("#4  Findports - ", err)
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
			log.Println("#5 Findports - serialOpen: ", err)
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
			log.Println("#8 Findports - Found PTrak at port ", theport)
		}
		port.Close()
	}
	if theport > "" {
		pt.PTrakport = theport
		return "PTrak:" + theport, nil
	}
	return "", err
}

package vasinstruments

/*
module for reading PTrak data
*/

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	jserial "github.com/jacobsa/go-serial/serial"
	bugserial "go.bug.st/serial"
)

func (pt *PTraktype) FindPTrak() (string, error) {
	// 1. Hämta alla tillgängliga serieportar på ett plattformsoberoende sätt
	ports, err := bugserial.GetPortsList()
	if err != nil {
		return "", fmt.Errorf("#1 Findports - Could not list serial ports: %w", err)
	}

	if len(ports) == 0 {
		log.Println("No serial ports found on system.")
		return "", errors.New("no serial ports found")
	}

	log.Println("Found available serial ports:", ports)

	var theport string = ""

	// 2. Loopa igenom och testa alla hittade portar
	for _, portName := range ports {
		options := jserial.OpenOptions{
			PortName:              portName,
			BaudRate:              9600,
			DataBits:              8,
			StopBits:              1,
			MinimumReadSize:       1,
			InterCharacterTimeout: 200,
		}

		port, err := jserial.Open(options)
		if err != nil {
			// Om en port är upptagen eller saknar rättigheter, hoppa vidare
			log.Printf("Could not open port %s: %v\n", portName, err)
			continue
		}

		// Skicka identifieringsfråga till instrumentet
		cmd := "QLMODL\r\n"
		_, err = port.Write([]byte(cmd))
		if err != nil {
			log.Printf("Failed writing to port %s: %v\n", portName, err)
			_ = port.Close()
			continue
		}

		// Ge PTrak lite tid att hinna svara (250 ms är säkrare för serieomvandlare)
		time.Sleep(250 * time.Millisecond)

		// Gör en säkrad läsning i bakgrunden för att undvika att port.Read hänger sig
		type readResult struct {
			n   int
			err error
			buf []byte
		}
		ch := make(chan readResult, 1)

		go func() {
			buf := make([]byte, 64)
			n, err := port.Read(buf)
			ch <- readResult{n: n, err: err, buf: buf}
		}()

		var res readResult
		select {
		case res = <-ch:
			// Läsningen slutfördes
		case <-time.After(500 * time.Millisecond):
			// Porten svarade inte inom 500 ms (t.ex. en tyst serieport)
			log.Printf("Timeout reading from port %s\n", portName)
			_ = port.Close()
			continue
		}

		_ = port.Close() // Stäng alltid porten efter avläsningen

		if res.err != nil || res.n == 0 {
			continue
		}

		reply := strings.TrimSpace(strings.Split(string(res.buf[:res.n]), "\r")[0])
		log.Printf("Port %s replied: %s\n", portName, reply)

		// Kolla om svar/modell stämmer överens med PTrak (8525)
		if strings.Contains(reply, "QLMODL8525") || reply == "QLMODL8525" {
			theport = portName
			log.Println("Found PTrak at port:", theport)

			// Säker lagring i preferences utan kraschrisk
			if app := fyne.CurrentApp(); app != nil {
				app.Preferences().SetString("PTrak", theport)
			}
			break
		}
	}

	if theport != "" {
		pt.PTrakport = theport
		return "PTrak:" + theport, nil
	}

	return "", errors.New("PTrak instrument not found on any serial port")
}

package vasinstruments

/*
module for reading PTrak data
*/

import (
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

type PTraktype struct {
	SimulatePTrak   bool
	PTrakrunning    bool
	PTrakconnection io.ReadWriteCloser
	PTrakolddata    int
	PTrakport       string
	mdata           int64
}

func (pt *PTraktype) OpenPTrak() io.ReadWriteCloser {
	if pt.PTrakport == "" {
		log.Println("#1 OpenPTrak: no port configured in settings")
		return nil
	}

	options := serial.OpenOptions{
		PortName:              pt.PTrakport,
		BaudRate:              9600,
		DataBits:              8,
		StopBits:              1,
		MinimumReadSize:       1,
		InterCharacterTimeout: 200,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Println("#1 OpenPTrak serial.Open error:", err.Error())
		return nil
	}
	return port
}

func (pt *PTraktype) GetPTrakdata() int32 {
	if pt.PTrakconnection == nil {
		pt.PTrakconnection = pt.OpenPTrak()
		if pt.PTrakconnection == nil {
			log.Println("#2 GetPTrakdata: could not open PTrak port")
			return -1
		}
	}

	var val int = -1

	for attempts := 0; attempts < 3; attempts++ {
		// Skicka kommando
		b := []byte("QLCONC\r\n")
		_, err := pt.PTrakconnection.Write(b)
		if err != nil {
			log.Println("#2 GetPTrakdata Write error:", err)
			pt.PTrakstop()
			break
		}

		// Läs in data tills vi hittar en radbrytning (\r eller \n) eller får timeout
		var rawResponse strings.Builder
		buf := make([]byte, 1)
		deadline := time.Now().Add(400 * time.Millisecond)

		for time.Now().Before(deadline) {
			n, err := pt.PTrakconnection.Read(buf)
			if err == nil && n > 0 {
				rawResponse.WriteByte(buf[0])
				// När vi ser vagnretur eller radbrytning har vi hela svaret!
				if buf[0] == '\r' || buf[0] == '\n' {
					break
				}
			} else {
				time.Sleep(10 * time.Millisecond)
			}
		}

		reply := strings.TrimSpace(rawResponse.String())

		// Förväntat format: t.ex. "QLCONC 1234" eller "QLCONC1234"
		if strings.HasPrefix(reply, "QLCONC") {
			cleanVal := strings.TrimPrefix(reply, "QLCONC")
			cleanVal = strings.TrimSpace(cleanVal)

			parsedVal, parseErr := strconv.Atoi(cleanVal)
			if parseErr == nil {
				val = parsedVal
				break // Giltigt värde hittat!
			} else {
				log.Println("#4 GetPTrakdata conversion error:", parseErr, "from reply:", reply)
			}
		}
	}

	pt.PTrakolddata = val
	return int32(val)
}

func (pt *PTraktype) PTrakstop() {
	if pt.PTrakconnection != nil {
		err := pt.PTrakconnection.Close()
		if err != nil {
			log.Println("#1 PTrakstop Close error:", err)
		}
		pt.PTrakconnection = nil
	}
}

package vasinstruments

/*
module for reading PTrak data
*/

import (
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"github.com/jacobsa/go-serial/serial"
)

type PTraktype struct {
	SimulatePTrak   bool
	PTrakrunning    bool
	PTrakconnection io.ReadWriteCloser
	PTrakolddata    int
	PTrakport       string
	mdata           int64
	debugPTraK      bool
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
		MinimumReadSize:       0,   // 💡 VIKTIGT: 0 gör att Read() inte blockerar oändligt!
		InterCharacterTimeout: 100, // Timeout i ms om inga fler tecken kommer
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Println("#1 OpenPTrak serial.Open error:", err.Error())
		return nil
	}
	return port
}

func (pt *PTraktype) readLineWithTimeout(timeout time.Duration) (string, error) {
	var rawResponse strings.Builder
	buf := make([]byte, 1)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		n, err := pt.PTrakconnection.Read(buf)
		if err != nil {
			return "", err
		}
		if n > 0 {
			rawResponse.WriteByte(buf[0])
			if buf[0] == '\r' || buf[0] == '\n' {
				break
			}
		} else {
			// MinimumReadSize=0 gör att Read returnerar n=0 direkt om inget finns,
			// så vi pausar kort för att inte belasta CPU:n 100%
			time.Sleep(10 * time.Millisecond)
		}
	}

	return strings.TrimSpace(rawResponse.String()), nil
}

func (pt *PTraktype) GetPTrakdata() (int32, error) {
	if pt.PTrakconnection == nil {
		pt.PTrakconnection = pt.OpenPTrak()
		if pt.PTrakconnection == nil {
			log.Println("#2 GetPTrakdata: could not open PTrak port")
			return -1, errors.New("could not open PTrak port")
		}
	}

	var val int = -1

	for attempts := 0; attempts < 3; attempts++ {
		// Skicka kommando
		b := []byte("QLCONC\r\n")
		_, err := pt.PTrakconnection.Write(b)
		if err != nil {
			log.Println("#2 GetPTrakdata Write error:", err)
			pt.PTrakStop()
			return -1, err
		}
		if pt.debugPTraK {
			log.Println("#1 GetPTrakdata Write command:", string(b))
		}

		// Läs med en säker 500ms timeout
		reply, err := pt.readLineWithTimeout(500 * time.Millisecond)
		if err != nil {
			log.Println("#3 GetPTrakdata Read error:", err)
			pt.PTrakStop()
			return -1, err
		}

		if pt.debugPTraK {
			log.Println("#3 GetPTrakdata raw reply:", reply)
		}

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
				return -1, parseErr
			}
		}
	}

	pt.PTrakolddata = val
	if pt.debugPTraK {
		log.Println("#5 GetPTrakdata returning value:", val)
	}
	return int32(val), nil
}

func (pt *PTraktype) PTrakStop() error {
	var pref = fyne.CurrentApp().Preferences()
	pt.PTrakport = pref.StringWithFallback("PTrak", "")
	if pt.PTrakconnection != nil {
		err := pt.PTrakconnection.Close()
		if err != nil {
			log.Println("#1 PTrakStop Close error:", err)
			return err
		}
		pt.PTrakconnection = nil
	} else {
		log.Println("#2 PTrakStop: connection was already nil")
	}
	return nil
}
func (pt *PTraktype) PTrakStart() error {
	var pref = fyne.CurrentApp().Preferences()
	pt.PTrakport = pref.StringWithFallback("PTrak", "")
	if pt.PTrakconnection == nil {
		pt.PTrakconnection = pt.OpenPTrak()
		if pt.PTrakconnection == nil {
			log.Println("#2 GetPTrakdata: could not open PTrak port")
			return errors.New("could not open PTrak port")
		}
	}

	for attempts := 0; attempts < 3; attempts++ {
		// Skicka kommando
		b := []byte("QLCONC\r\n")
		_, err := pt.PTrakconnection.Write(b)
		if pt.debugPTraK {
			log.Println("#1 GetPTrakdata Write command:", string(b))
		}
		if err != nil {
			log.Println("#2 GetPTrakdata Write error:", err)
			pt.PTrakStop()
			return err
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
		if pt.debugPTraK {
			log.Println("#3 GetPTrakdata raw reply:", reply)
		}

		// Förväntat format: t.ex. "QLCONC 1234" eller "QLCONC1234"
		if strings.HasPrefix(reply, "QLCONC") {
			cleanVal := strings.TrimPrefix(reply, "QLCONC")
			cleanVal = strings.TrimSpace(cleanVal)

			_, parseErr := strconv.Atoi(cleanVal)
			if parseErr == nil {
				return nil
			} else {
				log.Println("#4 GetPTrakdata conversion error:", parseErr, "from reply:", reply)
			}
		}
	}
	return errors.New("failed to start PTrak after multiple attempts")
}

package vasinstruments

/*
TCP routines to read from TSI DustTrak
*/

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type DustTraktype struct {
	SimulateDustTrak bool
	DustTrakrunning  bool
	DustTrakconn     net.Conn
	DustTraklastin   int
	DustTrakport     string
}

func (dt *DustTraktype) tcpcommand(cmd string) (string, error) {
	if cmd == "WSDATETIME" {
		cmd = "WSDATETIME " + dt.setdusttrakdatetime()
	}

	if dt.DustTrakconn == nil {
		return "", fmt.Errorf("#1 tcpcommand: connection is nil")
	}

	// Sätt en deadline för både skrivning och läsning så att vi aldrig låser tråden
	_ = dt.DustTrakconn.SetDeadline(time.Now().Add(2 * time.Second))

	// TSI förväntar sig \r\n (CRLF)
	b := []byte(cmd + "\r\n")
	_, err := dt.DustTrakconn.Write(b)
	if err != nil {
		log.Println("#1 tcpcommand write error:", err)
		dt.closeConn() // Stäng död anslutning
		return "", err
	}

	tmp := make([]byte, 1024)
	n, err := dt.DustTrakconn.Read(tmp)
	if err != nil {
		log.Println("#1 tcpcommand read error:", err)
		dt.closeConn()
		return "", err
	}

	reply := strings.TrimSpace(string(tmp[:n]))
	if reply == "FAIL" {
		log.Printf("DustTrak command #%s# -> FAIL", cmd)
		return "", fmt.Errorf("command failed")
	}

	return reply, nil
}

func (dt *DustTraktype) setdusttrakdatetime() string {
	// DustTrak time format: "01/18/2011,11:44:00"
	t := time.Now()
	return fmt.Sprintf("%02d/%02d/%04d,%02d:%02d:%02d",
		int(t.Month()), t.Day(), t.Year(),
		t.Hour(), t.Minute(), t.Second())
}

func (dt *DustTraktype) tcpdusttrakopen() error {
	var err error
	if dt.DustTrakport == "" {
		log.Println("#1 tcpdusttrak: no DustTrak port set in preferences")
		return fmt.Errorf("no port set")
	}

	// 🟢 Höj timeout från 10ms till 2 sekunder för stabil nätverksuppkoppling
	dt.DustTrakconn, err = net.DialTimeout("tcp", dt.DustTrakport, 2*time.Second)
	if err != nil {
		log.Println("#2 tcpdusttrak open error:", err.Error())
		dt.DustTrakconn = nil
		return err
	}

	return nil
}

func (dt *DustTraktype) closeConn() {
	if dt.DustTrakconn != nil {
		_ = dt.DustTrakconn.Close()
		dt.DustTrakconn = nil
	}
}

func (dt *DustTraktype) GetDustTrakdata() int32 {
	var reply string
	var samples []int32
	var lastmeasurementnumber int
	var err error

	dt.DustTrakrunning = false

	if dt.DustTrakconn == nil {
		if err = dt.tcpdusttrakopen(); err != nil {
			log.Println("#1 GetDustTrakdata: could not open port:", err.Error())
			return -1
		}
	}

	retry := 0
	for len(samples) < 1 && retry < 5 {
		reply, err = dt.tcpcommand("MSTATUS")
		if err != nil || reply != "Running" {
			log.Println("#2 GetDustTrakdata MSTATUS:", reply)

			if reply == "Log Error" {
				_, _ = dt.tcpcommand("WMODECURLOG0 12:15:0,09/30/2021,0:1,0:2:0,3,0:0:1,5,0,0,0:15,0,LOG MODE 1")
				time.Sleep(1 * time.Second)
				_, _ = dt.tcpcommand("MUPDATE")
				time.Sleep(1 * time.Second)
			}

			// Starta mätning på instrumentet
			reply, err = dt.tcpcommand("MSTART")
			if err != nil {
				log.Println("#3 GetDustTrakdata could not start measuring:", err)
				dt.closeConn()
				return -1
			}
			time.Sleep(500 * time.Millisecond)
		}

		// Läs av mätdata
		reply, err = dt.tcpcommand("RMMEAS")
		if err == nil && len(reply) > 2 && reply != "FAIL" {
			t := strings.Split(reply, ",")
			if len(t) > 2 {
				lastmeasurementnumber, err = strconv.Atoi(t[0])
				if err == nil && dt.DustTraklastin < lastmeasurementnumber {
					// Extrahera mätvärdena
					for _, rawVal := range t[1 : len(t)-1] {
						if rawVal == "" {
							continue
						}
						f, parseErr := strconv.ParseFloat(rawVal, 64)
						if parseErr == nil {
							samples = append(samples, int32(f*1000)) // Konvertera mg/m³ till µg/m³
						}
					}
					dt.DustTraklastin = lastmeasurementnumber
				}
			}
		}

		time.Sleep(200 * time.Millisecond)
		retry++
	}

	dt.DustTrakrunning = true

	if len(samples) > 0 {
		return samples[0]
	}

	return -1
}

func (dt *DustTraktype) DustTrakstop() error {
	if dt.DustTrakconn == nil {
		if err := dt.tcpdusttrakopen(); err != nil {
			return err
		}
	}

	reply, err := dt.tcpcommand("MSTOP")
	if err != nil || reply != "OK" {
		log.Println("#1 DustTrakstop - retry MSTOP")
		_, err = dt.tcpcommand("MSTOP")
	}

	dt.closeConn()
	return err
}

func (dt *DustTraktype) GetDustTrakinfo(port string) string {
	dt.DustTrakport = port
	if err := dt.tcpdusttrakopen(); err != nil {
		return "Could not connect"
	}
	defer dt.closeConn()

	var t string
	reply, _ := dt.tcpcommand("MSTOP")
	t += "MSTOP: " + reply + "\n"

	reply, _ = dt.tcpcommand("RDMN")
	t += "Model: " + reply + "\n"

	reply, _ = dt.tcpcommand("RDSN")
	t += "Serial: " + reply + "\n"

	reply, _ = dt.tcpcommand("MSTATUS")
	t += "Status: " + reply + "\n"

	reply, _ = dt.tcpcommand("RSDATETIME")
	t += "DateTime: " + reply + "\n"

	log.Println(t)
	return t
}

package instruments

/*

2021-04-06
fixed so DustTrak reading works!

2021-04-03 23:23
TCP routines to read from DustTrak/AeroTrak...
working ok

*/

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

type DustTraktype struct {
	simulateDustTrak bool
	DustTrakrunning  bool
	DustTrakconn     net.Conn
	DustTraklastin   int
	DustTrakport     string
}

func (dt *DustTraktype) tcpcommand(cmd string) (string, error) {
	var reply string = ""
	var err error
	var n int
	if cmd == "WSDATETIME" {
		cmd = "WSDATETIME " + dt.setdusttrakdatetime()
	}
	if dt.DustTrakconn == nil {
		log.Println("#1 connection closed!!!")
		return "", err
	}
	b := []byte(cmd + "\n\r")
	n, err = dt.DustTrakconn.Write(b)
	if err != nil {
		log.Println("#1 tcpcommand read error:", err, n)
		return "", err
	}
	buf := make([]byte, 0, 4096) // big buffer
	tmp := make([]byte, 256)     // using small tmo buffer for demonstrating
	n, err = dt.DustTrakconn.Read(tmp)
	if err != nil {
		log.Println("#1 tcpcommand read error:", err, n)
		return "", err
	} else {
		reply = string(append(buf, tmp[:n]...))
		reply = strings.TrimSpace(reply)
		if reply == "FAIL" {
			log.Print("DustTrak #", cmd, "#->FAIL")
			return "", err
		}
	}
	return reply, err
}
func (dt *DustTraktype) setdusttrakdatetime() string {
	// DustTrak time format:  "01/18/2011,11:44:00"
	t := time.Now()
	formatted := string(fmt.Sprintf("%d/%d/%02d,%02d:%02d:%02d", t.Month(), t.Day(), t.Year(), t.Hour(), t.Minute(), t.Second()))
	return formatted
}
func (dt *DustTraktype) tcpdusttrakopen() error {
	var err error
	if dt.DustTrakport == "" {
		log.Println("#1 tcpdusttrak:tcpcommand DustTrak port set")
		return nil
	}
	dt.DustTrakconn, err = net.DialTimeout("tcp", dt.DustTrakport, 10*time.Millisecond)
	if err != nil {
		log.Println("#2 tcpdusttrak:tcpcommand open ", err.Error())
	}
	return err
}
func (dt *DustTraktype) GetDustTrakdata() int32 {
	/*
		RMMEAS reads DustTrak measuremments!
		measurements format read 2,0.013,0.013,0.014,0.025,0.074,
		where first number is samplenumber
	*/
	var reply string
	var samples = []int32{}
	var lastmeasurementnumber int
	var err error
	var t []string
	var f float64 = 0
	dt.DustTrakrunning = false
	if dt.DustTrakconn == nil {
		err = dt.tcpdusttrakopen()
		if err != nil {
			log.Println("#1 GetDustTrakdata could not open port", err.Error())
			dt.DustTrakrunning = true
			return -1
		}
	}
	retry := 0
	for len(samples) < 2 && retry < 10 {
		reply, err = dt.tcpcommand("MSTATUS")
		if err != nil {
			log.Println("#2 GetDustTrakdata DustTrak could not start measuring!", reply, err.Error())
		}
		if reply != "Running" {
			log.Println("#4 GetDustTrakdata MSTATUS=#", reply, "#")
			if reply == "Log Error" {
				//				reply, _ = dt.tcpcommand("WMODELOG3 12:15:0,09/30/2021,0:1,0:2:0,3,0:0:1,5,0,0,0:15,0,LOG MODE 1\n")    //set Log 0
				reply, _ = dt.tcpcommand("WMODECURLOG0 12:15:0,09/30/2021,0:1,0:2:0,3,0:0:1,5,0,0,0:15,0,LOG MODE 1\n") //set Log 0
				time.Sleep(time.Second * 4)
				log.Println("WMODELOG2_>", reply)
				reply, _ = dt.tcpcommand("MUPDATE") //set Log 0
				time.Sleep(time.Second * 2)
				log.Println("MUPDATE_>", reply)
				reply, _ = dt.tcpcommand("RMLOGINFO") //set Log 0
				log.Println("RMLOGINFO_>", reply)
			}
			log.Print("#2 GetDustTrakdata DustTrak restart")
			time.Sleep(time.Second * 3)
			reply, err = dt.tcpcommand("MSTART") //start Measuring
			if err != nil {
				log.Println("#2 GetDustTrakdata DustTrak could not start measuring!", reply, err.Error())
				dt.DustTrakrunning = true
				return -1
			}
		}
		reply, err = dt.tcpcommand("RMMEAS")
		if reply == "FAIL" || err != nil {
			err = dt.tcpdusttrakopen()
			if err != nil {
				log.Println("#3 GetDustTrakdata could not open port", err.Error())
				dt.DustTrakrunning = true
				return -1
			}
			reply, err = dt.tcpcommand("MSTATUS")
			if err != nil || reply != "Running" {
				reply, err = dt.tcpcommand("MSTART") //start Measuring
				if err != nil || reply != "OK" {
					log.Println("#4 GetDustTrakdata DastTrak could not start measuring!", err.Error())
					dt.DustTrakrunning = true
					return -1
				}
			}
		} else {
			if len(reply) > 2 {
				t = strings.Split(reply, ",")
				if len(t) > 2 {
					lastmeasurementnumber, err = strconv.Atoi(t[0])
					if err != nil {
						log.Println("#5 GetDustTrakdata error converting lastmeasurementnumber: ", reply, "Error:", err.Error())
					}
					if dt.DustTraklastin < lastmeasurementnumber {
						t = t[1:]            //remove samplenumber
						t = t[:(len(t) - 1)] // remove last sample (nothing after comma)
						for _, t2 := range t {
							f, err = strconv.ParseFloat(t2, 64)
							if err != nil {
								log.Println("#5 GetDustTrakdata error converting data: ", reply, "Error:", err.Error())
							}
							samples = append(samples, int32(f*1000))
						}
					}
					dt.DustTraklastin = lastmeasurementnumber
				}
			}
		}
		time.Sleep(time.Second / 10)
		retry++
	}
	dt.DustTrakrunning = true
	if len(samples) > 0 {
		return samples[0]
	} else {
		fmt.Print(" - DustTrak retry>10 ")
		return -1
	}
}
func (dt *DustTraktype) DustTrakstop() error {
	var err error
	var reply string
	if dt.DustTrakconn == nil {
		err = dt.tcpdusttrakopen()
		if err != nil {
			log.Println("#1 DustTrakstop -  could not open port", err.Error())
			return err
		}
	}
	reply, err = dt.tcpcommand("MSTOP") // stop measuring
	if err != nil {
		log.Println("#2 DustTrakstop - Error response on MSTOP->", err.Error())
	}
	if reply == "OK" {
		err = dt.DustTrakconn.Close()
		dt.DustTrakconn = nil
		return err
	} else {
		reply, err = dt.tcpcommand("MSTOP") // stop measuring
		if err != nil {
			log.Println("#3 DustTrakstop - DustTrakconn.Close: ", err.Error())
		}
		if reply == "OK" {
			err = dt.DustTrakconn.Close()
			dt.DustTrakconn = nil
			return err
		}
	}
	dt.DustTrakconn = nil
	return err
}
func SimulatedDustTrak() int32 {
	return rand.Int31n(1000)
}
func (dt *DustTraktype) GetDustTrakinfo() string {
	dt.tcpdusttrakopen()
	var t, reply string
	reply, _ = dt.tcpcommand("MSTOP") // stop measuring
	t = t + reply
	reply, _ = dt.tcpcommand("RDMN") //MODEL number
	t = t + reply
	reply, _ = dt.tcpcommand("RDSN") //serialnumber
	t = t + reply
	reply, _ = dt.tcpcommand("MSTATUS") // Current status
	t = t + reply
	reply, _ = dt.tcpcommand("RSDATETIME") //read datetime in instrument
	t = t + reply
	log.Println(t)
	//	dt.tcpdusttrakstop()
	return t
}

func (dt *DustTraktype) TestDustTraktest(T *testing.T) {
	var d int32
	for i := 0; i < 100; i++ {
		for d < 0 {
			time.Sleep(time.Second / 10)
			d = dt.GetDustTrakdata()
		}
		fmt.Println(d)
	}
	// conn := dt.tcpdusttrakopen()
	// t := dt.GetDustTrakdata()
	// t = t + dt.tcpcommand(conn, "MSTOP") // stop measuring
	// t = t + dt.readdusttrakdata(conn)
	// t = t + dt.readdusttrakdata(conn)
	// t = t + "RSDATETIME  \t> " + dt.tcpcommand(conn, "RSDATETIME")
	// t = t + g.tcpcommand(conn, "WSDATETIME "+dt.setdusttrakdatetime())
	// t = t + "RSDATETIME  \t> " + g.tcpcommand(conn, "RSDATETIME")

	// t = t + "RDMN        \t> " + dt.tcpcommand(conn, "RDMN")       //MODEL number
	// t = t + "RDSN        \t> " + dt.tcpcommand(conn, "RDSN")       //serialnumber
	// t = t + "MSTATUS     \t>" + dt.tcpcommand(conn, "MSTATUS")     // Current status
	// t = t + "RSDATETIME  \t> " + dt.tcpcommand(conn, "RSDATETIME") //read datetime in instrument
	// t = t + "RSDATETIME  \t> " + dt.tcpcommand(conn, "RSDATETIME")
	// t = t + "RSDATETIME  \t> " + dt.tcpcommand(conn, "RSDATETIME")
	// t = t + dt.tcpcommand(conn, "MSTART") //start Measuring
	// t = t + dt.tcpcommand(conn, "MSTOP")  // stop measuring
	// log.Println(t)
	// dt.setdusttrakdatetime()
	// conn.Close()
}

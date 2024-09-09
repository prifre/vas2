/*

2021-04-03 23:23
TCP routines to handle TSI AeroTrak
2021-04-10 got it to work.

*/

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

type AeroTraktype struct {
	simulateAeroTrak  bool
	AeroTrakrunning   bool
	AeroTrakclient    modbus.Client
	AeroTrakhandler   *modbus.TCPClientHandler
	AeroTraklastin    int
	AeroTraksetupdone bool
	AeroTrakport      string
	defdelay          int64
	showcmd           bool
	showdata          bool
}

func bytestoint32(b []byte) int32 {
	var r int32
	if len(b) == 4 {
		r = int32(uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24)
	}
	if len(b) == 2 {
		r = int32(uint32(b[1]) | uint32(b[0])<<8)
	}
	return r
}
func TSIdate(b []byte) string {
	y := int(b[0]) * 256
	y = y + int(b[1])
	m := int(b[3])
	d := int(b[5])
	return fmt.Sprintf("%4d-%02d-%02d", y, m, d)
}
func TSItime(b []byte) string {
	t := int(b[1])
	m := int(b[3])
	s := int(b[5])
	return fmt.Sprintf("%02d:%02d:%02d", t, m, s)
}
func TSIbyteorder(b []byte) []byte {
	for i := 0; i < len(b); i = i + 2 {
		b[i], b[i+1] = b[i+1], b[i]
	}
	return b
}
func setTSItime() []uint16 {
	t := time.Now()
	var rtc []uint16
	rtc = append(rtc, uint16(t.Year()))
	rtc = append(rtc, uint16(t.Month()))
	rtc = append(rtc, uint16(t.Day()))
	rtc = append(rtc, uint16(t.Hour()))
	rtc = append(rtc, uint16(t.Minute()))
	rtc = append(rtc, uint16(t.Second()))
	return rtc
}
func (at *AeroTraktype) modbusAeroTrakReadHoldingRegisters(adr uint16, count uint16) ([]byte, error) {
	var results []byte
	var err error
	tries := 100
	n := 100
	//	ShowMyDebug("Reading " + fmt.Sprintf("%v, %v", adr, at.AeroTrakclient))
	if at.AeroTrakclient == nil {
		log.Printf("AeroTrakclient nil!")
		return nil, nil
	}
	time.Sleep(time.Duration(at.defdelay)) //time.Sleep(time.Duration(n) * time.Millisecond)
	results, err = at.AeroTrakclient.ReadHoldingRegisters(adr, count)
	//	ShowMyDebug("Read " + fmt.Sprintf("%v, %v", adr, at.AeroTrakclient))
	if err != nil {
		//		log.Printf("error reading ReadHoldingRegisters %v", err.Error())
		return nil, err
	}
	for len(results) < int(count) && tries > 0 {
		if len(results) != int(count) {
			//	ShowMyDebug(fmt.Sprintf("Got too little data!, waiting... %v,%v", n, tries))
			time.Sleep(time.Duration(at.defdelay)) //time.Sleep(time.Duration(n) * time.Millisecond)
			//	ShowMyDebug("Waited!")
			n = n * 2
		}
	}
	return results, err
}
func (at *AeroTraktype) ModbusAeroTrakgetinfo() (string, error) {
	var err error = nil
	var results []byte
	var reply string = ""
	if at.AeroTrakclient == nil {
		err = at.modbusAeroTrakopen()
		if err != nil {
			log.Printf("#1 GetAeroTrakdata Failed to open AeroTrak port %v\n", err)
			return "", err
		}
	}
	err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{6}) //AeroTrak START
	if err != nil {
		log.Printf("#2 GetAeroTrakData AeroTrak START! failed:  %v\n", err)
	}
	// get devicestatus to check if started!!!!
	results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	if err != nil {
		log.Printf("#3 getdevicestatus1 %v\n", err)
		return "", err

	}
	if len(results) > 0 {
		for i := 0; i < 10; i++ {
			for results[1] != 3 {
				time.Sleep(time.Second / 10) //time.Sleep(1 * time.Second)
				results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
				if err != nil {
					log.Printf("#3 getdevicestatus2 %v\n", err)
				}
			}
		}
	}
	log.Printf("Device status: %v", results)
	// Device Info group 40001...
	results, err = at.modbusAeroTrakReadHoldingRegisters(0, 32)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters %v", err.Error())
		return "", err
	}
	reply = reply + fmt.Sprintf("Firmware version = %v\n", results[2])                       // Firmware version
	reply = reply + fmt.Sprintf("Model = %v\n", string(TSIbyteorder(results[4:20])))         // Model
	reply = reply + fmt.Sprintf("Serialnumber = %v\n", string(TSIbyteorder(results[20:36]))) // get Serialnumber
	// get DateofManufacture / Datelastcalibration / Date next calibration due
	results, err = at.modbusAeroTrakReadHoldingRegisters(40024-40001, 32)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(40024-40001, 32) %v ", err.Error())
		return "", err
	}
	reply = reply + fmt.Sprintf("Manufactureddate = %v\nLastcalibrationdate = %v\nNextcalibrationduedate = %v\n", TSIdate(results[0:6]), TSIdate(results[6:12]), TSIdate(results[12:18]))
	// Configuration group 41082...
	results, err = at.modbusAeroTrakReadHoldingRegisters(41082-40001, 20)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(40082-40001, 32) %v ", err.Error())
		return "", err
	}
	reply = reply + fmt.Sprintf("DHCP enabled %v\n", results[0:2])
	reply = reply + fmt.Sprintf("IP address %v\n", results[2:6])
	reply = reply + fmt.Sprintf("Subnet mask %v\n", results[6:10])
	reply = reply + fmt.Sprintf("Gateway %v\n", results[10:14])
	reply = reply + fmt.Sprintf("Modbus server port %v\n", results[14:16])
	reply = reply + fmt.Sprintf("Modbus unit address %v\n", results[16:18])
	for i := 18; i < 30; i = i + 2 {
		reply = reply + fmt.Sprintf("Cutoff points = %v\n", int(results[i])*256+int(results[i+1]))
	}
	// Sample data block 42001...
	results, err = at.modbusAeroTrakReadHoldingRegisters(42001-40001, 52)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(42001-40001, 52) %v", err.Error())
		return "", err
	}
	reply = reply + fmt.Sprintf("Sample Record Count = %v\n", bytestoint32(results[0:4]))
	reply = reply + fmt.Sprintf("Index of last record count = %v\n", bytestoint32(results[4:8]))
	reply = reply + fmt.Sprintf("RecordID of record being read %v\n", bytestoint32(results[8:12]))
	reply = reply + fmt.Sprintf("Record Timestamp %v %v\n", TSIdate(results[12:18]), TSItime(results[18:24]))
	reply = reply + fmt.Sprintf("Device Status = %v\n", results[35:37])
	reply = reply + fmt.Sprintf("Alarm Status = %v\n", results[37:39])
	reply = reply + fmt.Sprintf("Flow rate X100 = %v\n", results[39:41])
	reply = reply + fmt.Sprintf("Elapsed sample time %v\n", results[41:43])
	reply = reply + fmt.Sprintf("Elapsed sample time units %v (0=mSec, 1=sec, 2= 1/10 sec\n", results[6])
	reply = reply + fmt.Sprintf("Count mode %v (0=differential, 1=cumulative)\n", results[7])
	reply = reply + fmt.Sprintf("Count type = %v (0=Raw count, 1=FT3, 2=m3)\n", results[8])
	reply = reply + fmt.Sprintf("Location number = %v\n", results[9])
	results, err = at.modbusAeroTrakReadHoldingRegisters(41001-40001, 64)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(42001-40001, 64) %v", err.Error())
		return "", err
	}
	reply = reply + fmt.Sprintf("Device running state = %v  (1=stopped, 2=start delay, 3=holding, 4=sampling, other=not ready to send data)\n", results[3])
	reply = reply + fmt.Sprintf("Battery status = %v\n", results[5])
	reply = reply + fmt.Sprintf("Battery charge = %v\n", results[7])
	reply = reply + fmt.Sprintf("Realtimeflowstatus = %v (0=OK, 1=Error, 2=Stopped)\n", string(results[8:10]))
	reply = reply + fmt.Sprintf("AeroTrak Real-Time clock = %v %v\n", TSIdate(results[10:16]), TSItime(results[16:22]))
	reply = reply + fmt.Sprintf("Start Delay = %v\n", bytestoint32(results[22:26]))
	reply = reply + fmt.Sprintf("Hold Time = %v\n", bytestoint32(results[26:30]))
	reply = reply + fmt.Sprintf("Sample Time = %v seconds\n", bytestoint32(results[30:34]))
	reply = reply + fmt.Sprintf("Cycles Count = %v\n", int(results[28])*256+int(results[29]))
	reply = reply + fmt.Sprintf("Channel Data Enable = %v\n\n", results[30:31])
	at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{7}) //AeroTrak STOP
	results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	if err != nil {
		log.Printf("#3 getdevicestatus1 %v\n", err)
		return "", err
	}
	for i := 0; i < 10; i++ {
		for results[1] != 0 {
			time.Sleep(time.Duration(at.defdelay)) //time.Sleep(1 * time.Second)
			results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
			if err != nil {
				log.Printf("#3 getdevicestatus2 %v\n", err)
				return "", err
			}
		}
	}
	return reply, err
}
func (at *AeroTraktype) modbuswrite(client modbus.Client, adr uint16, v []uint16) error {
	var err error
	var r []byte
	var i uint16
	if client == nil {
		err = at.modbusAeroTrakopen()
		if err != nil {
			return err
		}
	}
	for i = 0; i < uint16(len(v)); i++ {
		if client == nil {
			return err
		}
		r, err = client.WriteSingleRegister(adr+i, v[i])
		if err != nil {
			log.Printf("#modbuswrite Error adr %v %v %v", (adr + 40001), r, err.Error())
		}
	}
	return err
}

func (at *AeroTraktype) modbusAeroTrakopen() error {
	var err error
	if at.AeroTrakclient != nil {
		return nil
	}
	for i := 0; i < 20; i++ {
		_, err = net.DialTimeout("tcp", at.AeroTrakport, time.Duration(time.Second))
		if err == nil {
			break
		}
	}
	if at.AeroTrakport == "" {
		log.Printf("No AeroTrak port set")
		return fmt.Errorf("no aerotrak port set")
	}
	at.AeroTrakhandler = modbus.NewTCPClientHandler(at.AeroTrakport)
	// at.AeroTrakhandler.Timeout = 1000
	//	handler.SlaveId = 0xFF
	//	handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)
	// Connect manually so that multiple requests are handled in one connection session
	err = at.AeroTrakhandler.Connect() // Connect manually so that multiple requests are handled in one session
	if err != nil {
		log.Printf("#1 AeroTrakopen error modbus.NewTCPClientHandler %v\n", err)
		return err
	}
	at.AeroTrakclient = modbus.NewClient(at.AeroTrakhandler)
	if err != nil {
		log.Printf("#2 AeroTrakopen error modbus.NewClient %v\n", err)
		return err
	}
	return err
}
func (at *AeroTraktype) GetAeroTrakdata() ([6]int32, error) {
	// Configuration group 41082...
	var data [6]int32 = [6]int32{-1, -1, -1, -1, -1, -1}
	var err error
	var results []byte
	var lastrecordindatabase int
	if at.AeroTrakclient == nil {
		err = at.Setup()
		if err != nil {
			log.Printf("#1 AeroTrakSetup failed %v", err.Error())
			at.AeroTrakclient = nil
			return data, err
		} else {
			at.AeroTraksetupdone = true
		}
	}
	results, err = at.modbusAeroTrakReadHoldingRegisters(42003-40001, 2)
	if err != nil {
		log.Printf("#4 Error reading last record: %v \n %v", results, err.Error())
		return data, err
	}
	lastrecordindatabase = int(bytestoint32(results[0:4]))
	if at.AeroTraklastin < lastrecordindatabase {
		//read the record
		results, err = at.modbusAeroTrakReadHoldingRegisters(42026-40001, 64)
		if err != nil {
			log.Printf("#5 getAeroTrakdata ReadHoldingregisters failed: %v, %v", results, err.Error())
			return data, err
		}
		for i := 0; i < 6; i = i + 1 {
			data[i] = bytestoint32(results[i*4 : i*4+4]) // convert 4 bytes to int32
		}
		at.AeroTraklastin = lastrecordindatabase
		at.afterread()
	} else {
		at.AeroTraklastin = 0
		at.beforeread()
	}
	if lastrecordindatabase > 9999 {
		err = at.AeroTrakstart()
		if err != nil {
			log.Printf("#2 GetAeroTrakData AeroTrak START! failed:  %v\n", err)
		}
	}
	return data, err
}
func (at *AeroTraktype) Setup() error {
	var err error
	var cmd string = g.app.Preferences().StringWithFallback("aerotrakcmd", at.setupaerotrakcode())
	//findsection "setup"
	// err = at.modbusAeroTrakopen()
	// if err != nil {
	// 	log.Println("#1 GetAeroTrakdata Failed to open AeroTrak port", err.Error())
	// 	return err
	// }
	if !strings.Contains(cmd, "setup:") {
		log.Print("Bad AeroTrak command program, 'setup:' missing")
		return nil
	}
	cmd = strings.TrimSpace(string(strings.Split(strings.Split(cmd, "setup:")[1], ":")[0]))
	err = at.aerotrakcode(cmd)
	if err != nil {
		return err
	}
	return err
}
func (at *AeroTraktype) beforeread() error {
	var err error
	var cmd string = g.app.Preferences().StringWithFallback("aerotrakcmd", at.setupaerotrakcode())
	//findsection "setup"
	if !strings.Contains(cmd, "beforeread:") {
		log.Print("Bad AeroTrak command program, 'beforeread:' missing")
		return nil
	}
	cmd = strings.TrimSpace(string(strings.Split(strings.Split(cmd, "beforeread:")[1], ":")[0]))
	err = at.aerotrakcode(cmd)
	if err != nil {
		return err
	}
	return err
}
func (at *AeroTraktype) afterread() error {
	var err error
	var cmd string = g.app.Preferences().StringWithFallback("aerotrakcmd", at.setupaerotrakcode())
	//findsection "setup"
	if !strings.Contains(cmd, "afterread:") {
		log.Print("Bad AeroTrak command program, 'afterread:' missing")
		return nil
	}
	cmd = strings.TrimSpace(string(strings.Split(strings.Split(cmd, "afterread:")[1], ":")[0]))
	err = at.aerotrakcode(cmd)
	if err != nil {
		return err
	}
	return err
}
func (at *AeroTraktype) AeroTrakstop() error {
	var err error
	// getproglines "stop"
	var cmd string = g.app.Preferences().StringWithFallback("aerotrakcmd", at.setupaerotrakcode())
	//findsection "stop"
	if !strings.Contains(cmd, "stop:") {
		log.Print("Bad AeroTrak command program, 'stop:' missing")
		return nil
	}
	cmd = strings.TrimSpace(string(strings.Split(strings.Split(cmd, "stop:")[1], ":")[0]))
	err = at.aerotrakcode(cmd)
	if err != nil {
		return err
	}
	at.AeroTrakclient = nil
	at.AeroTrakhandler = nil
	return err
}

func (at *AeroTraktype) AeroTrakstart() error {
	var err error
	var cmd string = g.app.Preferences().StringWithFallback("aerotrakcmd", at.setupaerotrakcode())
	//findsection "start"
	if !strings.Contains(cmd, "restart:") {
		log.Print("Bad AeroTrak command program, 'restart:' missing")
		return nil
	}
	cmd = strings.TrimSpace(string(strings.Split(strings.Split(cmd, "restart:")[1], ":")[0]))
	err = at.aerotrakcode(cmd)
	return err
}
func SimulatedAeroTrak() [6]int32 {
	return [6]int32{rand.Int31n(1000), rand.Int31n(500), rand.Int31n(100), rand.Int31n(50), rand.Int31n(10), rand.Int31n(5)}
}

func (at *AeroTraktype) getAeroTrakrecipes() ([]string, error) {
	var err error
	var r []string
	var results []byte
	var mydebug bool = g.app.Preferences().BoolWithFallback("mydebug", true)
	var numofrecipes, labellength, utf16support int32
	results, err = at.modbusAeroTrakReadHoldingRegisters(40065-40001, 2) // get num of recipes
	if mydebug && err != nil {
		fmt.Printf("#1 getAeroTrakrecipes:modbusAeroTrakReadHoldingRegisters(40065 %v\n", err)
	}
	numofrecipes = bytestoint32(results[0:2]) // convert 4 bytes to int32
	if numofrecipes < 1 {
		log.Printf("#2 getAeroTrakrecipes:Error reading 40065 NO RECIPE SUPPORT! %v", err)
		return nil, err
	}

	results, err = at.modbusAeroTrakReadHoldingRegisters(40066-40001, 2) // get labelsupport of recipes
	labellength = bytestoint32(results[0:2])                             // convert 4 bytes to int32
	if labellength < 1 {
		log.Printf("#3 getAeroTrakrecipes:Error reading 40066 NO RECIPE LABEL SUPPORT! %v", err)
		return nil, err
	}

	results, err = at.modbusAeroTrakReadHoldingRegisters(40035-40001, 2) // get UTF-16 of recipes
	utf16support = bytestoint32(results[0:2])                            // convert 4 bytes to int32
	if utf16support < 1 {
		log.Printf("#4 getAeroTrakrecipes:Error reading 40035 UTF-16 SUPPORT! %v", err)
		return nil, err
	}
	for i := 1; i < int(numofrecipes); i++ {
		err = at.modbuswrite(at.AeroTrakclient, 43018-40001, []uint16{uint16(i)}) // Select label 1
		if mydebug && err != nil {
			fmt.Printf("#5 getAeroTrakrecipes:Labelselect error %v\n", err)
		}

		results, err = at.modbusAeroTrakReadHoldingRegisters(43018+1-40001, 32) // get recipe labels
		if err != nil {
			log.Printf("#6 getAeroTrakrecipes:Error reading 43018 Recipe labels! %v", err)
		}
		s := string(results)
		s = strings.ReplaceAll(s, string(rune(0)), "")
		if s > "" {
			r = append(r, s)
		}
	}
	return r, err
}
func (at *AeroTraktype) getAeroTraklocations() ([]string, error) {
	var err error
	var r []string
	var results []byte
	var mydebug bool = g.app.Preferences().BoolWithFallback("mydebug", true)
	var numoflocations, labellength, utf16support int32
	results, err = at.modbusAeroTrakReadHoldingRegisters(40063-40001, 2) // get num of locations
	if mydebug && err != nil {
		fmt.Printf("#1 getAeroTraklocations:modbusAeroTrakReadHoldingRegisters(40063 %v\n", err)
	}
	numoflocations = bytestoint32(results[0:2]) // convert 4 bytes to int32
	fmt.Printf("\nSupported num of Locations: %v\n", numoflocations)
	if numoflocations < 1 {
		log.Printf("#2 getAeroTraklocations:Error reading 40063 NO LOCATION SUPPORT! %v", err)
		return nil, err
	}

	results, err = at.modbusAeroTrakReadHoldingRegisters(40064-40001, 2) // get labelsupport of recipes
	labellength = bytestoint32(results[0:2])                             // convert 4 bytes to int32
	if labellength < 1 {
		log.Printf("#3 getAeroTraklocations:Error reading 40064 NO RECIPE LABEL SUPPORT! %v", err)
		return nil, err
	}

	results, err = at.modbusAeroTrakReadHoldingRegisters(40035-40001, 2) // get UTF-16 of recipes
	utf16support = bytestoint32(results[0:2])                            // convert 4 bytes to int32
	if utf16support < 1 {
		log.Printf("#4 getAeroTraklocations:Error reading 40035 UTF-16 SUPPORT! %v", err)
		return nil, err
	}
	for i := 1; i < int(numoflocations); i++ {
		err = at.modbuswrite(at.AeroTrakclient, 43001-40001, []uint16{uint16(i)}) // Select label 1
		if mydebug && err != nil {
			fmt.Printf("#5 getAeroTrakrecipes:Labelselect error %v\n", err)
		}

		results, err = at.modbusAeroTrakReadHoldingRegisters(43001+1-40001, 32) // get recipe labels
		if err != nil {
			log.Printf("#6 getAeroTraklocations:Error reading 430101 Location labels! %v", err)
		}
		s := string(results)
		s = strings.ReplaceAll(s, string(rune(0)), "")
		s = strings.ReplaceAll(s, " ", "")
		if s > "" {
			r = append(r, s)
		}
	}
	return r, err
}
func (at *AeroTraktype) aerotrakcode(prog string) error {
	var cmd string
	var val int
	var proglines []string
	var err error = nil
	var reg *regexp.Regexp
	err = at.modbusAeroTrakopen()
	if err != nil {
		log.Printf("#1 cmdloop failed to open %v ", err.Error())
		return err
	}
	proglines = strings.Split(prog, "\n")
	for i := 0; i < len(proglines); i++ {
		cmd = string(proglines[i])
		if cmd == "" {
			return nil
		}
		if at.showcmd {
			fmt.Println("##########:", cmd)
		}
		if len(strings.Split(cmd, " ")) > 1 {
			v1 := strings.Split(cmd, " ")[1]
			reg, err = regexp.Compile("[^0-9]+")
			v1 = reg.ReplaceAllString(v1, "")
			val, _ = strconv.Atoi(v1)
			cmd = strings.Split(cmd, " ")[0]
		}
		cmd = strings.ToUpper(cmd)
		switch cmd {
		case "AEROTRAKSTOP":
			err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{7}) //AeroTrak STOP
		case "SHOWCMD":
			at.showcmd = (val == 1)
		case "SHOWDATA":
			at.showdata = (val == 1)
		case "AEROTRAKSTART":
			err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{6}) //AeroTrak START
		case "ENABLELOCALCONTROL":
			err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{13}) //ENABLE_LOCAL_CONTROL
		case "DISABLELOCALCONTROL":
			err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{12}) // DISABLE_LOCAL_CONTROL
		case "SETTSITIME":
			err = at.modbuswrite(at.AeroTrakclient, 41006-40001, setTSItime()) // Set AeroTrak Real-Time Clock!
		case "CMDSETRTC":
			err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{8}) // CMD_SET_RTC
		case "SAMPLINGMODE":
			err = at.modbuswrite(at.AeroTrakclient, 41055-40001, []uint16{uint16(val)}) // SamplingMode 0=Auto, 1 = Manual
		case "STARTDELAY":
			err = at.modbuswrite(at.AeroTrakclient, 41012-40001, []uint16{0, uint16(val)}) // Start Delay
		case "HOLDTIME":
			err = at.modbuswrite(at.AeroTrakclient, 41014-40001, []uint16{0, uint16(val)}) // Hold time
		case "CLEARDATARECORDS":
			err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{1}) // 1 = CMD_CLEAR ALL DATARECORDS
		case "SAMPLETIME":
			err = at.modbuswrite(at.AeroTrakclient, 41016-40001, []uint16{0, uint16(val)}) // Sample time
		case "FLOWUNIT":
			err = at.modbuswrite(at.AeroTrakclient, 41064-40001, []uint16{0, uint16(val)}) // Sample time
		case "CYCLES":
			err = at.modbuswrite(at.AeroTrakclient, 41018-40001, []uint16{uint16(val)}) // Cykles
		case "CHANNELDATAENABLE":
			err = at.modbuswrite(at.AeroTrakclient, 41020-40001, []uint16{uint16(val)}) // Channeldataenable (63  !!)
		case "CUTPOINTS":
			err = at.modbuswrite(at.AeroTrakclient, 41091-40001, []uint16{300, 500, 1000, 3000, 5000, 10000}) // CutPoints
		case "SETSAMPLINGCONFIG":
			err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9}) // 9 = CMD_SET_SAMPLING_CONFIG
		case "SLEEP":
			time.Sleep(time.Duration(val * int(time.Millisecond)))
		case "AEROTRAKHANDLER":
			at.AeroTrakhandler.Timeout = time.Duration(val)
		case "DIALOUT":
			for i := 0; i < val; i++ {
				_, err = net.DialTimeout("tcp", at.AeroTrakport, time.Duration(time.Second))
				if err == nil {
					break
				}
			}
		case "READHOLDINGREGISTERS":
			var results []byte
			results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
			log.Println("READHOLDINGREGISTERS: ", results)
		case "SHOWINFO":
			log.Println(at.ModbusAeroTrakgetinfo())
		case "SHOWRECIPES":
			var results []string
			results, err = at.getAeroTrakrecipes()
			log.Printf("RECIPES: \n %v", results)
		case "SHOWLOCATIONS":
			var results []string
			results, err = at.getAeroTraklocations()
			log.Printf("LOCATIONS: \n%v", results)
		case "QUIT":
			os.Exit(0)
		}
	}
	return err
}

func (at *AeroTraktype) setupaerotrakcode() string {
	var s string
	s = ""
	s += "setup:\n"
	//	s += "AEROTRAKSTOP\n"
	s += "SHOWCMD 0 // special to show or hide command in text window 1=yes, 0=no\n"
	s += "SHOWDATA 1 // special to show or hide measurement data in text window 1=yes, 0=no\n"
	s += "AEROTRAKHANDLER 0 // Timeout setting for communication protocol. 0 is waiting mode!\n"
	s += "#SHOWINFO // special command to dump some device info\n"
	s += "#SHOWRECIPES // special command to dump recipes info\n"
	s += "#SHOWLOCATIONS // special command to dump locations info\n"
	s += "AEROTRAKSTART // Start AeroTrak measuring\n"
	s += "AEROTRAKSTOP // special command to dump locations info\n"
	s += "SETSAMPLINGCONFIG // Must be sent to activate configuration.\n"
	s += "DISABLELOCALCONTROL // makes AeroTrak buttons computercontrolled only\n"
	s += "SAMPLINGMODE 0 // 0=Auto, 1=Manual\n"
	s += "SETTSITIME // calculates time to set based on computer time\n"
	s += "CMDSETRTC // sets Real-Time-Clock on AeroTrak to calculated time\n"
	s += "#ENABLELOCALCONTROL // makes AeroTrak buttons user controlled again\n"
	s += "STARTDELAY 10 // Start Delay\n"
	s += "HOLDTIME 0 // Hold time between sampling\n"
	s += "CLEARDATARECORDS // removes all saved data from AeroTrak\n"
	s += "SAMPLETIME 1 // 1 second interval is default and fastest possible\n"
	s += "CYCLES 9999 // 9999 is default. After 9999 RESTART is done automatically.\n"
	s += "FLOWUNIT 8 // 8 is default (m3/h).\n"
	s += "CHANNELDATAENABLE 63 // 63 (bits for all 6 channels) is default.\n"
	s += "CUTPOINTS // this is automatically set to 300,500,1000,3000,5000,10000 if used.\n"
	s += "SETSAMPLINGCONFIG // Must be sent to activate configuration.\n"
	s += "AEROTRAKSTART // Start AeroTrak measuring\n"
	s += "SLEEP 100 // if AeroTrak is found to not cope, it is possible to insert SLEEP between commands. Time in milliseconds.\n"
	s += "\n"
	s += "stop:\n"
	s += "ENABLELOCALCONTROL // makes AeroTrak buttons user controlled again\n"
	s += "SAMPLINGMODE 1 // 0=Auto, 1=Manual\n"
	s += "AEROTRAKSTOP // Stop AeroTrak measuring\n"
	s += "\n"
	s += "restart:\n"
	s += "AEROTRAKSTOP // Stop AeroTrak measuring\n"
	s += "CLEARDATARECORDS // Remove all datarecords\n"
	s += "AEROTRAKSTART // Start AeroTrak measuring\n"
	s += "\n"
	s += "beforeread:\n"
	s += "SETSAMPLINGCONFIG // Must be sent to activate configuration.\n"
	s += "AEROTRAKSTART // Start AeroTrak measuring\n"
	s += "\n"
	s += "afterread:\n"
	s += "AEROTRAKSTART\n"
	return s
}

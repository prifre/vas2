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
	"testing"
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
	cycles            int
}

func bytestoint32(b []byte) int32 {
	var r = int32(uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24)
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
	results, err = at.AeroTrakclient.ReadHoldingRegisters(adr, count)
	//	ShowMyDebug("Read " + fmt.Sprintf("%v, %v", adr, at.AeroTrakclient))
	if err != nil {
		log.Println("error reading ReadHoldingRegisters", err)
	}
	for len(results) < int(count) && tries > 0 {
		if len(results) != int(count) {
			//	ShowMyDebug(fmt.Sprintf("Got too little data!, waiting... %v,%v", n, tries))
			time.Sleep(time.Duration(n) * time.Millisecond)
			//	ShowMyDebug("Waited!")
			n = n * 2
		}
	}
	return results, err
}
func (at *AeroTraktype) ModbusAeroTrakgetinfo() string {
	var err error
	var results []byte
	var reply string = ""

	// Device Info group 40001...
	results, err = at.modbusAeroTrakReadHoldingRegisters(0, 32)
	if err != nil {
		log.Println(err.Error())
	}
	reply = reply + fmt.Sprintf("Firmware version = %v\n", results[2])                       // Firmware version
	reply = reply + fmt.Sprintf("Model = %v\n", string(TSIbyteorder(results[4:20])))         // Model
	reply = reply + fmt.Sprintf("Serialnumber = %v\n", string(TSIbyteorder(results[20:36]))) // get Serialnumber

	// get DateofManufacture / Datelastcalibration / Date next calibration due
	results, err = at.modbusAeroTrakReadHoldingRegisters(40024-40001, 32)
	if err != nil {
		log.Println(err.Error())
	}
	reply = reply + fmt.Sprintf("Manufactureddate = %v\nLastcalibrationdate = %v\nNextcalibrationduedate = %v\n", TSIdate(results[0:6]), TSIdate(results[6:12]), TSIdate(results[12:18]))

	// Configuration group 41082...
	results, err = at.modbusAeroTrakReadHoldingRegisters(41082-40001, 20)
	if err != nil {
		log.Println(err.Error())
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
		log.Println(err.Error())
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
		log.Println(err.Error())
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
	return reply
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
		r, err = client.WriteSingleRegister(adr+i, v[i])
		if err != nil {
			log.Println("#modbuswrite Error adr ", (adr + 40001), r, err.Error())
		}
	}
	return err
}
func (at *AeroTraktype) modbusAeroTraksetup() {
	at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{7}) //AeroTrak STOP
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41006-40001, setTSItime()) // Set AeroTrak Real-Time Clock!
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{8}) // CMD_SET_RTC
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41012-40001, []uint16{0, 0}) // Start Delay
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41014-40001, []uint16{0, 0}) // Hold time
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41016-40001, []uint16{0, 1}) // Sample time
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41018-40001, []uint16{255 * 255}) // Cykles
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{12}) // Enable Modbus & user config
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9}) // CMD_SET_SAMPLING_CONFIG
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41091-40001, []uint16{300, 500, 1000, 3000, 5000, 10000}) // CutPoints
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41055-40001, []uint16{0}) // SamplingMode 0=Auto
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41020-40001, []uint16{63}) // Channeldataenable
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{1}) // 1 = CMD_CLEAR ALL DATARECORDS
	time.Sleep(time.Second / 100)
	at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9}) // 9 = CMD_SET_SAMPLING_CONFIG
	time.Sleep(time.Second / 100)

	// at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{1})  // 1 = CMD_CLEAR ALL DATARECORDS
	// at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{7})  //AeroTrak STOP
	// at.modbuswrite(at.AeroTrakclient, 41006-40001, setTSItime()) // Set AeroTrak Real-Time Clock!
	// at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9})  // SET AUTOMODE CONFIG
	// at.modbuswrite(at.AeroTrakclient, 41080-40001, []uint16{7})
	// at.modbuswrite(at.AeroTrakclient, 41091-40001, []uint16{300, 500, 1000, 3000, 5000, 10000}) // CutPoints
	// at.modbuswrite(at.AeroTrakclient, 41020-40001, []uint16{63})                                // Channel data enable
	// at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9})                                 // SET AUTOMODE CONFIG
	// at.modbuswrite(at.AeroTrakclient, 41018-40001, []uint16{0})                                 // Cykles
	// at.modbuswrite(at.AeroTrakclient, 41019-40001, []uint16{998})                               //
	// at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9})                                 // SET AUTOMODE CONFIG
	// at.modbuswrite(at.AeroTrakclient, 41016-40001, []uint16{0})                                 //
	// at.modbuswrite(at.AeroTrakclient, 41017-40001, []uint16{5})                                 //
	// at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9})                                 // SET AUTOMODE CONFIG

	// //	modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{8})                                 // CMD_SET_RTC
	// //	modbuswrite(at.AeroTrakclient, 41012-40001, []uint16{0, 0}) // Start Delay
	// //	modbuswrite(at.AeroTrakclient, 41014-40001, []uint16{0, 0}) // Hold time
	// //	modbuswrite(at.AeroTrakclient, 41016-40001, []uint16{0, 1}) // Sample time
	// //	modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9})                                 // SET AUTOMODE CONFIG
	// //	modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9})                                 // 9 = CMD_SET_SAMPLING_CONFIG
	// //	modbuswrite(at.AeroTrakclient, 41055-40001, []uint16{0}) // SamplingMode
	// //	modbuswrite(at.AeroTrakclient, 41020-40001, []uint16{63})                                // Channel data enable
	// //	modbuswrite(at.AeroTrakclient, 41020-40001, []uint16{63}) // Channeldataenable
	// //	modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{1})  // 1 = CMD_CLEAR ALL DATARECORDS
	// //	modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{9})  // 9 = CMD_SET_SAMPLING_CONFIG
}
func (at *AeroTraktype) modbusAeroTrakopen() error {
	var err error
	if at.AeroTrakport == "" {
		log.Println("No AeroTrak port set")
		return fmt.Errorf("no aerotrak port set")
	}
	at.AeroTrakhandler = modbus.NewTCPClientHandler(at.AeroTrakport)
	// AeroTrakhandler.Timeout = 20 * time.Millisecond
	//	handler.SlaveId = 0xFF
	//	handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)
	// Connect manually so that multiple requests are handled in one connection session
	err = at.AeroTrakhandler.Connect() // Connect manually so that multiple requests are handled in one session
	if err != nil {
		log.Println("#1 AeroTrakopen error modbus.NewTCPClientHandler", err.Error())
		return err
	}
	at.AeroTrakclient = modbus.NewClient(at.AeroTrakhandler)
	if err != nil {
		log.Println("#2 AeroTrakopen error modbus.NewClient", err.Error())
		return err
	}
	return err
}
func (at *AeroTraktype) GetAeroTrakdata() [6]int32 {
	// Configuration group 41082...
	var data [6]int32 = [6]int32{-1, -1, -1, -1, -1, -1}
	var err error
	var results []byte
	if at.AeroTrakclient == nil {
		err = at.modbusAeroTrakopen()
		if err != nil {
			log.Println("#1 GetAeroTrakdata Failed to open AeroTrak port", err.Error())
			return data
		}
		if !at.AeroTraksetupdone {
			//			ShowMyDebug("AeroTrak setup started")
			time.Sleep(time.Second / 10)
			at.modbusAeroTraksetup()
			time.Sleep(time.Second / 10)
			time.Sleep(time.Second / 10)
			at.AeroTraklastin = 0
			//			ShowMyDebug("AeroTrak setup done")
			if at.AeroTrakclient == nil {
				log.Println("#1 getAeroTrakdata no modbus.NewClient", err.Error())
				return data
			}
			at.AeroTraksetupdone = true
		}
		err = at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{6}) //AeroTrak START
		if err != nil {
			log.Println("#GetAeroTrakData Modbus AeroTrak Start failed: ", err.Error())
		}
	}
	//42001 SampleRecordCountHigh R Number of records in database, high word
	//42002 SampleRecordCountLow R Number of records in database, low word
	// read records in database:
	// results, err = at.modbusAeroTrakReadHoldingRegisters(42001-40001, 2)
	// if err != nil {
	// 	log.Println("#1 Error reading records in database", err)
	// 	return data
	// }
	// recordsindatabase := int(bytestoint32(results[0:4]))
	// if recordsindatabase < 1 {
	// 	return data
	// }

	//42003 SampleRecordIndexHigh R Index of last recorded record, high word
	//42004 SampleRecordIndexLow R Index of last recorded record, high word
	results, err = at.modbusAeroTrakReadHoldingRegisters(42003-40001, 2)
	if err != nil {
		log.Println("#1 Error reading last record in database", err)
		return data
	}
	lastrecordindatabase := int(bytestoint32(results[0:4]))
	//	fmt.Println("Checking last record...:", lastrecordindatabase, at.AeroTraklastin)
	if at.AeroTraklastin < lastrecordindatabase {
		//read the record
		results, err = at.modbusAeroTrakReadHoldingRegisters(42026-40001, 64)
		if err != nil {
			log.Println("#3 getAeroTrakdata ReadHoldingregisters failed: ", results, err.Error())
		}
		at.AeroTraklastin = lastrecordindatabase
		for i := 0; i < 6; i = i + 1 {
			data[i] = bytestoint32(results[i*4 : i*4+4]) // convert 4 bytes to int32
		}
		at.cycles = 0
	}
	at.cycles++
	if at.cycles > 10 && at.AeroTraklastin == lastrecordindatabase {
		fmt.Print(" - AeroTrak setup...")
		at.AeroTrakstop()
		time.Sleep(time.Second)
		at.modbusAeroTrakopen()
		time.Sleep(time.Second)
		at.modbusAeroTraksetup()
		at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{6}) //AeroTrak START
		at.AeroTraklastin = 0
		at.cycles = 0
	}
	return data
}
func (at *AeroTraktype) AeroTrakstop() error {
	var err error
	at.AeroTraksetupdone = false
	err = at.modbusAeroTrakopen()
	if err != nil {
		log.Println("#1 AeroTrakstop failed to open to stop")
	} else {
		at.modbuswrite(at.AeroTrakclient, 41001-40001, []uint16{7}) //AeroTrak STOP
	}
	at.AeroTrakclient = nil
	at.AeroTrakhandler = nil
	return err
}
func SimulatedAeroTrak() [6]int32 {
	return [6]int32{rand.Int31n(1000), rand.Int31n(500), rand.Int31n(100), rand.Int31n(50), rand.Int31n(10), rand.Int31n(5)}
}

func (at *AeroTraktype) TestAeroTrak(T *testing.T) {
	if at.modbusAeroTrakopen() != nil {
		fmt.Println("error opening AeroTrakport")
		T.Fail()
		return
	}
	c0 := at.AeroTrakclient
	at.modbuswrite(c0, 41001-40001, []uint16{7}) //AeroTrak STOP
	at.modbusAeroTraksetup()
	log.Println(at.ModbusAeroTrakgetinfo())
	at.modbuswrite(c0, 41001-40001, []uint16{6}) //AeroTrak START
	for r := 0; r < 4; r++ {
		fmt.Println(at.GetAeroTrakdata())
	}
	time.Sleep(5 * time.Second)
}

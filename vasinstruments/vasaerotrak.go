package vasinstruments

/*

2021-04-03 23:23
TCP routines to handle TSI AeroTrak
2021-04-10 got it to work.

*/

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"github.com/goburrow/modbus"
)

type AeroTraktype struct {
	SimulateAeroTrak  bool
	AeroTrakrunning   bool
	AeroTrakclient    modbus.Client
	AeroTrakhandler   *modbus.TCPClientHandler
	AeroTraklastin    int
	AeroTraksetupdone bool
	debugAeroTrak     bool
	AeroTrakport      string
	Defdelay          int64
	showcmd           bool
	Showdata          bool
}

func (at *AeroTraktype) GetAeroTrakdata() ([6]int32, error) {
	var err error
	defaultData := [6]int32{-1, -1, -1, -1, -1, -1}

	// 1. Säkerställ anslutning
	if at.AeroTrakclient == nil {
		err = at.AeroTrakSetup()
		if err != nil {
			log.Printf("#1 AeroTrakSetup failed: %v", err)
			return defaultData, err
		}
		at.AeroTraksetupdone = true
	}

	// 2. Läs senast registrerade index i databasen
	results, err := at.modbusAeroTrakReadHoldingRegisters(42003-40001, 2)
	if at.debugAeroTrak {
		log.Printf("#2 GetAeroTrakdata ReadHoldingRegisters results: %v, err: %v", results, err)
	}
	if err != nil || len(results) < 4 {
		log.Printf("#4 Error reading last record: %v, err: %v", results, err)
		return defaultData, err
	}

	lastrecordindatabase := int(bytestoint32(results[0:4]))

	// 3. Om det finns ett nytt rekord -> läs ut det
	if at.AeroTraklastin < lastrecordindatabase {
		results, err = at.modbusAeroTrakReadHoldingRegisters(42026-40001, 64)
		if err != nil || len(results) < 24 { // 6 st int32 kräver minst 24 bytes
			log.Printf("#5 getAeroTrakdata ReadHoldingregisters failed: %v, err: %v", results, err)
			return defaultData, err
		}

		var data [6]int32
		for i := 0; i < 6; i++ {
			data[i] = bytestoint32(results[i*4 : i*4+4])
		}

		at.AeroTraklastin = lastrecordindatabase
		at.afterread()

		// Starta om mätaren om databasindexet blivit för högt
		if lastrecordindatabase > 9999 {
			if startErr := at.AeroTrakStart(); startErr != nil {
				log.Printf("#2 GetAeroTrakData AeroTrak START! failed: %v", startErr)
			}
		}

		return data, nil
	}

	// 4. Inget nytt data fanns
	at.beforeread()

	if lastrecordindatabase > 9999 {
		if startErr := at.AeroTrakStart(); startErr != nil {
			log.Printf("#2 GetAeroTrakData AeroTrak START! failed: %v", startErr)
		}
	}

	return defaultData, nil
}
func (at *AeroTraktype) AeroTrakSetup() error {
	var err error
	var pref = fyne.CurrentApp().Preferences()
	at.AeroTrakport = pref.StringWithFallback("AeroTrak", "")
	var cmd string = pref.StringWithFallback("aerotrakcmd", at.Setupaerotrakcode())
	//findsection "setup"
	// err = at.modbusAeroTrakopen()
	// if err != nil {
	// 	log.Println("#1 GetAeroTrakdata Failed to open AeroTrak port", err.Error())
	// 	return err
	// }
	if !strings.Contains(cmd, "setup:") {
		log.Print("Bad AeroTrak command program, 'setup:' missing")
		return errors.New("Bad AeroTrak command program, 'setup:' missing")
	}
	cmd = strings.TrimSpace(string(strings.Split(strings.Split(cmd, "setup:")[1], ":")[0]))
	err = at.aerotrakcode(cmd)
	return err
}
func (at *AeroTraktype) AeroTrakStop() error {
	var err error
	// getproglines "stop"
	var cmd string = fyne.CurrentApp().Preferences().StringWithFallback("aerotrakcmd", at.Setupaerotrakcode())
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
func (at *AeroTraktype) AeroTrakStart() error {
	var err error
	var pref = fyne.CurrentApp().Preferences()
	var cmd string = pref.StringWithFallback("aerotrakcmd", at.Setupaerotrakcode())
	at.AeroTrakport = pref.StringWithFallback("AeroTrak", "")
	//findsection "start"
	if !strings.Contains(cmd, "restart:") {
		log.Print("Bad AeroTrak command program, 'restart:' missing")
		return nil
	}
	cmd = strings.TrimSpace(string(strings.Split(strings.Split(cmd, "restart:")[1], ":")[0]))
	err = at.aerotrakcode(cmd)
	return err
}
func (at *AeroTraktype) aerotrakcode(prog string) error {
	var err error

	if err = at.modbusAeroTrakopen(); err != nil {
		log.Printf("#1 cmdloop failed to open: %v", err)
		return err
	}

	proglines := strings.Split(prog, "\n")
	var reg = regexp.MustCompile(`[^0-9]+`)
	for _, rawLine := range proglines {
		// 1. Ta bort kommentarer (både // och #) samt trimma mellanslag
		line := rawLine
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)

		// Hoppa över tomma rader istället för att göra 'return'
		if line == "" {
			continue
		}

		if at.showcmd {
			fmt.Println("##########:", line)
		}

		// 2. Dela upp kommando och eventuellt värde
		parts := strings.Fields(line)
		cmd := strings.ToUpper(parts[0])
		val := 0

		if len(parts) > 1 {
			vStr := reg.ReplaceAllString(parts[1], "")
			if vStr != "" {
				val, _ = strconv.Atoi(vStr)
			}
		}

		// 3. Exekvera kommando
		if at.debugAeroTrak {
			log.Printf("Executing command: %s with value: %d", cmd, val)
		}
		switch cmd {
		case "AEROTRAKSTOP":
			err = at.modbuswrite(41001-40001, []uint16{7})
		case "SHOWCMD":
			at.showcmd = (val == 1)
		case "SHOWDATA":
			at.Showdata = (val == 1)
		case "AEROTRAKSTART":
			err = at.modbuswrite(41001-40001, []uint16{6})
		case "ENABLELOCALCONTROL":
			err = at.modbuswrite(41001-40001, []uint16{13})
		case "DISABLELOCALCONTROL":
			err = at.modbuswrite(41001-40001, []uint16{12})
		case "SETTSITIME":
			err = at.modbuswrite(41006-40001, setTSItime())
		case "CMDSETRTC":
			err = at.modbuswrite(41001-40001, []uint16{8})
		case "SAMPLINGMODE":
			err = at.modbuswrite(41055-40001, []uint16{uint16(val)})
		case "STARTDELAY":
			err = at.modbuswrite(41012-40001, []uint16{0, uint16(val)})
		case "HOLDTIME":
			err = at.modbuswrite(41014-40001, []uint16{0, uint16(val)})
		case "CLEARDATARECORDS":
			err = at.modbuswrite(41001-40001, []uint16{1})
		case "SAMPLETIME":
			err = at.modbuswrite(41016-40001, []uint16{0, uint16(val)})
		case "FLOWUNIT":
			err = at.modbuswrite(41064-40001, []uint16{0, uint16(val)})
		case "CYCLES":
			err = at.modbuswrite(41018-40001, []uint16{uint16(val)})
		case "CHANNELDATAENABLE":
			err = at.modbuswrite(41020-40001, []uint16{uint16(val)})
		case "CUTPOINTS":
			err = at.modbuswrite(41091-40001, []uint16{300, 500, 1000, 3000, 5000, 10000})
		case "SETSAMPLINGCONFIG":
			err = at.modbuswrite(41001-40001, []uint16{9})
		case "SLEEP":
			time.Sleep(time.Duration(val) * time.Millisecond)
		case "AEROTRAKHANDLER":
			if at.AeroTrakhandler != nil {
				at.AeroTrakhandler.Timeout = time.Duration(val) * time.Millisecond
			}
		case "READHOLDINGREGISTERS":
			var results []byte
			results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
			log.Println("READHOLDINGREGISTERS: ", results)
		case "SHOWINFO":
			info, _ := ModbusAeroTrakgetinfo(at.AeroTrakport)
			log.Println(info)
		case "SHOWRECIPES":
			results, _ := at.GetAeroTrakrecipes()
			log.Printf("RECIPES: \n %v", results)
		case "SHOWLOCATIONS":
			results, _ := at.getAeroTraklocations()
			log.Printf("LOCATIONS: \n%v", results)
		case "QUIT":
			os.Exit(0)
		}

		if err != nil {
			log.Printf("Error executing command '%s': %v", cmd, err)
		}
	}

	return err
}

package vasinstruments

import (
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

func ModbusAeroTrakgetinfo(port string) (string, error) {
	var err error
	var results []byte
	var reply string

	at := new(AeroTraktype)
	at.AeroTrakport = port

	if err = at.modbusAeroTrakopen(); err != nil {
		log.Printf("#1 GetAeroTrakdata Failed to open AeroTrak port %v\n", err)
		return "", err
	}

	// Stäng anslutningen snyggt när funktionen är klar och nollställ pekare
	defer func() {
		if at.AeroTrakhandler != nil {
			_ = at.AeroTrakhandler.Close()
			at.AeroTrakhandler = nil
		}
		at.AeroTrakclient = nil
	}()

	// AeroTrak START
	if err = at.modbuswrite(41001-40001, []uint16{6}); err != nil {
		log.Printf("#2 GetAeroTrakData AeroTrak START! failed: %v\n", err)
	}

	// Vänta säkert på att status blir 3 (holding/sampling)
	for retries := 0; retries < 20; retries++ {
		results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
		if err == nil && len(results) >= 2 && results[1] == 3 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Device Info group (32 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(0, 32)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters (0, 32) %v", err)
		return "", err
	}

	// Ändrat från >= 36 till >= 32
	if len(results) >= 32 {
		reply += fmt.Sprintf("Firmware version = %v\n", results[2])
		reply += fmt.Sprintf("Model = %v\n", string(TSIbyteorder(results[4:20])))
		reply += fmt.Sprintf("Serialnumber = %v\n", string(TSIbyteorder(results[20:32])))
	}

	// Manufacture / Calibration dates
	results, err = at.modbusAeroTrakReadHoldingRegisters(40024-40001, 32)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(40024-40001, 32) %v", err)
		return "", err
	}
	if len(results) >= 18 {
		reply += fmt.Sprintf("Manufactureddate = %v\nLastcalibrationdate = %v\nNextcalibrationduedate = %v\n",
			TSIdate(results[0:6]), TSIdate(results[6:12]), TSIdate(results[12:18]))
	}

	// Configuration group (30 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(41082-40001, 30)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(41082-40001, 30) %v", err)
		return "", err
	}
	if len(results) >= 18 {
		reply += fmt.Sprintf("DHCP enabled %v\n", results[0:2])
		reply += fmt.Sprintf("IP address %v\n", results[2:6])
		reply += fmt.Sprintf("Subnet mask %v\n", results[6:10])
		reply += fmt.Sprintf("Gateway %v\n", results[10:14])
		reply += fmt.Sprintf("Modbus server port %v\n", results[14:16])
		reply += fmt.Sprintf("Modbus unit address %v\n", results[16:18])
	}
	if len(results) >= 30 {
		for i := 18; i < 30; i += 2 {
			reply += fmt.Sprintf("Cutoff points = %v\n", int(results[i])*256+int(results[i+1]))
		}
	}

	// Sample data block 42001...
	results, err = at.modbusAeroTrakReadHoldingRegisters(42001-40001, 52)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(42001-40001, 52) %v", err)
		return "", err
	}
	if len(results) >= 43 {
		reply += fmt.Sprintf("Sample Record Count = %v\n", bytestoint32(results[0:4]))
		reply += fmt.Sprintf("Index of last record count = %v\n", bytestoint32(results[4:8]))
		reply += fmt.Sprintf("RecordID of record being read = %v\n", bytestoint32(results[8:12]))
		reply += fmt.Sprintf("Record Timestamp = %v %v\n", TSIdate(results[12:18]), TSItime(results[18:24]))
		reply += fmt.Sprintf("Device Status = %v\n", results[35:37])
		reply += fmt.Sprintf("Alarm Status = %v\n", results[37:39])
		reply += fmt.Sprintf("Flow rate X100 = %v\n", results[39:41])
		reply += fmt.Sprintf("Elapsed sample time = %v\n", results[41:43])
	}

	// Status and RTC settings
	results, err = at.modbusAeroTrakReadHoldingRegisters(41001-40001, 64)
	if err != nil {
		log.Printf("modbusAeroTrakReadHoldingRegisters(41001-40001, 64) %v", err)
		return "", err
	}
	if len(results) >= 34 {
		reply += fmt.Sprintf("Device running state = %v\n", results[3])
		reply += fmt.Sprintf("Battery status = %v\n", results[5])
		reply += fmt.Sprintf("Battery charge = %v\n", results[7])
		reply += fmt.Sprintf("Realtimeflowstatus = %v\n", string(results[8:10]))
		reply += fmt.Sprintf("AeroTrak Real-Time clock = %v %v\n", TSIdate(results[10:16]), TSItime(results[16:22]))
		reply += fmt.Sprintf("Start Delay = %v\n", bytestoint32(results[22:26]))
		reply += fmt.Sprintf("Hold Time = %v\n", bytestoint32(results[26:30]))
		reply += fmt.Sprintf("Sample Time = %v seconds\n", bytestoint32(results[30:34]))
	}

	// AeroTrak STOP
	_ = at.modbuswrite(41001-40001, []uint16{7})

	// Vänta tills enheten bekräftat STOP (status 0 eller 1)
	for retries := 0; retries < 20; retries++ {
		results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
		if err == nil && len(results) >= 2 && (results[1] == 0 || results[1] == 1) {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return reply, nil
}
func (at *AeroTraktype) GetAeroTrakrecipes() ([]string, error) {
	var err error
	var r []string
	var results []byte

	// Säker kontroll för Fyne preferences
	mydebug := true
	if fyne.CurrentApp() != nil {
		mydebug = fyne.CurrentApp().Preferences().BoolWithFallback("mydebug", true)
	}

	// 1. Hämta antal recept (Läs 2 register = 4 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(40065-40001, 2)
	if err != nil || len(results) < 4 {
		if mydebug {
			log.Printf("#1 getAeroTrakrecipes: Error reading 40065 (numofrecipes): %v", err)
		}
		return nil, err
	}
	numofrecipes := bytestoint32(results[0:4])
	if numofrecipes < 1 {
		log.Printf("#2 getAeroTrakrecipes: NO RECIPE SUPPORT or 0 recipes found (count: %d)", numofrecipes)
		return nil, fmt.Errorf("no recipe support or 0 recipes")
	}

	// 2. Hämta recept-etikettstöd (Läs 2 register = 4 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(40066-40001, 2)
	if err != nil || len(results) < 4 {
		log.Printf("#3 getAeroTrakrecipes: Error reading 40066 (labellength): %v", err)
		return nil, err
	}
	labellength := bytestoint32(results[0:4])
	if labellength < 1 {
		log.Printf("#4 getAeroTrakrecipes: NO RECIPE LABEL SUPPORT!")
		return nil, fmt.Errorf("no recipe label support")
	}

	// 3. Hämta UTF-16 stöd (Läs 2 register = 4 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(40035-40001, 2)
	if err != nil || len(results) < 4 {
		log.Printf("#5 getAeroTrakrecipes: Error reading 40035 (utf16support): %v", err)
		return nil, err
	}
	utf16support := bytestoint32(results[0:4])
	if utf16support < 1 {
		log.Printf("#6 getAeroTrakrecipes: NO UTF-16 SUPPORT!")
		return nil, fmt.Errorf("no utf-16 support")
	}

	// 4. Loopa igenom och hämta receptnamnen (0 till numofrecipes-1)
	for i := 0; i < int(numofrecipes); i++ {
		// Välj recept-index (utan at.AeroTrakclient som argument)
		err = at.modbuswrite(43018-40001, []uint16{uint16(i)})
		if err != nil {
			if mydebug {
				log.Printf("#7 getAeroTrakrecipes: Labelselect error for index %d: %v", i, err)
			}
			continue
		}

		// Läs receptetiketten (16 register = 32 bytes)
		results, err = at.modbusAeroTrakReadHoldingRegisters(43019-40001, 16)
		if err != nil {
			log.Printf("#8 getAeroTrakrecipes: Error reading Recipe label %d: %v", i, err)
			continue
		}

		// Rensa noll-bytes (null terminators) och städa strängen
		s := string(results)
		s = strings.ReplaceAll(s, string(rune(0)), "")
		s = strings.TrimSpace(s)

		if len(s) > 0 {
			r = append(r, s)
		}
	}

	return r, nil
}
func (at *AeroTraktype) getAeroTraklocations() ([]string, error) {
	var err error
	var r []string
	var results []byte

	// Säker kontroll för Fyne preferences
	mydebug := true
	if fyne.CurrentApp() != nil {
		mydebug = fyne.CurrentApp().Preferences().BoolWithFallback("mydebug", true)
	}

	// 1. Hämta antal platser/locations (Läs 2 register = 4 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(40063-40001, 2)
	if err != nil || len(results) < 4 {
		if mydebug {
			log.Printf("#1 getAeroTraklocations: Error reading 40063 (numoflocations): %v", err)
		}
		return nil, err
	}
	numoflocations := bytestoint32(results[0:4])
	if mydebug {
		log.Printf("Supported num of Locations: %d", numoflocations)
	}
	if numoflocations < 1 {
		log.Printf("#2 getAeroTraklocations: NO LOCATION SUPPORT or 0 locations found")
		return nil, fmt.Errorf("no location support or 0 locations")
	}

	// 2. Hämta etikettlängd (Läs 2 register = 4 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(40064-40001, 2)
	if err != nil || len(results) < 4 {
		log.Printf("#3 getAeroTraklocations: Error reading 40064 (labellength): %v", err)
		return nil, err
	}
	labellength := bytestoint32(results[0:4])
	if labellength < 1 {
		log.Printf("#4 getAeroTraklocations: NO LOCATION LABEL SUPPORT!")
		return nil, fmt.Errorf("no location label support")
	}

	// 3. Hämta UTF-16 stöd (Läs 2 register = 4 bytes)
	results, err = at.modbusAeroTrakReadHoldingRegisters(40035-40001, 2)
	if err != nil || len(results) < 4 {
		log.Printf("#5 getAeroTraklocations: Error reading 40035 (utf16support): %v", err)
		return nil, err
	}
	utf16support := bytestoint32(results[0:4])
	if utf16support < 1 {
		log.Printf("#6 getAeroTraklocations: NO UTF-16 SUPPORT!")
		return nil, fmt.Errorf("no utf-16 support")
	}

	// 4. Loopa igenom alla platser (0 till numoflocations - 1)
	for i := 0; i < int(numoflocations); i++ {
		// Välj plats-index (utan at.AeroTrakclient som argument)
		err = at.modbuswrite(43001-40001, []uint16{uint16(i)})
		if err != nil {
			if mydebug {
				log.Printf("#7 getAeroTraklocations: Labelselect error for index %d: %v", i, err)
			}
			continue
		}

		// Läs platsens etikett (16 register = 32 bytes)
		results, err = at.modbusAeroTrakReadHoldingRegisters(43002-40001, 16)
		if err != nil {
			log.Printf("#8 getAeroTraklocations: Error reading Location label %d: %v", i, err)
			continue
		}

		// Rensa noll-bytes och städa strängen
		s := string(results)
		s = strings.ReplaceAll(s, string(rune(0)), "")
		s = strings.TrimSpace(s)

		if len(s) > 0 {
			r = append(r, s)
		}
	}

	return r, nil
}

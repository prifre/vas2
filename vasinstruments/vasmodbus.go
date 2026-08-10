package vasinstruments

import (
	"encoding/binary"
	"fmt"
	"log"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"github.com/goburrow/modbus"
)

func (at *AeroTraktype) Setupaerotrakcode() string {
	return `setup:
SHOWCMD 0 // special to show or hide command in text window 1=yes, 0=no
SHOWDATA 1 // special to show or hide measurement data in text window 1=yes, 0=no
AEROTRAKHANDLER 0 // Timeout setting for communication protocol. 0 is waiting mode!
#SHOWINFO // special command to dump some device info
#SHOWRECIPES // special command to dump recipes info
#SHOWLOCATIONS // special command to dump locations info
AEROTRAKSTART // Start AeroTrak measuring
AEROTRAKSTOP // Stop AeroTrak measuring
SETSAMPLINGCONFIG // Must be sent to activate configuration.
DISABLELOCALCONTROL // makes AeroTrak buttons computercontrolled only
SAMPLINGMODE 0 // 0=Auto, 1=Manual
SETTSITIME // calculates time to set based on computer time
CMDSETRTC // sets Real-Time-Clock on AeroTrak to calculated time
#ENABLELOCALCONTROL // makes AeroTrak buttons user controlled again
STARTDELAY 10 // Start Delay
HOLDTIME 0 // Hold time between sampling
CLEARDATARECORDS // removes all saved data from AeroTrak
SAMPLETIME 1 // 1 second interval is default and fastest possible
CYCLES 9999 // 9999 is default. After 9999 RESTART is done automatically.
FLOWUNIT 8 // 8 is default (m3/h).
CHANNELDATAENABLE 63 // 63 (bits for all 6 channels) is default.
CUTPOINTS // this is automatically set to 300,500,1000,3000,5000,10000 if used.
SETSAMPLINGCONFIG // Must be sent to activate configuration.
AEROTRAKSTART // Start AeroTrak measuring
SLEEP 100 // Time in milliseconds.

stop:
ENABLELOCALCONTROL // makes AeroTrak buttons user controlled again
SAMPLINGMODE 1 // 0=Auto, 1=Manual
AEROTRAKSTOP // Stop AeroTrak measuring

restart:
AEROTRAKSTOP // Stop AeroTrak measuring
CLEARDATARECORDS // Remove all datarecords
AEROTRAKSTART // Start AeroTrak measuring

beforeread:
SETSAMPLINGCONFIG // Must be sent to activate configuration.
AEROTRAKSTART // Start AeroTrak measuring

afterread:
AEROTRAKSTART
`
}

// Hjälpfunktion för att säkert plocka ut en sektion ur skriptsträngen
func (at *AeroTraktype) getScriptSection(section string) string {
	cmd := at.Setupaerotrakcode()
	if fyne.CurrentApp() != nil {
		cmd = fyne.CurrentApp().Preferences().StringWithFallback("aerotrakcmd", at.Setupaerotrakcode())
	}

	if !strings.Contains(cmd, section) {
		log.Printf("Bad AeroTrak command program, '%s' missing", section)
		return ""
	}

	parts := strings.Split(cmd, section)
	if len(parts) < 2 {
		return ""
	}

	body := parts[1]

	// Sluta läsa om vi stöter på en ny sektions-etikett (t.ex. "stop:", "afterread:")
	lines := strings.Split(body, "\n")
	var sectionLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Om raden slutar med ':' och inte är en kommentar, har vi nått nästa sektion
		if strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") {
			break
		}
		sectionLines = append(sectionLines, line)
	}

	return strings.TrimSpace(strings.Join(sectionLines, "\n"))
}

func (at *AeroTraktype) beforeread() error {
	section := at.getScriptSection("beforeread:")
	if section == "" {
		return nil
	}
	return at.aerotrakcode(section)
}

func (at *AeroTraktype) afterread() error {
	section := at.getScriptSection("afterread:")
	if section == "" {
		return nil
	}
	return at.aerotrakcode(section)
}

func (at *AeroTraktype) modbuswrite(adr uint16, v []uint16) error {
	var err error
	if len(v) == 0 {
		return nil
	}

	// Säkerställ anslutning
	if at.AeroTrakclient == nil {
		if err := at.modbusAeroTrakopen(); err != nil {
			return fmt.Errorf("modbuswrite failed to open connection: %w", err)
		}
	}

	// Om vi bara har 1 register -> WriteSingleRegister
	// Om vi har flera register -> WriteMultipleRegisters (mycket snabbare!)
	if len(v) == 1 {
		_, err = at.AeroTrakclient.WriteSingleRegister(adr, v[0])
	} else {
		// Omvandla []uint16 till byte-slice (Modbus kräver BigEndian / 2 bytes per register)
		payload := make([]byte, len(v)*2)
		for i, val := range v {
			binary.BigEndian.PutUint16(payload[i*2:], val)
		}
		_, err = at.AeroTrakclient.WriteMultipleRegisters(adr, uint16(len(v)), payload)
	}

	// Felhantering vid skrivfel
	if err != nil {
		log.Printf("#modbuswrite Error adr %d (Holding %d, len %d): %v", adr, adr+40001, len(v), err)
		at.closeModbusConnection()
		return err
	}

	return nil
}

// Hjälpmetod för att städa anslutningen säkert vid fel
func (at *AeroTraktype) closeModbusConnection() {
	if at.AeroTrakhandler != nil {
		_ = at.AeroTrakhandler.Close()
		at.AeroTrakhandler = nil
	}
	at.AeroTrakclient = nil
}

func (at *AeroTraktype) modbusAeroTrakopen() error {
	if at.AeroTrakclient != nil {
		return nil
	}

	if at.AeroTrakport == "" {
		log.Printf("No AeroTrak port set in preferences")
		return fmt.Errorf("No AeroTrak port set")
	}

	at.AeroTrakhandler = modbus.NewTCPClientHandler(at.AeroTrakport)
	at.AeroTrakhandler.Timeout = 2 * time.Second
	at.AeroTrakhandler.SlaveId = 0xFF

	err := at.AeroTrakhandler.Connect()
	if err != nil {
		log.Printf("#1 AeroTrakopen error Connect: %v\n", err)
		at.AeroTrakhandler = nil
		return err
	}

	at.AeroTrakclient = modbus.NewClient(at.AeroTrakhandler)
	if at.AeroTrakclient == nil {
		log.Printf("#2 AeroTrakopen error modbus.NewClient: failed to create client\n")
		return fmt.Errorf("failed to create modbus client")
	}

	return nil
}

func (at *AeroTraktype) modbusAeroTrakReadHoldingRegisters(adr uint16, count uint16) ([]byte, error) {
	// 1. Validera indata enligt Modbus-standarden (max 125 register per anrop)
	if count == 0 || count > 125 {
		return nil, fmt.Errorf("invalid Modbus count %d (must be 1-125)", count)
	}

	// 2. Säkerställ anslutning och skydda mot nil pointer
	if at.AeroTrakclient == nil {
		if err := at.modbusAeroTrakopen(); err != nil || at.AeroTrakclient == nil {
			log.Printf("AeroTrakclient is nil and could not open: %v", err)
			if err == nil {
				err = fmt.Errorf("client remains nil after open attempt")
			}
			return nil, err
		}
	}

	// 3. Valfri fördröjning för buss-stabilisering
	if at.Defdelay > 0 {
		time.Sleep(time.Duration(at.Defdelay) * time.Millisecond)
	}

	// 4. Utför läsningen
	results, err := at.AeroTrakclient.ReadHoldingRegisters(adr, count)
	if err != nil {
		log.Printf("Error reading ReadHoldingRegisters (adr: %d, count: %d): %v", adr, count, err)
		at.closeModbusConnection() // Rensa anslutningsstatus säkert
		return nil, err
	}

	return results, nil
}

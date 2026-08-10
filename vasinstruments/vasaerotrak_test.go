package vasinstruments

/*
TCP routines to handle TSI AeroTrak - Unit Tests
*/

import (
	"fmt"
	"log"
	"testing"
	"time"
)

const theaerotrakport = "192.168.0.171:502"

func TestAeroTrak(t *testing.T) {
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport

	fmt.Println("at.modbusAeroTrakopen")
	if err := at.modbusAeroTrakopen(); err != nil {
		fmt.Println("error opening AeroTrakport:", err)
		t.FailNow()
	}
	fmt.Println("at.modbusAeroTrakopen done")

	fmt.Println("AeroTrak STOP")
	at.modbuswrite(41001-40001, []uint16{7}) // AeroTrak STOP
	fmt.Println("AeroTrak STOP done")

	fmt.Println("AeroTrak SETUP")
	_ = at.AeroTrakSetup()

	fmt.Println("AeroTrak START")
	at.modbuswrite(41001-40001, []uint16{6}) // AeroTrak START

	fmt.Print("AeroTrak devicerunningstate: ")
	results, err := at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	fmt.Println(results, err, "AeroTrak devicerunningstate done")

	// Säker loop-kontroll så att inte testerna kraschar om results är tom
	for len(results) >= 2 && results[1] != 3 {
		fmt.Println("time.second sleep")
		time.Sleep(time.Second)
		fmt.Print("AeroTrak devicerunningstate: ")
		results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
		fmt.Println(results, err, "AeroTrak devicerunningstate")
	}

	fmt.Println("GetAeroTrakdata 3 times")
	for r := 1; r < 4; r++ {
		fmt.Println("GetAeroTrakdata")
		d, err := at.GetAeroTrakdata()
		if err != nil {
			log.Print(err.Error())
		}
		for d[0] < 0 {
			fmt.Println("GetAeroTrakdata waiting")
			time.Sleep(500 * time.Millisecond)
			d, err = at.GetAeroTrakdata()
			if err != nil {
				log.Print(err.Error())
			}
		}
		fmt.Println("AeroTrak Data: ", d)
	}
	fmt.Println("GetAeroTrakdata 3 times done")

	fmt.Println("AeroTrak STOP")
	at.modbuswrite(41001-40001, []uint16{7}) // AeroTrak STOP
	fmt.Print("AeroTrak devicerunningstate: ")
	results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	fmt.Println(results, err, "AeroTrak devicerunningstate done")

	for len(results) >= 2 && results[1] != 0 {
		fmt.Print("AeroTrak devicerunningstate: ")
		results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
		fmt.Println(results, err, "AeroTrak devicerunningstate")
	}
}

func TestAeroTrakSTOP(t *testing.T) {
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport

	fmt.Println("at.modbusAeroTrakopen")
	if err := at.modbusAeroTrakopen(); err != nil {
		fmt.Println("error opening AeroTrakport:", err)
		t.FailNow()
	}
	fmt.Println("at.modbusAeroTrakopen done")

	fmt.Println("AeroTrak STOP")
	at.modbuswrite(41001-40001, []uint16{7}) // AeroTrak STOP
	fmt.Println("AeroTrak STOP done")

	fmt.Println("5 second sleep")
	time.Sleep(5 * time.Second)
	fmt.Println("AeroTrak STOPPED??")
}

func TestAeroTrakSTART(t *testing.T) {
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport

	fmt.Println("at.modbusAeroTrakopen")
	if err := at.modbusAeroTrakopen(); err != nil {
		fmt.Println("error opening AeroTrakport:", err)
		t.FailNow()
	}
	fmt.Println("at.modbusAeroTrakopen done")

	fmt.Println("AeroTrak setup")
	_ = at.AeroTrakSetup()
	fmt.Println("AeroTrak SETUP done")

	fmt.Println("AeroTrak START")
	at.modbuswrite(41001-40001, []uint16{6}) // AeroTrak START
	fmt.Println("AeroTrak START done")

	fmt.Println("5 second sleep")
	time.Sleep(5 * time.Second)
	fmt.Println("STARTED????")
}

func TestAeroTrakmodbusinfo(t *testing.T) {
	info, err := ModbusAeroTrakgetinfo(theaerotrakport)
	if err != nil {
		t.Logf("Error getting info: %v", err)
	}
	fmt.Println(info)
	fmt.Println("-----------------------------")
}

func TestAeroTrakValues(t *testing.T) {
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport

	if err := at.modbusAeroTrakopen(); err != nil {
		fmt.Println("error opening AeroTrakport:", err)
		t.FailNow()
	}

	_ = at.AeroTrakSetup()
	at.modbuswrite(41001-40001, []uint16{6}) // AeroTrak START

	results, _ := at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	for len(results) >= 2 && results[1] != 3 {
		time.Sleep(time.Second)
		fmt.Print("AeroTrak devicerunningstate: ", results)
		results, _ = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	}

	n := 5
	fmt.Println("GetAeroTrakdata ", n, " times")
	for r := 1; r <= n; r++ {
		d, _ := at.GetAeroTrakdata()
		for d[0] < 0 {
			time.Sleep(200 * time.Millisecond)
			d, _ = at.GetAeroTrakdata()
		}
		fmt.Println("AeroTrak Data ", r, ": ", d)
	}
	fmt.Println("GetAeroTrakdata ", n, " times done")

	at.modbuswrite(41001-40001, []uint16{7}) // AeroTrak STOP
}

func TestAeroTrakMethod(t *testing.T) {
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport

	if err := at.modbusAeroTrakopen(); err != nil {
		fmt.Println("error opening AeroTrakport:", err)
		t.FailNow()
	}

	at.modbuswrite(41001-40001, []uint16{7}) // AeroTrak STOP
	_ = at.AeroTrakSetup()

	info, _ := ModbusAeroTrakgetinfo(theaerotrakport)
	log.Println(info)

	at.modbuswrite(41001-40001, []uint16{6}) // AeroTrak START
	for r := 0; r < 4; r++ {
		fmt.Println(at.GetAeroTrakdata())
	}
	time.Sleep(5 * time.Second)
}

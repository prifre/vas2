package vasinstruments

/*

2021-04-03 23:23
TCP routines to handle TSI AeroTrak
2021-04-10 got it to work.

*/

import (
	"fmt"
	"log"
	"testing"
	"time"
)

const theaerotrakport = "192.168.0.171:502"

func TestAeroTrak(T *testing.T) {
	// Testing function of AeroTrak
	var err error
	var results []byte
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport
	fmt.Println("at.modbusAeroTrakopen")
	if at.modbusAeroTrakopen() != nil {
		fmt.Println("error opening AeroTrakport")
		T.Fail()
		return
	}
	fmt.Println("at.modbusAeroTrakopen done")
	c0 := at.AeroTrakclient
	fmt.Println("AeroTrak STOP")
	at.modbuswrite(c0, 41001-40001, []uint16{7}) //AeroTrak STOP
	fmt.Println("AeroTrak STOP done")
	fmt.Println("AeroTrak SETUP")
	at.Setup()
	fmt.Println("AeroTrak START")
	at.modbuswrite(c0, 41001-40001, []uint16{6}) //AeroTrak START
	fmt.Print("AeroTrak devicerunningstate: ")
	results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	fmt.Println(results, err, "AeroTrak devicerunningstate done")
	for results[1] != 3 {
		fmt.Println("time.second sleep")
		time.Sleep(time.Second)
		fmt.Print("AeroTrak devicerunningstate: ")
		results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
		fmt.Println(results, err, "AeroTrak devicerunningstate")
	}
	fmt.Println("GetAeroTrakdata 3 times")
	for r := 1; r < 4; r++ {
		fmt.Println("GetAeroTrakdata")
		d,err := at.GetAeroTrakdata()
		if err!=nil {
			log.Print(err.Error)
		}
		for d[0] < 0 {
			fmt.Println("GetAeroTrakdata waiting")
			d,err = at.GetAeroTrakdata()
			if err!=nil {
				log.Print(err.Error)
			}
			}
		fmt.Println("AeroTrak Data: ", d)
	}
	fmt.Println("GetAeroTrakdata 3 times done")
	fmt.Println("AeroTrak STOP")
	at.modbuswrite(c0, 41001-40001, []uint16{7}) //AeroTrak STOP
	fmt.Print("AeroTrak devicerunningstate: ")
	results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	fmt.Println(results, err, "AeroTrak devicerunningstate done")
	for results[1] != 0 {
		fmt.Print("AeroTrak devicerunningstate: ")
		results, err = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
		fmt.Println(results, err, "AeroTrak devicerunningstate")
	}
}
func TestAeroTrakSTOP(T *testing.T) {
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport
	fmt.Println("at.modbusAeroTrakopen")
	if at.modbusAeroTrakopen() != nil {
		fmt.Println("error opening AeroTrakport")
		T.Fail()
		return
	}
	fmt.Println("at.modbusAeroTrakopen done")
	c0 := at.AeroTrakclient
	fmt.Println("AeroTrak STOP")
	at.modbuswrite(c0, 41001-40001, []uint16{7}) //AeroTrak STOP
	fmt.Println("AeroTrak STOP done")
	fmt.Println("5 second sleep")
	time.Sleep(5 * time.Second)
	fmt.Println("AeroTrak STOPPED??")
}

func TestAeroTrakSTART(T *testing.T) {
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport
	fmt.Println("at.modbusAeroTrakopen")
	if at.modbusAeroTrakopen() != nil {
		fmt.Println("error opening AeroTrakport")
		T.Fail()
		return
	}
	fmt.Println("at.modbusAeroTrakopen done")
	c0 := at.AeroTrakclient
	fmt.Println("AeroTrak setup")
	at.Setup()
	fmt.Println("AeroTrak SETUP done")
	fmt.Println("AeroTrak START")
	at.modbuswrite(c0, 41001-40001, []uint16{6}) //AeroTrak START
	fmt.Println("AeroTrak START done")
	fmt.Println("3 second sleep")
	time.Sleep(5 * time.Second)
	fmt.Println("STARTED????")
}

func TestAeroTrakmodbusinfo(T *testing.T) {
	fmt.Println(ModbusAeroTrakgetinfo(theaerotrakport))
	fmt.Println("-----------------------------")
}

func TestAeroTrakValues(T *testing.T) {
	var results []byte
	at := new(AeroTraktype)
	at.AeroTrakport = theaerotrakport
	if at.modbusAeroTrakopen() != nil {
		fmt.Println("error opening AeroTrakport")
		T.Fail()
		return
	}
	c0 := at.AeroTrakclient
	at.Setup()
	at.modbuswrite(c0, 41001-40001, []uint16{6}) //AeroTrak START
	results, _ = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	for results[1] != 3 {
		time.Sleep(time.Second)
		fmt.Print("AeroTrak devicerunningstate: ", results)
		results, _ = at.modbusAeroTrakReadHoldingRegisters(41002-40001, 1)
	}
	n := 5
	fmt.Println("GetAeroTrakdata ", n, " times")
	for r := 1; r <= n; r++ { // get 10 values!
		// fmt.Println("GetAeroTrakdata")
		d,_ := at.GetAeroTrakdata()
		for d[0] < 0 {
			// fmt.Println("GetAeroTrakdata waiting")
			d,_ = at.GetAeroTrakdata()
		}
		fmt.Println("AeroTrak Data ", r, ": ", d)
	}
	fmt.Println("GetAeroTrakdata ", n, " times done")
	// fmt.Println("AeroTrak STOP")
	at.modbuswrite(c0, 41001-40001, []uint16{7}) //AeroTrak STOP
}

func (at *AeroTraktype) TestAeroTrak(T *testing.T) {
	if at.modbusAeroTrakopen() != nil {
		fmt.Println("error opening AeroTrakport")
		T.Fail()
		return
	}
	c0 := at.AeroTrakclient
	at.modbuswrite(c0, 41001-40001, []uint16{7}) //AeroTrak STOP
	at.Setup()
	log.Println(ModbusAeroTrakgetinfo(theaerotrakport))
	at.modbuswrite(c0, 41001-40001, []uint16{6}) //AeroTrak START
	for r := 0; r < 4; r++ {
		fmt.Println(at.GetAeroTrakdata())
	}
	time.Sleep(5 * time.Second)
}

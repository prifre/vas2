package vasinstruments

import (
	"fmt"
	"testing"
	"time"
)

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

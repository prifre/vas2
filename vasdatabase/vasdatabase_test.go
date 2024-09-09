package vasdatabase

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestCreateTables(t *testing.T) {
	t.Log("Running database CreateTables test")
	db := new(DBtype)
	db.Setupdb()
	db.mdata[0] = 123
	db.nanostamp = 1621802710870044900
	db.Createtables()
	db.Addmeasurement()
	// delete database
	//create new database}

}

func TestAddMeasurement(t *testing.T) {
	var sq string
	var n1 []string
	var err error
	t.Log("AddMeasurement test")
	db := new(DBtype)
	db.Setupdb()
	db.nanostamp = 1621802710870044900
	db.tstamp = time.Now().Format(time.RFC3339)
	db.mname = "Measurement" + db.tstamp
	db.mname = strings.Replace(db.mname, "-", "", 99)
	db.mname = strings.Replace(db.mname, "/", "", 99)
	db.mname = strings.Replace(db.mname, ":", "", 99)
	db.mname = strings.Replace(db.mname, ".", "", 99)
	db.mname = db.mname[:26]
	db.mdata = [8]int32{-1, -1, -1, -1, -1, -1, -1, -1}
	db.Addmeasurement()
	for i := 0; i < len(db.mdata); i++ {
		db.mdata[i] = 123 // PTrak = 0
	}
	db.Addmeasurement()
	sq = fmt.Sprintf("SELECT COUNT(*) AS mname FROM tblMain WHERE nanostamp=%v ", db.nanostamp)
	n1, err = db.Getsql(sq)
	if n1[0] != "1" {
		t.Error("Mname failed:", n1, " != ", db.mname, err)
		return
	}
	sq = fmt.Sprintf("SELECT mdata FROM tblPTrak WHERE nanostamp=%v ", db.nanostamp)
	n1, err = db.Getsql(sq)
	if n1[0] != "123" {
		t.Error("PTrak mdata failed:", n1, err)
	}
}

func Testexporttotext(t *testing.T) {
	t.Log("Test Exporttotext")
	db := new(DBtype)
	db.Setupdb()
	db.Exporttotext()
}

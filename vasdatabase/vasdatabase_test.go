package vasdatabase

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2/app"
)

func TestCreateTables(t *testing.T) {
	t.Log("Running database CreateTables test")
	db := new(dbtype)
	db.Setupdb()
	db.mdata[0] = 123
	db.nanostamp = 1621802710870044900
	db.CreateTables()
	db.AddMeasurement()
	// delete database
	//create new database}

}

func TestAddMeasurement(t *testing.T) {
	var sq string
	var n1 []string
	var err error
	t.Log("AddMeasurement test")
	db := new(dbtype)
	g := new(game)
	g.d = new(dbtype)
	g.d = db
	g.app = app.NewWithID("particles.in.air")
	db.Setupdb(g)
	db.nanostamp = 1621802710870044900
	db.tstamp = time.Now().Format(time.RFC3339)
	db.mname = "Measurement" + g.d.tstamp
	db.mname = strings.Replace(g.d.mname, "-", "", 99)
	db.mname = strings.Replace(g.d.mname, "/", "", 99)
	db.mname = strings.Replace(g.d.mname, ":", "", 99)
	db.mname = strings.Replace(g.d.mname, ".", "", 99)
	db.mname = g.d.mname[:26]
	db.mdata = [8]int32{-1, -1, -1, -1, -1, -1, -1, -1}
	db.AddMeasurement()
	for i := 0; i < len(db.mdata); i++ {
		db.mdata[i] = 123 // PTrak = 0
	}
	db.AddMeasurement()
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

func TestExporttotext(t *testing.T) {
	t.Log("Test Exporttotext")
	db := new(dbtype)
	g := new(game)
	g.d = new(dbtype)
	g.d = db
	g.app = app.NewWithID("particles.in.air")
	g.window = g.app.NewWindow("ax")
	db.Setupdb(g)
	db.Exporttotext(g)
}

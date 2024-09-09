package vasdatabase

import (
	"testing"
)

func TestFixtime(t *testing.T) {
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

func TestExporttotext(t *testing.T) {
	t.Log("Test Exporttotext")
	db := new(DBtype)
	db.Setupdb()
	db.Exporttotext()
}

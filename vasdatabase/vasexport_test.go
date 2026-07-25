package vasdatabase

import (
	"testing"
)

func TestExportonetotext(t *testing.T) {
	t.Log("Test Exportonetotext")
	db := new(DBtype)
	db.SetupDatabase()
	db.Exportonetotext()
}

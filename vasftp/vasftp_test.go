package vasftp

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
	"vas2/general"

	"fyne.io/fyne/v2/app"
)

func TestDoftp(t *testing.T) {
	//  Doscreenshot()
	app.NewWithID("vas2")

	// func Doftp(fn,ftpserver,ftpuser,ftppassword,ftpdir string) string {
	// Skapa dummy-filer för tabellerna du vill testa
	tables := []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}
	for _, tbl := range tables {
		_ = CreateTestFile(tbl)
	}

	// Kör sedan din FTP-uppladdning!
	for _, tbl := range tables {
		fn := tbl + ".txt"
		err := Doftp(fn, "ftpcluster.loopia.se", "vas000", "fwzIn4oz!", ".")
		if err != nil {
			fmt.Println(err.Error())
			t.Fail()
		}
	}
}

// CreateTestFile skapar en textfil med en tidsstämpel för test
func CreateTestFile(tableName string) error {
	filePath := filepath.Join(general.GetHomeDir(), tableName+".txt")

	// Innehållet i testfilen
	content := fmt.Sprintf("Testdata för %s\nSkapad: %s\n1;2;3;4;5",
		tableName,
		time.Now().Format("2006-01-02 15:04:05"))

	// Skriv filen (0644 ger vanliga läs/skriv-rättigheter)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("kunde inte skapa testfil %s: %w", filePath, err)
	}

	fmt.Println("Testfil skapad:", filePath)
	return nil
}

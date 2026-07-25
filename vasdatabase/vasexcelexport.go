package vasdatabase

import (
	"fmt"
	"log"
	"strconv"
	"time"
	"vas2/vascharts"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func Fixtime(t string) string {
	parsedtime, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return t // Returnera originalsträngen om parsningsfelet uppstår
	}
	s := fmt.Sprintf("%v", parsedtime)
	if len(s) > 19 {
		s = s[11:19]
	}
	return s
}

func (db *DBtype) Exporttoexcel(nanostamp string, fn string) {
	var mname, n1, t1 []string
	var err error

	// Validera nanostamp så det faktiskt är ett nummer
	nanoInt, err := strconv.ParseInt(nanostamp, 10, 64)
	if err != nil {
		log.Println("#Exporttoexcel Invalid nanostamp:", nanostamp)
		return
	}

	xlsx := excelize.NewFile()

	// 1. Hämta huvudinfo säkert
	mname, _ = db.Getsql(fmt.Sprintf("SELECT mname FROM tblMain WHERE nanostamp=%d", nanoInt))
	n1, _ = db.Getsql(fmt.Sprintf("SELECT note FROM tblMain WHERE nanostamp=%d", nanoInt))
	t1, _ = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblMain WHERE nanostamp=%d", nanoInt))

	// Bygg rubriksträng säkert utan krasch
	headerTstamp := ""
	headerMname := ""
	headerNote := ""

	if len(t1) > 0 {
		headerTstamp = t1[0]
	}
	if len(mname) > 0 {
		headerMname = mname[0]
	}
	if len(n1) > 0 {
		headerNote = n1[0]
	}

	xlsx.SetCellValue("Sheet1", "A1", fmt.Sprintf("%s  %s, %s", headerTstamp, headerMname, headerNote))

	// 2. PTrak
	t1, err = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblPTrak WHERE nanostamp=%d", nanoInt))
	if err == nil && len(t1) > 0 {
		d1, _ := db.Getsql(fmt.Sprintf("SELECT mdata FROM tblPTrak WHERE nanostamp=%d", nanoInt))

		xlsx.SetCellValue("Sheet1", "B2", "PTrak")
		xlsx.SetCellValue("Sheet1", "C2", vascharts.ChartTitles[0])

		for i := 0; i < len(d1) && i < len(t1); i++ {
			v, _ := strconv.Atoi(d1[i])
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+3), Fixtime(t1[i]))
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+3), v)
		}
	}

	// 3. DustTrak
	t1, err = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblDustTrak WHERE nanostamp=%d", nanoInt))
	if err == nil && len(t1) > 0 {
		d1, _ := db.Getsql(fmt.Sprintf("SELECT mdata FROM tblDustTrak WHERE nanostamp=%d", nanoInt))

		xlsx.SetCellValue("Sheet1", "D2", "DustTrak")
		xlsx.SetCellValue("Sheet1", "E2", vascharts.ChartTitles[1])

		for i := 0; i < len(d1) && i < len(t1); i++ {
			v, _ := strconv.Atoi(d1[i])
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+3), Fixtime(t1[i]))
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+3), v)
		}
	}

	// 4. AeroTrak (6 kanaler)
	t1, err = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblAeroTrak WHERE nanostamp=%d", nanoInt))
	if err == nil && len(t1) > 0 {
		xlsx.SetCellValue("Sheet1", "F2", "AeroTrak")

		// Sätt tidsstämplarna i F-kolumnen en gång
		for i := 0; i < len(t1); i++ {
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+3), Fixtime(t1[i]))
		}

		// Hämta och sätt kanaler G-L
		colLetters := []string{"G", "H", "I", "J", "K", "L"}
		for c := 0; c < 6; c++ {
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s2", colLetters[c]), vascharts.ChartTitles[2+c])

			d1, err := db.Getsql(fmt.Sprintf("SELECT ch%d FROM tblAeroTrak WHERE nanostamp=%d", c+1, nanoInt))
			if err == nil {
				for i := 0; i < len(d1); i++ {
					v, _ := strconv.Atoi(d1[i])
					xlsx.SetCellValue("Sheet1", fmt.Sprintf("%s%d", colLetters[c], i+3), v)
				}
			}
		}
	}

	// Spara Excel-filen
	if err := xlsx.SaveAs(fn); err != nil {
		log.Println("#Exporttoexcel SaveAs Error:", err.Error())
	}
}

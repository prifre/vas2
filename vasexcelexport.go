package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

func fixtime(t string) string {
	var parsedtime time.Time
	//	t = strings.Replace(t, "T", " ", 1)
	//	t = t[:19]
	parsedtime, _ = time.Parse(time.RFC3339, t)
	s := fmt.Sprintf("%v", parsedtime)
	if len(s) > 19 {
		s = s[11:19]
	}
	return s
}

func (db *dbtype) exporttoexcel(nanostamp string, fn string) {
	var mname, n1, d1, t1 []string
	var err error
	var v int
	xlsx := excelize.NewFile()
	//	var style int
	//	style, err = xlsx.NewStyle(`{"number_format": 22}`)
	if err != nil {
		log.Println("0# Exportexcel ", err.Error())
	}
	if mname, err = db.Getsql("SELECT mname FROM tblMain WHERE nanostamp=" + nanostamp); err != nil {
		log.Println("#1 Exporttoexcel Error", err.Error())
	}
	if n1, err = db.Getsql("SELECT note FROM tblMain WHERE nanostamp=" + nanostamp); err != nil {
		log.Println("#1 Exporttoexcel Error", err.Error())
	}
	if t1, err = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblMain WHERE nanostamp=%v", nanostamp)); err != nil {
		log.Println("#2 Exporttoexcel PTrakError", err.Error())
	}
	xlsx.SetCellValue("Sheet1", "A1", t1[0]+"  "+mname[0]+", "+n1[0])

	if t1, err = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblPTrak WHERE nanostamp=%v", nanostamp)); err != nil {
		log.Println("#2 Exporttoexcel PTrakError", err.Error())
	}
	if d1, err = db.Getsql("SELECT mdata FROM tblPTrak WHERE nanostamp=" + nanostamp); err != nil {
		log.Println("#2.1 Exporttoexcel PTrak Error", err.Error())
	}
	for i := 0; i < len(d1); i++ {
		v, _ = strconv.Atoi(d1[i])
		xlsx.SetCellValue("Sheet1", "B2", "PTrak")
		xlsx.SetCellValue("Sheet1", "C2", ChartTitles[0])
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("B%v", i+3), fixtime(t1[i]))
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("C%v", i+3), v)
	}

	if t1, err = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblDustTrak WHERE nanostamp=%v", nanostamp)); err != nil {
		log.Println("#3 Exporttoexcel DustTrakError", err.Error())
	}
	if d1, err = db.Getsql("SELECT mdata FROM tblDustTrak WHERE nanostamp=" + nanostamp); err != nil {
		log.Println("#3.1 Exporttoexcel DustTrak Error", err.Error())
	}
	for i := 0; i < len(d1); i++ {
		v, _ = strconv.Atoi(d1[i])
		xlsx.SetCellValue("Sheet1", "D2", "DustTrak")
		xlsx.SetCellValue("Sheet1", "E2", ChartTitles[1])
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("D%v", i+3), fixtime(t1[i]))
		xlsx.SetCellValue("Sheet1", fmt.Sprintf("E%v", i+3), v)
	}

	if t1, err = db.Getsql(fmt.Sprintf("SELECT tstamp FROM tblAeroTrak WHERE nanostamp=%v", nanostamp)); err != nil {
		log.Println("#4 Exporttoexcel AeroTrakError", err.Error())
	}
	for c := 0; c < 6; c++ {
		if d1, err = db.Getsql(fmt.Sprintf("SELECT ch%v FROM tblAeroTrak WHERE nanostamp=%v", c+1, nanostamp)); err != nil {
			log.Println("#4 Exporttoexcel AeroTrakError", err.Error())
		}
		for i := 0; i < len(d1); i++ {
			v, _ = strconv.Atoi(d1[i])
			xlsx.SetCellValue("Sheet1", "F2", "AeroTrak")
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("%v2", "GHIJKL"[c:c+1]), ChartTitles[2+c])
			//			xlsx.SetCellStyle("Sheet1", fmt.Sprintf("F%v", i+3), fmt.Sprintf("F%v", i+3), style)
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("F%v", i+3), fixtime(t1[i]))
			xlsx.SetCellValue("Sheet1", fmt.Sprintf("%v%v", "GHIJKL"[c:c+1], i+3), v)
		}
	}
	// Set active sheet of the workbook.
	// Save xlsx file by the given path.
	if err := xlsx.SaveAs(fn); err != nil {
		log.Println(err.Error())
	}
}

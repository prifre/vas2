package vasdatabase

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"

	_ "github.com/mattn/go-sqlite3"
)

type DBtype struct {
	conn      *sql.DB
	Tstamp    string
	Nanostamp int64
	Mname     string
	Note      string
	Mdata     []int32
}

// Opendb säkerställer att anslutningen lever utan att stänga den i onödan.
func (db *DBtype) Opendb() error {
	if db.conn != nil {
		if err := db.conn.Ping(); err == nil {
			return nil // Anslutningen är redan igång!
		}
	}

	fname := "/home/prifre/dev/go/src/vas2/resources/vasdatabase.db"
	if fyne.CurrentApp() != nil {
		if pref := fyne.CurrentApp().Preferences().String("dbfilename"); pref != "" {
			fname = pref
		}
	}

	// Sätt timeout och WAL-mode för att undvika locking-problem i SQLite
	connStr := fname
	if !strings.Contains(connStr, "_busy_timeout") {
		if strings.Contains(connStr, "?") {
			connStr += "&_busy_timeout=5000&_journal_mode=WAL"
		} else {
			connStr += "?_busy_timeout=5000&_journal_mode=WAL"
		}
	}

	var err error
	db.conn, err = sql.Open("sqlite3", connStr)
	if err != nil {
		log.Println("#1 Opendb open error:", err)
		return err
	}

	// Tillåt Go att behålla anslutningen i sin pool
	db.conn.SetMaxOpenConns(1)
	db.conn.SetMaxIdleConns(1)
	db.conn.SetConnMaxIdleTime(time.Hour * 2)
	db.conn.SetConnMaxLifetime(time.Hour * 2)

	return db.conn.Ping()
}

// Closedb anropas ELAST när hela programmet avslutas (inte under körning!)
func (db *DBtype) Closedb() {
	if db.conn != nil {
		db.conn.Close()
		db.conn = nil
	}
}

// Addmeasurement sparar mätdata tryggt
func (db *DBtype) Addmeasurement() error {
	if err := db.Opendb(); err != nil {
		return fmt.Errorf("#1 AddMeasurement open error: %w", err)
	}

	if len(db.Mdata) == 0 {
		return fmt.Errorf("#1 AddMeasurement: Mdata is empty")
	}

	// Om inga instrument har giltiga mätvärden än, se till att tblMain har raden
	if db.Mdata[0] <= -1 && db.Mdata[1] <= -1 && db.Mdata[2] <= -1 {
		n1, err := db.Getsql(fmt.Sprintf("SELECT mname FROM tblMain WHERE nanostamp=%d", db.Nanostamp))
		if err != nil {
			return fmt.Errorf("#2 AddMeasurement main check error: %w", err)
		}
		if len(n1) == 0 || n1[0] == "" || n1[0] == "0" {
			stmt, err := db.conn.Prepare("INSERT INTO tblMain (nanostamp, tstamp, mname, note) VALUES (?, ?, ?, ?)")
			if err != nil {
				return err
			}
			defer stmt.Close() // Stänger enbart SQL-satsen, EINTE databasen!
			if _, err := stmt.Exec(db.Nanostamp, db.Tstamp, db.Mname, db.Note); err != nil {
				return err
			}
		}
		return nil
	}

	// PTrak
	if db.Mdata[0] > -1 {
		stmt, err := db.conn.Prepare("INSERT INTO tblPTrak (nanostamp, tstamp, mdata) VALUES (?, ?, ?)")
		if err == nil {
			_, err = stmt.Exec(db.Nanostamp, db.Tstamp, db.Mdata[0])
			stmt.Close()
		}
		if err != nil {
			log.Println("Error inserting PTrak:", err)
		}
	}

	// DustTrak
	if db.Mdata[1] > -1 {
		stmt, err := db.conn.Prepare("INSERT INTO tblDustTrak (nanostamp, tstamp, mdata) VALUES (?, ?, ?)")
		if err == nil {
			_, err = stmt.Exec(db.Nanostamp, db.Tstamp, db.Mdata[1])
			stmt.Close()
		}
		if err != nil {
			log.Println("Error inserting DustTrak:", err)
		}
	}

	// AeroTrak
	if db.Mdata[2] > -1 && len(db.Mdata) >= 8 {
		stmt, err := db.conn.Prepare("INSERT INTO tblAeroTrak (nanostamp, tstamp, ch1, ch2, ch3, ch4, ch5, ch6) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
		if err == nil {
			_, err = stmt.Exec(db.Nanostamp, db.Tstamp, db.Mdata[2], db.Mdata[3], db.Mdata[4], db.Mdata[5], db.Mdata[6], db.Mdata[7])
			stmt.Close()
		}
		if err != nil {
			log.Println("Error inserting AeroTrak:", err)
		}
	}

	return nil
}

func (db *DBtype) Closemeasurement() error {
	if err := db.Opendb(); err != nil {
		return err
	}

	stmt, err := db.conn.Prepare("UPDATE tblMain SET tend=? WHERE nanostamp=?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	nowStr := time.Now().Format(time.RFC3339)
	_, err = stmt.Exec(nowStr, db.Nanostamp)
	return err
}

func (db *DBtype) Getsql(sq string) ([]string, error) {
	if err := db.Opendb(); err != nil {
		return nil, err
	}

	rows, err := db.conn.Query(sq)
	if err != nil {
		log.Println("#2 Getsql Query error:", err)
		return nil, err
	}
	defer rows.Close() // Stänger rad-pekaren så SQLite-låset släpps

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if len(cols) > 1 {
		return nil, fmt.Errorf("too many columns in query: %d", len(cols))
	}

	var result []string
	for rows.Next() {
		var s sql.NullString
		if err := rows.Scan(&s); err != nil {
			continue
		}
		if s.Valid {
			result = append(result, s.String)
		} else {
			result = append(result, "")
		}
	}

	return result, rows.Err()
}

func (db *DBtype) Deleteall(nanostamp string) error {
	if err := db.Opendb(); err != nil {
		return err
	}

	tables := []string{"tblDustTrak", "tblPTrak", "tblAeroTrak", "tblMain"}
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE nanostamp=?", table)
		stmt, err := db.conn.Prepare(query)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(nanostamp)
		stmt.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DBtype) Updatedetails(nanostamp string) string {
	if strings.TrimSpace(nanostamp) == "" {
		return "Ingen mätning vald."
	}

	var d strings.Builder

	if n1, err := db.Getsql("SELECT tstamp FROM tblMain WHERE nanostamp=" + nanostamp); err == nil && len(n1) > 0 {
		d.WriteString(fmt.Sprintf("Measurement start: %v\n\n", n1[0]))
	}
	if n1, err := db.Getsql("SELECT tend FROM tblMain WHERE nanostamp=" + nanostamp); err == nil && len(n1) > 0 {
		d.WriteString(fmt.Sprintf("Measurement end: %v\n\n", n1[0]))
	}
	if n1, err := db.Getsql("SELECT note FROM tblMain WHERE nanostamp=" + nanostamp); err == nil && len(n1) > 0 {
		d.WriteString(fmt.Sprintf("Note: %v\n\n", n1[0]))
	}

	d.WriteString(fmt.Sprintf("Nanostamp: %v\n\n", nanostamp))

	getCount := func(table string) string {
		q := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE nanostamp=%s", table, nanostamp)
		if n1, err := db.Getsql(q); err == nil && len(n1) > 0 {
			return n1[0]
		}
		return "0"
	}

	d.WriteString(fmt.Sprintf("AeroTrak data: %v\n", getCount("tblAeroTrak")))
	d.WriteString(fmt.Sprintf("DustTrak data: %v\n", getCount("tblDustTrak")))
	d.WriteString(fmt.Sprintf("PTrak data: %v\n", getCount("tblPTrak")))

	return d.String()
}

func (db *DBtype) Pruning() error {
	if db.conn == nil {
		if err := db.Opendb(); err != nil {
			return err
		}
	}

	tables := []string{"tblAeroTrak", "tblDustTrak", "tblPTrak"}
	savefrequency := fyne.CurrentApp().Preferences().IntWithFallback("savefrequency", 0)

	for _, table := range tables {
		// 1. Räkna rader före rensning
		s, err := db.Getsql(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE nanostamp=%d", table, db.Nanostamp))
		if err != nil || len(s) == 0 {
			continue
		}
		bef, _ := strconv.Atoi(s[0])

		if bef > 0 {
			var s1 string
			switch savefrequency {
			case 0: // Spara var 5:e sek
				s1 = "unixepoch(tstamp) % 5 != 0"
			case 1: // Spara var 10:e sek
				s1 = "unixepoch(tstamp) % 10 != 0"
			case 2: // Spara varje hel minut
				s1 = "unixepoch(tstamp) % 60 != 0"
			}
			if s1 != "" {
				// Kör rensningen
				query := fmt.Sprintf("DELETE FROM %s WHERE %s AND nanostamp=?", table, s1)
				res, err := db.conn.Exec(query, db.Nanostamp)
				if err != nil {
					log.Println("#1 pruning exec error:", err)
					continue
				}

				rowsAffected, _ := res.RowsAffected()
				if rowsAffected > 0 {
					log.Printf("Records for nanostamp %d removed from %s: %d\n", db.Nanostamp, table, rowsAffected)
				}
			}
		}
	}
	return nil
}
func (db *DBtype) UpdateMeasurementNameNote(ID string, mname string, note string) error {
	if err := db.Opendb(); err != nil {
		return err
	}

	stmt1, err := db.conn.Prepare("UPDATE tblMain SET mname=? WHERE nanostamp=?")
	if err != nil {
		return err
	}
	defer stmt1.Close()
	if _, err := stmt1.Exec(mname, ID); err != nil {
		return err
	}

	stmt2, err := db.conn.Prepare("UPDATE tblMain SET note=? WHERE nanostamp=?")
	if err != nil {
		return err
	}
	defer stmt2.Close()
	_, err = stmt2.Exec(note, ID)

	return err
}

func (db *DBtype) FillDatabase() {
	db.Nanostamp = time.Now().UnixNano()
	db.Tstamp = time.Now().Format(time.RFC3339)

	cleanTstamp := db.Tstamp
	for _, char := range []string{"-", "/", ":", "."} {
		cleanTstamp = strings.ReplaceAll(cleanTstamp, char, "")
	}
	db.Mname = "Measurement" + cleanTstamp
	if len(db.Mname) > 26 {
		db.Mname = db.Mname[:26]
	}

	db.Mdata = []int32{-1, -1, -1, -1, -1, -1, -1, -1}
	_ = db.Addmeasurement()

	for i := 0; i < len(db.Mdata); i++ {
		db.Mdata[i] = int32(rand.Intn(1001))
	}
	_ = db.Addmeasurement()
}

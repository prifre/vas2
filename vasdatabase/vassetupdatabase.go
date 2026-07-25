package vasdatabase

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
)

func (db *DBtype) SetupDatabase() error {
	fname := "/home/prifre/dev/go/src/vas2/resources/vasdatabase.db"
	if fyne.CurrentApp() != nil {
		if pref := fyne.CurrentApp().Preferences().String("dbfilename"); pref != "" {
			fname = pref
		}
	}

	// Sätt inställningen om den inte fanns
	if fyne.CurrentApp() != nil {
		fyne.CurrentApp().Preferences().SetString("dbfilename", fname)
	}

	// Kolla om filen finns på disken
	fileExisted := true
	if _, err := os.Stat(fname); os.IsNotExist(err) {
		fileExisted = false
		log.Println("Database file missing, creating new:", fname)
	}

	// Lägg på timeout och WAL-mode
	connStr := fname
	if !strings.Contains(connStr, "_busy_timeout") {
		connStr += "?_busy_timeout=5000&_journal_mode=WAL"
	}

	var err error
	db.conn, err = sql.Open("sqlite3", connStr)
	if err != nil {
		log.Println("#1 SetupDatabase Failed to open db:", err)
		return err
	}

	// Om filen var helt ny ELLER om tabeller saknas, skapa tabellerna
	if !fileExisted || CheckTablesExist(db.conn) != nil {
		log.Println("Creating missing database tables...")
		if err = db.Createtables(); err != nil {
			log.Println("#2 SetupDatabase CreateTables error:", err)
			return err
		}
	}

	return nil
}

func CheckTablesExist(conn *sql.DB) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	sq := "SELECT name FROM sqlite_master WHERE type='table'"
	rows, err := conn.Query(sq)
	if err != nil {
		log.Println("#1 CheckTablesExist Query error:", err)
		return err
	}
	defer rows.Close() // 👈 VIKTIGT: Släpp rader direkt!

	existingTables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			existingTables[name] = true
		}
	}

	requiredTables := []string{"tblMain", "tblAeroTrak", "tblDustTrak", "tblPTrak"}
	for _, tbl := range requiredTables {
		if !existingTables[tbl] {
			return fmt.Errorf("missing table: %s", tbl)
		}
	}

	return nil // Alla tabeller finns!
}

func (db *DBtype) Createtables() error {
	if db.conn == nil {
		return fmt.Errorf("db.conn is nil in Createtables")
	}

	sq := []string{
		`CREATE TABLE IF NOT EXISTS tblMain (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
			nanostamp INTEGER, 
			tstamp TEXT, 
			tend TEXT, 
			mname TEXT, 
			note TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS tblAeroTrak (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
			nanostamp INTEGER, 
			tstamp TEXT, 
			ch1 INTEGER, ch2 INTEGER, ch3 INTEGER, ch4 INTEGER, ch5 INTEGER, ch6 INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS tblDustTrak (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
			nanostamp INTEGER, 
			tstamp TEXT, 
			mdata INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS tblPTrak (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, 
			nanostamp INTEGER, 
			tstamp TEXT, 
			mdata INTEGER
		);`,
	}

	for _, query := range sq {
		stmt, err := db.conn.Prepare(query)
		if err != nil {
			log.Println("#1 CreateTables Prepare failed:", err)
			return err
		}
		_, err = stmt.Exec()
		stmt.Close() // 👈 Stäng direkt efter Exec
		if err != nil {
			log.Println("#2 CreateTables Exec failed:", err)
			return err
		}
	}

	return nil
}

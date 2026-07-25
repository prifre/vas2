package vasdatabase

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

func (db *DBtype) Exportonetotext() error {
	tbl := []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}
	var cnt []string

	dir := fyne.CurrentApp().Preferences().String("homedir")
	if db.Nanostamp == 0 {
		return errors.New("No current measurement, no data exported.")
	}

	err := db.Opendb()
	if err != nil {
		msg := fmt.Sprintf("#1 Exportonetotext open Failed: %v", err)
		log.Println(msg)
		return errors.New(msg)
	}

	// 1. Räkna rader säkert
	for i := 0; i < len(tbl); i++ {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE nanostamp=%v", tbl[i], db.Nanostamp)

		err := db.conn.QueryRow(query).Scan(&count)
		if err != nil {
			log.Printf("Error counting rows for %s: %v", tbl[i], err)
			cnt = append(cnt, "0")
		} else {
			cnt = append(cnt, fmt.Sprintf("%d", count))
		}
	}

	// 2. Exportera data
	for i := 0; i < len(tbl); i++ {
		tableName := tbl[i]

		// Använd en anonym funktion per tabell för att nyttja defer säkert
		err := func() error {
			f, err := os.OpenFile(filepath.Join(dir, tableName+".txt"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("#2 Export Could not create textfile: %w", err)
			}
			defer f.Close()

			var sq string
			switch tableName {
			case "tblDustTrak", "tblPTrak":
				sq = fmt.Sprintf("SELECT id, nanostamp, tstamp, mdata FROM %s WHERE nanostamp=%v", tableName, db.Nanostamp)
			case "tblAeroTrak":
				sq = fmt.Sprintf("SELECT id, nanostamp, tstamp, ch1, ch2, ch3, ch4, ch5, ch6 FROM tblAeroTrak WHERE nanostamp=%v", db.Nanostamp)
			case "tblMain":
				sq = fmt.Sprintf("SELECT id, nanostamp, tstamp, tend, mname, note FROM tblMain WHERE nanostamp=%v", db.Nanostamp)
			}

			rows, err := db.conn.Query(sq)
			if err != nil {
				return fmt.Errorf("#3 Export query error: %w", err)
			}
			defer rows.Close()

			var id int
			var nanostamp int64
			var tstamp, tend, mname, note sql.NullString
			var mdata, ch1, ch2, ch3, ch4, ch5, ch6 int

			for rows.Next() {
				s := ""
				mn, ts, te, nt := "", "", "", ""

				switch tableName {
				case "tblDustTrak", "tblPTrak":
					if err = rows.Scan(&id, &nanostamp, &tstamp, &mdata); err != nil {
						return fmt.Errorf("ERROR_%s: %w", tableName, err)
					}
					if tstamp.Valid {
						ts = tstamp.String
					}
					s = fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", tableName, id, nanostamp, ts, mdata)

				case "tblAeroTrak":
					if err = rows.Scan(&id, &nanostamp, &tstamp, &ch1, &ch2, &ch3, &ch4, &ch5, &ch6); err != nil {
						return fmt.Errorf("ERROR_%s: %w", tableName, err)
					}
					if tstamp.Valid {
						ts = tstamp.String
					}
					s = fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tableName, id, nanostamp, ts, ch1, ch2, ch3, ch4, ch5, ch6)

				case "tblMain":
					if err = rows.Scan(&id, &nanostamp, &tstamp, &tend, &mname, &note); err != nil {
						return fmt.Errorf("ERROR_tblMain: %w", err)
					}
					if mname.Valid {
						mn = mname.String
					}
					if tstamp.Valid {
						ts = tstamp.String
					}
					if tend.Valid {
						te = tend.String
					}
					if note.Valid {
						nt = note.String
					}
					s = fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tableName, id, nanostamp, ts, te, mn, nt)
				}

				if _, err := f.Write([]byte(s)); err != nil {
					log.Println("error writing textfile ", tableName, err)
				}
			}

			if err := rows.Err(); err != nil {
				return fmt.Errorf("rows iteration error for %s: %w", tableName, err)
			}

			log.Printf("Exported %v from %v to textfile %v ok.", cnt[i], tableName, f.Name())
			return nil
		}()

		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (db *DBtype) Exporttotext() error {
	tbl := []string{"tblAeroTrak", "tblDustTrak", "tblPTrak", "tblMain"}
	var cnt []int

	dir := fyne.CurrentApp().Preferences().String("homedir")
	err := db.Opendb()
	if err != nil {
		log.Printf("#1 Exporttotext open Failed: %v", err)
		return fmt.Errorf("#1 Exporttotext open Failed: %w", err)
	}

	// 1. Räkna rader säkert
	for i := 0; i < len(tbl); i++ {
		var count int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl[i])
		err := db.conn.QueryRow(countQuery).Scan(&count)
		if err != nil {
			log.Printf("Error selecting COUNT(*) from tbl %s: %v", tbl[i], err)
			cnt = append(cnt, 0)
		} else {
			cnt = append(cnt, count)
		}
	}

	// 2. Exportera data
	for i := 0; i < len(tbl); i++ {
		tableName := tbl[i]

		err := func() error {
			f, err := os.OpenFile(filepath.Join(dir, tableName+".txt"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("#2 Export Could not create textfile: %w", err)
			}
			defer f.Close()

			var sq string
			switch tableName {
			case "tblDustTrak", "tblPTrak":
				sq = fmt.Sprintf("SELECT id, nanostamp, tstamp, mdata FROM %s", tableName)
			case "tblAeroTrak":
				sq = "SELECT id, nanostamp, tstamp, ch1, ch2, ch3, ch4, ch5, ch6 FROM tblAeroTrak"
			case "tblMain":
				sq = "SELECT id, nanostamp, tstamp, tend, mname, note FROM tblMain"
			}

			rows, err := db.conn.Query(sq)
			if err != nil {
				return fmt.Errorf("#3 Export query error for table %s: %w", tableName, err)
			}
			defer rows.Close()

			var id int
			var nanostamp int64
			var tstamp, tend, mname, note sql.NullString
			var mdata, ch1, ch2, ch3, ch4, ch5, ch6 int

			for rows.Next() {
				s := ""
				mn, ts, te, nt := "", "", "", ""

				switch tableName {
				case "tblDustTrak", "tblPTrak":
					if err = rows.Scan(&id, &nanostamp, &tstamp, &mdata); err != nil {
						return fmt.Errorf("ERROR %s: %w", tableName, err)
					}
					if tstamp.Valid {
						ts = tstamp.String
					}
					s = fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", tableName, id, nanostamp, ts, mdata)

				case "tblAeroTrak":
					if err = rows.Scan(&id, &nanostamp, &tstamp, &ch1, &ch2, &ch3, &ch4, &ch5, &ch6); err != nil {
						return fmt.Errorf("ERROR %s: %w", tableName, err)
					}
					if tstamp.Valid {
						ts = tstamp.String
					}
					s = fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tableName, id, nanostamp, ts, ch1, ch2, ch3, ch4, ch5, ch6)

				case "tblMain":
					if err = rows.Scan(&id, &nanostamp, &tstamp, &tend, &mname, &note); err != nil {
						return fmt.Errorf("ERROR %s: %w", tableName, err)
					}
					if mname.Valid {
						mn = mname.String
					}
					if tstamp.Valid {
						ts = tstamp.String
					}
					if tend.Valid {
						te = tend.String
					}
					if note.Valid {
						nt = note.String
					}
					s = fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tableName, id, nanostamp, ts, te, mn, nt)
				}

				if _, err := f.Write([]byte(s)); err != nil {
					log.Println("error writing textfile ", tableName, err)
				}
			}

			if err := rows.Err(); err != nil {
				return fmt.Errorf("rows iteration error for %s: %w", tableName, err)
			}

			log.Printf("Exported %v from %v to textfile %v ok.", cnt[i], tableName, f.Name())
			return nil
		}()

		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

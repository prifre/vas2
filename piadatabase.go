package main // package

import (
	//	"database/sql"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	//	"github.com/mxk/go-sqlite/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

type dbtype struct {
	conn         *sql.DB
	statement    *sql.Stmt
	reply        sql.Result
	Databasepath string
	mname        string
	tstamp       string
	//	tend      string
	nanostamp int64
	note      string
	mdata     [8]int32
	dbstart   time.Time
}

// add measurement, based on table and data as []int32
func (db *dbtype) AddMeasurement() error {
	// var statement *sqlite3.Stmt
	// var result sqlite3.Result
	var sq []string
	var err error
	err = db.Opendb()
	if err != nil {
		log.Println("error opening", err.Error())
	}
	if db.mdata[0] > -1 {
		sq = append(sq, fmt.Sprintf("INSERT INTO tblPTrak (nanostamp,tstamp,mdata) VALUES (%v,'%v',%v) ", db.nanostamp, db.tstamp, db.mdata[0]))
	}
	if db.mdata[1] > -1 {
		sq = append(sq, fmt.Sprintf("INSERT INTO tblDustTrak (nanostamp,tstamp,mdata) VALUES (%v,'%v',%v) ", db.nanostamp, db.tstamp, db.mdata[0]))
	}
	if db.mdata[2] > -1 {
		sq = append(sq, fmt.Sprintf("INSERT INTO tblAeroTrak (nanostamp,tstamp,ch1,ch2,ch3,ch4,ch5,ch6) VALUES (%v,'%v',%v,%v,%v,%v,%v,%v) ",
			db.nanostamp, db.tstamp, db.mdata[2], db.mdata[3], db.mdata[4], db.mdata[5], db.mdata[6], db.mdata[7]))
	}
	if len(sq) == 0 {
		var n1 []string
		n1, err = db.Getsql(fmt.Sprintf("SELECT mname FROM tblMain WHERE nanostamp=%v ", db.nanostamp))
		if err != nil {
			log.Println("error getting tblMain ", err.Error())
		}
		if n1 == nil || n1[0] == "0" || n1[0] == "" {
			sq = append(sq, fmt.Sprintf("INSERT INTO tblMain (nanostamp,tstamp,mname,note) VALUES (%v,'%v','%v','%v')",
				db.nanostamp, db.tstamp, db.mname, db.note))
		}

	}
	err = db.Opendb()
	//	db.conn.BusyTimeout(time.Second)
	if err != nil {
		log.Println("error opening2", err.Error())
	}
	//	for _, s := range sq {
	for i := 0; i < len(sq); i++ {
		db.statement, err = db.conn.Prepare(sq[i]) // Prepare SQL Statement
		if fmt.Sprintf("%v", err) == "sqlite3: database is locked [5]" {
			return nil
		}
		if err != nil {
			log.Println("#2 execsql prepare failed:  ", err.Error())
			return err
		}
		_, err = db.statement.Exec() // Execute SQL Statements
		if err != nil {
			if fmt.Sprintf("%v", err) == "sqlite3: database is locked [5]" {
				return nil
			}
		}
		if err != nil {
			log.Println("#2 execsql exec failed: ", err.Error())
			return err
		}
	}
	return err
}
func (db *dbtype) Setupdb(g *game) error {
	var err error
	var temp fyne.URI
	temp, err = storage.Child(g.app.Storage().RootURI(), g.app.UniqueID()+".db")
	db.Databasepath = temp.Path()
	if err != nil {
		log.Fatal("#1 Setupdb storage.Child error", err.Error())
	}
	if _, err = os.Stat(db.Databasepath); err == nil {
		err = db.Opendb()
		if err != nil {
			log.Println("#1 setupdb Failed to open db '"+db.Databasepath+"'", db.conn)
		}
	} else {
		log.Println("#2 pia.db not found, creating new db: " + db.Databasepath)
		var file *os.File
		file, err = os.Create(db.Databasepath) // Create SQLite file
		if err != nil {
			log.Println("#3 setupdb Failed to create db", err.Error())
		}
		file.Close()
		err = db.CreateTables() // Create Database Tables
		if err != nil {
			log.Println("#4 Could not create tables!", err.Error())
		} else {
			log.Println("pia.db tables created")
		}
		err = db.Opendb()
		if err != nil {
			log.Println("#5 setupdb Failed to open db", db.conn)
		}
	}
	return err
}
func (db *dbtype) CreateTables() error {
	var err error
	var sq []string

	err = db.Opendb()
	if err != nil {
		log.Println("#1 CreateTables failed Opendb: ", err.Error())
		return err
	}
	// check if table exists
	_, table_check := db.conn.Query("select * from tblMain;")

	if table_check == nil {
		return nil
		//table tblMain exists, so probably all is well...
	}
	//create tables...
	sq = append(sq, "CREATE TABLE tblMain (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, nanostamp integer, tstamp TEXT, "+
		"tend TEXT, mname TEXT, note TEXT);")
	sq = append(sq, `CREATE TABLE tblAeroTrak (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, nanostamp integer, tstamp TEXT,`+
		` ch1 integer, ch2 integer, ch3 integer, ch4 integer, ch5 integer, ch6 integer);`)
	sq = append(sq, `CREATE TABLE tblDustTrak (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, nanostamp integer, tstamp TEXT, mdata integer);`)
	sq = append(sq, `CREATE TABLE tblPTrak (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, nanostamp integer, tstamp TEXT, mdata integer);`)
	for _, s := range sq {
		db.statement, err = db.conn.Prepare(s) // Prepare SQL Statement
		if err != nil {
			if err.Error() == "table tblMain already exists" {
				err = nil
				return err
			}
			log.Println("#1 CreateTables: ", err.Error())
		}
		db.reply, err = db.statement.Exec() // Execute SQL Statements
		if err != nil {
			log.Println("#2 CreateTables failed: ", sq, " ", err.Error(), db.reply)
			return err
		}
	}
	return err
}

func (db *dbtype) Opendb() error {
	var err error
	// var temp fyne.URI
	if db.conn != nil {
		return nil
	}
	db.conn, err = sql.Open("sqlite3", db.Databasepath) // Open the created SQLite File
	db.conn.SetMaxOpenConns(1)
	db.conn.SetMaxIdleConns(0)
	db.conn.SetConnMaxIdleTime(time.Hour * 2)
	db.conn.SetConnMaxLifetime(time.Hour * 2)
	if err != nil {
		log.Fatal("setupdatabase storage.Child error", err.Error())
	}
	// if db.conn.Stats().OpenConnections == 0 {
	// 	db.conn, err = sqlite3.Open("sqlite3", db.Databasepath+"?cache=shared&mode=rwc") // Open the created SQLite File
	// 	if err != nil {
	// 		log.Fatal("setupdatabase storage.Child error", err.Error())
	// 	}
	// }
	return nil
}

func (db *dbtype) CloseMeasurement() error {
	var sq string
	var err error
	sq = fmt.Sprintf("UPDATE tblMain SET tend='%v' WHERE nanostamp=%v", fmt.Sprintf("%v", time.Now().Format(time.RFC3339)), db.nanostamp)
	err = db.Opendb()
	if err != nil {
		log.Println("#1 CloseMeasurement open Failed", err.Error())
		return err
	}
	db.statement, err = db.conn.Prepare(sq) // Prepare SQL Statement
	if err != nil {
		log.Println("#2 CloseMeasurement prepare failed:", sq, " ", err.Error())
		return err
	}
	db.reply, err = db.statement.Exec() // Execute SQL Statements
	if err != nil {
		log.Println("#2 CloseMeasurement exec failed: ", sq, " ", err.Error(), db.reply)
		return err
	}
	db.conn.Close()
	db.conn = nil
	return err
}
func (db *dbtype) Exporttotext(g *game) (string, error) {
	var err error
	var id int
	var cnt []string
	var nanostamp int64 = 0
	var msg, sq, mn, ts, te, nt string
	var tstamp, tend, mname, note sql.NullString
	var mdata, ch1, ch2, ch3, ch4, ch5, ch6 int
	var f *os.File
	var s []string
	var dir string = filepath.Dir(db.Databasepath)
	err = db.Opendb()
	if err != nil {
		log.Println("#1 Exporttotext open Failed", err.Error())
		return "", err
	}
	for i := 0; i < len(tbl); i++ {
		s, err = db.Getsql(fmt.Sprintf("SELECT COUNT(*) from %v", tbl[i]))
		cnt = append(cnt, s[0])
		if err != nil {
			return "", err
		}
	}
	for i := 0; i < len(tbl); i++ {
		f, err = os.OpenFile(filepath.Join(dir, tbl[i]+".txt"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Println("#2 Export Could not create textfile", err.Error())
			return "", err
		}
		sq = "SELECT * FROM " + tbl[i]
		rows, err := db.conn.Query(sq)
		if err != nil {
			log.Println("#3 Export query error ", err.Error())
			return "", err
		}
		for rows.Next() {
			s := ""
			mn = ""
			ts = ""
			te = ""
			nt = ""
			switch tbl[i] {
			case "tblDustTrak":
				if err = rows.Scan(&id, &nanostamp, &tstamp, &mdata); err != nil {
					log.Println("ERROR_DustTrak", err.Error())
					return "", err
				}
				if tstamp.Valid {
					ts = tstamp.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, mdata)
			case "tblPTrak":
				if err = rows.Scan(&id, &nanostamp, &tstamp, &mdata); err != nil {
					log.Println("ERROR_PTrak", err.Error())
					return "", err
				}
				if tstamp.Valid {
					ts = tstamp.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, mdata)
			case "tblAeroTrak":
				err = rows.Scan(&id, &nanostamp, &tstamp, &ch1, &ch2, &ch3, &ch4, &ch5, &ch6)
				if err != nil {
					log.Printf("ERROR_%v, %v", tbl[i], err.Error())
				}
				if tstamp.Valid {
					ts = tstamp.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, ch1, ch2, ch3, ch4, ch5, ch6)
			case "tblMain":
				err = rows.Scan(&id, &nanostamp, &tstamp, &tend, &mname, &note)
				if err != nil {
					log.Println("ERROR_tblMain", err.Error())
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
					nt = mname.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, te, mn, nt)
			}
			_, err := f.Write([]byte(s))
			if err != nil {
				log.Println("error writing textfile ", tbl[i], err.Error())
			}
		}
		f.Close()
		log.Printf("Exported %v from %v to textfile %v ok.", cnt[i], tbl[i], f.Name())
	}
	if err != nil {
		log.Println("Problems exporting database to text ", err.Error())
		return "", err
	} else {
		msg = "All measurement data has been exported to textfiles:\n"
		msg = msg + "tblAeroTrak.txt\n"
		msg = msg + "tblDustTrak.txt\n"
		msg = msg + "tblPTrak.txt\n"
	}
	return msg, err
}
func (db *dbtype) Exportonetotext(g *game) (string, error) {
	var err error
	var id int
	var cnt []string
	var nanostamp int64 = 0
	var msg, sq, mn, ts, te, nt string
	var tstamp, tend, mname, note sql.NullString
	var mdata, ch1, ch2, ch3, ch4, ch5, ch6 int
	var f *os.File
	var s []string
	var dir string = filepath.Dir(db.Databasepath)
	err = db.Opendb()
	if err != nil {
		log.Println("#1 Exporttotext open Failed", err.Error())
		return "", err
	}
	for i := 0; i < len(tbl); i++ {
		s, err = db.Getsql(fmt.Sprintf("SELECT COUNT(*) from %v WHERE nanostamp=%v", tbl[i], db.nanostamp))
		cnt = append(cnt, s[0])
		if err != nil {
			return "", err
		}
	}
	for i := 0; i < len(tbl); i++ {
		f, err = os.OpenFile(filepath.Join(dir, tbl[i]+".txt"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Println("#2 Export Could not create textfile", err.Error())
			return "", err
		}
		sq = fmt.Sprintf("SELECT * FROM %v WHERE nanostamp=%v", tbl[i], db.nanostamp)
		rows, err := db.conn.Query(sq)
		if err != nil {
			log.Println("#3 Export query error ", err.Error())
			return "", err
		}
		for rows.Next() {
			s := ""
			mn = ""
			ts = ""
			te = ""
			nt = ""
			switch tbl[i] {
			case "tblDustTrak":
				if err = rows.Scan(&id, &nanostamp, &tstamp, &mdata); err != nil {
					log.Println("ERROR_DustTrak", err.Error())
					return "", err
				}
				if tstamp.Valid {
					ts = tstamp.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, mdata)
			case "tblPTrak":
				if err = rows.Scan(&id, &nanostamp, &tstamp, &mdata); err != nil {
					log.Println("ERROR_PTrak", err.Error())
					return "", err
				}
				if tstamp.Valid {
					ts = tstamp.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, mdata)
			case "tblAeroTrak":
				err = rows.Scan(&id, &nanostamp, &tstamp, &ch1, &ch2, &ch3, &ch4, &ch5, &ch6)
				if err != nil {
					log.Printf("ERROR_%v, %v", tbl[i], err.Error())
				}
				if tstamp.Valid {
					ts = tstamp.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, ch1, ch2, ch3, ch4, ch5, ch6)
			case "tblMain":
				err = rows.Scan(&id, &nanostamp, &tstamp, &tend, &mname, &note)
				if err != nil {
					log.Println("ERROR_tblMain", err.Error())
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
					nt = mname.String
				}
				s = s + fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v\t%v\n", tbl[i], id, nanostamp, ts, te, mn, nt)
			}
			_, err := f.Write([]byte(s))
			if err != nil {
				log.Println("error writing textfile ", tbl[i], err.Error())
			}
		}
		f.Close()
		log.Printf("Exported %v from %v to textfile %v ok.", cnt[i], tbl[i], f.Name())
	}
	if err != nil {
		log.Println("Problems exporting database to text ", err.Error())
		return "", err
	} else {
		msg = fmt.Sprintf("All measurement data for measurement %v has been exported to textfiles:\n,", db.nanostamp)
		msg = msg + "tblAeroTrak.txt\n"
		msg = msg + "tblDustTrak.txt\n"
		msg = msg + "tblPTrak.txt\n"
	}
	return msg, err
}

// get one value from database quickly...
func (db *dbtype) Getsql(sq string) ([]string, error) {
	var err error
	var k []string
	var s sql.NullString
	var s2 string
	err = db.Opendb()
	if err != nil {
		log.Println("#1 Getsql opendb error: ", err.Error())
		return nil, err
	}
	rows, err := db.conn.Query(sq)
	if err != nil {
		fmt.Println("#2 Getsql Query error", err.Error())
		return nil, err
	}
	col, err := rows.Columns()
	if err != nil {
		fmt.Println("#3 Getsql Col error", err.Error())
		return nil, err
	}
	if len(col) > 1 {
		log.Println("#4 Getsql too many columns in query! Do your own query!")
		return nil, fmt.Errorf("too many columns!%v", "")
	}
	var ct []*sql.ColumnType
	ct, err = rows.ColumnTypes()
	if err != nil {
		fmt.Println("#5 Getsql CT error", err.Error())
	}
	for rows.Next() {
		switch strings.ToUpper(ct[0].DatabaseTypeName()) {
		case "INTEGER":
			var x int64
			err = rows.Scan(&x)
			if err != nil {
				fmt.Println("#6 Getsql Scan error", err.Error())
			}
			s2 = fmt.Sprintf("%v", x)
		case "TEXT":
			err = rows.Scan(&s)
			s2 = ""
			if s.Valid {
				s2 = fmt.Sprintf("%v", s.String)
			}
			if err != nil {
				fmt.Println("#6 Getsql Scan error", err.Error())
			}
		default:
			err = rows.Scan(&s)
			// COUNT(*)...
			if s.Valid {
				s2 = fmt.Sprintf("%v", s.String)
			}
		}
		k = append(k, s2)
	}
	return k, err
}

func (db *dbtype) deleteall(n string) error {
	var err error
	var sq []string
	// remove from database
	err = db.Opendb()
	if err != nil {
		log.Println("#1 deleteall open Failed", err.Error())
	}
	sq = append(sq, "DELETE FROM tblDustTrak WHERE nanostamp="+n)
	sq = append(sq, "DELETE FROM tblPTrak WHERE nanostamp="+n)
	sq = append(sq, "DELETE FROM tblAeroTrak WHERE nanostamp="+n)
	sq = append(sq, "DELETE FROM tblMain WHERE nanostamp="+n)
	for i := 0; i < len(sq); i++ {
		db.statement, err = db.conn.Prepare(sq[i]) // Prepare SQL Statement
		if err != nil {
			log.Println("#2 deleteall prepare failed: ", sq[i], " ", err.Error())
			return err
		}
		db.reply, err = db.statement.Exec() // Execute SQL Statements
		if err != nil {
			log.Println("#3 deleteall exec failed: ", sq[i], " ", err.Error(), db.reply)
			return err
		}
	}
	return err
}

// // fixes all tend
// func DorepairDatabase(db *dbtype, a fyne.App) error {
// 	var err error
// 	err = db.Opendb()
// 	w := a.NewWindow("repair")
// 	dialog.ShowInformation("Database is never bad, so repair is not needed! ", "...hahahahaha", w)
// 	msg := "Reparation of database can be done anytime.\n"
// 	msg += "Reparation means:\n"
// 	msg += "  - all non-finished measurements will be added and end time.\n"
// 	msg += "  - all measurements not containing any measurement data will be removed."
// 	msg += "  - "
// 	msg += "Reparation means all non-finished measurements will be added and end time.\n"
// 	// 	msg += "Before reparation, you must Stop Measurement!"
// 	// 	dialog.ShowInformation("Database reparation", msg, *w)
// 	// 	return
// 	// }
// 	// get []nanostamp for all tblMain where tend=''
// 	var n1, n2 []string
// 	n1, err = db.Getsql("SELECT nanostamp FROM tblMain WHERE tend=''")
// 	if err != nil {
// 		log.Println("#1 updatedetails SELECT ", err.Error())
// 		dialog.ShowInformation("Database reparation failed: ", msg, w)
// 		return err
// 	}
// 	cnf := dialog.NewConfirm("Fix measurement ends.", "Proceed?", func(reply bool) {
// 		if reply {
// 			for i := 0; i < len(n1); i++ {
// 				n2, err = db.Getsql("SELECT tstamp FROM tblAeroTrak WHERE nanostamp='" + n1[i] + "' ORDER BY tstamp DESC")
// 				if err != nil && len(n2[0]) > 0 {
// 					err = db.Opendb()
// 					if err != nil {
// 						log.Println("#1 DoRepair open Failed", err.Error())
// 					}
// 					sq := "UPDATE tblMain SET tend='" + n2[0] + "' WHERE nanostamp='" + n1[i] + "'"
// 					db.statement, err = db.conn.Prepare(sq) // Prepare SQL Statement
// 					if err != nil {
// 						log.Println("#2 execsql prepare failed: >>", sq, " ", err.Error())
// 						return
// 					}
// 					db.reply, err = db.statement.Exec() // Execute SQL Statements
// 					if err != nil {
// 						log.Println("#2 execsql exec failed: >>", sq, " ", err.Error(), db.reply)
// 						return
// 					}
// 				}
// 			}
// 		}
// 	}, w)
// 	cnf.SetDismissText("No")
// 	cnf.SetConfirmText("Yes")
// 	cnf.Show()
// 	return err
// }

func (db *dbtype) updatedetails(nanostamp string, mname string) string {
	var n1 []string
	var err error
	var d string
	db.Opendb()
	d = fmt.Sprintf("Measurement name: %v", mname)
	if n1, err = db.Getsql("SELECT tstamp FROM tblMain WHERE nanostamp=" + nanostamp); err != nil {
		log.Println("#1 updatedetails SELECT ", err.Error())
	} else {
		d += fmt.Sprintf("\n\nMeasurement start: %v", n1[0])
	}
	if n1, err = db.Getsql("SELECT tend FROM tblMain WHERE nanostamp=" + nanostamp); err != nil {
		log.Println("#2 updatedetails SELECT ", err.Error())
	} else {
		d += fmt.Sprintf("\n\nMeasurement end: %v", n1[0])
	}
	n1, err = db.Getsql("SELECT note FROM tblMain WHERE nanostamp=" + nanostamp)
	if err != nil {
		log.Println("#3 updatedetails SELECT ", err.Error())
	} else {
		d += fmt.Sprintf("\n\nNote: %v\n\n", n1[0])
	}
	n1, err = db.Getsql("SELECT nanostamp FROM tblMain WHERE nanostamp=" + nanostamp)
	if err != nil {
		log.Println("#3 updatedetails SELECT ", err.Error())
	} else {
		d += fmt.Sprintf("\n\nNanostamp: %v\n\n", n1[0])
	}
	n1, err = db.Getsql("SELECT COUNT(*) FROM tblAeroTrak WHERE nanostamp=" + nanostamp)
	if err != nil {
		log.Println("#3 updatedetails SELECT ", err.Error())
	} else {
		d += fmt.Sprintf("\n\nAeroTrak data: %v", n1[0])
	}
	n1, err = db.Getsql("SELECT COUNT(*) FROM tblDustTrak WHERE nanostamp=" + nanostamp)
	if err != nil {
		log.Println("#3 updatedetails SELECT ", err.Error())
	} else {
		d += fmt.Sprintf("\n\nDustTrak data: %v", n1[0])
	}
	n1, err = db.Getsql("SELECT COUNT(*) FROM tblPTrak WHERE nanostamp=" + nanostamp)
	if err != nil {
		log.Println("#3 updatedetails SELECT ", err.Error())
	} else {
		d += fmt.Sprintf("\n\nPTrak data: %v", n1[0])
	}
	return d
}

// Maintainance for measurements
func (db *dbtype) DoMeasurements(g *game) error {
	var sq string
	var d1, nanostring, n1 []string
	var c0, c2 fyne.Container
	var err error
	var myID int = -1
	w := g.app.NewWindow("Measurements")
	w.SetCloseIntercept(func() {
		w.Close()
		w = nil
	})
	details := *widget.NewLabel("")
	d1, err = g.d.Getsql("SELECT mname FROM tblMain ORDER BY nanostamp ASC")
	if err != nil {
		log.Println("#1 handlemeasurements SELECT ", err.Error())
	}
	if len(d1) == 0 {
		return err
	}
	srclist := binding.BindStringList(&d1)
	nanostring, err = g.d.Getsql("SELECT nanostamp FROM tblMain ORDER BY nanostamp ASC")
	if err != nil {
		log.Println("#1 handlemeasurements SELECT ", err.Error())
	}

	list := widget.NewListWithData(srclist, func() fyne.CanvasObject {
		co99 := widget.NewLabel("Placeholder")
		return co99
	}, func(item binding.DataItem, obj fyne.CanvasObject) {
		t1 := obj.(*widget.Label)
		t1.Bind(item.(binding.String))
	})
	list.OnSelected = func(id int) {
		myID = id
		details.Text = db.updatedetails(nanostring[myID], d1[myID])
		details.Refresh()
	}
	wt := widget.NewToolbar(
		//Edit Selected Measurement
		widget.NewToolbarAction(theme.DocumentCreateIcon(),
			func() {
				if myID > -1 {
					d := widget.NewEntry()
					d.SetText(d1[myID])
					d.MultiLine = false
					note := widget.NewEntry()
					note.MultiLine = true
					// GET NOTE FROM DB
					n1, err = g.d.Getsql("SELECT note FROM tblMain WHERE nanostamp=" + nanostring[myID])
					if err != nil {
						log.Print("Getting note failed", err.Error())
						return
					}
					note.SetText(n1[0])
					dialog.ShowForm("Measurement information"+strings.Repeat(" ", 100), "Enter", "Cancel", []*widget.FormItem{
						widget.NewFormItem("Measurement name:", d),
						widget.NewFormItem("Note: ", note)}, func(bool) {
						d1[myID] = d.Text
						sq = "UPDATE tblMain SET mname='" + d.Text + "' WHERE nanostamp=" + nanostring[myID]
						db.statement, err = db.conn.Prepare(sq) // Prepare SQL Statement
						if err != nil {
							log.Println("#2 prepare failed: >> ", err.Error())
							return
						}
						_, err = db.statement.Exec() // Execute SQL Statements
						if err != nil {
							log.Println("#3 exec failed: >> ", err.Error())
							return
						}
						sq = "UPDATE tblMain SET note='" + note.Text + "' WHERE nanostamp=" + nanostring[myID]
						db.statement, err = db.conn.Prepare(sq) // Prepare SQL Statement
						if err != nil {
							log.Println("#2 prepare failed: >> ", err.Error())
							return
						}
						_, err = db.statement.Exec() // Execute SQL Statements
						if err != nil {
							log.Println("#3 exec failed: >> ", err.Error())
							return
						}
						srclist.Reload()
					}, w)
				}
			}),
		widget.NewToolbarAction(theme.DeleteIcon(),
			func() {
				if nanostring[myID] == fmt.Sprintf("%v", g.d.nanostamp) {
					dialog.ShowInformation("Warning", "Active measurement cannot be removed!", w)
					return
				}
				// Remove the element at index i from a.
				if myID > -1 && len(d1) > 0 {
					g.d.deleteall(nanostring[myID])
					if myID < len(d1) {
						copy(d1[myID:], d1[myID+1:])                 // Shift a[i+1:] left one index.
						copy(nanostring[myID:], nanostring[myID+1:]) // Shift a[i+1:] left one index.
					}
					d1[len(d1)-1] = ""  // Erase last element (write zero value).
					d1 = d1[:len(d1)-1] // Truncate slice.
					if myID == len(d1) {
						myID--
					}
					srclist.Reload()
					nanostring[len(nanostring)-1] = ""          // Erase last element (write zero value).
					nanostring = nanostring[:len(nanostring)-1] // Truncate slice.
					// details.Text = db.updatedetails(nanostring[myID], d1[myID])
					// details.Refresh()
					list.Select(myID)
					list.Refresh()
				}
			}),
		// Export to Excel!!!
		widget.NewToolbarAction(theme.DocumentSaveIcon(),
			func() {
				if myID > -1 {
					fn := filepath.Join(g.app.Storage().RootURI().Path(), "pia.xlsx")
					g.d.exporttoexcel(nanostring[myID], fn)
					log.Println("Exported to Excel: " + fn)
					dialog.ShowInformation("Export to Excel!", "Export finished ok.", w)
				} else {
					dialog.ShowInformation("Export to Excel!", "Select a measurement first!", w)
				}
			}),
		//Start Measurement
		widget.NewToolbarAction(theme.SearchReplaceIcon(),
			func() {
				db.nanostamp = Getint64(nanostring[myID])
				db.Pruning(g)
			}),
		//Start Measurement
		widget.NewToolbarAction(theme.MediaPlayIcon(),
			func() {
				if nanostring[myID] > "" {
					g.app.Preferences().SetBool("autostart", true)
					g.app.Preferences().SetString("nanostamp", nanostring[myID])
					g.StartMeasurement()
					w.Close()
				}
			}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HomeIcon(),
			func() {}),
		widget.NewToolbarAction(theme.HelpIcon(),
			func() {}),
	)

	c2 = *container.New(layout.NewGridLayoutWithColumns(2), list, &details)
	c0 = *container.New(layout.NewBorderLayout(wt, nil, nil, nil), wt, &c2)
	w.SetContent(&c0)
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()
	w.Show()
	return err
}

func (db *dbtype) Pruning(g *game) error {
	// Values can be saved per 5 seconds, 10 seconds or per minute...
	// count datapoints for the measurement
	var err error
	var s []string
	var sq string
	var savefrequency int = g.app.Preferences().IntWithFallback("savefrequency", 0)
	for i := 0; i < len(tbl); i++ {
		if tbl[i] == "tblMain" {
			continue
		}
		s, _ = db.Getsql(fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE nanostamp=%v", tbl[i], db.nanostamp))
		bef, _ := strconv.Atoi(s[0])
		// fmt.Printf("Number of records in %v = %v\n", tbl[i], bef)
		// sq = fmt.Sprintf("SELECT SUBSTRING(tstamp,12,8) FROM %v WHERE nanostamp=%v", tbl[i], db.nanostamp)
		// // SUBSTR(tstamp,12,8) == hh:mm:ss
		// s, _ = db.Getsql(sq)
		if bef > 0 {
			s1 := ""
			switch savefrequency {
			case 0: // sample every 5 seconds
				s1 += "(SUBSTR(tstamp,19,1)='1' OR "
				s1 += "SUBSTR(tstamp,19,1)='2' OR "
				s1 += "SUBSTR(tstamp,19,1)='3' OR "
				s1 += "SUBSTR(tstamp,19,1)='4' OR "
				s1 += "SUBSTR(tstamp,19,1)='6' OR "
				s1 += "SUBSTR(tstamp,19,1)='7' OR "
				s1 += "SUBSTR(tstamp,19,1)='8' OR "
				s1 += "SUBSTR(tstamp,19,1)='9')"
			case 1: // sample every 10 seconds
				s1 += "(SUBSTR(tstamp,19,1)='1' OR "
				s1 += "SUBSTR(tstamp,19,1)='2' OR "
				s1 += "SUBSTR(tstamp,19,1)='3' OR "
				s1 += "SUBSTR(tstamp,19,1)='4' OR "
				s1 += "SUBSTR(tstamp,19,1)='5' OR "
				s1 += "SUBSTR(tstamp,19,1)='6' OR "
				s1 += "SUBSTR(tstamp,19,1)='7' OR "
				s1 += "SUBSTR(tstamp,19,1)='8' OR "
				s1 += "SUBSTR(tstamp,19,1)='9')"
			case 2: // sample every minute
				s1 += "SUBSTR(tstamp,18,2)='00'"
			}
			sq = fmt.Sprintf("DELETE FROM %v WHERE %v AND nanostamp=%v", tbl[i], s1, db.nanostamp)
			db.statement, err = db.conn.Prepare(sq) // Prepare SQL Statement
			if err != nil {
				log.Println("#1 pruning error ", sq, err.Error())
				return nil
			}
			_, err = db.statement.Exec() // Execute SQL Statements
			if err != nil {
				log.Println("#2 pruning error ", db.statement, err.Error())
				return nil
			}
			s, _ = db.Getsql(fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE nanostamp=%v", tbl[i], db.nanostamp))
			aft, _ := strconv.Atoi(s[0])
			fmt.Printf("Records for nanostamp %v removed from %v: %v \n", db.nanostamp, tbl[i], bef-aft)
		}
	}
	return nil
}

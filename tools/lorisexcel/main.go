// lorisexcel takes a LORIS database, and dumps
// all of the instruments in it into an excel file.
//
// It's intended to work identically to the old
// PHP version, but more efficiently.
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/driusan/goloris/config"
	"github.com/jmoiron/sqlx"
	"github.com/tealeg/xlsx"
)

type TestName string

func (t TestName) String() string {
	return string(t)
}

func fileExists(f string) bool {
	_, err := os.Stat(f)

	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}
	return true
}

// Check if the instrument (contained in instrumentpath) should have validity
// enabled. It is the caller's responsibility to verify that the instrument
// exists before calling this function, otherwise it will return true
func (t TestName) ValidityEnabled(instrumentpath string) (bool, error) {
	// If the linst file exists, they always contain validity
	if fileExists(instrumentpath + "/" + t.String() + ".linst") {
		return true, nil
	}

	// If the PHP instrument exists, use a regex to search for
	// 'ValidityEnabled = true'
	// It's theoretically not as accurate as the PHP excel dump in edge cases,
	// but in practice the only way this should get tripped up is if someone
	// is trying to be malicious.
	if file := instrumentpath + "/NDB_BVL_Instrument_" + t.String() + ".class.inc"; fileExists(file) {
		f, err := ioutil.ReadFile(file)
		if err != nil {
			return false, err
		}

		matched, _ := regexp.Match("ValidityEnabled(\\s*)=(\\s*)true", f)
		if matched {
			return true, nil
		}
		matched, _ = regexp.Match(`ValidityEnabled(\s*)=(\s*)false`, f)
		if matched {
			return false, nil
		}

		// the default for instruments is true
		return true, nil
	}
	return false, fmt.Errorf("No such instrument: %s", t)
}

// TODO: Move this to config package
func getConfigFromDB(db *sqlx.DB, configname string) (string, error) {
	var val string
	err := db.QueryRow(
		`SELECT Value FROM ConfigSettings cs JOIN Config c ON (cs.ID=c.ConfigID) WHERE cs.Name=?`,
		configname,
	).Scan(&val)
	return val, err
}

// Gets all test names from the LORIS database that
// db is connected to.
func getAllTestNames(db *sqlx.DB) ([]TestName, error) {
	rows, err := db.Query("SELECT Test_name FROM test_names")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []TestName
	for rows.Next() {
		var tn string
		if err := rows.Scan(&tn); err != nil {
			log.Print(err)
			continue
		}
		tests = append(tests, TestName(tn))
	}

	return tests, nil

}

func newFile(db *sqlx.DB, t TestName) (*xlsx.File, error) {
	var q, vldy string
	switch t {
	case "prefrontal_task":
		// Hack for IBIS. This is carried forward from the PHP excelDump
		vldy = `'See validity_of_data field' as Validity, `
	case "radiology_review":
		// Hack to include radiological review columns in radiology_review dump
		// (Carried over from PHP based excel dump)
		q = `SELECT c.PSCID, c.CandID, sp.title as Subproject, s.Visit_label, s.Submitted,
			s.Current_stage, s.Screening, s.Visit, f.Administration,
			e.full_name as Examiner_name, f.Data_entry, f.Validity,
			'Site review:', i.*,
			'Final Review:', COALESCE(fr.Review_Done, 0) as Review_Done,
			fr.Final_Review_Results, fr.Final_Exclusionary, fr.Final_Incidental_Findings,
			fre.full_name as Final_Examiner_Name, fr.Final_Review_Results2,
			fre2.full_name as Final_Examiner2_Name, fr.Final_Exclusionary2,
			COALESCE(fr.Review_Done2, 0) as Review_Done2,
			fr.Final_Incidental_Findings2, fr.Finalized
		FROM candidate c JOIN session s ON (s.CandID=c.CandID)
			LEFT JOIN subproject sp ON (s.SubprojectID=sp.SubprojectID)
			JOIN flag f ON (s.ID=f.sessionID)
			JOIN ` + t.String() + ` i ON (f.CommentID=i.CommentID)
			LEFT JOIN final_radiological_review fr ON (fr.CommentID=i.CommentID)
			LEFT JOIN examiners e on (i.Examiner = e.examinerID)
			LEFT JOIN examiners fre ON (fr.Final_Examiner=fre.examinerID)
			LEFT JOIN examiners fre2 ON (fre2.examinerID=fr.Final_Examiner2)
		WHERE c.CenterID <> 1 AND c.Entity_type != 'Scanner'
			AND i.CommentID NOT LIKE 'DDE%'
			AND c.Active='Y' AND s.Active='Y'
		ORDER BY s.Visit_label, c.PSCID`
	default:
		idir, err := getConfigFromDB(db, "base")
		if idir == "" || err != nil {
			return nil, fmt.Errorf("Could not find LORIS base directory.")
		}
		venabled, err := t.ValidityEnabled(idir + "/project/instruments")
		if err != nil {
			return nil, err
		}
		if vldy == "" && venabled {
			vldy = "f.Validity, "
		}

	}
	if q == "" {
		q = `SELECT c.PSCID, c.CandID, sp.title as Subproject, s.Visit_label,
				 s.Submitted, s.Current_stage, s.Screening, s.Visit,
				f.Administration, e.full_name as Examiner_name,
				f.Data_entry, ` + vldy + `i.*
			FROM candidate c JOIN session s ON (s.CandID=c.CandID)
			LEFT JOIN subproject sp ON (s.SubprojectID=sp.SubprojectID)
				JOIN flag f ON (f.sessionID=s.ID)
				JOIN ` + t.String() + ` i ON (f.CommentID=i.CommentID)
				LEFT JOIN examiners e ON (i.Examiner = e.examinerID)
			WHERE c.CenterID <> 1 AND c.Entity_type != 'Scanner' 
				AND i.CommentID NOT LIKE 'DDE%' 
				AND c.Active='Y' AND s.Active='Y'
			ORDER BY s.Visit_label, c.PSCID`
	}

	rows, err := db.Queryx(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	f := xlsx.NewFile()

	sheet, err := f.AddSheet(t.String())
	if err != nil {
		return nil, err
	}

	headers, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	for i, h := range headers {
		sheet.Cell(0, i).Value = h
	}
	for row := 1; rows.Next(); row++ {
		cols, err := rows.SliceScan()
		if err != nil {
			log.Print(err)
			continue
		}
		for i, val := range cols {
			c := sheet.Cell(row, i)
			if val != nil {
				c.Value = fmt.Sprintf("%s", val)
			} else {
				c.Value = "."
			}
		}
	}
	return f, nil
}

func main() {
	zipo := flag.String("zip", "", "if non-empty, zip xlsx into filename specified")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] configfile\n\n", os.Args[0])

		flag.PrintDefaults()
		os.Exit(1)
	}
	cr, err := config.GetConfigReader(args[0])
	if err != nil {
		log.Fatal(err)

	}
	defer cr.Close()

	dbc, err := config.GetDBConnection(cr)
	if err != nil {
		log.Fatal(err)
	}
	dsn := fmt.Sprintf(
		"%v:%v@tcp(%v:%d)/%v",
		dbc.Username,
		dbc.Password,
		dbc.Host,
		3306,
		dbc.Database,
	)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	testnames, err := getAllTestNames(db)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(testnames))
	zipMu := &sync.Mutex{}
	var zipr *zip.Writer
	if *zipo != "" {
		f, err := os.Create(*zipo)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()

		zipr = zip.NewWriter(f)
		defer zipr.Close()
	}
	for _, test := range testnames {
		// Using goroutines here is about 50% slower than the synchronous
		// version, because it's thrashing the CPU cache. Leaving the
		// closure for now (even if it's not using goroutines), so I can
		//investigate why later..
		func(test TestName) {
			defer wg.Done()
			f, err := newFile(db, test)

			if err != nil {
				log.Print(err)
				return
			}

			if *zipo == "" {
				// If not zipping, just write the xlsx per
				// test name.
				f.Save(test.String() + ".xlsx")
			} else {
				// Zipping, so acquire a lock to write to the
				// zip file.
				zipMu.Lock()
				defer zipMu.Unlock()

				zw, err := zipr.Create(test.String() + ".xlsx")
				if err != nil {
					log.Println(err)
					return
				}
				err = f.Write(zw)
				if err != nil {
					log.Println(err)
					return
				}

			}
		}(test)
	}
	wg.Wait()
}

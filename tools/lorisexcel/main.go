// lorisexcel takes a LORIS database, and dumps
// all of the instruments in it into an excel file.
//
// It's intended to work identically to the old
// PHP version, but more efficiently.
package main

import (
	"log"
	"github.com/jmoiron/sqlx"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/driusan/goloris/config"
	"github.com/tealeg/xlsx"
)

type TestName string

func (t TestName) String() string {
	return string(t)
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

	/*
		PHP:
		//Query to pull the data from the DB
	$Test_name = $instrument['Test_name'];
    if ($Test_name == 'prefrontal_task') {
	    $query = "select c.PSCID, c.CandID, s.SubprojectID, s.Visit_label, s.Submitted, s.Current_stage, s.Screening, s.Visit, f.Administration, e.full_name as Examiner_name, f.Data_entry, 'See validity_of_data field' as Validity, i.* from candidate c, session s, flag f, $Test_name i left outer join examiners e on i.Examiner = e.examinerID where c.PSCID not like 'dcc%' and c.PSCID not like '0%' and c.PSCID not like '1%' and c.PSCID not like '2%' and c.Entity_type != 'Scanner' and i.CommentID not like 'DDE%' and c.CandID = s.CandID and s.ID = f.sessionID and f.CommentID = i.CommentID AND c.Active='Y' AND s.Active='Y' order by s.Visit_label, c.PSCID";
    } else if ($Test_name == 'radiology_review') {
        $query = "select c.PSCID, c.CandID, s.SubprojectID, s.Visit_label, s.Submitted, s.Current_stage, s.Screening, s.Visit, f.Administration, e.full_name as Examiner_name, f.Data_entry, f.Validity, 'Site review:', i.*, 'Final Review:', COALESCE(fr.Review_Done, 0) as Review_Done, fr.Final_Review_Results, fr.Final_Exclusionary, fr.Final_Incidental_Findings, fre.full_name as Final_Examiner_Name, fr.Final_Review_Results2, fre2.full_name as Final_Examiner2_Name, fr.Final_Exclusionary2, COALESCE(fr.Review_Done2, 0) as Review_Done2, fr.Final_Incidental_Findings2, fr.Finalized from candidate c, session s, flag f, $Test_name i left join final_radiological_review fr ON (fr.CommentID=i.CommentID) left outer join examiners e on (i.Examiner = e.examinerID) left join examiners fre ON (fr.Final_Examiner=fre.examinerID) left join examiners fre2 ON (fre2.examinerID=fr.Final_Examiner2) where c.PSCID not like 'dcc%' and c.PSCID not like '0%' and c.PSCID not like '1%' and c.PSCID not like '2%' and c.Entity_type != 'Scanner' and i.CommentID not like 'DDE%' and c.CandID = s.CandID and s.ID = f.sessionID and f.CommentID = i.CommentID AND c.Active='Y' AND s.Active='Y' order by s.Visit_label, c.PSCID";
    } else {
        if (is_file("../project/instruments/NDB_BVL_Instrument_$Test_name.class.inc")) {
            $instrument =& NDB_BVL_Instrument::factory($Test_name, '', false);
            if ($instrument->ValidityEnabled == true) {
	            $query = "select c.PSCID, c.CandID, s.SubprojectID, s.Visit_label, s.Submitted, s.Current_stage, s.Screening, s.Visit, f.Administration, e.full_name as Examiner_name, f.Data_entry, f.Validity, i.* from candidate c, session s, flag f, $Test_name i left outer join examiners e on i.Examiner = e.examinerID where c.PSCID not like 'dcc%' and c.PSCID not like '0%' and c.PSCID not like '1%' and c.PSCID not like '2%' and c.Entity_type != 'Scanner' and i.CommentID not like 'DDE%' and c.CandID = s.CandID and s.ID = f.sessionID and f.CommentID = i.CommentID AND c.Active='Y' AND s.Active='Y' order by s.Visit_label, c.PSCID";
            } else {
	            $query = "select c.PSCID, c.CandID, s.SubprojectID, s.Visit_label, s.Submitted, s.Current_stage, s.Screening, s.Visit, f.Administration, e.full_name as Examiner_name, f.Data_entry, i.* from candidate c, session s, flag f, $Test_name i left outer join examiners e on i.Examiner = e.examinerID where c.PSCID not like 'dcc%' and c.PSCID not like '0%' and c.PSCID not like '1%' and c.PSCID not like '2%' and c.Entity_type != 'Scanner' and i.CommentID not like 'DDE%' and c.CandID = s.CandID and s.ID = f.sessionID and f.CommentID = i.CommentID AND c.Active='Y' AND s.Active='Y' order by s.Visit_label, c.PSCID";
            }
        } else {
	    $query = "select c.PSCID, c.CandID, s.SubprojectID, s.Visit_label, s.Submitted, s.Current_stage, s.Screening, s.Visit, f.Administration, e.full_name as Examiner_name, f.Data_entry, f.Validity, i.* from candidate c, session s, flag f, $Test_name i left outer join examiners e on i.Examiner = e.examinerID where c.PSCID not like 'dcc%' and c.PSCID not like '0%' and c.PSCID not like '1%' and c.PSCID not like '2%' and c.Entity_type != 'Scanner' and i.CommentID not like 'DDE%' and c.CandID = s.CandID and s.ID = f.sessionID and f.CommentID = i.CommentID AND c.Active='Y' AND s.Active='Y' order by s.Visit_label, c.PSCID";
        }
    }
	$DB->select($query, $instrument_table);
    MapSubprojectID($instrument_table);
	writeExcel($Test_name, $instrument_table, $dataDir);
	*/

	var q string
	switch t {
		case "prefrontal_task":
		case "radiology_review":
		default:
			// TODO: Use polymorphism to get the correct query.
			linst := true
			if linst {
	    		q = `SELECT c.PSCID, c.CandID, s.SubprojectID, s.Visit_label,
	s.Submitted, s.Current_stage, s.Screening, s.Visit, f.Administration,
	e.full_name as Examiner_name, f.Data_entry, f.Validity, i.* from candidate c,
	session s, flag f, ` + string(t) + ` i LEFT OUTER JOIN examiners e ON(i.Examiner = e.examinerID)
	WHERE c.CenterID <> 1
		AND c.Entity_type != 'Scanner' 
		AND i.CommentID NOT LIKE 'DDE%' 
		AND (c.CandID = s.CandID) and (s.ID = f.sessionID) AND (f.CommentID = i.CommentID)
		AND c.Active='Y' AND s.Active='Y'
	ORDER BY s.Visit_label, c.PSCID`;
			} else {
				// PHP instrument
			}
	}

	if q == "" {
		return nil, fmt.Errorf("QUERY NOT SPECIFIED")
	}	


	rows, err := db.Queryx(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	f := xlsx.NewFile()

	sheet, err := f.AddSheet(string(t))
	if err != nil {
		return nil, err
	}

	for row := 0; rows.Next(); row++ {
		cols, err := rows.SliceScan()
		if err != nil {
			log.Print(err)
			continue
		}
		for i, val := range cols {
			c := sheet.Cell(row, i)
			c.Value = fmt.Sprintf("%s", val)
		}
	}
	return f, nil
}

func main() {
	cr, err := config.GetConfigReader()
	if err != nil {
		log.Fatal(err)

	}

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

	// Make sure the database connection is valid before
	// going on, so that we don't need to wait until the
	// first query to connect
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	

	testnames, err := getAllTestNames(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, test := range testnames {
		f, err := newFile(db, test)

		if err != nil {
			log.Print(err)
			continue
		}
		f.Save(test.String() + ".xlsx")
	}
} 
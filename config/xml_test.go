package config

import (
	"testing"
	"strings"

	"encoding/xml"
)

func TestGetDBConnection(t *testing.T) {
	testcases := []struct {
		XML      string
		Expected DBConnInfo
	}{
		{`
<?xml version="1.0" encoding="ISO-8859-1" ?>
<!--

	This is a minimal 
  NB: because study and sites elements get merged in a recursive and
  overwriting manner, any time when multiple elements of the same type
  (such as <item/> <item/>) occurs in the study or sites tree, the
  sites tree will overwrite the element entirely instead of simply
  merging - i.e., if the multiple elements are branches, the sites
  branch in its entirely will override the study branch.
-->
<config>
    <!-- database settings -->
<database>
        <host>127.0.0.1</host>
        <username>testuser</username>
        <password>fakepassword!</password>
        <database>testdb</database>
        <name>Test database</name>
    </database>
</config>
`, DBConnInfo{XMLName: xml.Name{"", "database"}, Host: "127.0.0.1", Username: "testuser", Password: "fakepassword!", Database: "testdb"},
		},
	}


	for i, tc := range testcases {
		r := strings.NewReader(tc.XML)
		con, err := GetDBConnection(r)
		if err != nil {
			t.Fatal(err)
		}
		if *con != tc.Expected {
			t.Errorf("Unexpected result for test case %v: got %v want %v", i, *con, tc.Expected)
		}
	}
}

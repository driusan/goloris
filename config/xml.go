package config

import (
	"encoding/xml"
	"io"

	"golang.org/x/net/html/charset"
)

type ConfigXML struct {
	XMLName xml.Name `xml:"config"`
	DB      DBConnInfo
}

type DBConnInfo struct {
	XMLName  xml.Name `xml:"database"`
	Host     string   `xml:"host"`
	Username string   `xml:"username"`
	Password string   `xml:"password"`
	Database string   `xml:"database"`
}

// GetDBConnection extracts the database connection information
// from the reader r, which is assumed to be a reader into a valid
// XML file.
func GetDBConnection(r io.Reader) (*DBConnInfo, error) {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel

	var c ConfigXML
	if err := decoder.Decode(&c); err != nil {
		return nil, err
	}
	return &c.DB, nil
}

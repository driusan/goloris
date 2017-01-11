package config

import (
	"io"
	"os"
)

func GetConfigReader(filename string) (io.ReadCloser, error) {
	if filename == "" {
		return os.Open("config.xml")
	}
	return os.Open(filename)
}

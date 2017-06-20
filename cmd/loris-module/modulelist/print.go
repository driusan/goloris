package modulelist

import (
	"fmt"
	"log"

	"io/ioutil"
)

// PrintAll will print all
func PrintAll(dirs []string) {
	for _, d := range dirs {
		files, err := ioutil.ReadDir(d)
		if err != nil {
			//log.Fatalln(err)
			log.Printf("Could not read LORIS module directory %v\n", d)
			continue
		}
		for _, f := range files {
			if f.IsDir() {
				fmt.Println(f.Name())
			}
		}
	}
}

func PrintValid() {
	files, err := ioutil.ReadDir("modules")
	if err != nil {
		//log.Fatalln(err)
		log.Fatalln("Could not read LORIS module directory. This command must be run from the top level of LORIS.")
	}
	for _, f := range files {
		if f.IsDir() {
			fmt.Println(f.Name())
		}
	}
	files, err = ioutil.ReadDir("project/modules")
	if err != nil {
		log.Fatalln("Could not read LORIS project modules directory.")
	}
	for _, f := range files {
		if f.IsDir() {
			fmt.Println(f.Name())
		}
	}
}

// lorisexcel takes a LORIS database, and dumps
// all of the instruments in it into an excel file.
//
// It's intended to work identically to the old
// PHP version, but more efficiently.
package main

import (
	"flag"
	"log"

	"github.com/driusan/goloris/cmd/loris-module/modulelist"
	//	_ "github.com/go-sql-driver/mysql"
	//	"github.com/driusan/goloris/config"
	//	"github.com/jmoiron/sqlx"
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"list"}
	}
	switch cmd := args[0]; cmd {
	case "show-patches":
		// List all available patches and their status (installed, unsourced, etc)
		// Options for module or version?
		// TODO: Determine where to get the list of available patches. From the filesystem?
	case "install-patch":
		// Install a patch and update the table with list of patches
		// usage: loris-module install-patch filename (filename should match show-patches)
	case "find-patches":
		// Update the patches table with a list of all SQL patches on the filesystem which
		// aren't in the database.
	case "update":
		// usage: loris-module update modulename
		// - sources all unsourced patches for modulename
		// need a way to exclude some?
		// (Can be combined with standard unix tools like xargs to update all modules)
	case "list":
		// List all installed LORIS modules. Should we have a $LORIS variable so that it knows
		// what directory to look at?
		// Add options for --all, --valid (has module.class.inc defined), --project?
		modulelist.PrintAll([]string{"modules", "project/modules"})
	default:
		log.Println("Invalid command", cmd)
	}
}

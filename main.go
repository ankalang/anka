package main

import (
	"fmt"
	"os"

	"github.com/iscosmos/anka/install"
	"github.com/iscosmos/anka/repl"
	"github.com/iscosmos/anka/util"
)

// Version of the ANK interpreter
var Version = "1.0.0"

// The ANK interpreter
func main() {
	args := os.Args
	if len(args) == 2 && args[1] == "sürüm" {
		if newver, update := util.UpdateAvailable(Version); update {
			fmt.Printf("yeni sürüm: %s (sendeki sürüm: %s)\n", newver, Version)
			os.Exit(1)
		} else {
			fmt.Println(Version)
			return
		}
	}

	if len(args) == 3 && args[1] == "indir" {
		install.Install(args[2])
		return
	}

	// begin the REPL
	repl.BeginRepl(args, Version)
}

package main

import (
	"fmt"
	"os"

	"github.com/Tualua/pk_ctladm/pk_scst"

	"github.com/akamensky/argparse"
)

func GetDevList(xml bool) {
	if iScsiTargets, err := pk_scst.ScstGetIscsiTargets(); err != nil {
		fmt.Println("Error getting iscsi targets")
	} else {
		fmt.Print(iScsiTargets)
	}

}

func main() {
	var (
		err error
	)
	parser := argparse.NewParser("ctladm", "Prints provided string to stdout")

	parserDevlist := parser.NewCommand("devlist", "List devices")
	argDevListXml := parserDevlist.Flag("x", "xml", &argparse.Options{Help: "Enable XML Output"})

	parserPortlist := parser.NewCommand("portlist", "List ports")
	argPortListQuiet := parserPortlist.Flag("q", "quiet", &argparse.Options{Help: "Omit header"})
	argPortListVerbose := parserPortlist.Flag("v", "verbose", &argparse.Options{Help: "Verbose output"})
	argPortListPort := parserPortlist.Int("p", "port", &argparse.Options{Help: "Port Number"})
	argPortListXml := parserPortlist.Flag("x", "xml", &argparse.Options{Help: "Enable XML Output"})

	// Parse input
	if err = parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(err))
	} else {
		if parserDevlist.Happened() {
			fmt.Println("devlist")
			GetDevList(*argDevListXml)
		} else if parserPortlist.Happened() {
			fmt.Println("portlist")
			fmt.Println(*argPortListQuiet)
			fmt.Println(*argPortListVerbose)
			fmt.Println(*argPortListPort)
			fmt.Println(*argPortListXml)
		}
	}

}

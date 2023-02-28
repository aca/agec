package main

import (
	"log"
	"os"
)

var version = "devel"

func main() {
	if version == "devel" {
		log.SetFlags(log.LstdFlags | log.Llongfile)
	} else {
		log.SetFlags(0)
	}

	err := cmdMain()
	if err != nil {
		log.Printf("agec: %v", err)
		os.Exit(1)
	}
}

package main

import (
	"log"
	"os"
)

var version = "devel"

func main() {
	log.SetFlags(0)
	err := cmdMain()
	if err != nil {
		log.Printf("agec: %v", err)
		os.Exit(1)
	}
}

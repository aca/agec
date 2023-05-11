package main

import (
	"log"
)

var version = "devel"

func main() {
	log.SetFlags(0)
	err := cmdMain()
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"os"
)

var version = "devel"

func main() {
	err := cmdMain()
	if err != nil {
		os.Exit(1)
	}
}

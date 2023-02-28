package main

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	version = "test"
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"agec": func() int {
			err := cmdMain()
			if err == nil {
				return 0
			}
			return 1
		},
	}))
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "testdata",
	})
}

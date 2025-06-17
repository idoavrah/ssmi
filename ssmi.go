package main

import (
	"os"

	"github.com/idoavrah/ssmi/internal"
)

var version = "X.Y.Z"

func main() {

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--version" {
		println("SSMI v" + version)
		return
	}

	internal.StartApplication()
}

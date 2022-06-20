package main

import (
	"os"

	"github.com/pteichman/ahoy"
)

func main() {
	os.Exit(ahoy.CLI(os.Args[1:]))
}

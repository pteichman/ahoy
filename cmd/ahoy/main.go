package main

import "os"

func main() {
	os.Exit(ahoy.CLI(os.Args[1:], os.Stdout, os.Stderr))
}

package main

import (
	"flag"
	"fmt"
)

var flagVerbose = flag.String("v", "none", "help flag")
var username = flag.String("u", "none", "user flag")

var systemUser string

func cliFunc() int {
	flag.Parse()
	fmt.Printf("Parsed Flags: %s\n", *flagVerbose)
	systemUser = *username
	return 0
}

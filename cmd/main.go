package main

import (
	"flag"
	"fmt"
	"skyline/internal"

	log "github.com/sirupsen/logrus"
)

var flagVerbose = flag.String("v", "none", "help flag")
var username = flag.String("u", "", "user flag")
var mode = flag.String("m", "", "server or client")

func main() {
	internal.GetLogger()
	flag.Parse()
	log.Debugf("parsed flags: %s", *flagVerbose)
	mode := *mode
	switch mode {
	case "server":
		internal.ServerInit()
	case "client":
		if len(*username) == 0 {
			log.Fatal("Username required")
		}
		internal.ClientStart(*username)
	default:
		fmt.Println("Choose client or server")
	}
}

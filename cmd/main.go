package main

import (
	"flag"
	"fmt"
	"skyline/config"
	"skyline/internal/client"
	"skyline/internal/server"

	log "github.com/sirupsen/logrus"
)

var flagVerbose = flag.String("v", "none", "help flag")
var username = flag.String("u", "", "user flag")
var mode = flag.String("m", "", "server or client")

func main() {
	config.GetLogger()
	flag.Parse()
	log.Debugf("parsed flags: %s", *flagVerbose)
	mode := *mode
	switch mode {
	case "server":
		server.ServerInit()
	case "client":
		if len(*username) == 0 {
			log.Error("username required")
			log.Info("generating a user id")
			username = client.GenerateUserID()
			log.Info("username:", *username)
		}
		client.ClientStart(*username)
	default:
		fmt.Println("choose client or server")
	}
}

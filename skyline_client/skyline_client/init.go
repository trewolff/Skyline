package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
)

func initFunc() {
	username, err := user.Current()
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(fmt.Sprintf("/home/%s/skyline", username.Name)); errors.Is(err, os.ErrNotExist) {
		initConfig()
	}
}

func initConfig() {
	fmt.Println("Loading Config")
}

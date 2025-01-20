package config

import (
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

type Specification struct {
	SERVER_HOST_PORT string    `default:"localhost:8082"`
	SERVER_URL       string    `default:"wss://localhost:8082/socket"`
	LOG_LEVEL        log.Level `default:"debug" log:"log.SetLevel()"`
}

func GetConfig() (Specification, error) {
	var s Specification
	err := envconfig.Process("skyline", &s)
	if err != nil {
		log.Fatal(err.Error())
	}
	return s, nil
}

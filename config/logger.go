package config

import (
	log "github.com/sirupsen/logrus"
)

func GetLogger() {
	conf, _ := GetConfig()
	log.SetLevel(conf.LOG_LEVEL)
}

package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func setupLog() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

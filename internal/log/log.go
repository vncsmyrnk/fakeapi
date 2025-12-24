package log

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func SetupLog() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

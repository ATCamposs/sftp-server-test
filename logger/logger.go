package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var loggerInstance *zerolog.Logger

func InitLogger() {
	l := log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	loggerInstance = &l
}

func GetLogger() *zerolog.Logger {
	return loggerInstance
}

package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	Logger *zerolog.Logger
}

func NewZeroLogger(debug bool) *ZeroLogger {
	var logOutput = os.Stderr
	var logger zerolog.Logger
	if debug {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        logOutput,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        logOutput,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return &ZeroLogger{
		Logger: &logger,
	}
}

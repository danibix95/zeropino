package model

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog"
)

// PinoLevel is a string representation of levels adopted PinoJS logger
type PinoLevel string

// Represents all the Pino log levels accepted
// Note: the Panic level has been added to cover this log level not available in Javascript
const (
	Trace PinoLevel = "10"
	Debug PinoLevel = "20"
	Info  PinoLevel = "30"
	Warn  PinoLevel = "40"
	Error PinoLevel = "50"
	Fatal PinoLevel = "60"
	Panic PinoLevel = "70"
)

// ConvertLevel Convert a zerolog log level into the corresponding pino one
func ConvertLevel(level zerolog.Level) string {
	var pinoLevel PinoLevel
	switch level {
	case zerolog.TraceLevel:
		pinoLevel = Trace
	case zerolog.DebugLevel:
		pinoLevel = Debug
	case zerolog.InfoLevel:
		pinoLevel = Info
	case zerolog.WarnLevel:
		pinoLevel = Warn
	case zerolog.ErrorLevel:
		pinoLevel = Error
	case zerolog.FatalLevel:
		pinoLevel = Fatal
	case zerolog.PanicLevel:
		pinoLevel = Panic
	case zerolog.Disabled:
		fallthrough
	case zerolog.NoLevel:
		fallthrough
	default:
		return ""
	}

	return string(pinoLevel)
}

// ParseLevel Parse a string name of the log level and return the corresponding zerolog level
func ParseLevel(level string) (zerolog.Level, error) {
	if len(level) > 0 {
		switch strings.ToLower(level) {
		case "trace":
			return zerolog.TraceLevel, nil
		case "debug":
			return zerolog.DebugLevel, nil
		case "info":
			return zerolog.InfoLevel, nil
		case "warn":
			return zerolog.WarnLevel, nil
		case "error":
			return zerolog.ErrorLevel, nil
		case "fatal":
			return zerolog.FatalLevel, nil
		case "panic":
			return zerolog.PanicLevel, nil
		default:
			return zerolog.NoLevel, fmt.Errorf("level %s is not recognized", level)
		}
	}

	return zerolog.InfoLevel, nil
}

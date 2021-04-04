package internal

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
	PinoTrace PinoLevel = "10"
	PinoDebug PinoLevel = "20"
	PinoInfo  PinoLevel = "30"
	PinoWarn  PinoLevel = "40"
	PinoError PinoLevel = "50"
	PinoFatal PinoLevel = "60"
	PinoPanic PinoLevel = "70"
)

// ConvertLevel Convert a zerolog log level into the corresponding pino one
func ConvertLevel(level zerolog.Level) string {
	var pinoLevel PinoLevel
	switch level {
	case zerolog.TraceLevel:
		pinoLevel = PinoTrace
	case zerolog.DebugLevel:
		pinoLevel = PinoDebug
	case zerolog.InfoLevel:
		pinoLevel = PinoInfo
	case zerolog.WarnLevel:
		pinoLevel = PinoWarn
	case zerolog.ErrorLevel:
		pinoLevel = PinoError
	case zerolog.FatalLevel:
		pinoLevel = PinoFatal
	case zerolog.PanicLevel:
		pinoLevel = PinoPanic
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

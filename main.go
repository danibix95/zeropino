package zerologmia

import (
	"os"

	"github.com/rs/zerolog"
)

// Init function creates a zerolog logger with default properties and custom style
func Init() (zerolog.Logger, error) {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.LevelFieldMarshalFunc = pinoLevel

	hostname, _ := os.Hostname()
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Int("pid", os.Getpid()).
		Str("hostname", hostname).
		Logger().
		Level(zerolog.InfoLevel)

	return log, nil
}

func pinoLevel(l zerolog.Level) string {
	switch l {
	case zerolog.TraceLevel:
		return "10"
	case zerolog.DebugLevel:
		return "20"
	case zerolog.InfoLevel:
		return "30"
	case zerolog.WarnLevel:
		return "40"
	case zerolog.ErrorLevel:
		return "50"
	case zerolog.FatalLevel:
		fallthrough
	case zerolog.PanicLevel:
		return "60"
	case zerolog.NoLevel:
		return ""
	}
	return ""
}

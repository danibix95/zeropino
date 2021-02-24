package zerologmia

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	pinoModel "github.com/danibix95/zerolog-mia/internal"
)

// InitOptions These are the possible options that can be used to initialize the logger
type InitOptions struct {
	Level     string
	UseTimeMs bool
	Writer    io.Writer
}

// Init Creates a zerolog logger with custom default properties and custom style
func Init(options InitOptions) (*zerolog.Logger, error) {
	var logWriter io.Writer = os.Stdout
	if options.Writer != nil {
		logWriter = options.Writer
	}

	logLevel, err := pinoModel.ParseLevel(options.Level)
	if err != nil {
		return nil, err
	}

	return createLogger(logWriter, logLevel, options.UseTimeMs), nil
}

// InitDefault Creates a zerolog logger with custom default properties
// and custom style using predefined writer and log level
func InitDefault() *zerolog.Logger {
	return createLogger(os.Stdout, zerolog.InfoLevel, false)
}

func createLogger(writer io.Writer, level zerolog.Level, useTimeMs bool) *zerolog.Logger {
	// global default configuration
	zerolog.MessageFieldName = "msg"
	zerolog.LevelFieldMarshalFunc = pinoModel.ConvertLevel
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if useTimeMs {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	}

	// ignore hostname in case of error
	hostname, _ := os.Hostname()

	log := zerolog.New(writer).With().
		Timestamp().
		Int("pid", os.Getpid()).
		Str("hostname", hostname).
		Logger().
		Level(level)

	return &log
}

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

// Init Creates a zerolog logger with default properties and custom style
func Init(options InitOptions) (*zerolog.Logger, error) {
	// global default configuration
	zerolog.MessageFieldName = "msg"
	zerolog.LevelFieldMarshalFunc = pinoModel.ConvertLevel
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if options.UseTimeMs {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	}

	var logWriter io.Writer = os.Stdout
	if options.Writer != nil {
		logWriter = options.Writer
	}

	logLevel, err := pinoModel.ParseLevel(options.Level)
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()
	log := zerolog.New(logWriter).With().
		Timestamp().
		Int("pid", os.Getpid()).
		Str("hostname", hostname).
		Logger().
		Level(logLevel)

	return &log, nil
}

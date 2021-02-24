package zerologmia

import (
	"context"

	"github.com/rs/zerolog"
)

// heavily inspired by https://github.com/mia-platform/glogger/blob/master/context.go

type loggerKey struct{}

var defaultLogger *zerolog.Logger = InitDefault()

// WithLogger returns a new context with the provided logger. Use in
// combination with logger.WithField(s) for great effect.
func WithLogger(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// Get retrieves the current logger from the context.
// If no logger is available, the default logger is returned.
func Get(ctx context.Context) *zerolog.Logger {
	logger := ctx.Value(loggerKey{})

	if logger == nil {
		return defaultLogger
	}

	entry, ok := logger.(*zerolog.Logger)
	if !ok {
		return defaultLogger
	}
	return entry
}

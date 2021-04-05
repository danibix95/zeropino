package gorillamux

import (
	"context"
	"testing"

	zp "github.com/danibix95/zeropino"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {
	t.Run("Test WithLogger when no logger is given", func(t *testing.T) {
		ctx := context.TODO()

		ctx = WithLogger(ctx, nil)
		assert.Nil(t, ctx.Value(loggerKey{}))
	})

	t.Run("Test WithLogger when a logger is given", func(t *testing.T) {
		ctx := context.TODO()
		log := zp.InitDefault()

		ctx = WithLogger(ctx, log)
		contextLog := ctx.Value(loggerKey{})
		assert.NotNil(t, contextLog)
		assert.IsType(t, &zerolog.Logger{}, contextLog)
	})
}

func TestGet(t *testing.T) {
	t.Run("Test Get context when no logger was set", func(t *testing.T) {
		ctx := context.TODO()
		logger := Get(ctx)

		assert.NotNil(t, logger)
		assert.IsType(t, &zerolog.Logger{}, logger, "Return the default logger since no logger was previously provided")
		assert.Equal(t, logger.GetLevel(), zerolog.InfoLevel)
	})

	t.Run("Test Get context when a logger was set", func(t *testing.T) {
		ctx := context.TODO()
		contextLogger, err := zp.Init(zp.InitOptions{Level: "debug"})
		assert.Nil(t, err)

		ctx = context.WithValue(ctx, loggerKey{}, contextLogger)
		logger := Get(ctx)

		assert.NotNil(t, logger)
		assert.IsType(t, &zerolog.Logger{}, logger, "Return the logger previously set")
		assert.Equal(t, logger.GetLevel(), zerolog.DebugLevel)
	})

	t.Run("Test Get context when a value different from zerolog.Logger was set", func(t *testing.T) {
		ctx := context.TODO()

		ctx = context.WithValue(ctx, loggerKey{}, "Am I a logger?")
		logger := Get(ctx)

		assert.NotNil(t, logger)
		assert.IsType(t, &zerolog.Logger{}, logger, "Return the default logger since given one was not a logger")
		assert.Equal(t, logger.GetLevel(), zerolog.InfoLevel)
	})
}

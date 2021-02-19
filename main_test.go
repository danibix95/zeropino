package zerologmia

import (
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"
)

func TestInit(t *testing.T) {
	t.Run("Return default logger when initialized", func(t *testing.T) {
		logger, err := Init()

		assert.Assert(t, err == nil, "Error getting default logger")
		assert.Equal(t, logger.GetLevel(), zerolog.InfoLevel, "Default level value")

		logger.Info().Msg("Hello Mia!")
	})
}

package zerologmia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"

	pinoModel "github.com/danibix95/zerolog-mia/internal"
)

type miaLog struct {
	Level    string      `json:"level,omitempty"`
	Pid      int         `json:"pid,omitempty"`
	Hostname string      `json:"hostname,omitempty"`
	Time     int         `json:"time,omitempty"`
	Msg      string      `json:"msg,omitempty"`
	Stack    interface{} `json:"error,omitempty"`
}

func TestInit(t *testing.T) {
	message := "Hello Mia!"

	t.Run("Initialize Default Logger", func(t *testing.T) {
		logger, err := Init(InitOptions{})

		assert.NilError(t, err)
		assert.Assert(t, err == nil, "Error getting default logger")
		assert.Equal(t, logger.GetLevel(), zerolog.InfoLevel, "Default level value")

		logger.Info().Msg(message)
	})

	t.Run("Initialize a Logger with a writer", func(t *testing.T) {
		// passing a buffer allows to read what the logger is actually outputting
		out := &bytes.Buffer{}
		logger, err := Init(InitOptions{Writer: out})

		assert.NilError(t, err, "No init error should be encountered")
		assert.Equal(t, logger.GetLevel(), zerolog.InfoLevel, "Default level value")

		logger.Info().Msg(message)

		result := miaLog{}
		assert.NilError(t, json.Unmarshal(out.Bytes(), &result))

		assert.Equal(t, result.Level, string(pinoModel.PinoInfo), "Message level is correctly converted")
		assert.Equal(t, result.Msg, message)
		assert.Equal(t, len(strconv.Itoa(result.Time)), 10, "Time is an Unix timestamp")
	})

	t.Run("Initialize a Logger with custom log level", func(t *testing.T) {
		// passing a buffer allows to read what the logger is actually outputting
		out := &bytes.Buffer{}
		logger, err := Init(InitOptions{
			Level:  "error",
			Writer: out,
		})

		assert.NilError(t, err, "No init error should be encountered")
		assert.Equal(t, logger.GetLevel(), zerolog.ErrorLevel, "Default level value")

		message := "Hello Mia!"
		logger.Info().Msg(message)

		assert.Equal(t, out.Len(), 0, "No output should be produced due to msg logged using lower level than enabled one")
	})

	t.Run("Initialize a Logger with unrecognized log level", func(t *testing.T) {
		// passing a buffer allows to read what the logger is actually outputting
		out := &bytes.Buffer{}
		levelString := "custom"
		logger, err := Init(InitOptions{
			Level:  "custom",
			Writer: out,
		})

		var emptyPointer *zerolog.Logger
		assert.Error(t, err, fmt.Sprintf("level %s is not recognized", levelString))
		assert.Equal(t, logger, emptyPointer)
	})

	t.Run("Initialize a Logger with timestamp in milliseconds", func(t *testing.T) {
		// passing a buffer allows to read what the logger is actually outputting
		out := &bytes.Buffer{}
		logger, err := Init(InitOptions{
			Writer:    out,
			Level:     "warn",
			UseTimeMs: true,
		})

		assert.NilError(t, err, "No init error should be encountered")
		assert.Equal(t, logger.GetLevel(), zerolog.WarnLevel, "Default level value")

		logger.Warn().Msg(message)

		result := miaLog{}
		assert.NilError(t, json.Unmarshal(out.Bytes(), &result))

		assert.Equal(t, result.Level, string(pinoModel.PinoWarn), "Message level is correctly converted")
		assert.Equal(t, result.Msg, message)
		assert.Equal(t, len(strconv.Itoa(result.Time)), 13, "Time is an Unix timestamp in milliseconds")
	})
}

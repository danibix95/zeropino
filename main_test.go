package zeropino

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"

	pino "github.com/danibix95/zeropino/internal/model"
)

type miaLog struct {
	Level    string      `json:"level,omitempty"`
	Pid      int         `json:"pid,omitempty"`
	Hostname string      `json:"hostname,omitempty"`
	Time     int         `json:"time,omitempty"`
	Msg      string      `json:"msg,omitempty"`
	Stack    interface{} `json:"error,omitempty"`
}

const message = "Hello Mia!"
const unixTimestampLen = 10
const unixTimestampMsLen = 13

func TestInit(t *testing.T) {
	t.Run("Initialize Default Logger", func(t *testing.T) {
		logger := InitDefault()

		assert.Equal(t, logger.GetLevel(), zerolog.InfoLevel, "Default level value")
		logger.Info().Msg(message)
	})

	t.Run("Initialize Logger without user options", func(t *testing.T) {
		logger, err := Init(InitOptions{})

		verifyInit(t, logger, err, zerolog.InfoLevel)
		logger.Info().Msg(message)
	})

	t.Run("Initialize a Logger with a writer", func(t *testing.T) {
		// passing a buffer allows to read what the logger is actually outputting
		out := &bytes.Buffer{}
		logger, err := Init(InitOptions{Writer: out})

		verifyInit(t, logger, err, zerolog.InfoLevel)
		logger.Info().Msg(message)

		result := miaLog{}
		assert.NilError(t, json.Unmarshal(out.Bytes(), &result))

		verifyLog(t, &result, message, string(pino.Info), unixTimestampLen)
	})

	t.Run("Initialize a Logger with custom log level", func(t *testing.T) {
		// passing a buffer allows to read what the logger is actually outputting
		out := &bytes.Buffer{}
		logger, err := Init(InitOptions{
			Level:  "error",
			Writer: out,
		})

		verifyInit(t, logger, err, zerolog.ErrorLevel)
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

		verifyInit(t, logger, err, zerolog.WarnLevel)
		logger.Warn().Msg(message)

		result := miaLog{}
		assert.NilError(t, json.Unmarshal(out.Bytes(), &result))

		verifyLog(t, &result, message, string(pino.Warn), unixTimestampMsLen)
	})
}

func verifyLog(t testing.TB, log *miaLog, msg, level string, timeLen int) {
	t.Helper()

	assert.Equal(t, level, log.Level, "Message level is correctly converted")
	assert.Equal(t, msg, log.Msg)
	assert.Equal(t, timeLen, len(strconv.Itoa(log.Time)), "Time is an Unix timestamp of specified length")
}

func verifyInit(t testing.TB, logger *zerolog.Logger, err error, expected zerolog.Level) {
	t.Helper()

	assert.NilError(t, err, "No init error should be encountered")
	assert.Equal(t, expected, logger.GetLevel(), "Level value")
}

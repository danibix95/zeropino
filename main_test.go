/*
 *   Copyright 2021 Daniele Bissoli
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

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

const message = "Follow the spiders!"
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

		verifyLog(t, &result, message, string(pino.Info), unixTimestampMsLen)
	})

	t.Run("Initialize a Logger with custom log level", func(t *testing.T) {
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

	t.Run("Initialize a Logger with timestamp in seconds instead of the milliseconds default", func(t *testing.T) {
		out := &bytes.Buffer{}
		logger, err := Init(InitOptions{
			Writer:        out,
			Level:         "warn",
			DisableTimeMs: true,
		})

		verifyInit(t, logger, err, zerolog.WarnLevel)
		logger.Warn().Msg(message)

		result := miaLog{}
		assert.NilError(t, json.Unmarshal(out.Bytes(), &result))

		verifyLog(t, &result, message, string(pino.Warn), unixTimestampLen)
	})
}

func BenchmarkZeropino(b *testing.B) {
	logger, _ := Init(InitOptions{Level: "trace"})

	for i := 0; i < b.N; i++ {
		logger.Warn().Msg(message)
	}
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

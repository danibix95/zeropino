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
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	pino "github.com/danibix95/zeropino/internal/model"
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

	logLevel, err := pino.ParseLevel(options.Level)
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
	zerolog.LevelFieldMarshalFunc = pino.ConvertLevel
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

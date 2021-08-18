/*
 * Copyright 2019 Mia srl
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package std

import (
	"context"

	"github.com/rs/zerolog"

	zp "github.com/danibix95/zeropino"
)

type loggerKey struct{}

var defaultLogger *zerolog.Logger = zp.InitDefault()

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

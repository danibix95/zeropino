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

package gorillamux

import (
	"context"
	"testing"

	zp "github.com/danibix95/zeropino"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestWithLogger(t *testing.T) {
	t.Run("Test WithLogger when no logger is given", func(t *testing.T) {
		ctx := context.TODO()

		ctx = WithLogger(ctx, nil)
		require.Nil(t, ctx.Value(loggerKey{}))
	})

	t.Run("Test WithLogger when a logger is given", func(t *testing.T) {
		ctx := context.TODO()
		log := zp.InitDefault()

		ctx = WithLogger(ctx, log)
		contextLog := ctx.Value(loggerKey{})
		require.NotNil(t, contextLog)
		require.IsType(t, &zerolog.Logger{}, contextLog)
	})
}

func TestGet(t *testing.T) {
	t.Run("Test Get context when no logger was set", func(t *testing.T) {
		ctx := context.TODO()
		logger := Get(ctx)

		require.NotNil(t, logger)
		require.IsType(t, &zerolog.Logger{}, logger, "Return the default logger since no logger was previously provided")
		require.Equal(t, logger.GetLevel(), zerolog.InfoLevel)
	})

	t.Run("Test Get context when a logger was set", func(t *testing.T) {
		ctx := context.TODO()
		contextLogger, err := zp.Init(zp.InitOptions{Level: "debug"})
		require.Nil(t, err)

		ctx = context.WithValue(ctx, loggerKey{}, contextLogger)
		logger := Get(ctx)

		require.NotNil(t, logger)
		require.IsType(t, &zerolog.Logger{}, logger, "Return the logger previously set")
		require.Equal(t, logger.GetLevel(), zerolog.DebugLevel)
	})

	t.Run("Test Get context when a value different from zerolog.Logger was set", func(t *testing.T) {
		ctx := context.TODO()

		ctx = context.WithValue(ctx, loggerKey{}, "Am I a logger?")
		logger := Get(ctx)

		require.NotNil(t, logger)
		require.IsType(t, &zerolog.Logger{}, logger, "Return the default logger since given one was not a logger")
		require.Equal(t, logger.GetLevel(), zerolog.InfoLevel)
	})
}

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

package fiber

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const testPath = "/no-log"

func TestReqLogger(t *testing.T) {
	app := fiber.New()

	app.Get(testPath, func(c *fiber.Ctx) error {
		logger := ReqLogger(c)

		require.IsType(t, &zerolog.Logger{}, logger)
		require.NotNil(t, logger, "Return the default logger since no one was set before")
		require.Equal(t, zerolog.InfoLevel, logger.GetLevel(), "default logger has info level")

		return c.JSON(fiber.Map{"msg": "Hello, World!"})
	})

	request := httptest.NewRequest("GET", testPath, nil)
	app.Test(request, requestTimeoutMs)
}

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
	zp "github.com/danibix95/zeropino"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const loggerKey = "request-logger"

func ReqLogger(c *fiber.Ctx) *zerolog.Logger {
	if logger, ok := c.Locals(loggerKey).(*zerolog.Logger); ok {
		return logger
	}
	return zp.InitDefault()
}

func WithLogger(c *fiber.Ctx, l *zerolog.Logger) {
	c.Locals(loggerKey, l)
}

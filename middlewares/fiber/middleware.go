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
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const million float64 = 1000000
const (
	contentLengthHeaderKey = "Content-Length"
	userAgentHeaderKey     = "User-Agent"
	requestIDHeaderKey     = "X-Request-ID"
	forwardedHostHeaderKey = "X-Forwarded-Host"
	forwardedForHeaderKey  = "X-Forwarded-For"
)

// RequestLogger is a fiber middleware to log all requests with a custom zerolog Logger
// It logs both when requests arrive and when they are completed, adding request latency
func RequestLogger(l *zerolog.Logger) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		sub := l.With().Str("reqId", extractRequestID(l, c)).Logger()
		WithLogger(c, &sub)

		logIncoming(c)
		err := c.Next()
		logCompleted(c, start)

		return err
	}
}

func logIncoming(c *fiber.Ctx) {
	ReqLogger(c).Trace().
		Dict("http", zerolog.Dict().
			Dict("request", zerolog.Dict().
				Str("method", c.Method()).
				Dict("userAgent", zerolog.Dict().
					Str("original", c.Get(userAgentHeaderKey)),
				),
			),
		).
		Dict("url", zerolog.Dict().
			Str("path", c.Path()),
		).
		Dict("host", zerolog.Dict().
			Str("hostname", removePort(c.Hostname())).
			Str("forwardedHost", c.Get(forwardedHostHeaderKey)).
			Str("ip", c.Get(forwardedForHeaderKey)),
		).
		Msg("incoming request")
}

func logCompleted(c *fiber.Ctx, start time.Time) {
	ReqLogger(c).Info().
		Dict("http", zerolog.Dict().
			Dict("request", zerolog.Dict().
				Str("method", c.Method()).
				Dict("userAgent", zerolog.Dict().
					Str("original", c.Get(userAgentHeaderKey)),
				),
			).
			Dict("response", zerolog.Dict().
				Int("statusCode", c.Response().StatusCode()).
				Dict("body", zerolog.Dict().
					Int("bytes", c.Response().Header.ContentLength()),
				),
			),
		).
		Dict("url", zerolog.Dict().
			Str("path", c.Path()),
		).
		Dict("host", zerolog.Dict().
			Str("hostname", removePort(c.Hostname())).
			Str("forwardedHost", c.Get(forwardedHostHeaderKey)).
			Str("ip", c.Get(forwardedForHeaderKey)),
		).
		Float64("responseTime", float64(time.Since(start).Nanoseconds())/million).
		Msg("request completed")
}

func removePort(host string) string {
	return strings.Split(host, ":")[0]
}

func extractRequestID(logger *zerolog.Logger, c *fiber.Ctx) string {
	if requestID := c.Get(requestIDHeaderKey); requestID != "" {
		return requestID
	}

	// Generate a random uuid string. e.g. 16c9c1f2-c001-40d3-bbfe-48857367e7b5
	requestID, err := uuid.NewRandom()
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error generating request id")
	}

	return requestID.String()
}

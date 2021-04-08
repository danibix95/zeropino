package fiber

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func RequestLogger(logger *zerolog.Logger) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		child := logger.With().Str("reqId", c.Get("x-request-id")).Logger()
		c.Locals("log", &child)
		c.Locals("startTime", time.Now())

		child.Trace().
			Dict("http", zerolog.Dict().
				Dict("request", zerolog.Dict().
					Str("method", c.Method()).
					Dict("userAgent", zerolog.Dict().
						Str("original", c.Get("user-agent")),
					),
				),
			).
			Dict("url", zerolog.Dict().
				Str("path", c.Path()),
			).
			Dict("host", zerolog.Dict().
				Str("hostname", removePort(c.Hostname())).
				Str("forwardedHost", c.Get("x-forwarded-host")).
				Str("ip", c.Get("x-forwarded-for")),
			).
			Msg("incoming request")

		return c.Next()
	}
}

func ResponseLogger(logger *zerolog.Logger) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		child, ok := c.Locals("log").(*zerolog.Logger)
		if !ok {
			child = logger
		}

		child.Info().
			Dict("http", zerolog.Dict().
				Dict("request", zerolog.Dict().
					Str("method", c.Method()).
					Dict("userAgent", zerolog.Dict().
						Str("original", c.Get("user-agent")),
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
				Str("forwardedHost", c.Get("x-forwarded-host")).
				Str("ip", c.Get("x-forwarded-for")),
			).
			Float64("responseTime", getResponseTime(c.Locals("startTime"))).
			Msg("request completed")

		return c.Next()
	}
}

func removePort(host string) string {
	return strings.Split(host, ":")[0]
}

func getResponseTime(start interface{}) float64 {
	if startTime, isTime := start.(time.Time); isTime {
		return float64(time.Since(startTime).Nanoseconds()) / 1000000
	}

	// do not provide any relevant information about response time
	// in case the start time is not available
	return 0
}

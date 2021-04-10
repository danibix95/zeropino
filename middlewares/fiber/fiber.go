package fiber

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	mux "github.com/danibix95/zeropino/middlewares/gorillamux"
)

const million float64 = 1000000

const (
	forwardedHostHeaderKey = "X-Forwarded-Host"
	forwardedForHeaderKey  = "X-Forwarded-For"
)

func LogMiddleware(l *zerolog.Logger) func(*fiber.Ctx) error {
	return adaptor.HTTPMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			child := l.With().Str("reqId", r.Header.Get("X-Request-Id")).Logger()
			ctx := mux.WithLogger(r.Context(), &child)

			logIncoming(ctx, r)

			h.ServeHTTP(w, r.WithContext(ctx))

			logOutgoing(ctx, r, w, start)
		})
	})
}

func logIncoming(ctx context.Context, r *http.Request) {
	mux.Get(ctx).Trace().
		Dict("http", zerolog.Dict().
			Dict("request", zerolog.Dict().
				Str("method", r.Method).
				Dict("userAgent", zerolog.Dict().
					Str("original", r.Header.Get("User-Agent")),
				),
			),
		).
		Dict("url", zerolog.Dict().
			Str("path", r.URL.RequestURI()),
		).
		Dict("host", zerolog.Dict().
			Str("hostname", removePort(r.Host)).
			Str("forwardedHost", r.Header.Get(forwardedHostHeaderKey)).
			Str("ip", r.Header.Get(forwardedForHeaderKey)),
		).
		Msg("incoming request")
}

func logOutgoing(ctx context.Context, r *http.Request, w http.ResponseWriter, start time.Time) {
	mux.Get(ctx).Info().
		Dict("http", zerolog.Dict().
			Dict("request", zerolog.Dict().
				Str("method", r.Method).
				Dict("userAgent", zerolog.Dict().
					Str("original", r.Header.Get("User-Agent")),
				),
			).
			Dict("response", zerolog.Dict().
				Int("statusCode", 200).
				Dict("body", zerolog.Dict().
					Int("bytes", getBodyLength(w)),
				),
			),
		).
		Dict("url", zerolog.Dict().
			Str("path", r.URL.RequestURI()),
		).
		Dict("host", zerolog.Dict().
			Str("hostname", removePort(r.Host)).
			Str("forwardedHost", r.Header.Get(forwardedHostHeaderKey)).
			Str("ip", r.Header.Get(forwardedForHeaderKey)),
		).
		Float64("responseTime", float64(time.Since(start).Milliseconds())).
		Msg("request completed")
}

// RequestLogger logs details about incoming requests with a trace level
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

// ResponseLogger logs details about outgoing responses with an info level.
// In order to work properly with fiber, previous routes must call fiber.Ctx.Next() method
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
		return float64(time.Since(startTime).Nanoseconds()) / million
	}

	// do not provide any relevant information about response time
	// in case the start time is not available
	return 0
}

// func getReqID(logger *zerolog.Logger, headers http.Header) string {
// 	if requestID := headers.Get("X-Request-Id"); requestID != "" {
// 		return requestID
// 	}

// 	// Generate a random uuid string. e.g. 16c9c1f2-c001-40d3-bbfe-48857367e7b5
// 	requestIDRaw, err := uuid.NewRandom()
// 	if err != nil {
// 		logger.Error().Stack().Err(err).Msg("error generating request id")
// 	}

// 	requestID := requestIDRaw.String()
// 	logger.Trace().Str("reqId", requestID).Msg("generated request id")

// 	return requestID
// }

func getBodyLength(w http.ResponseWriter) int {
	if content := w.Header().Get("Content-Length"); content != "" {
		if length, err := strconv.Atoi(content); err == nil {
			return length
		}
	}
	return 0
}

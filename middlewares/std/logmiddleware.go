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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const million float64 = 1000000
const (
	contentLengthHeaderKey = "Content-Length"
	requestIDHeaderKey     = "X-Request-ID"
	forwardedHostHeaderKey = "X-Forwarded-Host"
	forwardedForHeaderKey  = "X-Forwarded-For"
)

// RequestLogger is a gorilla/mux middleware to log all requests with zeropino
// It logs the incoming request and when request is completed, adding latency of the request
func RequestLogger(logger *zerolog.Logger, excludedPrefix []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := getReqID(logger, r.Header)
			reqLogger := logger.With().Str("reqId", requestID).Logger()
			ctx := WithLogger(r.Context(), &reqLogger)
			customRW := readableResponseWriter{writer: w, statusCode: http.StatusOK}

			// Skip logging for excluded routes
			for _, prefix := range excludedPrefix {
				if strings.HasPrefix(r.URL.RequestURI(), prefix) {
					next.ServeHTTP(&customRW, r.WithContext(ctx))
					return
				}
			}

			logIncoming(ctx, r)

			next.ServeHTTP(&customRW, r.WithContext(ctx))

			logOutgoing(ctx, r, &customRW, start)
		})
	}
}

func logIncoming(ctx context.Context, r *http.Request) {
	Get(ctx).Trace().
		Dict("http", zerolog.Dict().
			Dict("request", zerolog.Dict().
				Str("method", r.Method).
				Dict("userAgent", zerolog.Dict().
					Str("original", r.UserAgent()),
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

func logOutgoing(ctx context.Context, r *http.Request, myw *readableResponseWriter, start time.Time) {
	Get(ctx).Info().
		Dict("http", zerolog.Dict().
			Dict("request", zerolog.Dict().
				Str("method", r.Method).
				Dict("userAgent", zerolog.Dict().
					Str("original", r.UserAgent()),
				),
			).
			Dict("response", zerolog.Dict().
				Int("statusCode", myw.statusCode).
				Dict("body", zerolog.Dict().
					Int("bytes", getBodyLength(myw)),
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
		Float64("responseTime", float64(time.Since(start).Nanoseconds())/million).
		Msg("request completed")
}

func removePort(host string) string {
	return strings.Split(host, ":")[0]
}

func getBodyLength(customRW *readableResponseWriter) int {
	if content := customRW.Header().Get(contentLengthHeaderKey); content != "" {
		if length, err := strconv.Atoi(content); err == nil {
			return length
		}
	}
	return customRW.Length()
}

func getReqID(logger *zerolog.Logger, headers http.Header) string {
	if requestID := headers.Get(requestIDHeaderKey); requestID != "" {
		return requestID
	}

	// Generate a random uuid string. e.g. 16c9c1f2-c001-40d3-bbfe-48857367e7b5
	requestIDRaw, err := uuid.NewRandom()
	if err != nil {
		logger.Error().Stack().Err(err).Msg("error generating request id")
	}

	requestID := requestIDRaw.String()
	logger.Trace().Str("reqId", requestID).Msg("generated request id")

	return requestID
}

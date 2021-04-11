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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	zp "github.com/danibix95/zeropino"
	pino "github.com/danibix95/zeropino/internal/model"
	zpm "github.com/danibix95/zeropino/middlewares"
)

const (
	hostname    = "kind-host"
	userAgent   = "goHttp"
	requestPath = "/cool-path"
	clientHost  = "client-host"
	requestID   = "req-id"
	method      = "GET"
)

const noContentLength = 0

// for tests purposes
const requestTimeoutMs = 500

var defaultRequestURL = fmt.Sprintf("http://%s:3000%s", hostname, requestPath)

type logFields struct {
	Level         string
	Msg           string
	Method        string
	RequestID     string
	Path          string
	Hostname      string
	ForwardedHost string
	Original      string
	IP            string
	StatusCode    int
	Bytes         int
}

type expectedBodyData struct {
	Bytes int
}

func TestRequestLogger(t *testing.T) {
	t.Run("trace log level - log both incoming request and response details", func(t *testing.T) {
		// use a buffer to avoid printing the logs on screen during tests
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "trace", Writer: buffer})

		middleware := RequestLogger(logger)
		app := createFiberApp(t, middleware, fiber.StatusOK, noContentLength)

		request := getRequestWithHeaders(method, defaultRequestURL, nil)

		response, err := app.Test(request, requestTimeoutMs)
		require.Nil(t, err)
		response.Body.Close()

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		require.Equal(t, 2, len(entries))

		expectedRequestLog := logFields{
			Level:         string(pino.Trace),
			Msg:           "incoming request",
			RequestID:     requestID,
			Method:        method,
			Original:      userAgent,
			Path:          requestPath,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			IP:            removePort(request.RemoteAddr),
		}
		assertRequestLog(t, expectedRequestLog, bytes.NewBufferString(entries[0]))

		expectedResponseLog := logFields{
			Level:         string(pino.Info),
			Msg:           "request completed",
			RequestID:     requestID,
			Method:        method,
			Original:      userAgent,
			Path:          requestPath,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			StatusCode:    fiber.StatusOK,
			IP:            removePort(request.RemoteAddr),
		}
		assertResponseLog(t, expectedResponseLog, bytes.NewBufferString(entries[1]))
	})

	t.Run("log level is debug or info - log only response details", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "debug", Writer: buffer})

		middleware := RequestLogger(logger)
		app := createFiberApp(t, middleware, fiber.StatusTeapot, noContentLength)

		request := getRequestWithHeaders(method, defaultRequestURL, nil)

		response, err := app.Test(request, requestTimeoutMs)
		require.Nil(t, err)
		response.Body.Close()

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		require.Equal(t, 1, len(entries))

		expected := logFields{
			Level:         string(pino.Info),
			Msg:           "request completed",
			RequestID:     requestID,
			Method:        method,
			Original:      userAgent,
			Path:          requestPath,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			StatusCode:    fiber.StatusTeapot,
			IP:            removePort(request.RemoteAddr),
		}
		assertResponseLog(t, expected, buffer)
	})

	t.Run("Content-Length is logged in case the header is found", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "debug", Writer: buffer})

		bodySize := 312
		middleware := RequestLogger(logger)
		app := createFiberApp(t, middleware, fiber.StatusTeapot, bodySize)

		request := getRequestWithHeaders(method, defaultRequestURL, nil)

		response, err := app.Test(request, requestTimeoutMs)
		require.Nil(t, err)
		response.Body.Close()

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		require.Equal(t, 1, len(entries))

		expected := logFields{
			Level:         string(pino.Info),
			Msg:           "request completed",
			RequestID:     requestID,
			Method:        method,
			Original:      userAgent,
			Path:          requestPath,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			StatusCode:    fiber.StatusTeapot,
			IP:            removePort(request.RemoteAddr),
			Bytes:         bodySize,
		}
		assertResponseLog(t, expected, buffer)
	})

	t.Run("log level higher than info - no log produced", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "warn", Writer: buffer})

		middleware := RequestLogger(logger)
		app := createFiberApp(t, middleware, fiber.StatusOK, noContentLength)

		request := httptest.NewRequest(method, defaultRequestURL, nil)

		response, err := app.Test(request, requestTimeoutMs)
		require.Nil(t, err)
		response.Body.Close()

		require.Equal(t, 0, buffer.Len(), "no log output should be produced")
	})
}

func BenchmarkRequestLogger(b *testing.B) {
	logger, _ := zp.Init(zp.InitOptions{Level: "trace"})

	middleware := RequestLogger(logger)
	app := createFiberApp(b, middleware, fiber.StatusOK, noContentLength)

	request := getRequestWithHeaders(method, defaultRequestURL, nil)

	for i := 0; i < b.N; i++ {
		response, _ := app.Test(request, requestTimeoutMs)
		response.Body.Close()
	}
}

func assertRequestLog(t testing.TB, expected logFields, actual *bytes.Buffer) *zpm.LogFormat {
	t.Helper()

	var logOutput zpm.LogFormat
	err := json.Unmarshal(actual.Bytes(), &logOutput)

	require.Nil(t, err)
	require.Equal(t, expected.Level, logOutput.Level)
	require.Equal(t, expected.Msg, logOutput.Msg)
	require.Equal(t, expected.RequestID, logOutput.RequestID)
	require.Equal(t, expected.Method, logOutput.HTTP.Request.Method)
	require.Equal(t, expected.Original, logOutput.HTTP.Request.UserAgent["original"])
	require.Equal(t, expected.Path, logOutput.URL.Path)
	require.Equal(t, expected.Hostname, logOutput.Host.Hostname)
	require.Equal(t, expected.ForwardedHost, logOutput.Host.ForwardedHost)
	require.Equal(t, expected.IP, logOutput.Host.IP)

	return &logOutput
}

func assertResponseLog(t testing.TB, expected logFields, actual *bytes.Buffer) {
	t.Helper()

	logOutput := assertRequestLog(t, expected, actual)

	require.Equal(t, expected.StatusCode, logOutput.HTTP.Response.StatusCode)
	require.Greater(t, logOutput.ResponseTime, 0.0, "Response time is not null")

	if expected.Bytes >= 0 {
		binaryData, _ := json.Marshal(logOutput.HTTP.Response.Body)
		var structBody expectedBodyData
		require.Nil(t, json.Unmarshal(binaryData, &structBody))
		require.Equal(t, expected.Bytes, structBody.Bytes, "Body size is reported when set")
	}
}

func getRequestWithHeaders(method, path string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, path, body)
	ip := removePort(request.RemoteAddr)
	request.Header.Set(requestIDHeaderKey, requestID)
	request.Header.Set(userAgentHeaderKey, userAgent)
	request.Header.Set(forwardedForHeaderKey, ip)
	request.Header.Set(forwardedHostHeaderKey, clientHost)

	return request
}

func createFiberApp(t testing.TB, middleware func(*fiber.Ctx) error, statusCode int, contentLength int) *fiber.App {
	t.Helper()
	app := fiber.New()

	// apply the middleware
	app.Use(middleware)

	app.Get(requestPath, func(c *fiber.Ctx) error {
		if contentLength > 0 {
			c.Set(contentLengthHeaderKey, strconv.Itoa(contentLength))
		}
		return c.Status(statusCode).JSON(fiber.Map{"msg": "Hello, World!"})
	})

	return app
}

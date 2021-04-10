package fiber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gotest.tools/assert"

	zp "github.com/danibix95/zeropino"
	pino "github.com/danibix95/zeropino/internal/model"
	types "github.com/danibix95/zeropino/middlewares"
)

const hostname = "my-host"
const userAgent = "goHttp"

// const bodyBytes = 21
const requestPath = "/my-req"
const clientHost = "client-host"
const requestID = "req-id"

const baseURL = "http://my-host:3000"
const method = "GET"

// request timemout for tests in milliseconds
const testTimeout = 500

const doNotCheckBytes = -1

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
	Bytes         int
	StatusCode    int
}

type expectedBodyData struct {
	Bytes int
}

func TestLogMiddleware(t *testing.T) {
	t.Run("when trace log level, it logs both incoming request and outgoing response details", func(t *testing.T) {
		// use a buffer to avoid printing the logs on screen during tests
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "trace", Writer: buffer})

		middleware := LogMiddleware(logger)
		app := createFiberApp(t, middleware)

		request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)
		ip := removePort(request.RemoteAddr)
		request.Header.Set("X-Request-Id", requestID)
		request.Header.Set("User-Agent", userAgent)
		request.Header.Set("X-Forwarded-For", ip)
		request.Header.Set("X-Forwarded-Host", clientHost)

		response, err := app.Test(request, testTimeout)
		assert.NilError(t, err)
		response.Body.Close()

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 2)

		expectedRequestLog := logFields{
			Level:         string(pino.Trace),
			Msg:           "incoming request",
			RequestID:     requestID,
			Method:        method,
			Original:      userAgent,
			Path:          requestPath,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			IP:            ip,
			Bytes:         doNotCheckBytes,
		}
		assertLog(t, expectedRequestLog, bytes.NewBufferString(entries[0]))

		expectedResponseLog := logFields{
			Level:         string(pino.Info),
			Msg:           "request completed",
			RequestID:     requestID,
			Method:        method,
			Original:      userAgent,
			Path:          requestPath,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			IP:            ip,
			Bytes:         doNotCheckBytes,
		}
		assertLog(t, expectedResponseLog, bytes.NewBufferString(entries[1]))
	})

	t.Run("when log level is debug or higher, it logs only outgoing response details", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "debug", Writer: buffer})

		middleware := LogMiddleware(logger)
		app := createFiberApp(t, middleware)

		request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)
		ip := removePort(request.RemoteAddr)
		request.Header.Set("X-Request-Id", requestID)
		request.Header.Set("User-Agent", userAgent)
		request.Header.Set("X-Forwarded-For", ip)
		request.Header.Set("X-Forwarded-Host", clientHost)

		response, err := app.Test(request, testTimeout)
		assert.NilError(t, err)
		response.Body.Close()

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 1)

		expected := logFields{
			Level:         string(pino.Info),
			Msg:           "request completed",
			RequestID:     requestID,
			Method:        method,
			Original:      userAgent,
			Path:          requestPath,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			IP:            ip,
			Bytes:         doNotCheckBytes,
		}
		assertLog(t, expected, buffer)
	})

	t.Run("when log level is higher than info, it does not log anything", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "warn", Writer: buffer})

		middleware := LogMiddleware(logger)
		app := createFiberApp(t, middleware)

		request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)

		response, err := app.Test(request, testTimeout)
		assert.NilError(t, err)
		response.Body.Close()

		assert.Equal(t, 0, buffer.Len(), "No log output should be produced")
	})
}

func BenchmarkFiberMiddleware(b *testing.B) {
	logger, _ := zp.Init(zp.InitOptions{Level: "trace"})

	middleware := LogMiddleware(logger)
	app := createFiberApp(b, middleware)

	request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)
	ip := removePort(request.RemoteAddr)
	request.Header.Set("X-Request-Id", requestID)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("X-Forwarded-For", ip)
	request.Header.Set("X-Forwarded-Host", clientHost)

	for i := 0; i < b.N; i++ {
		response, _ := app.Test(request, testTimeout)
		response.Body.Close()
	}
}

func assertLog(t testing.TB, expected logFields, actual *bytes.Buffer) {
	t.Helper()

	var logOutput types.MiddlewareLog
	err := json.Unmarshal(actual.Bytes(), &logOutput)

	assert.NilError(t, err)
	assert.Equal(t, expected.Level, logOutput.Level)
	assert.Equal(t, expected.Msg, logOutput.Msg)
	assert.Equal(t, expected.RequestID, logOutput.RequestID)
	assert.Equal(t, expected.Method, logOutput.HTTP.Request.Method)
	assert.Equal(t, expected.Original, logOutput.HTTP.Request.UserAgent["original"])
	assert.Equal(t, expected.Path, logOutput.URL.Path)
	assert.Equal(t, expected.Hostname, logOutput.Host.Hostname)
	assert.Equal(t, expected.ForwardedHost, logOutput.Host.ForwardedHost)
	assert.Equal(t, expected.IP, logOutput.Host.IP)

	if expected.Bytes >= 0 {
		binaryData, _ := json.Marshal(logOutput.HTTP.Response.Body)
		var structBody expectedBodyData
		assert.NilError(t, json.Unmarshal(binaryData, &structBody))
		assert.Equal(t, structBody.Bytes, expected.Bytes, "Body size is reported when set")
	}
}

func createFiberApp(t testing.TB, middleware func(*fiber.Ctx) error) *fiber.App {
	t.Helper()

	app := fiber.New()

	// apply the middleware
	app.Use(middleware)

	app.Get(requestPath, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "Hello, World!"})
	})

	return app
}

// func fiberAppAfter(middleware func(*fiber.Ctx) error) *fiber.App {
// 	app := fiber.New()

// 	app.Get(requestPath, func(c *fiber.Ctx) error {
// 		c.Response().Header.Set("Content-Length", strconv.Itoa(bodyBytes))
// 		c.JSON(fiber.Map{"msg": "Hello, World!"})

// 		return c.Next()
// 	})

// 	// apply the middleware
// 	app.Use(middleware)

// 	return app
// }

// func getRequestWithHeaders(method, path string, body io.Reader) *http.Request {
// 	request:= httptest.NewRequest(method, path, body)
// 	ip := removePort(request.RemoteAddr)
// 	request.Header.Set("X-Request-Id", requestID)
// 	request.Header.Set("User-Agent", userAgent)
// 	request.Header.Set("X-Forwarded-For", ip)
// 	request.Header.Set("X-Forwarded-Host", clientHost)

// 	return request
// }

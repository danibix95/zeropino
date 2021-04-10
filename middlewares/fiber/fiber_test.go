package fiber

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strconv"
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
const bodyBytes = 21
const requestPath = "/my-req"
const clientHost = "client-host"
const requestID = "req-id"

const baseURL = "http://my-host:3000"
const method = "GET"

// request timemout for tests in milliseconds
const testTimeout = 500

const doNotCheckBytes = -1

type logFields struct {
	Level             string
	Msg               string
	Method            string
	RequestID         string
	Path              string
	Hostname          string
	ForwardedHost     string
	Original          string
	IP                string
	Bytes             int
	StatusCode        int
	CheckResponseTime bool
}

type expectedBodyData struct {
	Bytes int
}

func TestRequestLogger(t *testing.T) {
	t.Run("when trace log level it logs the incoming request details", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "trace", Writer: buffer})

		middleware := RequestLogger(logger)

		app := fiberAppBefore(middleware)

		request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)
		ip := removePort(request.RemoteAddr)
		request.Header.Set("x-request-id", requestID)
		request.Header.Set("user-agent", userAgent)
		request.Header.Set("x-forwarded-for", ip)
		request.Header.Set("x-forwarded-host", clientHost)

		_, err := app.Test(request, testTimeout)
		assert.NilError(t, err)

		expected := logFields{
			Level:             string(pino.Trace),
			Msg:               "incoming request",
			RequestID:         requestID,
			Method:            method,
			Original:          userAgent,
			Path:              requestPath,
			Hostname:          hostname,
			ForwardedHost:     clientHost,
			IP:                ip,
			Bytes:             doNotCheckBytes,
			CheckResponseTime: false,
		}
		assertLog(t, expected, buffer)
	})

	t.Run("when log level is debug or higher it does not logs the incoming request details", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "info", Writer: buffer})

		middleware := RequestLogger(logger)

		app := fiberAppBefore(middleware)

		request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)

		_, err := app.Test(request, testTimeout)
		assert.NilError(t, err)

		assert.Equal(t, 0, buffer.Len(), "No log output should be produced")
	})
}

func TestResponseLogger(t *testing.T) {
	t.Run("logs the outgoing response details without responseTime", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "trace", Writer: buffer})

		middleware := ResponseLogger(logger)

		app := fiberAppAfter(middleware)

		request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)
		ip := removePort(request.RemoteAddr)
		request.Header.Set("x-request-id", requestID)
		request.Header.Set("user-agent", userAgent)
		request.Header.Set("x-forwarded-for", ip)
		request.Header.Set("x-forwarded-host", clientHost)

		_, err := app.Test(request, testTimeout)
		assert.NilError(t, err)

		expected := logFields{
			Level:             string(pino.Info),
			Msg:               "request completed",
			RequestID:         "",
			Method:            method,
			Original:          userAgent,
			Path:              requestPath,
			Hostname:          hostname,
			ForwardedHost:     clientHost,
			IP:                ip,
			Bytes:             bodyBytes,
			CheckResponseTime: false,
		}
		assertLog(t, expected, buffer)
	})

	t.Run("logs the outgoing response details with responseTime", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "trace", Writer: buffer})

		before := RequestLogger(logger)
		after := ResponseLogger(logger)

		app := fiberAppBeforeAfter(before, after)

		request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)
		ip := removePort(request.RemoteAddr)
		request.Header.Set("x-request-id", requestID)
		request.Header.Set("user-agent", userAgent)
		request.Header.Set("x-forwarded-for", ip)
		request.Header.Set("x-forwarded-host", clientHost)

		_, err := app.Test(request, testTimeout)
		assert.NilError(t, err)

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 2)

		expectedRequestLog := logFields{
			Level:             string(pino.Trace),
			Msg:               "incoming request",
			RequestID:         requestID,
			Method:            method,
			Original:          userAgent,
			Path:              requestPath,
			Hostname:          hostname,
			ForwardedHost:     clientHost,
			IP:                ip,
			Bytes:             doNotCheckBytes,
			CheckResponseTime: false,
		}
		assertLog(t, expectedRequestLog, bytes.NewBufferString(entries[0]))

		expectedResponseLog := logFields{
			Level:             string(pino.Info),
			Msg:               "request completed",
			RequestID:         requestID,
			Method:            method,
			Original:          userAgent,
			Path:              requestPath,
			Hostname:          hostname,
			ForwardedHost:     clientHost,
			IP:                ip,
			Bytes:             doNotCheckBytes,
			CheckResponseTime: true,
		}
		assertLog(t, expectedResponseLog, bytes.NewBufferString(entries[1]))
	})
}

func BenchmarkFiberMiddleware(b *testing.B) {
	logger, _ := zp.Init(zp.InitOptions{Level: "trace"})

	before := RequestLogger(logger)
	after := ResponseLogger(logger)

	app := fiberAppBeforeAfter(before, after)

	request := httptest.NewRequest(method, fmt.Sprintf("%s%s", baseURL, requestPath), nil)
	ip := removePort(request.RemoteAddr)
	request.Header.Set("x-request-id", requestID)
	request.Header.Set("user-agent", userAgent)
	request.Header.Set("x-forwarded-for", ip)
	request.Header.Set("x-forwarded-host", clientHost)

	for i := 0; i < b.N; i++ {
		app.Test(request, testTimeout)
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

	if expected.CheckResponseTime {
		assert.Check(t, logOutput.ResponseTime > 0, "Request takes some time to be processed")
	}

	if expected.Bytes >= 0 {
		binaryData, _ := json.Marshal(logOutput.HTTP.Response.Body)
		var structBody expectedBodyData
		assert.NilError(t, json.Unmarshal(binaryData, &structBody))

		assert.Equal(t, structBody.Bytes, expected.Bytes, "Request takes some time to be processed")
	}
}

func fiberAppBefore(middleware func(*fiber.Ctx) error) *fiber.App {
	app := fiber.New()

	// apply the middleware
	app.Use(middleware)

	app.Get(requestPath, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "Hello, World!"})
	})

	return app
}

func fiberAppAfter(middleware func(*fiber.Ctx) error) *fiber.App {
	app := fiber.New()

	app.Get(requestPath, func(c *fiber.Ctx) error {
		c.Response().Header.Set("Content-Length", strconv.Itoa(bodyBytes))
		c.JSON(fiber.Map{"msg": "Hello, World!"})

		return c.Next()
	})

	// apply the middleware
	app.Use(middleware)

	return app
}

func fiberAppBeforeAfter(before, after func(*fiber.Ctx) error) *fiber.App {
	app := fiber.New()

	// request logger
	app.Use(before)

	app.Get(requestPath, func(c *fiber.Ctx) error {
		c.JSON(fiber.Map{"msg": "Hello, World!"})

		return c.Next()
	})

	// response logger
	app.Use(after)

	return app
}

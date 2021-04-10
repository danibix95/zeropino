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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"gotest.tools/assert"

	zp "github.com/danibix95/zeropino"
	pino "github.com/danibix95/zeropino/internal/model"
	types "github.com/danibix95/zeropino/middlewares"
)

const hostname = "my-host.com"
const port = "3030"
const userAgent = "goHttp"
const bodyBytes int = 0
const path = "/my-req"
const clientHost = "client-host"

var ip string
var defaultRequestPath = fmt.Sprintf("http://%s:%s/my-req", hostname, port)

type ExpectedLogFields struct {
	Level     string
	RequestID string
	Message   string
}

type ExpectedIncomingLogFields struct {
	Method        string
	Path          string
	Hostname      string
	ForwardedHost string
	Original      string
	IP            string
}

type ExpectedOutcomingLogFields struct {
	Method        string
	Path          string
	Hostname      string
	ForwardedHost string
	Original      string
	IP            string
	Bytes         int
	StatusCode    int
}

type ExpectedLogBody struct {
	Bytes int
}

func TestLogMiddleware(t *testing.T) {
	t.Run("create a middleware", func(t *testing.T) {
		called := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})
		testMockMiddlewareInvocation(handler, "", nil, "")

		assert.Assert(t, called, "handler must be called")
	})

	t.Run("log is a JSON also with trouble getting logger from context", func(t *testing.T) {
		logger, _ := zp.Init(zp.InitOptions{Level: "trace"})

		const logMessage string = "New log message"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), loggerKey{}, "notALogger")
			Get(ctx).Info().Msg(logMessage)
		})
		buffer := testMockMiddlewareInvocation(handler, "", logger, "")

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 3, "Number of logs is not 4 - %q", entries)

		for _, value := range entries {
			assertJSON(t, value)
		}
	})

	t.Run("middleware correctly passing configured logger with request id from request header", func(t *testing.T) {
		const statusCode = 400
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		buffer := testMockMiddlewareInvocation(handler, requestID, nil, "")

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:     string(pino.Trace),
			Message:   "incoming request",
			RequestID: requestID,
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     string(pino.Info),
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")
	})

	t.Run("passing a content-length header by default", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Header().Set("content-length", "10")
		})
		buffer := testMockMiddlewareInvocation(handler, requestID, nil, "")

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:     string(pino.Trace),
			Message:   "incoming request",
			RequestID: requestID,
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     string(pino.Info),
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         10,
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")
	})

	t.Run("without content-length in the header", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		contentToWrite := []byte("testing\n")
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Write(contentToWrite)
		})
		buffer := testMockMiddlewareInvocation(handler, requestID, nil, "")

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		i := 0
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:     string(pino.Trace),
			Message:   "incoming request",
			RequestID: requestID,
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     string(pino.Info),
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         len(contentToWrite),
		})

		assert.Equal(t, incomingRequestID, outcomingRequestID, "Data reqId of request and response log must be the same")
	})

	t.Run("using info level returning only outcomingRequest", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		logger := zp.InitDefault()
		buffer := testMockMiddlewareInvocation(handler, requestID, logger, "")

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")

		i := 0
		outcomingRequest := entries[i]
		logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     string(pino.Info),
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})
	})

	t.Run("test getHostname with request path without port", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		var requestPathWithoutPort = fmt.Sprintf("http://%s/my-req", hostname)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		logger := zp.InitDefault()
		buffer := testMockMiddlewareInvocation(handler, requestID, logger, requestPathWithoutPort)

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")

		i := 0
		outcomingRequest := entries[i]
		logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:     string(pino.Info),
			Message:   "request completed",
			RequestID: requestID,
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})
	})

	t.Run("test getHostname with request path with query", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		const pathWithQuery = "/my-req?foo=bar&some=other"
		var requestPathWithoutPort = fmt.Sprintf("http://%s%s", hostname, pathWithQuery)

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		logger := zp.InitDefault()
		buffer := testMockMiddlewareInvocation(handler, requestID, logger, requestPathWithoutPort)

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")

		logValue := assertJSON(t, entries[0])
		assert.Equal(t, logValue.URL.Path, "/my-req?foo=bar&some=other")
	})

	t.Run("middleware correctly passing configured logger with request id from request header", func(t *testing.T) {
		const statusCode = 200
		const requestID = "my-req-id"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		buffer := testMockMiddlewareInvocation(handler, requestID, nil, "/-/healthz")

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 1, "Unexpected entries length.")
	})

	t.Run("middleware correctly create request id if not present in header", func(t *testing.T) {
		const statusCode = 400
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
		})
		buffer := testMockMiddlewareInvocation(handler, "", nil, "")

		entries := strings.Split(strings.TrimSpace(buffer.String()), "\n")
		assert.Equal(t, len(entries), 3, "Unexpected entries length.")

		i := 1
		incomingRequest := entries[i]
		incomingRequestID := logAssertions(t, incomingRequest, ExpectedLogFields{
			Level:   string(pino.Trace),
			Message: "incoming request",
		})
		incomingRequestAssertions(t, incomingRequest, ExpectedIncomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
		})

		i++
		outcomingRequest := entries[i]
		outcomingRequestID := logAssertions(t, outcomingRequest, ExpectedLogFields{
			Level:   string(pino.Info),
			Message: "request completed",
		})
		outcomingRequestAssertions(t, outcomingRequest, ExpectedOutcomingLogFields{
			Method:        http.MethodGet,
			Path:          path,
			Hostname:      hostname,
			ForwardedHost: clientHost,
			Original:      userAgent,
			IP:            ip,
			StatusCode:    statusCode,
			Bytes:         bodyBytes,
		})

		assert.Equal(
			t,
			incomingRequestID,
			outcomingRequestID,
			fmt.Sprintf("Data reqId of request and response log must be the same. for log %d", i),
		)
	})
}

func BenchmarkMuxMiddleware(b *testing.B) {
	logger, _ := zp.Init(zp.InitOptions{Level: "trace"})

	const requestID string = "req-id"

	// prepare the request
	request := httptest.NewRequest("GET", defaultRequestPath, nil)
	ip := removePort(request.RemoteAddr)
	request.Header.Set("x-request-id", requestID)
	request.Header.Set("user-agent", userAgent)
	request.Header.Set("x-forwarded-for", ip)
	request.Header.Set("x-forwarded-host", clientHost)

	// prepare the server
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	handler := RequestMiddlewareLogger(logger, []string{"/-/"})
	// invoke the handler
	server := handler(next)
	// Create a response writer
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		// Serve HTTP server
		server.ServeHTTP(writer, request)
	}
}

func testMockMiddlewareInvocation(next http.HandlerFunc, requestID string,
	logger *zerolog.Logger, requestPath string) *bytes.Buffer {
	buffer := &bytes.Buffer{}
	if requestPath == "" {
		requestPath = defaultRequestPath
	}
	// create a request
	req := httptest.NewRequest(http.MethodGet, requestPath, nil)
	req.Header.Add("x-request-id", requestID)
	req.Header.Add("user-agent", userAgent)
	req.Header.Add("x-forwarded-for", ip)
	req.Header.Add("x-forwarded-host", clientHost)
	ip = removePort(req.RemoteAddr)

	// in case it is not provided, create a null logger
	if logger == nil {
		logger, _ = zp.Init(zp.InitOptions{Level: "trace", Writer: buffer})
	} else {
		testLogger := logger.Output(buffer)
		logger = &testLogger
	}

	handler := RequestMiddlewareLogger(logger, []string{"/-/"})
	// invoke the handler
	server := handler(next)
	// Create a response writer
	writer := httptest.NewRecorder()
	// Serve HTTP server
	server.ServeHTTP(writer, req)

	return buffer
}

func assertJSON(t *testing.T, str string) types.MiddlewareLog {
	var properties types.MiddlewareLog

	err := json.Unmarshal([]byte(str), &properties)
	assert.Equal(t, err, nil, "log %q is not expected JSON", str)

	return properties
}

func logAssertions(t *testing.T, rawLog string, expected ExpectedLogFields) string {
	logEntry := assertJSON(t, rawLog)

	assert.Equal(t, logEntry.Level, expected.Level, "Unexpected level of log for log in incoming request")
	assert.Equal(t, logEntry.Msg, expected.Message, "Unexpected message of log for log in incoming request")

	requestID := logEntry.RequestID
	if expected.RequestID != "" {
		assert.Assert(t, len(requestID) > 0, "Empty requestID for log in incoming request")
		assert.Equal(t, requestID, expected.RequestID, "Unexpected requestID for log in incoming request")
	}

	return requestID
}

func incomingRequestAssertions(t *testing.T, incomingRequestLogEntry string, expected ExpectedIncomingLogFields) {
	logValue := assertJSON(t, incomingRequestLogEntry)

	http := logValue.HTTP
	assert.Equal(t, http.Request.Method, expected.Method,
		"Unexpected http method for log in incoming request")
	assert.Equal(t, http.Request.UserAgent["original"], expected.Original,
		"Unexpected original userAgent for log of request completed")

	url := logValue.URL
	assert.Equal(t, url.Path, expected.Path,
		"Unexpected http uri path for log in incoming request")

	host := logValue.Host
	assert.Equal(t, host.Hostname, expected.Hostname,
		"Unexpected hostname for log of request completed")
	assert.Equal(t, host.ForwardedHost, expected.ForwardedHost,
		"Unexpected forwaded hostname for log of request completed")
	assert.Equal(t, host.IP, expected.IP,
		"Unexpected ip for log of request completed")
}

func outcomingRequestAssertions(t *testing.T, outcomingRequestLogEntry string, expected ExpectedOutcomingLogFields) {
	logValue := assertJSON(t, outcomingRequestLogEntry)

	http := logValue.HTTP
	assert.Equal(t, http.Request.Method, expected.Method,
		"Unexpected http method for log in incoming request")
	assert.Equal(t, http.Request.UserAgent["original"], expected.Original,
		"Unexpected original userAgent for log of request completed")
	assert.Equal(t, http.Response.StatusCode, expected.StatusCode,
		"Unexpected status code for log of request completed")

	binaryData, _ := json.Marshal(http.Response.Body)
	var structBody ExpectedLogBody
	assert.NilError(t, json.Unmarshal(binaryData, &structBody))
	assert.Equal(t, structBody.Bytes, expected.Bytes,
		"Unexpected body size for log of request completed")

	url := logValue.URL
	assert.Equal(t, url.Path, expected.Path,
		"Unexpected http uri path for log in incoming request")

	host := logValue.Host
	assert.Equal(t, host.Hostname, expected.Hostname,
		"Unexpected hostname for log of request completed")
	assert.Equal(t, host.ForwardedHost, expected.ForwardedHost,
		"Unexpected forwaded hostname for log of request completed")
	assert.Equal(t, host.IP, expected.IP,
		"Unexpected ip for log of request completed")
}

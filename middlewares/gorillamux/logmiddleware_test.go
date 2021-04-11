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

package gorillamux

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

	"github.com/gorilla/mux"
	"gotest.tools/assert"

	zp "github.com/danibix95/zeropino"
	pino "github.com/danibix95/zeropino/internal/model"
	zpm "github.com/danibix95/zeropino/middlewares"
)

const hostname = "my-host"
const userAgent = "goHttp"

const bodyBytes = 370
const requestPath = "/my-req"
const clientHost = "client-host"
const requestID = "req-id"
const doNotCheckBytes = -1

const method = "GET"

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
	Bytes         int
	StatusCode    int
}

type expectedBodyData struct {
	Bytes int
}

func TestRequestLogger(t *testing.T) {
	t.Run("trace log level - log both incoming request and response details", func(t *testing.T) {
		// use a buffer to avoid printing the logs on screen during tests
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "trace", Writer: buffer})

		middleware := RequestLogger(logger, []string{})
		app := createMuxApp(t, middleware, http.StatusOK, false)

		request := getRequestWithHeaders(method, defaultRequestURL, nil)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Result().StatusCode)

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
			IP:            removePort(request.RemoteAddr),
			StatusCode:    http.StatusOK,
			Bytes:         doNotCheckBytes,
		}
		assertResponseLog(t, expectedResponseLog, bytes.NewBufferString(entries[1]))
	})

	t.Run("log level is debug or info - log only response details", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "debug", Writer: buffer})

		middleware := RequestLogger(logger, []string{})
		app := createMuxApp(t, middleware, http.StatusBadRequest, false)

		request := getRequestWithHeaders(method, defaultRequestURL, nil)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusBadRequest, recorder.Result().StatusCode)

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
			IP:            removePort(request.RemoteAddr),
			StatusCode:    http.StatusBadRequest,
			Bytes:         doNotCheckBytes,
		}
		assertResponseLog(t, expected, buffer)
	})

	t.Run("log Content-Length value when set", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "debug", Writer: buffer})

		middleware := RequestLogger(logger, []string{})
		app := createMuxApp(t, middleware, http.StatusOK, true)

		request := getRequestWithHeaders(method, defaultRequestURL, nil)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Result().StatusCode)

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
			IP:            removePort(request.RemoteAddr),
			StatusCode:    http.StatusOK,
			Bytes:         bodyBytes,
		}
		assertResponseLog(t, expected, buffer)
	})

	t.Run("log level higher than info - no log produced", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		logger, _ := zp.Init(zp.InitOptions{Level: "warn", Writer: buffer})

		middleware := RequestLogger(logger, []string{})
		app := createMuxApp(t, middleware, http.StatusOK, false)

		request := httptest.NewRequest(method, defaultRequestURL, nil)

		recorder := httptest.NewRecorder()
		app.ServeHTTP(recorder, request)
		assert.Equal(t, http.StatusOK, recorder.Result().StatusCode)

		assert.Equal(t, 0, buffer.Len(), "no log output should be produced")
	})
}

func BenchmarkRequestLogger(b *testing.B) {
	logger, _ := zp.Init(zp.InitOptions{Level: "trace"})

	middleware := RequestLogger(logger, []string{"/-/"})
	app := createMuxApp(b, middleware, http.StatusOK, false)

	request := getRequestWithHeaders(method, defaultRequestURL, nil)

	recorder := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		app.ServeHTTP(recorder, request)
	}
}

func assertRequestLog(t testing.TB, expected logFields, actual *bytes.Buffer) zpm.LogFormat {
	t.Helper()

	var logOutput zpm.LogFormat
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

	return logOutput
}

func assertResponseLog(t testing.TB, expected logFields, actual *bytes.Buffer) {
	t.Helper()

	logOutput := assertRequestLog(t, expected, actual)

	assert.Equal(t, expected.StatusCode, logOutput.HTTP.Response.StatusCode)
	assert.Check(t, logOutput.ResponseTime > 0.0, "Response time is not null")

	if expected.Bytes >= 0 {
		binaryData, _ := json.Marshal(logOutput.HTTP.Response.Body)
		var structBody expectedBodyData
		assert.NilError(t, json.Unmarshal(binaryData, &structBody))
		assert.Equal(t, structBody.Bytes, expected.Bytes, "Body size is reported when set")
	}
}

func getRequestWithHeaders(method, path string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, path, body)
	ip := removePort(request.RemoteAddr)
	request.Header.Set(requestIDHeaderKey, requestID)
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set(forwardedForHeaderKey, ip)
	request.Header.Set(forwardedHostHeaderKey, clientHost)

	return request
}

func createMuxApp(t testing.TB, middleware mux.MiddlewareFunc, statusCode int, addContentLength bool) *mux.Router {
	t.Helper()
	router := mux.NewRouter()

	router.Use(middleware)

	response := struct {
		Msg string
	}{
		Msg: "Hello, World!",
	}

	router.HandleFunc(requestPath, func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		if addContentLength {
			w.Header().Add(contentLengthHeaderKey, strconv.Itoa(bodyBytes))
		}
		w.WriteHeader(statusCode)

		responseBody, _ := json.Marshal(&response)

		w.Write(responseBody)
	})

	return router
}

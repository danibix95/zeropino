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

package middleware

// HTTP is the struct of the log formatter.
type HTTP struct {
	Request  *Request  `json:"request,omitempty"`
	Response *Response `json:"response,omitempty"`
}

// Request contains the items of request info log.
type Request struct {
	Method    string                 `json:"method,omitempty"`
	UserAgent map[string]interface{} `json:"userAgent,omitempty"`
}

// Response contains the items of response info log.
type Response struct {
	StatusCode int                    `json:"statusCode,omitempty"`
	Body       map[string]interface{} `json:"body,omitempty"`
}

// Host has the host information.
type Host struct {
	Hostname      string `json:"hostname,omitempty"`
	ForwardedHost string `json:"forwardedHost,omitempty"`
	IP            string `json:"ip,omitempty"`
}

// URL info
type URL struct {
	Path string `json:"path,omitempty"`
}

// LogFormat represents the final log structure adopter by provided middlewares
type LogFormat struct {
	Level        string      `json:"level,omitempty"`
	Pid          int         `json:"pid,omitempty"`
	Hostname     string      `json:"hostname,omitempty"`
	Time         int         `json:"time,omitempty"`
	Msg          string      `json:"msg,omitempty"`
	Stack        interface{} `json:"error,omitempty"`
	RequestID    string      `json:"reqId,omitempty"`
	HTTP         HTTP        `json:"http,omitempty"`
	URL          URL         `json:"url,omitempty"`
	Host         Host        `json:"host,omitempty"`
	ResponseTime float64     `json:"responseTime,omitempty"`
}

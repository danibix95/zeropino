package middlewares

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

// MiddlewareLog represents the final log structure adopter by provided middlewares
type MiddlewareLog struct {
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

package http

//go:generate stringer -type=Method
type Method int

//go:generate stringer -type=Version
type Version int

const (
	GET Method = iota
	HEAD
	POST
	PUT
	DELETE
	CONNECT
	OPTIONS
	TRACE
	PATCH
)

const (
	HTTP1_0 Version = iota
	HTTP1_1
	HTTP2
	HTTP3
)

type Status string

const (
	// 1xx Informational
	StatusContinue           Status = "100 Continue"
	StatusSwitchingProtocols Status = "101 Switching Protocols"
	StatusProcessing         Status = "102 Processing"  // WebDAV
	StatusEarlyHints         Status = "103 Early Hints" // RFC 8297

	// 2xx Success
	StatusOK                   Status = "200 OK"
	StatusCreated              Status = "201 Created"
	StatusAccepted             Status = "202 Accepted"
	StatusNonAuthoritativeInfo Status = "203 Non-Authoritative Information"
	StatusNoContent            Status = "204 No Content"
	StatusResetContent         Status = "205 Reset Content"
	StatusPartialContent       Status = "206 Partial Content"
	StatusMultiStatus          Status = "207 Multi-Status"     // WebDAV
	StatusAlreadyReported      Status = "208 Already Reported" // WebDAV
	StatusIMUsed               Status = "226 IM Used"          // RFC 3229

	// 3xx Redirection
	StatusMultipleChoices   Status = "300 Multiple Choices"
	StatusMovedPermanently  Status = "301 Moved Permanently"
	StatusFound             Status = "302 Found"
	StatusSeeOther          Status = "303 See Other"
	StatusNotModified       Status = "304 Not Modified"
	StatusUseProxy          Status = "305 Use Proxy"
	_                              = "306 (Unused)"
	StatusTemporaryRedirect Status = "307 Temporary Redirect"
	StatusPermanentRedirect Status = "308 Permanent Redirect"

	// 4xx Client Errors
	StatusBadRequest                  Status = "400 Bad Request"
	StatusUnauthorized                Status = "401 Unauthorized"
	StatusPaymentRequired             Status = "402 Payment Required"
	StatusForbidden                   Status = "403 Forbidden"
	StatusNotFound                    Status = "404 Not Found"
	StatusMethodNotAllowed            Status = "405 Method Not Allowed"
	StatusNotAcceptable               Status = "406 Not Acceptable"
	StatusProxyAuthRequired           Status = "407 Proxy Authentication Required"
	StatusRequestTimeout              Status = "408 Request Timeout"
	StatusConflict                    Status = "409 Conflict"
	StatusGone                        Status = "410 Gone"
	StatusLengthRequired              Status = "411 Length Required"
	StatusPreconditionFailed          Status = "412 Precondition Failed"
	StatusPayloadTooLarge             Status = "413 Payload Too Large"
	StatusURITooLong                  Status = "414 URI Too Long"
	StatusUnsupportedMediaType        Status = "415 Unsupported Media Type"
	StatusRangeNotSatisfiable         Status = "416 Range Not Satisfiable"
	StatusExpectationFailed           Status = "417 Expectation Failed"
	StatusTeapot                      Status = "418 I'm a teapot" // RFC 7168
	StatusMisdirectedRequest          Status = "421 Misdirected Request"
	StatusUnprocessableEntity         Status = "422 Unprocessable Entity"
	StatusLocked                      Status = "423 Locked"
	StatusFailedDependency            Status = "424 Failed Dependency"
	StatusTooEarly                    Status = "425 Too Early"
	StatusUpgradeRequired             Status = "426 Upgrade Required"
	StatusPreconditionRequired        Status = "428 Precondition Required"
	StatusTooManyRequests             Status = "429 Too Many Requests"
	StatusRequestHeaderFieldsTooLarge Status = "431 Request Header Fields Too Large"
	StatusUnavailableForLegalReasons  Status = "451 Unavailable For Legal Reasons"

	// 5xx Server Errors
	StatusInternalServerError           Status = "500 Internal Server Error"
	StatusNotImplemented                Status = "501 Not Implemented"
	StatusBadGateway                    Status = "502 Bad Gateway"
	StatusServiceUnavailable            Status = "503 Service Unavailable"
	StatusGatewayTimeout                Status = "504 Gateway Timeout"
	StatusHTTPVersionNotSupported       Status = "505 HTTP Version Not Supported"
	StatusVariantAlsoNegotiates         Status = "506 Variant Also Negotiates"
	StatusInsufficientStorage           Status = "507 Insufficient Storage"
	StatusLoopDetected                  Status = "508 Loop Detected"
	StatusNotExtended                   Status = "510 Not Extended"
	StatusNetworkAuthenticationRequired Status = "511 Network Authentication Required"
)

package httpserver

import "errors"

// ErrServerFailed marks a failure returned by the HTTP listener.
var ErrServerFailed = errors.New("HTTP server failed")

package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"codecollab/utils"
)

type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
	written    int64
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	n, err := rc.ResponseWriter.Write(b)
	rc.written += int64(n)
	return n, err
}

func LoggingMiddleware(logger *utils.LokiLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var requestBody []byte
			if r.Body != nil && r.ContentLength > 0 && r.ContentLength < 1024*1024 {
				requestBody, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
			}

			rc := &responseCapture{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(rc, r)

			duration := time.Since(start)

			headers := make(map[string]string)
			for key, values := range r.Header {
				if len(values) > 0 {
					headers[key] = values[0]
				}
			}

			responseBody := rc.body.String()
			if len(responseBody) > 10000 {
				responseBody = responseBody[:10000] + "... (truncated)"
			}

			requestBodyStr := string(requestBody)
			if len(requestBodyStr) > 10000 {
				requestBodyStr = requestBodyStr[:10000] + "... (truncated)"
			}

			logger.LogRequest(utils.RequestLog{
				Timestamp:      start,
				Method:         r.Method,
				Path:           r.URL.Path,
				Query:          r.URL.RawQuery,
				Headers:        headers,
				RequestBody:    requestBodyStr,
				ResponseBody:   responseBody,
				StatusCode:     rc.statusCode,
				DurationMs:     duration.Milliseconds(),
				RemoteAddr:     r.RemoteAddr,
				UserAgent:      r.UserAgent(),
				ContentLength:  r.ContentLength,
				ResponseSize:   rc.written,
			})
		})
	}
}

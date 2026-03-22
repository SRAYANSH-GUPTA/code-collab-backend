package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type RequestLog struct {
	Timestamp      time.Time
	Method         string
	Path           string
	Query          string
	Headers        map[string]string
	RequestBody    string
	ResponseBody   string
	StatusCode     int
	DurationMs     int64
	RemoteAddr     string
	UserAgent      string
	ContentLength  int64
	ResponseSize   int64
}

type LokiLogger struct {
	lokiURL string
	labels  map[string]string
	client  *http.Client
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

type lokiRequest struct {
	Streams []lokiStream `json:"streams"`
}

func NewLokiLogger(lokiURL string, labels map[string]string) *LokiLogger {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["job"] = "codecollab-backend"
	labels["service"] = "codecollab"

	return &LokiLogger{
		lokiURL: lokiURL,
		labels:  labels,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (l *LokiLogger) LogRequest(log RequestLog) {
	logData := map[string]interface{}{
		"timestamp":       log.Timestamp.Format(time.RFC3339Nano),
		"method":          log.Method,
		"path":            log.Path,
		"query":           log.Query,
		"headers":         log.Headers,
		"request_body":    log.RequestBody,
		"response_body":   log.ResponseBody,
		"status_code":     log.StatusCode,
		"duration_ms":     log.DurationMs,
		"remote_addr":     log.RemoteAddr,
		"user_agent":      log.UserAgent,
		"content_length":  log.ContentLength,
		"response_size":   log.ResponseSize,
	}

	logJSON, err := json.Marshal(logData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log: %v\n", err)
		return
	}

	labels := make(map[string]string)
	for k, v := range l.labels {
		labels[k] = v
	}
	labels["method"] = log.Method
	labels["path"] = log.Path
	labels["status"] = fmt.Sprintf("%d", log.StatusCode)

	stream := lokiStream{
		Stream: labels,
		Values: [][]string{
			{
				fmt.Sprintf("%d", log.Timestamp.UnixNano()),
				string(logJSON),
			},
		},
	}

	lokiReq := lokiRequest{
		Streams: []lokiStream{stream},
	}

	go l.sendToLoki(lokiReq)
}

func (l *LokiLogger) sendToLoki(req lokiRequest) {
	if l.lokiURL == "" {
		return
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal loki request: %v\n", err)
		return
	}

	httpReq, err := http.NewRequest("POST", l.lokiURL+"/loki/api/v1/push", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create loki request: %v\n", err)
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to send to loki: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "Loki returned non-2xx status: %d\n", resp.StatusCode)
	}
}

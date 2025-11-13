package metrics

import (
	"runtime"
	"time"
)

func StartSystemMetricsCollector() {
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			collectSystemMetrics()
		}
	}()
}

func collectSystemMetrics() {
	GoRoutinesCount.Set(float64(runtime.NumGoroutine()))
}

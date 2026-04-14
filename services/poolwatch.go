package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// PoolWatchMetrics holds the latest snapshot from PoolWatch /api/v1/metrics.
type PoolWatchMetrics struct {
	CollectedAt time.Time              `json:"collected_at"`
	Postgres    map[string]interface{} `json:"postgres"`
	PgBouncer   map[string]interface{} `json:"pgbouncer"`
	Derived     map[string]interface{} `json:"derived"`
}

// PoolWatchAlert represents a single alert entry from PoolWatch /api/v1/alerts.
type PoolWatchAlert struct {
	Code      string    `json:"code"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// PoolWatchAlertsResponse is the response envelope for /api/v1/alerts.
type PoolWatchAlertsResponse struct {
	Alerts []PoolWatchAlert `json:"alerts"`
}

// PoolWatchStatus holds the runtime status from PoolWatch /api/v1/status.
type PoolWatchStatus struct {
	Mode      string    `json:"mode"`
	StartedAt time.Time `json:"started_at"`
	Healthy   bool      `json:"healthy"`
	Extra     map[string]interface{} `json:"-"`
}

// PoolWatchSnapshot is the combined in-memory cache of the latest data.
type PoolWatchSnapshot struct {
	Metrics   *PoolWatchMetrics
	Alerts    []PoolWatchAlert
	Status    map[string]interface{}
	FetchedAt time.Time
	Error     error
}

// PoolWatchService continuously polls PoolWatch and caches the latest data.
type PoolWatchService struct {
	baseURL  string
	client   *http.Client
	interval time.Duration

	mu       sync.RWMutex
	snapshot PoolWatchSnapshot
}

// NewPoolWatchService creates a new PoolWatchService. If baseURL is empty, the
// service is disabled and all getters return nil / empty results.
func NewPoolWatchService(baseURL string, pollInterval time.Duration) *PoolWatchService {
	if pollInterval <= 0 {
		pollInterval = 15 * time.Second
	}
	return &PoolWatchService{
		baseURL:  baseURL,
		interval: pollInterval,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Enabled reports whether PoolWatch integration is configured.
func (s *PoolWatchService) Enabled() bool {
	return s.baseURL != ""
}

// Start begins the polling loop, which runs until ctx is cancelled.
func (s *PoolWatchService) Start(ctx context.Context) {
	if !s.Enabled() {
		log.Println("[PoolWatch] No POOLWATCH_URL configured – integration disabled")
		return
	}
	log.Printf("[PoolWatch] Starting — polling %s every %s", s.baseURL, s.interval)
	go s.loop(ctx)
}

func (s *PoolWatchService) loop(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Poll immediately on start.
	s.poll()

	for {
		select {
		case <-ctx.Done():
			log.Println("[PoolWatch] Polling stopped")
			return
		case <-ticker.C:
			s.poll()
		}
	}
}

func (s *PoolWatchService) poll() {
	metrics, metricsErr := s.fetchJSON(s.baseURL + "/api/v1/metrics")
	alerts, alertsErr := s.fetchJSON(s.baseURL + "/api/v1/alerts")
	status, statusErr := s.fetchJSON(s.baseURL + "/api/v1/status")

	snap := PoolWatchSnapshot{
		FetchedAt: time.Now(),
	}

	if metricsErr != nil || alertsErr != nil || statusErr != nil {
		err := metricsErr
		if err == nil {
			err = alertsErr
		}
		if err == nil {
			err = statusErr
		}
		snap.Error = err
		log.Printf("[PoolWatch] Poll error: %v", err)
	} else {
		// Parse metrics.
		if m, ok := metrics["derived"]; ok {
			_ = m // keep raw map, expose it as-is
		}
		snap.Status = status

		// Try to extract alerts list.
		if rawAlerts, ok := alerts["alerts"]; ok {
			b, _ := json.Marshal(rawAlerts)
			var alertList []PoolWatchAlert
			if jsonErr := json.Unmarshal(b, &alertList); jsonErr == nil {
				snap.Alerts = alertList
			}
		}

		// Wrap the metrics map loosely.
		mBytes, _ := json.Marshal(metrics)
		var pm PoolWatchMetrics
		if jsonErr := json.Unmarshal(mBytes, &pm); jsonErr == nil {
			snap.Metrics = &pm
		} else {
			// Store raw fields anyway.
			snap.Metrics = &PoolWatchMetrics{
				CollectedAt: time.Now(),
				Postgres:    toStringMap(metrics["postgres"]),
				PgBouncer:   toStringMap(metrics["pgbouncer"]),
				Derived:     toStringMap(metrics["derived"]),
			}
		}
	}

	s.mu.Lock()
	s.snapshot = snap
	s.mu.Unlock()
}

func (s *PoolWatchService) fetchJSON(url string) (map[string]interface{}, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body %s: %w", url, err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GET %s: HTTP %d: %s", url, resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode %s: %w", url, err)
	}
	return result, nil
}

// GetSnapshot returns a copy of the latest cached snapshot.
func (s *PoolWatchService) GetSnapshot() PoolWatchSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot
}

// ProxyEnable calls POST /api/v1/proxy/enable on PoolWatch.
func (s *PoolWatchService) ProxyEnable() error {
	return s.postAction(s.baseURL + "/api/v1/proxy/enable")
}

// ProxyDisable calls POST /api/v1/proxy/disable on PoolWatch.
func (s *PoolWatchService) ProxyDisable() error {
	return s.postAction(s.baseURL + "/api/v1/proxy/disable")
}

func (s *PoolWatchService) postAction(url string) error {
	resp, err := s.client.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("POST %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s: HTTP %d: %s", url, resp.StatusCode, string(b))
	}
	return nil
}

func toStringMap(v interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return nil
}

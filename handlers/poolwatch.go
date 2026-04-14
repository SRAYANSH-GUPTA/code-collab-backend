package handlers

import (
	"encoding/json"
	"net/http"

	"codecollab/services"
)

var globalPoolWatch *services.PoolWatchService

// InitPoolWatchHandler wires the shared PoolWatchService into this handler package.
func InitPoolWatchHandler(svc *services.PoolWatchService) {
	globalPoolWatch = svc
}

// HandlePoolMetrics serves the latest PoolWatch metrics snapshot.
// GET /pool/metrics
func HandlePoolMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if globalPoolWatch == nil || !globalPoolWatch.Enabled() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "PoolWatch integration is not configured",
		})
		return
	}

	snap := globalPoolWatch.GetSnapshot()
	if snap.Error != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":      "PoolWatch unreachable",
			"detail":     snap.Error.Error(),
			"fetched_at": snap.FetchedAt,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(snap.Metrics)
}

// HandlePoolAlerts serves the latest active alert list from PoolWatch.
// GET /pool/alerts
func HandlePoolAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if globalPoolWatch == nil || !globalPoolWatch.Enabled() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "PoolWatch integration is not configured",
		})
		return
	}

	snap := globalPoolWatch.GetSnapshot()
	if snap.Error != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":      "PoolWatch unreachable",
			"detail":     snap.Error.Error(),
			"fetched_at": snap.FetchedAt,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alerts":     snap.Alerts,
		"fetched_at": snap.FetchedAt,
	})
}

// HandlePoolStatus serves the PoolWatch runtime status.
// GET /pool/status
func HandlePoolStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if globalPoolWatch == nil || !globalPoolWatch.Enabled() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "PoolWatch integration is not configured",
		})
		return
	}

	snap := globalPoolWatch.GetSnapshot()
	if snap.Error != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":      "PoolWatch unreachable",
			"detail":     snap.Error.Error(),
			"fetched_at": snap.FetchedAt,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     snap.Status,
		"fetched_at": snap.FetchedAt,
	})
}

// HandlePoolProxyEnable triggers the PoolWatch guarded proxy mode.
// POST /pool/proxy/enable
func HandlePoolProxyEnable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if globalPoolWatch == nil || !globalPoolWatch.Enabled() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "PoolWatch integration is not configured"})
		return
	}

	if err := globalPoolWatch.ProxyEnable(); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "PoolWatch proxy mode enabled"})
}

// HandlePoolProxyDisable disables the PoolWatch guarded proxy mode.
// POST /pool/proxy/disable
func HandlePoolProxyDisable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if globalPoolWatch == nil || !globalPoolWatch.Enabled() {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "PoolWatch integration is not configured"})
		return
	}

	if err := globalPoolWatch.ProxyDisable(); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "PoolWatch proxy mode disabled"})
}

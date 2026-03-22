package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)



var (
	
	VoiceRoomsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "voice_rooms_active",
		Help: "Number of active voice rooms",
	})

	
	VoicePeersTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "voice_peers_total",
		Help: "Total number of peers in voice calls",
	})

	
	VoicePeersPerRoom = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "voice_peers_per_room",
		Help:    "Distribution of peers per voice room",
		Buckets: []float64{1, 2, 5, 10, 20, 50, 100},
	})

	
	VoiceConnectionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "voice_connection_duration_seconds",
		Help:    "Time taken to establish WebRTC connection",
		Buckets: prometheus.DefBuckets,
	})

	
	VoiceICEStateChanges = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "voice_ice_state_changes_total",
		Help: "Total number of ICE connection state changes",
	}, []string{"state"})

	
	VoiceOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "voice_operations_total",
		Help: "Total number of voice operations",
	}, []string{"operation", "status"})

	
	VoicePacketLoss = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "voice_packet_loss_percent",
		Help:    "Packet loss percentage in voice calls",
		Buckets: []float64{0, 1, 2, 5, 10, 20, 50},
	}, []string{"room_id"})

	
	VoiceJitter = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "voice_jitter_milliseconds",
		Help:    "Jitter in voice calls (milliseconds)",
		Buckets: []float64{0, 5, 10, 20, 50, 100, 200},
	}, []string{"room_id"})

	
	VoiceRTT = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "voice_rtt_milliseconds",
		Help:    "Round-trip time for voice connections",
		Buckets: []float64{0, 10, 25, 50, 100, 200, 500},
	}, []string{"room_id"})

	
	VoiceErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "voice_errors_total",
		Help: "Total number of voice-related errors",
	}, []string{"error_type"})
)


func RecordVoiceOperation(operation, status string) {
	VoiceOperations.WithLabelValues(operation, status).Inc()
}


func RecordICEStateChange(state string) {
	VoiceICEStateChanges.WithLabelValues(state).Inc()
}


func RecordVoiceError(errorType string) {
	VoiceErrors.WithLabelValues(errorType).Inc()
}


func UpdateVoiceStats(roomCount, peerCount int) {
	VoiceRoomsActive.Set(float64(roomCount))
	VoicePeersTotal.Set(float64(peerCount))
}

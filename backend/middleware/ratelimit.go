package middleware

import (
	"sync"
	"time"
)


type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int
	window   time.Duration
}




func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	
	go rl.cleanup()

	return rl
}


func (rl *RateLimiter) CheckRateLimit(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	
	requests, exists := rl.requests[userID]
	if !exists {
		rl.requests[userID] = []time.Time{now}
		return true
	}

	
	validRequests := []time.Time{}
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	
	if len(validRequests) >= rl.limit {
		return false
	}

	
	validRequests = append(validRequests, now)
	rl.requests[userID] = validRequests

	return true
}


func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for userID, requests := range rl.requests {
			validRequests := []time.Time{}
			for _, reqTime := range requests {
				if reqTime.After(windowStart) {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, userID)
			} else {
				rl.requests[userID] = validRequests
			}
		}
		rl.mu.Unlock()
	}
}

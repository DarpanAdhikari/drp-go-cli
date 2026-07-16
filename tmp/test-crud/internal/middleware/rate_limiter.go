package middleware

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter implements a per-IP token bucket rate limiter.
type IPRateLimiter struct {
	mu    sync.RWMutex
	ips   map[string]*rate.Limiter
	r     rate.Limit
	burst int
}

// NewIPRateLimiter creates a new IPRateLimiter.
func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		ips:   make(map[string]*rate.Limiter),
		r:     r,
		burst: burst,
	}
	go rl.cleanup()
	return rl
}

// Allow reports whether a request from the given IP should be allowed.
func (l *IPRateLimiter) Allow(ip string) bool {
	l.mu.RLock()
	limiter, exists := l.ips[ip]
	l.mu.RUnlock()
	if !exists {
		l.mu.Lock()
		limiter = rate.NewLimiter(l.r, l.burst)
		l.ips[ip] = limiter
		l.mu.Unlock()
	}
	return limiter.Allow()
}

func (l *IPRateLimiter) cleanup() {
	for {
		time.Sleep(10 * time.Minute)
		l.mu.Lock()
		for ip, limiter := range l.ips {
			if limiter.Tokens() == float64(l.burst) {
				delete(l.ips, ip)
			}
		}
		l.mu.Unlock()
	}
}

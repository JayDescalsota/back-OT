package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/clinicmanager/shared/logger"
)

type visitor struct {
	count   int
	resetAt time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	window   time.Duration
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
	}
	go func() {
		for {
			time.Sleep(window)
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.After(v.resetAt) {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		now := time.Now()
		if !exists || now.After(v.resetAt) {
			rl.visitors[ip] = &visitor{count: 1, resetAt: now.Add(rl.window)}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		if v.count > rl.rate {
			rl.mu.Unlock()
			logger.Warn(r.Context(), "rate limit exceeded", "ip", ip, "path", r.URL.Path)
			http.Error(w, `{"errors":[{"message":"too many requests"}]}`, http.StatusTooManyRequests)
			return
		}
		rl.mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

func extractIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

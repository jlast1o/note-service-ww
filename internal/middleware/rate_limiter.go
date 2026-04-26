package middleware

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiters sync.Map
	limit    rate.Limit
	burst    int
	requests int
}

func NewRateLimiter(requests int, periodSecond int) *RateLimiter {
	limit := rate.Limit(float64(requests) / float64(periodSecond))
	return &RateLimiter{
		limit:    limit,
		burst:    requests,
		requests: requests,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r.RemoteAddr)
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			w.Header().Set("X-RateLimit-Limit", formatLimit(rl))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("Retry-After", "60")

			slog.Warn("Слишком много запросов", "ip", ip)
			http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	if val, ok := rl.limiters.Load(ip); ok {
		return val.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(rl.limit, rl.burst)
	rl.limiters.Store(ip, limiter)

	return limiter
}

func formatLimit(rl *RateLimiter) string {
	return fmt.Sprintf("%d", rl.requests)
}

func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

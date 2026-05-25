package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimitStore struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	r        rate.Limit
	burst    int
}

func newRateLimitStore(requestsPerWindow, windowSeconds int) *rateLimitStore {
	s := &rateLimitStore{
		limiters: make(map[string]*ipLimiter),
		r:        rate.Limit(float64(requestsPerWindow) / float64(windowSeconds)),
		burst:    requestsPerWindow,
	}
	go s.cleanup()
	return s
}

func (s *rateLimitStore) getLimiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry, ok := s.limiters[ip]; ok {
		entry.lastSeen = time.Now()
		return entry.limiter
	}

	l := rate.NewLimiter(s.r, s.burst)
	s.limiters[ip] = &ipLimiter{limiter: l, lastSeen: time.Now()}
	return l
}

func (s *rateLimitStore) cleanup() {
	for {
		time.Sleep(5 * time.Minute)
		s.mu.Lock()
		for ip, entry := range s.limiters {
			if time.Since(entry.lastSeen) > 10*time.Minute {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimitMiddleware limits requests per client IP using a token-bucket algorithm.
// Applied only to the login route to mitigate brute-force attacks.
// The in-memory store resets on restart; a Redis store would be needed for
// multi-instance deployments.
func RateLimitMiddleware(requestsPerWindow, windowSeconds int) gin.HandlerFunc {
	store := newRateLimitStore(requestsPerWindow, windowSeconds)
	return func(c *gin.Context) {
		if !store.getLimiter(c.ClientIP()).Allow() {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, please try again later",
			})
			return
		}
		c.Next()
	}
}

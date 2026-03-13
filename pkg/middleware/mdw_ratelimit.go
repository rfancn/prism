package middleware

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hdget/sdk"
	"golang.org/x/time/rate"
	"github.com/rfancn/prism/repository"
)

// RateLimiter manages rate limiters per user.
type RateLimiter struct {
	mu           sync.RWMutex
	limiters     map[string]*rate.Limiter
	defaultRPS   rate.Limit
	defaultBurst int
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(defaultRPS, defaultBurst int) *RateLimiter {
	return &RateLimiter{
		limiters:     make(map[string]*rate.Limiter),
		defaultRPS:   rate.Limit(defaultRPS),
		defaultBurst: defaultBurst,
	}
}

// GetLimiter returns the rate limiter for a user.
func (r *RateLimiter) GetLimiter(userID string) *rate.Limiter {
	r.mu.RLock()
	limiter, exists := r.limiters[userID]
	r.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter with default or custom settings
	r.mu.Lock()
	defer r.mu.Unlock()

	// Double check
	if limiter, exists = r.limiters[userID]; exists {
		return limiter
	}

	// Try to get custom rate limit
	rps := r.defaultRPS
	burst := r.defaultBurst

	queries := repository.New()
	if queries != nil {
		customLimit, err := queries.GetRateLimitByUserID(context.Background(), userID)
		if err == nil {
			rps = rate.Limit(customLimit.RequestsPerSecond)
			burst = int(customLimit.Burst)
		}
	}

	limiter = rate.NewLimiter(rps, burst)
	r.limiters[userID] = limiter
	return limiter
}

var globalRateLimiter *RateLimiter

// SetGlobalRateLimiter sets the global rate limiter instance.
func SetGlobalRateLimiter(limiter *RateLimiter) {
	globalRateLimiter = limiter
}

// RateLimitMiddleware limits requests per user.
func NewRateLimitMiddleware() (gin.HandlerFunc, error) {
	return func(c *gin.Context) {
		if globalRateLimiter == nil {
			c.Next()
			return
		}

		userID := GetUserID(c)
		if userID == "" {
			// No user ID, skip rate limiting
			c.Next()
			return
		}

		limiter := globalRateLimiter.GetLimiter(userID)

		if !limiter.Allow() {
			sdk.Logger().Debug("rate limit exceeded", "user_id", userID)
			c.JSON(429, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}, nil
}
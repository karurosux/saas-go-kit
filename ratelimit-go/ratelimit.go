package ratelimit

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     int
	window   time.Duration
	cleanup  time.Duration
}

type visitor struct {
	lastSeen time.Time
	count    int
}

// Option configures the rate limiter
type Option func(*RateLimiter)

// WithCleanupInterval sets the cleanup interval
func WithCleanupInterval(interval time.Duration) Option {
	return func(rl *RateLimiter) {
		rl.cleanup = interval
	}
}

// New creates a new rate limiter
func New(rate int, window time.Duration, opts ...Option) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		window:   window,
		cleanup:  time.Minute, // default cleanup interval
	}

	for _, opt := range opts {
		opt(rl)
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if the request should be allowed
func (rl *RateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[identifier]
	now := time.Now()

	if !exists {
		rl.visitors[identifier] = &visitor{
			lastSeen: now,
			count:    1,
		}
		return true
	}

	// Reset count if window has passed
	if now.Sub(v.lastSeen) > rl.window {
		v.count = 1
		v.lastSeen = now
		return true
	}

	// Check rate limit
	if v.count >= rl.rate {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}

// Reset resets the count for a specific identifier
func (rl *RateLimiter) Reset(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.visitors, identifier)
}

// EchoMiddleware returns an Echo middleware
func (rl *RateLimiter) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identifier := c.RealIP()
			
			if !rl.Allow(identifier) {
				return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
					"error": map[string]interface{}{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "Too many requests. Please try again later.",
					},
				})
			}

			return next(c)
		}
	}
}

// HTTPMiddleware returns a standard HTTP middleware
func (rl *RateLimiter) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := r.RemoteAddr
		if ip := r.Header.Get("X-Real-IP"); ip != "" {
			identifier = ip
		} else if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
			identifier = ip
		}

		if !rl.Allow(identifier) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests. Please try again later."}}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// cleanupVisitors removes old visitor entries
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// KeyFunc defines a function to extract the rate limit key from a request
type KeyFunc func(c echo.Context) string

// Config holds rate limiter configuration
type Config struct {
	Rate     int
	Window   time.Duration
	KeyFunc  KeyFunc
	ErrorHandler func(c echo.Context) error
}

// NewWithConfig creates a rate limiter with advanced configuration
func NewWithConfig(cfg Config) echo.MiddlewareFunc {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c echo.Context) string {
			return c.RealIP()
		}
	}

	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c echo.Context) error {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error": map[string]interface{}{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests. Please try again later.",
				},
			})
		}
	}

	limiter := New(cfg.Rate, cfg.Window)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := cfg.KeyFunc(c)
			
			if !limiter.Allow(key) {
				return cfg.ErrorHandler(c)
			}

			return next(c)
		}
	}
}

// PerUserRateLimiter creates a rate limiter that limits by authenticated user
func PerUserRateLimiter(rate int, window time.Duration, userIDExtractor func(c echo.Context) string) echo.MiddlewareFunc {
	return NewWithConfig(Config{
		Rate:   rate,
		Window: window,
		KeyFunc: func(c echo.Context) string {
			if userID := userIDExtractor(c); userID != "" {
				return "user:" + userID
			}
			return c.RealIP()
		},
	})
}

// Store interface for distributed rate limiting
type Store interface {
	// Increment increments the counter for the given key and returns the new count
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
	// Reset resets the counter for the given key
	Reset(ctx context.Context, key string) error
}
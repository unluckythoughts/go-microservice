package ratelimiter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"github.com/unluckythoughts/go-microservice/v2/utils"
	"go.uber.org/zap"
)

// rateLimitScript atomically increments the request counter and sets the
// expiry only on the first increment so the window starts fresh each period.
var rateLimitScript = redis.NewScript(`
	local count = redis.call('INCR', KEYS[1])
	if count == 1 then
		redis.call('PEXPIRE', KEYS[1], ARGV[1])
	end
	return count
`)

type Options struct {
	RateLimitMax        int `env:"RATE_LIMIT_MAX" envDefault:"100"`
	RateLimitWindowSecs int `env:"RATE_LIMIT_WINDOW_SECS" envDefault:"60"`
	Logger              *zap.Logger
	Cache               *redis.Client
}

// RateLimiter enforces Redis-backed fixed-window rate limits.
type RateLimiter struct {
	maxRequests int
	window      time.Duration
	client      *redis.Client
	logger      *zap.Logger
}

// New creates a new RateLimiter backed by the given Redis client.
func New(opts Options) *RateLimiter {
	utils.ParseEnvironmentVars(&opts)
	return &RateLimiter{
		maxRequests: opts.RateLimitMax,
		window:      time.Duration(opts.RateLimitWindowSecs) * time.Second,
		client:      opts.Cache,
		logger:      opts.Logger,
	}
}

// GetMiddleware returns a web.Middleware that allows at most maxRequests within
// each window duration, keyed by the request path and client IP.
// If the Redis client is unavailable the check fails open (request is allowed).
func (rl *RateLimiter) GetMiddleware() web.Middleware {
	return func(r web.MiddlewareRequest) error {
		clientIP := getClientIP(r)
		key := fmt.Sprintf("rl:%s:%s", r.GetPath(), clientIP)

		count, err := rateLimitScript.Run(
			r.GetContext(),
			rl.client,
			[]string{key},
			int64(rl.window/time.Millisecond),
		).Int64()
		if err != nil {
			// Fail open: Redis unavailable should not block legitimate traffic.
			return nil
		}

		if count > int64(rl.maxRequests) {
			return web.NewError(http.StatusTooManyRequests,
				fmt.Errorf("rate limit exceeded, please try again later"))
		}

		return nil
	}
}

// getClientIP extracts the real client IP from well-known forwarding headers,
// falling back to the TCP remote address when no proxy headers are present.
func getClientIP(r web.Request) string {
	if ip := r.GetHeader("X-Real-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if fwd := r.GetHeader("X-Forwarded-For"); fwd != "" {
		// X-Forwarded-For can be a comma-separated list; the first entry is
		// the originating client.
		parts := strings.SplitN(fwd, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	return r.GetRemoteAddr()
}

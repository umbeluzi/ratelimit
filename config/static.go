package config

import (
    "context"
    "sync"
    "time"
)

// Static is a static implementation of the Config interface.
type Static struct {
    maxRequests int
    interval    time.Duration
    burstLimit  int
    tokens      int
    lastRefill  time.Time
    mu          sync.Mutex
}

// NewStatic creates a new Static.
func NewStatic(maxRequests int, interval time.Duration, burstLimit int, tokens int, lastRefill time.Time) *Static {
    return &Static{
        maxRequests: maxRequests,
        interval:    interval,
        burstLimit:  burstLimit,
        tokens:      tokens,
        lastRefill:  lastRefill,
    }
}

// MaxRequests returns the max requests from the static config.
func (c *Static) MaxRequests(ctx context.Context) (int, error) {
    return c.maxRequests, nil
}

// Interval returns the interval from the static config.
func (c *Static) Interval(ctx context.Context) (time.Duration, error) {
    return c.interval, nil
}

// BurstLimit returns the burst limit from the static config.
func (c *Static) BurstLimit(ctx context.Context) (int, error) {
    return c.burstLimit, nil
}

// Tokens returns the current token count from the static config.
func (c *Static) Tokens(ctx context.Context) (int, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.tokens, nil
}

// SetTokens sets the current token count in the static config.
func (c *Static) SetTokens(ctx context.Context, tokens int) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.tokens = tokens
    return nil
}

// LastRefill returns the last refill time from the static config.
func (c *Static) LastRefill(ctx context.Context) (time.Time, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.lastRefill, nil
}

// SetLastRefill sets the last refill time in the static config.
func (c *Static) SetLastRefill(ctx context.Context, lastRefill time.Time) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.lastRefill = lastRefill
    return nil
}

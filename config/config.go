package config

import (
    "context"
    "time"
)

// Config is the interface for rate limiter configuration.
type Config interface {
    MaxRequests(ctx context.Context) (int, error)
    Interval(ctx context.Context) (time.Duration, error)
    BurstLimit(ctx context.Context) (int, error)
    Tokens(ctx context.Context) (int, error)
    SetTokens(ctx context.Context, tokens int) error
    LastRefill(ctx context.Context) (time.Time, error)
    SetLastRefill(ctx context.Context, lastRefill time.Time) error
}

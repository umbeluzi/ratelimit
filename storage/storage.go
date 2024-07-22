package storage

import (
    "context"
    "time"
)

// Storage is the interface for rate limit storage.
type Storage interface {
    Increment(ctx context.Context, key string) (int, error)
    Reset(ctx context.Context, key string) error
    TTL(ctx context.Context, key string) (time.Duration, error)
    SetTTL(ctx context.Context, key string, ttl time.Duration) error
}

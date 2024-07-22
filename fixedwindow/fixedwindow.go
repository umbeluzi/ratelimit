package fixedwindow

import (
    "context"
    "sync"
    "time"

    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/storage"
)

// FixedWindow is an implementation of the fixed window rate limiting algorithm.
type FixedWindow struct {
    storage storage.Storage
    config  config.Config
    mu      sync.Mutex
}

// New creates a new FixedWindow rate limiter.
func New(storage storage.Storage, config config.Config) *FixedWindow {
    return &FixedWindow{
        storage: storage,
        config:  config,
    }
}

// Allow checks if a request is allowed for a given key using the fixed window algorithm.
func (fw *FixedWindow) Allow(ctx context.Context, key string) (bool, error) {
    fw.mu.Lock()
    defer fw.mu.Unlock()

    maxRequests, err := fw.config.MaxRequests(ctx)
    if err != nil {
        return false, err
    }

    window, err := fw.config.Interval(ctx)
    if err != nil {
        return false, err
    }

    burstLimit, err := fw.config.BurstLimit(ctx)
    if err != nil {
        return false, err
    }

    count, err := fw.storage.Increment(ctx, key)
    if err != nil {
        return false, err
    }

    if count == 1 {
        // Set a TTL if this is the first request
        err := fw.storage.SetTTL(ctx, key, window)
        if err != nil {
            return false, err
        }
    }

    if count > maxRequests+burstLimit {
        return false, nil
    }

    return true, nil
}

// Quota returns the current quota information.
func (fw *FixedWindow) Quota(ctx context.Context, key string) (int, int, int, error) {
    count, err := fw.storage.Get(ctx, key)
    if err != nil {
        return 0, 0, 0, err
    }

    maxRequests, err := fw.config.MaxRequests(ctx)
    if err != nil {
        return 0, 0, 0, err
    }

    burstLimit, err := fw.config.BurstLimit(ctx)
    if err != nil {
        return 0, 0, 0, err
    }

    return count, maxRequests, burstLimit, nil
}

// NextAllowed returns the time duration until the next allowed request.
func (fw *FixedWindow) NextAllowed(ctx context.Context, key string) (time.Duration, error) {
    ttl, err := fw.storage.TTL(ctx, key)
    if err != nil {
        return 0, err
    }
    return ttl, nil
}

package slidingwindow

import (
    "context"
    "sync"
    "time"

    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/storage"
)

// SlidingWindow is an implementation of the sliding window rate limiting algorithm.
type SlidingWindow struct {
    storage storage.Storage
    config  config.Config
    mu      sync.Mutex
}

// New creates a new SlidingWindow rate limiter.
func New(storage storage.Storage, config config.Config) *SlidingWindow {
    return &SlidingWindow{
        storage: storage,
        config:  config,
    }
}

// Allow checks if a request is allowed for a given key using the sliding window algorithm.
func (sw *SlidingWindow) Allow(ctx context.Context, key string) (bool, error) {
    sw.mu.Lock()
    defer sw.mu.Unlock()

    maxRequests, err := sw.config.MaxRequests(ctx)
    if err != nil {
        return false, err
    }

    window, err := sw.config.Interval(ctx)
    if err != nil {
        return false, err
    }

    burstLimit, err := sw.config.BurstLimit(ctx)
    if err != nil {
        return false, err
    }

    count, err := sw.storage.Increment(ctx, key)
    if err != nil {
        return false, err
    }

    if count == 1 {
        // Set a TTL if this is the first request
        err := sw.storage.SetTTL(ctx, key, window)
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
func (sw *SlidingWindow) Quota(ctx context.Context, key string) (int, int, int, error) {
    count, err := sw.storage.Get(ctx, key)
    if err != nil {
        return 0, 0, 0, err
    }

    maxRequests, err := sw.config.MaxRequests(ctx)
    if err != nil {
        return 0, 0, 0, err
    }

    burstLimit, err := sw.config.BurstLimit(ctx)
    if err != nil {
        return 0, 0, 0, err
    }

    return count, maxRequests, burstLimit, nil
}

// NextAllowed returns the time duration until the next allowed request.
func (sw *SlidingWindow) NextAllowed(ctx context.Context, key string) (time.Duration, error) {
    ttl, err := sw.storage.TTL(ctx, key)
    if err != nil {
        return 0, err
    }
    return ttl, nil
}

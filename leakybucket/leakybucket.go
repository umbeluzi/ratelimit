package leakybucket

import (
    "context"
    "sync"
    "time"

    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/storage"
)

// LeakyBucket is an implementation of the leaky bucket rate limiting algorithm.
type LeakyBucket struct {
    storage storage.Storage
    config  config.Config
    mu      sync.Mutex
}

// New creates a new LeakyBucket rate limiter.
func New(storage storage.Storage, config config.Config) *LeakyBucket {
    return &LeakyBucket{
        storage: storage,
        config:  config,
    }
}

// Allow checks if a request is allowed for a given key using the leaky bucket algorithm.
func (lb *LeakyBucket) Allow(ctx context.Context, key string) (bool, error) {
    lb.mu.Lock()
    defer lb.mu.Unlock()

    maxRequests, err := lb.config.MaxRequests(ctx)
    if err != nil {
        return false, err
    }

    interval, err := lb.config.Interval(ctx)
    if err != nil {
        return false, err
    }

    burstLimit, err := lb.config.BurstLimit(ctx)
    if err != nil {
        return false, err
    }

    count, err := lb.storage.Increment(ctx, key)
    if err != nil {
        return false, err
    }

    if count == 1 {
        // Set a TTL if this is the first request
        err := lb.storage.SetTTL(ctx, key, interval)
        if err != nil {
            return false, err
        }
    }

    if count > maxRequests+burstLimit {
        return false, nil
    }

    // Simulate leaking by decrementing the count after the interval
    go func() {
        time.Sleep(interval)
        lb.storage.Reset(ctx, key)
    }()

    return true, nil
}

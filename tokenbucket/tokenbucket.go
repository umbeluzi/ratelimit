package tokenbucket

import (
    "context"
    "sync"
    "time"

    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/storage"
)

// TokenBucket is an implementation of the token bucket rate limiting algorithm.
type TokenBucket struct {
    storage     storage.Storage
    config      config.Config
    mu          sync.Mutex
    ticker      *time.Ticker
    stopChannel chan struct{}
}

// New creates a new TokenBucket rate limiter.
func New(storage storage.Storage, config config.Config) *TokenBucket {
    tb := &TokenBucket{
        storage:     storage,
        config:      config,
        stopChannel: make(chan struct{}),
    }
    tb.startRefill()
    return tb
}

// startRefill starts the refill ticker.
func (tb *TokenBucket) startRefill() {
    interval, _ := tb.config.Interval(context.Background())
    tb.ticker = time.NewTicker(interval)
    go func() {
        for {
            select {
            case <-tb.ticker.C:
                tb.mu.Lock()
                tb.refillTokens()
                tb.mu.Unlock()
            case <-tb.stopChannel:
                tb.ticker.Stop()
                return
            }
        }
    }()
}

// refillTokens refills the bucket with tokens at the defined refill rate.
func (tb *TokenBucket) refillTokens() {
    interval, err := tb.config.Interval(context.Background())
    if err == nil {
        now := time.Now()
        lastRefill, _ := tb.config.LastRefill(context.Background())
        elapsed := now.Sub(lastRefill)
        tokensToAdd := int(elapsed / interval)
        tokens, _ := tb.config.Tokens(context.Background())
        tokens += tokensToAdd
        maxTokens, _ := tb.config.MaxRequests(context.Background())
        if tokens > maxTokens {
            tokens = maxTokens
        }
        tb.config.SetTokens(context.Background(), tokens)
        tb.config.SetLastRefill(context.Background(), now)
    }
}

// Stop stops the refill ticker for graceful shutdown.
func (tb *TokenBucket) Stop() {
    close(tb.stopChannel)
}

// Allow checks if a request is allowed for a given key using the token bucket algorithm.
func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    tokens, err := tb.config.Tokens(ctx)
    if err != nil {
        return false, err
    }

    burstLimit, err := tb.config.BurstLimit(ctx)
    if err != nil {
        return false, err
    }

    if tokens > 0 {
        tb.config.SetTokens(ctx, tokens-1)
        return true, nil
    }

    count, err := tb.storage.Increment(ctx, key)
    if err != nil {
        return false, err
    }

    if count > burstLimit {
        return false, nil
    }

    return true, nil
}

// Quota returns the current quota information.
func (tb *TokenBucket) Quota(ctx context.Context, key string) (int, int, int, error) {
    count, err := tb.storage.Get(ctx, key)
    if err != nil {
        return 0, 0, 0, err
    }

    maxRequests, err := tb.config.MaxRequests(ctx)
    if err != nil {
        return 0, 0, 0, err
    }

    burstLimit, err := tb.config.BurstLimit(ctx)
    if err != nil {
        return 0, 0, 0, err
    }

    return count, maxRequests, burstLimit, nil
}

// NextAllowed returns the time duration until the next allowed request.
func (tb *TokenBucket) NextAllowed(ctx context.Context, key string) (time.Duration, error) {
    ttl, err := tb.storage.TTL(ctx, key)
    if err != nil {
        return 0, err
    }
    return ttl, nil
}

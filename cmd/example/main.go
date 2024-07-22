package main

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/fixedwindow"
    "github.com/umbeluzi/ratelimit/leakybucket"
    "github.com/umbeluzi/ratelimit/slidingwindow"
    "github.com/umbeluzi/ratelimit/storage"
    "github.com/umbeluzi/ratelimit/tokenbucket"
)

type InMemoryStorage struct {
    data map[string]int
    ttl  map[string]time.Time
    mu   sync.Mutex
}

func NewInMemoryStorage() *InMemoryStorage {
    return &InMemoryStorage{
        data: make(map[string]int),
        ttl:  make(map[string]time.Time),
    }
}

func (s *InMemoryStorage) Increment(ctx context.Context, key string) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.data[key]++
    return s.data[key], nil
}

func (s *InMemoryStorage) Reset(ctx context.Context, key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.data[key] = 0
    return nil
}

func (s *InMemoryStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    return time.Until(s.ttl[key]), nil
}

func (s *InMemoryStorage) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.ttl[key] = time.Now().Add(ttl)
    return nil
}

func main() {
    ctx := context.Background()
    storage := NewInMemoryStorage()

    // Example configuration with burst limit
    config := config.NewStatic(5, time.Minute, 2, 0, time.Now())

    // Example using Leaky Bucket algorithm
    leakyBucket := leakybucket.New(storage, config)
    fmt.Println("Testing Leaky Bucket:")
    for i := 0; i < 10; i++ {
        allowed, err := leakyBucket.Allow(ctx, "leakybucket_key")
        if err != nil {
            fmt.Println("Error:", err)
        }

        if allowed {
            fmt.Println("Request allowed")
        } else {
            fmt.Println("Request denied")
        }

        time.Sleep(time.Second)
    }

    // Example using Token Bucket algorithm with burst support
    tokenBucket := tokenbucket.New(storage, config)
    fmt.Println("Testing Token Bucket:")
    for i := 0; i < 10; i++ {
        allowed, err := tokenBucket.Allow(ctx, "tokenbucket_key")
        if err != nil {
            fmt.Println("Error:", err)
        }

        if allowed {
            fmt.Println("Request allowed")
        } else {
            fmt.Println("Request denied")
        }

        time.Sleep(time.Second)
    }

    // Example using Fixed Window algorithm
    fixedWindow := fixedwindow.New(storage, config)
    fmt.Println("Testing Fixed Window:")
    for i := 0; i < 10; i++ {
        allowed, err := fixedWindow.Allow(ctx, "fixedwindow_key")
        if err != nil {
            fmt.Println("Error:", err)
        }

        if allowed {
            fmt.Println("Request allowed")
        } else {
            fmt.Println("Request denied")
        }

        time.Sleep(time.Second)
    }

    // Example using Sliding Window algorithm
    slidingWindow := slidingwindow.New(storage, config)
    fmt.Println("Testing Sliding Window:")
    for i := 0; i < 10; i++ {
        allowed, err := slidingWindow.Allow(ctx, "slidingwindow_key")
        if err != nil {
            fmt.Println("Error:", err)
        }

        if allowed {
            fmt.Println("Request allowed")
        } else {
            fmt.Println("Request denied")
        }

        time.Sleep(time.Second)
    }
}

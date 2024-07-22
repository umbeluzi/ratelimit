# Rate Limit Library

A Go library for rate limiting with various algorithms. The library supports the following rate limiting algorithms:

- [Fixed Window](fixedwindow)
- [Leaky Bucket](leakybucket)
- [Sliding Window](slidingwindow)
- [Token Bucket](tokenbucket)

## Installation

Use `go get` to install the library:

```sh
go get github.com/umbeluzi/ratelimit
```

## Usage

You can find example usage in the `cmd/example` directory.

## Implementing Storage

The `Storage` interface allows you to implement your own storage backend for rate limiting. The interface requires the following methods:

```go
type Storage interface {
    Increment(ctx context.Context, key string) (int, error)
    Reset(ctx context.Context, key string) error
    TTL(ctx context.Context, key string) (time.Duration, error)
    SetTTL(ctx context.Context, key string, ttl time.Duration) error
    Get(ctx context.Context, key string) (int, error)
}
```

### Example: In-Memory Storage

```go
package storage

import (
    "context"
    "sync"
    "time"
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

func (s *InMemoryStorage) Get(ctx context.Context, key string) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    return s.data[key], nil
}
```

### Example: Redis Storage

```go
package storage

import (
    "context"
    "github.com/redis/go-redis/v9"
    "time"
)

type RedisStorage struct {
    client *redis.Client
}

func NewRedisStorage(client *redis.Client) *RedisStorage {
    return &RedisStorage{client: client}
}

func (s *RedisStorage) Increment(ctx context.Context, key string) (int, error) {
    result, err := s.client.Incr(ctx, key).Result()
    return int(result), err
}

func (s *RedisStorage) Reset(ctx context.Context, key string) error {
    return s.client.Del(ctx, key).Err()
}

func (s *RedisStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
    result, err := s.client.TTL(ctx, key).Result()
    return result, err
}

func (s *RedisStorage) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
    return s.client.Expire(ctx, key, ttl).Err()
}

func (s *RedisStorage) Get(ctx context.Context, key string) (int, error) {
    result, err := s.client.Get(ctx, key).Int()
    return result, err
}
```

### Example: Memcached Storage

```go
package storage

import (
    "context"
    "github.com/bradfitz/gomemcache/memcache"
    "strconv"
    "time"
)

type MemcachedStorage struct {
    client *memcache.Client
}

func NewMemcachedStorage(client *memcache.Client) *MemcachedStorage {
    return &MemcachedStorage{client: client}
}

func (s *MemcachedStorage) Increment(ctx context.Context, key string) (int, error) {
    err := s.client.Increment(key, 1)
    if err == memcache.ErrCacheMiss {
        err = s.client.Set(&memcache.Item{Key: key, Value: []byte("1")})
        return 1, err
    }
    if err != nil {
        return 0, err
    }
    item, err := s.client.Get(key)
    if err != nil {
        return 0, err
    }
    result, err := strconv.Atoi(string(item.Value))
    return result, err
}

func (s *MemcachedStorage) Reset(ctx context.Context, key string) error {
    return s.client.Delete(key)
}

func (s *MemcachedStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
    return time.Minute, nil // Memcached does not support TTL retrieval
}

func (s *MemcachedStorage) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
    return s.client.Touch(key, int32(ttl.Seconds()))
}

func (s *MemcachedStorage) Get(ctx context.Context, key string) (int, error) {
    item, err := s.client.Get(key)
    if err != nil {
        return 0, err
    }
    result, err := strconv.Atoi(string(item.Value))
    return result, err
}
```

## Implementing Config

The `Config` interface allows you to implement your own configuration for rate limiting. The interface requires the following methods:

```go
type Config interface {
    MaxRequests(ctx context.Context) (int, error)
    Interval(ctx context.Context) (time.Duration, error)
    BurstLimit(ctx context.Context) (int, error)
    Tokens(ctx context.Context) (int, error)
    SetTokens(ctx context.Context, tokens int) error
    LastRefill(ctx context.Context) (time.Time, error)
    SetLastRefill(ctx context.Context, lastRefill time.Time) error
}
```

### Example: Static Config

```go
package config

import (
    "context"
    "sync"
    "time"
)

type Static struct {
    maxRequests int
    interval    time.Duration
    burstLimit  int
    tokens      int
    lastRefill  time.Time
    mu          sync.Mutex
}

func NewStatic(maxRequests int, interval time.Duration, burstLimit int, tokens int, lastRefill time.Time) *Static {
    return &Static{
        maxRequests: maxRequests,
        interval:    interval,
        burstLimit:  burstLimit,
        tokens:      tokens,
        lastRefill:  lastRefill,
    }
}

func (c *Static) MaxRequests(ctx context.Context) (int, error) {
    return c.maxRequests, nil
}

func (c *Static) Interval(ctx context.Context) (time.Duration, error) {
    return c.interval, nil
}

func (c *Static) BurstLimit(ctx context.Context) (int, error) {
    return c.burstLimit, nil
}

func (c *Static) Tokens(ctx context.Context) (int, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.tokens, nil
}

func (c *Static) SetTokens(ctx context.Context, tokens int) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.tokens = tokens
    return nil
}

func (c *Static) LastRefill(ctx context.Context) (time.Time, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.lastRefill, nil
}

func (c *Static) SetLastRefill(ctx context.Context, lastRefill time.Time) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.lastRefill = lastRefill
    return nil
}
```

## Algorithms

### Fixed Window

A fixed window rate limiter.

### Leaky Bucket

A leaky bucket rate limiter.

### Sliding Window

A sliding window rate limiter.

### Token Bucket

A token bucket rate limiter.

## Example Usage

See the `cmd/example` directory for usage examples.

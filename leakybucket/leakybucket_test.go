package leakybucket

import (
    "context"
    "testing"
    "time"

    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/storage"
)

type MockStorage struct {
    count int
}

func (ms *MockStorage) Increment(ctx context.Context, key string) (int, error) {
    ms.count++
    return ms.count, nil
}

func (ms *MockStorage) Reset(ctx context.Context, key string) error {
    ms.count = 0
    return nil
}

func (ms *MockStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
    return time.Minute, nil
}

func (ms *MockStorage) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
    return nil
}

func (ms *MockStorage) Get(ctx context.Context, key string) (int, error) {
    return ms.count, nil
}

func TestLeakyBucket_Allow(t *testing.T) {
    storage := &MockStorage{}
    config := config.NewStatic(5, time.Minute, 2, 0, time.Now())

    lb := New(storage, config)

    for i := 0; i < 7; i++ {
        allowed, err := lb.Allow(context.Background(), "test")
        if err != nil {
            t.Errorf("unexpected error: %v", err)
        }

        if i < 5 && !allowed {
            t.Errorf("request %d should be allowed", i+1)
        }

        if i >= 5 && allowed {
            t.Errorf("request %d should be denied", i+1)
        }
    }
}

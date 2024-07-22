# Token Bucket Rate Limiter

The Token Bucket rate limiting algorithm allows bursts of requests up to a maximum capacity and refills tokens at a steady rate.

## Usage

```go
import (
    "context"
    "fmt"
    "time"
    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/storage"
    "github.com/umbeluzi/ratelimit/tokenbucket"
)

func main() {
    ctx := context.Background()
    storage := storage.NewInMemoryStorage()
    config := config.NewStatic(5, time.Minute, 2, 0, time.Now())

    tokenBucket := tokenbucket.New(storage, config)
    defer tokenBucket.Stop() // Ensure the ticker is stopped for graceful shutdown

    allowed, err := tokenBucket.Allow(ctx, "test_key")
    if err != nil {
        fmt.Println("Error:", err)
    }
    if allowed {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request denied")
    }

    // Quota information
    count, maxRequests, burstLimit, err := tokenBucket.Quota(ctx, "test_key")
    if err != nil {
        fmt.Println("Error:", err)
    }
    fmt.Printf("Quota - Count: %d, MaxRequests: %d, BurstLimit: %d\n", count, maxRequests, burstLimit)

    // Retry-After header
    retryAfter, err := tokenBucket.NextAllowed(ctx, "test_key")
    if err != nil {
        fmt.Println("Error:", err)
    }
    fmt.Printf("Retry-After: %s\n", retryAfter)
}
```

## Implementing Storage

You can use any storage backend that implements the `Storage` interface. See the main project README for examples.

## Implementing Config

You can use any configuration that implements the `Config` interface. See the main project README for examples.

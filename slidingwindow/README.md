# Sliding Window Rate Limiter

The Sliding Window rate limiting algorithm allows a fixed number of requests within a given window of time, adjusted to the current time.

## Usage

```go
import (
    "context"
    "time"
    "github.com/umbeluzi/ratelimit/config"
    "github.com/umbeluzi/ratelimit/slidingwindow"
    "github.com/umbeluzi/ratelimit/storage"
)

func main() {
    ctx := context.Background()
    storage := NewInMemoryStorage()
    config := config.NewStatic(5, time.Minute, 2, 0, time.Now())

    slidingWindow := slidingwindow.New(storage, config)
    allowed, err := slidingWindow.Allow(ctx, "test_key")
    if err != nil {
        fmt.Println("Error:", err)
    }
    if allowed {
        fmt.Println("Request allowed")
    } else {
        fmt.Println("Request denied")
    }
}
```

## Implementing Storage

You can use any storage backend that implements the `Storage` interface. See the main project README for examples.

## Implementing Config

You can use any configuration that implements the `Config` interface. See the main project README for examples.

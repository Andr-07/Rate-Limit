# Rate Limiter Middleware for Go

A flexible and extensible rate limiter middleware for Go HTTP servers.  
Supports per-user and per-IP limits, Redis backend, optional Kafka logging, and can be used as middleware in any HTTP server.

---

## Features

- ✅ Rate limiting by **User ID** and **IP address**
- ✅ Redis-based sliding time window
- ✅ Temporary blocking after exceeding limits
- ✅ Optional logging to **Kafka**
- ✅ Simple integration with `http.Handler`

---

## Installation

Use `go get` to install the package:

```
go get github.com/your-username/rate-limiter
```
---

## Usage

### 1. Import and initialize

```
import (
    "time"
    "github.com/your-username/rate-limiter"
)

config := &limiter.RateLimiterConfig{
    MaxRequestsPerUser: 10,
    MaxRequestsPerIP:   20,
    TimeWindow:         time.Minute,
    BlockDuration:      5 * time.Minute,
    EnableKafkaLog:     true,
}

rateLimiter := limiter.New(redisClient, kafkaClient, config)
```

### 2. Apply as middleware

```
http.Handle("/api", rateLimiter.Middleware(myHandler))
```

---

## Request Headers

The middleware expects the header:

- `X-User-ID`: to uniquely identify the user

If the header is missing, it will return `400 Bad Request`.

---

## Context Usage

You can retrieve user ID and IP address from context in downstream handlers:

```
userID := r.Context().Value(limiter.ContextUserKey).(string)  
ip := r.Context().Value(limiter.ContextIpKey).(string)
```

---

## Example Output (with Kafka logging enabled)

```
{
  "user_id": "12345",
  "timestamp": "2025-04-29T12:00:00Z",
  "reason": "Rate limit exceeded"
}
```

---

## License

MIT License – see LICENSE for details.

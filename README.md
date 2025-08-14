
# 🚀 Project Rate Limiter

**Author:** Dave Russell

## 📦 How to Run

### 1️⃣ Using Docker (Recommended)

With Docker Compose, all application dependencies (including Redis) will be installed, built, and started automatically.

**Steps:**

1. Ensure **Docker** & **Docker Compose** are installed.
2. From the project root directory, run:

   ```bash
   docker-compose up --build
   ```
3. The application and Redis will start automatically (Default port: **8080**).

### 2️⃣ Manual Setup (Without Docker)

If you prefer to run it manually, ensure:

* **Go** version `>= 1.24.1` is installed.
* **Redis** (latest version recommended) is installed and running.

**Steps:**

1. Start Redis on `localhost:6379`.

2. Install Go dependencies:

   ```bash
   go mod tidy
   ```

3. Run the application:

   ```bash
   go run ./cmd/server
   ```

4. Available endpoints:

   * `GET http://localhost:8080//ipaddress/ping` → rate limiter using **IP address** as the key.
   * `GET http://localhost:8080//apikey/ping` → rate limiter using **API key** as the key.
   * `GET http://localhost:8080//ipaddress/ping` → rate limiter using **IP address** as the key.
   * `GET http://localhost:8080//apikey/ping` → rate limiter using **API key** as the key.

## 🧪 Running Tests

### Using Makefile

A **Makefile** is provided for convenience:

```bash
make test      # Run all tests + generate coverage output
make coverage  # Open the coverage results in HTML format
make clean     # Remove generated coverage output files
```

### Manual Test Execution

Run:

```bash
# Run tests with verbose output and create coverage.out
go test -v -coverprofile="coverage.out" ./...

# Open HTML coverage report
go tool cover -html="coverage.out"
```

## 🎨 Design Overview

### Project Structure
```
DOITPAY/
├── cmd/                     # Entry point of executable app
├── internal/                # Private application code
│   ├── config/              # Configuration management
│   ├── domain/              # Domain models and business logic
│   ├── memory/              # In-memory storage implementations
│   ├── rdb/                 # Redis storage implementations
│   ├── rest/                # REST API related
│   ├── service/             # Business logic services
│   └── util/                # Utility functions and helpers
```

### Design Pattern

I’m using a service + repository layer pattern, where the business logic code is written in the internal/service package, while the repository implementations are placed in separate packages according to the database or storage being used. For example, internal/memory contains repository implementations for storing data in the app’s memory, whereas internal/rdb contains repository implementations for storing data in Redis.

The interface definitions are placed where they are actually needed. For example, since the repository layer is used by the service layer, the service layer is responsible for defining the repository interfaces. This approach prevents the service layer from having a direct dependency on the repository layer, which helps reduce the risk of a dependency cycle. \
For example:
```go
type TokenBucketRepository interface {
	GetBucket(ctx context.Context, clientID string) (ratelimit.TokenBucket, error)
	SaveBucket(ctx context.Context, clientID string, bucket ratelimit.TokenBucket) error
}

type TokenBucketService struct {
	repo  TokenBucketRepository
	cfg   config.TokenBucket
	locks *util.StripedMutex
}
```

In the middleware itself, I define an interface for the RateLimiter, which only requires a single method: Allow(). This way, regardless of which algorithm is used, the core logic only needs to be implemented in the Allow() method to determine whether a request from a given key should be allowed or rejected.
```go
type RateLimiter interface {
	Allow(ctx context.Context, clientID string) (bool, error)
}

func RateLimit(rateLimiter RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := keyFunc(c)
		if clientID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing key"})
			c.Abort()
			return
		}

		allowed, err := rateLimiter.Allow(c.Request.Context(), clientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal rate limiter error"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

For configuration, i use **YAML files** to define rate-limiting settings.  
These configurations include parameters such as:

- **Max tokens** for the Fixed Window algorithm
- **Fill rate** for the Token Bucket algorithm

Using YAML allows us to easily adjust rate-limiting values without changing the code.
```yaml
server:
  host: 0.0.0.0
  port: 8080

redis:
  host: redis # Use redis as hostname if you use redis from docker
  port: 6379
  password: ""
  fixed-window-db: 0
  token-bucket-db: 1

rate-limiter:
  fixed-window:
    max-requests: 5
    time-frame-ms: 60000
  token-bucket:
    max-tokens: 2
    refill-rate: 1 # token/s
```

## ⚙️ Rate-Limiting Algorithms

In this project, we use **two rate-limiting algorithms** to control request traffic:

### 1. Fixed Window
- Counts the number of requests per client within a fixed time window (e.g., 1 minute).
- Once the limit is reached, all further requests are rejected until the next window.
- Simple and easy to implement but can cause spikes at the window boundaries.

### 2. Token Bucket (for handling burst traffic)
- Maintains a “bucket” of tokens, each request consumes a token.
- Tokens are refilled at a fixed rate.
- Allows occasional bursts of requests without rejecting them unnecessarily.
- Provides smoother traffic handling compared to fixed window.

### 🔒 Handling Concurrency

Since we are using a `map` for in-memory storage, we need to use **mutexes** to synchronize read and write operations
However, a simple mutex can still lead to **race conditions**: if two goroutines with the same `clientID` access the map simultaneously, one might read the data before the other writes it, causing incorrect rate-limit tracking.

To solve this, we use a **striped mutex** at the **service layer**, keyed by `clientID`.

### Benefits of using striped mutex:
- **Per-key locking:** Only goroutines operating on the same `clientID` will block each other.
- **No global lock:** Requests for different `clientID`s can proceed concurrently, improving throughput.
- Ensures correct rate-limit updates even under concurrent access.

## 📌 Assumptions & Limitations

- **Assumptions:**
  - The system is currently designed to run as a **single instance**
  - Rate limiting and concurrency control rely on the **in-memory striped mutex**, which only tracks locks within the same instance.

- **Limitations:**
  - If multiple instances of the service are running, the striped mutex **cannot coordinate locks across instances**.
  - This means concurrent requests with the same `clientID` on different instances may bypass the local lock, potentially causing inconsistent rate-limit tracking.
  - Handling bursts and distributed scenarios would require a **distributed locking mechanism** (e.g., Redis-based locks) to ensure consistency across instances.

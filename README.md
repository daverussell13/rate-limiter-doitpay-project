# 🚀 Project Rate Limiter
**Author:** Dave Russell

---

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

---

### 2️⃣ Manual Setup (Without Docker)
If you prefer to run it manually, ensure:
- **Go** version `>= 1.24.1` is installed.
- **Redis** (latest version recommended) is installed and running.

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
   - `GET http://localhost:8080/ipaddress/ping` → rate limiter using **IP address** as the key.
   - `GET http://localhost:8080/apikey/ping` → rate limiter using **API key** as the key.

---

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

---
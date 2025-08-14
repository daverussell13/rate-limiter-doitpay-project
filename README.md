# 🚀 Project Rate Limiter
**Author:** Dave Russell

## 📦 How to Run

### 1️⃣ Using Docker (Recommended)
Dengan Docker Compose, semua kebutuhan aplikasi (termasuk Redis) akan otomatis di-install, dibuild, dan dijalankan.

**Langkah:**
1. Pastikan **Docker** & **Docker Compose** sudah ter-install.
2. Di root directory project, jalankan:
   ```bash
   docker-compose up --build
   ```
3. Aplikasi dan Redis akan berjalan otomatis (Port yang digunakan 8080)

---

### 2️⃣ Manual Setup (Without Docker)
Kalau mau jalanin manual, pastikan:
- **Go** versi `>= 1.24.1` sudah ter-install.
- **Redis** versi terbaru sudah terpasang dan berjalan.

**Langkah:**
1. Jalankan Redis di `localhost:6379`.

2. Install dependencies Go:
   ```bash
   go mod tidy
   ```

3. Jalankan aplikasi:
   ```bash
   go run ./cmd/server
   ```

4. Akses endpoint:
   - `GET http://localhost:8080/ipaddress/ping` → endpoint rate limiter menggunakan ip address sebagai key
   - `GET http://localhost:8080/apikey/ping` → endpoint rate limiter menggunakan api key sebagai key
---
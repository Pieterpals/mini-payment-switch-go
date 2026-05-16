# 💳 Mini Payment Switch (Go)

Halo! Ini adalah proyek iseng tapi serius saya buat belajar bikin **Modular Monolith** yang *scalable* pakai Go. Intinya sih ini simulasi *payment switch* sederhana yang udah dilengkapi sama berbagai "mainan" keren biar berasa kayak aplikasi production beneran.

## 🚀 Apa aja sih isinya?
Tadinya aplikasi ini cuma punya satu endpoint buat proses bayar, tapi baru aja saya refactor jadi lebih rapi:
- **Inquiry**: Cek akun & hitung biaya admin (disimpen di Redis biar cepet).
- **Payment Execute**: Eksekusi transaksi, simpen ke Postgres, terus lempar event ke Kafka.
- **Status Check**: Buat ngecek transaksi tadi sukses atau nggak.

## 🛠 Tech Stack
- **Framework**: Echo (Go)
- **Database**: PostgreSQL (buat data transaksi)
- **Caching & Lock**: Redis (buat inquiry & cegah transaksi double/idempotency)
- **Messaging**: Kafka (buat kirim notifikasi async)
- **Observability**: Prometheus & Grafana (metrik) + Loki (log)
- **Distributed Tracing**: OpenTelemetry & Jaeger (biar bisa liat waterfall trace antar komponen)
- **API Docs**: Swagger/OpenAPI (biar nggak bingung cara nembak API-nya)

## 📦 Cara Jalaninnya
Gampang banget, tinggal pastiin Docker udah nyala:

1. **Nyalain Infrastruktur** (DB, Redis, Kafka, Jaeger, dll):
   ```bash
   docker-compose up -d
   ```

2. **Jalanin Aplikasinya**:
   ```bash
   go run cmd/api/main.go
   ```

3. **Cek Dokumentasi API**:
   Buka browser ke [http://localhost:8182/swagger/index.html](http://localhost:8182/swagger/index.html)

## 📊 Monitoring & Tracing
Kalo mau liat "jeroan" aplikasinya pas lagi jalan:
- **Jaeger (Tracing)**: [http://localhost:16686](http://localhost:16686) — Cek durasi eksekusi tiap fungsi di sini.
- **Grafana**: [http://localhost:3000](http://localhost:3000) (User/Pass: `admin/admin`) — Buat liat log & metrik.

## 🧪 Testing
Ada skrip buat simulasi banyak transaksi sekaligus di folder `scripts/loadtest.go`. Lumayan buat liat gimana sistem nanganin *concurrency* dan *race condition*.

---
Dibuat dengan ❤️ sambil ngopi.

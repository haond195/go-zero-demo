# Order Hub - Full-Stack Go-Zero Microservices Architecture

Dự án mẫu thực chiến chuẩn chỉ kiến trúc sạch (Clean Architecture & Microservices) sử dụng framework **go-zero**, tích hợp toàn diện 7 thành phần cốt lõi trong hệ sinh thái go-zero:

1. **Routes**: Thiết kế qua file `order-hub.api` với prefix `/api/v1`, declarative routing và grouping.
2. **Validate Request**: Kiểm tra dữ liệu đầu vào tự động ở tầng HTTP bằng mapping tag (`email`, `gte=1`, `gt=0`).
3. **Middleware**: Chuỗi Middleware xử lý `AuditLogMiddleware` (ghi log, tính thời gian, RBAC header) và `RateLimitMiddleware`.
4. **Database & Cache kép**: PostgreSQL qua `sqlx`, kết hợp Redis Cache qua `sqlc` (mô hình Cache-Aside tự động xóa cache khi cập nhật).
5. **Traffic Control & Rate Limiting**: Sử dụng `core/limit.PeriodLimit` phân tán bằng Redis Lua script nguyên tử (chống spam/DDoS, trả mã `HTTP 429`).
6. **Message Queue (Apache Kafka)**: Tích hợp thư viện `github.com/zeromicro/go-queue/kq` với `kq.Pusher` và worker `kq.MustNewQueue` độc lập.
7. **Microservices (zRPC & Protocol Buffers)**: Dịch vụ thanh toán tách rời `payment-rpc` giao tiếp nhị phân gRPC tốc độ cao (sub-millisecond ~0.3ms), tự động cân bằng tải P2C (Power of Two Choices) và truyền `trace_id` xuyên suốt.

---

## Kiến trúc hệ thống

```text
[ Client / Web / App ]
          │ (HTTP REST / JSON)
          ▼
┌──────────────────────────────────────────────────────────┐
│             ORDER-HUB API GATEWAY (Port 8888)            │
│  - Middleware: Audit Log, RBAC, Rate Limiter (Redis)     │
│  - Request Validation Tag                                │
│  - Prometheus Metrics (Port 9091)                        │
└─────┬───────────────────┬──────────────────────┬─────────┘
      │                   │                      │
(gRPC / Protobuf)   (Cache-Aside)         (Publish Event)
      ▼                   ▼                      ▼
┌──────────────┐    ┌──────────────┐       ┌──────────────┐
│ PAYMENT-RPC  │    │  POSTGRESQL  │       │ APACHE KAFKA │
│ (Port 8088)  │    │  (Port 5432) │       │ (Port 9092)  │
└──────────────┘    └──────┬───────┘       └──────┬───────┘
                           │                      │
                           ▼                      ▼
                    ┌──────────────┐       ┌──────────────┐
                    │  REDIS 7     │       │ KAFKA WORKER │
                    │  (Port 6379) │       │  (Consumer)  │
                    └──────────────┘       └──────────────┘
```

---

## Cấu trúc thư mục

```text
order-hub/
├── cmd/
│   └── consumer/main.go              <-- Kafka Consumer Worker độc lập (go-queue/kq)
├── etc/
│   └── orderhub-api.yaml             <-- Cấu hình API Gateway (Postgres, Redis, Kafka, RPC)
├── internal/
│   ├── config/config.go              <-- Khai báo Struct cấu hình hệ thống
│   ├── handler/                      <-- Tầng tiếp nhận HTTP Request
│   ├── logic/                        <-- Tầng nghiệp vụ (gọi gRPC, lưu DB, bắn Kafka)
│   ├── middleware/                   <-- AuditLog (RBAC) & RateLimiter (PeriodLimit)
│   ├── svc/servicecontext.go         <-- Túi quản lý tài nguyên chung (Context injection)
│   └── types/types.go                <-- Kiểu dữ liệu Request / Response
├── model/
│   └── ordermodel.go                 <-- Tầng CSDL (sqlc: PostgreSQL + Redis Cache)
├── rpc/
│   └── payment/                      <-- Microservice Payment độc lập
│       ├── client/payment/           <-- Client wrapper tự động sinh bởi goctl
│       ├── etc/payment.yaml          <-- Cấu hình Payment RPC (Port 8088)
│       ├── internal/logic/           <-- Nghiệp vụ xử lý thanh toán
│       ├── internal/server/          <-- Router gRPC Server
│       ├── payment/                  <-- Mã nguồn sinh từ Protobuf (*.pb.go)
│       ├── payment.proto             <-- File thiết kế hợp đồng dữ liệu Protocol Buffers v3
│       └── payment.go                <-- Entrypoint khởi động Payment RPC Server
├── schema.sql                        <-- DDL tạo bảng orders trong PostgreSQL
├── Dockerfile                        <-- Dockerfile multi-stage build cho API Gateway
├── Dockerfile.payment                <-- Dockerfile build cho Payment RPC Service
├── docker-compose.yml                <-- Khởi chạy toàn bộ cụm: Postgres, Redis, Kafka, RPC, App
├── order-hub.api                     <-- Bản thiết kế API DSL duy nhất
└── orderhub.go                       <-- Entrypoint API Gateway
```

---

## Hướng dẫn khởi chạy (1 lệnh duy nhất)

Toàn bộ hệ thống (PostgreSQL, Redis, Apache Kafka, Payment-RPC và API Gateway) được đóng gói sẵn qua Docker Compose:

```bash
# Khởi động toàn bộ cụm dịch vụ
docker compose up -d --build

# Xem log tổng thể thời gian thực
docker compose logs -f
```

Kiểm tra trạng thái các container đang chạy:
```bash
docker ps
```
Cụm bao gồm:
- `orderhub-app`: Port `8888` (API) & `9091` (Prometheus Metrics)
- `orderhub-payment-rpc`: Port `8088` (gRPC Microservice)
- `orderhub-kafka`: Port `9092` (Kafka Broker)
- `orderhub-redis`: Port `6379` (Redis Cache & Rate Limiter)
- `orderhub-postgres`: Port `5432` (PostgreSQL Database)

---

## Kịch bản kiểm thử thực chiến

### 1. Bật Kafka Consumer Worker (Cửa sổ Terminal riêng)
```powershell
go run cmd/consumer/main.go
```
Worker sẽ lắng nghe topic `order-created-topic` tại `localhost:9092`.

---

### 2. Tạo đơn hàng mới (Kích hoạt trọn vẹn chuỗi liên hoàn)
Request yêu cầu Header `X-Role: admin`:
```powershell
$body = '{"customer_name": "Nguyen Van A", "email": "nguyenvana@gmail.com", "product_code": "MACBOOK-PRO", "quantity": 1, "amount": 2500.0}'
Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders" -Method Post -Headers @{"X-Role"="admin"} -ContentType "application/json" -Body $body
```
**Luồng xử lý tự động ngầm**:
1. Middleware kiểm tra quyền admin và rate limit hợp lệ.
2. Lưu đơn hàng vào PostgreSQL và cập nhật Cache Redis.
3. Gọi sang **Payment-RPC qua gRPC** trừ tiền (xử lý siêu tốc trong `0.3ms`, đồng bộ `trace_id`).
4. Bắn Event vào **Kafka topic** `order-created-topic`.
5. Worker Consumer in ngay báo cáo hóa đơn trên terminal riêng.

---

### 3. Thử nghiệm Cache Hit vs Cache Miss (Tốc độ x16 lần)
```powershell
# Xóa cache chìa khóa trong Redis
docker exec orderhub-redis redis-cli del "cache:orders:orderId:ORD-523813"

# Lần 1: Cache Miss -> Truy vấn PostgreSQL (mất ~11.4ms)
Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders/ORD-523813" -Headers @{"X-Role"="admin"}

# Lần 2: Cache Hit -> Đọc trực tiếp từ Redis RAM (mất ~0.7ms, không tốn câu lệnh SQL)
Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders/ORD-523813" -Headers @{"X-Role"="admin"}
```

---

### 4. Thử nghiệm Distributed Rate Limiting (Chặn Spam HTTP 429)
Quy định hạn mức: Tối đa 3 request / 10 giây trên mỗi IP. Bắn liên tiếp 4 request:
```powershell
1..4 | ForEach-Object {
    $i = $_
    try {
        $res = Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders/ORD-523813" -Headers @{"X-Role"="admin"}
        Write-Host "Lan $i -> OK"
    } catch {
        Write-Host "Lan $i -> BI CHAN 429: $($_.ErrorDetails.Message)"
    }
}
```
*Kết quả*: Lần 1, 2, 3 thành công. Lần 4 bị chặn tức thì với mã lỗi `HTTP 429 Too Many Requests`.

---

### 5. Thử nghiệm Request Validation (Chặn lỗi dữ liệu rác HTTP 400)
Gửi đơn hàng thiếu trường bắt buộc `email`:
```powershell
try {
    Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders" -Method Post -Headers @{"X-Role"="admin"} -ContentType "application/json" -Body '{"customer_name":"A","quantity":1,"amount":10}'
} catch {
    $_.ErrorDetails.Message
}
```
*Kết quả*: Trả về `field "email" is not set` (HTTP 400) ngay tại tầng parse HTTP.

---

### 6. Thử nghiệm RBAC Middleware (Chặn truy cập trái phép HTTP 403)
Gửi request không kèm Header `X-Role: admin`:
```powershell
try {
    Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders" -Method Post -ContentType "application/json" -Body '{"customer_name":"B","email":"b@gmail.com","product_code":"P1","quantity":1,"amount":10}'
} catch {
    $_.ErrorDetails.Message
}
```
*Kết quả*: Trả về `Tu choi: Yeu cau Header X-Role: admin de tao don hang!` (HTTP 403).

---

### 7. Xem Metric Prometheus & Giám sát sức khỏe
Truy cập endpoint metrics được go-zero tự động cung cấp:
```powershell
Invoke-RestMethod -Uri "http://localhost:9091/metrics"
```
Bao gồm toàn bộ thống kê QPS, HTTP status latency histogram, CPU, Memory usage và Adaptive Load Shedding.

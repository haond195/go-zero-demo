# Order Hub - Go-Zero Production-Grade Demo

Dự án mẫu chuẩn chỉ kiến trúc sạch (Clean Architecture) sử dụng framework **go-zero**, tích hợp toàn diện 7 thành phần cốt lõi:

1. **Routes**: Thiết kế qua file `order-hub.api` với prefix `/api/v1` và grouping.
2. **Validate Request**: Kiểm tra dữ liệu đầu vào tự động qua tag `validate:"required,min=2,gte=1,gt=0"`.
3. **Middleware**: `AuditLogMiddleware` đo thời gian thực thi của từng request và chặn `403 Forbidden` nếu thiếu quyền.
4. **Database (PostgreSQL)**: Tầng Model `ordermodel.go` trừu tượng hóa CSDL, kèm file `schema.sql` cho PostgreSQL.
5. **Log**: Ghi log cấu trúc JSON qua `logx.WithContext(ctx)` tự động mang theo `trace_id` và `span_id`.
6. **Observe**: Kích hoạt tự động Tracing, Metrics, Load Shedding và thống kê CPU/RAM định kỳ.
7. **Queue (Apache Kafka + kq)**: Tích hợp package `github.com/zeromicro/go-queue/kq` với Pusher bắn event và Worker Consumer độc lập xử lý đơn hàng.

---

## Cấu trúc thư mục

```text
order-hub/
├── cmd/
│   └── consumer/main.go          <-- Kafka Consumer Worker độc lập
├── etc/
│   └── orderhub-api.yaml         <-- File cấu hình (Postgres, Redis Cache, Kafka)
├── internal/
│   ├── config/config.go          <-- Struct cấu hình
│   ├── handler/                  <-- Tiếp nhận HTTP Request
│   ├── logic/                    <-- Xử lý nghiệp vụ chính & Kafka Pusher
│   ├── middleware/               <-- AuditLogMiddleware (timing, audit, RBAC)
│   ├── response/response.go      <-- Chuẩn hóa JSON {code, msg, data}
│   ├── svc/servicecontext.go     <-- Túi tài nguyên chung (Postgres, Redis, Kafka Pusher)
│   └── types/types.go            <-- Struct Request & Response
├── model/
│   └── ordermodel.go             <-- Tầng CSDL (sqlc: PostgreSQL + Redis Cache)
├── schema.sql                    <-- Script tạo bảng PostgreSQL
├── docker-compose.yml            <-- Cụm PostgreSQL + Redis + Apache Kafka KRaft + App
├── Dockerfile                    <-- Multi-stage build Go-Zero siêu nhẹ
├── order-hub.api                 <-- Bản thiết kế API DSL duy nhất
└── orderhub.go                   <-- Entrypoint khởi động API Server
```

---

## Hướng dẫn chạy thử nghiệm

### Cách 1: Chạy bằng Docker Compose (Khuyên dùng - 1 lệnh duy nhất)
Tự động bật trọn bộ PostgreSQL, Redis Cache, Apache Kafka và Go-Zero API:
```bash
# Khởi động toàn bộ cụm dịch vụ
docker compose up -d

# Xem log thời gian thực
docker compose logs -f

# Tắt hệ thống khi xong
docker compose down
```

### Cách 2: Chạy Worker Kafka Consumer độc lập
Mở một Terminal riêng để xem Worker nhận tin nhắn trực tiếp từ Kafka:
```powershell
go run cmd/consumer/main.go
```

### Cách 3: Tạo đơn hàng (POST)
Yêu cầu có Header `X-Role: admin` và trường `email` hợp lệ:
```powershell
$body = '{"customer_name": "Nguyen Van A", "email": "nguyenvana@gmail.com", "product_code": "MACBOOK-PRO", "quantity": 1, "amount": 2500.0}'
Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders" -Method Post -Headers @{"X-Role"="admin"} -ContentType "application/json" -Body $body
```

### Cách 4: Xem đơn hàng (GET)
Mở trình duyệt web hoặc chạy Terminal:
```powershell
Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders/<MÃ_ĐƠN_HÀNG>" -Method Get
```
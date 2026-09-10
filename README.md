# Order Hub - Go-Zero Production-Grade Demo

Dự án mẫu chuẩn chỉ kiến trúc sạch (Clean Architecture) sử dụng framework **go-zero**, tích hợp toàn diện 7 thành phần cốt lõi:

1. **Routes**: Thiết kế qua file `order-hub.api` với prefix `/api/v1` và grouping.
2. **Validate Request**: Kiểm tra dữ liệu đầu vào tự động qua tag `validate:"required,min=2,gte=1,gt=0"`.
3. **Middleware**: `AuditLogMiddleware` đo thời gian thực thi của từng request và ghi nhận nhật ký truy cập.
4. **Database (PostgreSQL)**: Tầng Model `ordermodel.go` trừu tượng hóa CSDL, kèm file `schema.sql` cho PostgreSQL.
5. **Log**: Ghi log cấu trúc JSON qua `logx.WithContext(ctx)` tự động mang theo `trace_id` và `span_id`.
6. **Observe**: Kích hoạt tự động Tracing, Metrics, Load Shedding và thống kê CPU/RAM định kỳ.
7. **Queue**: Xử lý tác vụ gửi email ngầm bất đồng bộ (Background worker / async event).

---

## Cấu trúc thư mục

```text
order-hub/
├── etc/
│   └── orderhub-api.yaml         <-- File cấu hình (Host, Port, Log)
├── internal/
│   ├── config/config.go          <-- Struct cấu hình
│   ├── handler/                  <-- Tiếp nhận HTTP Request
│   ├── logic/                    <-- Xử lý nghiệp vụ chính
│   ├── middleware/               <-- AuditLogMiddleware (timing, audit)
│   ├── response/response.go      <-- Chuẩn hóa JSON {code, msg, data}
│   ├── svc/servicecontext.go     <-- Túi tài nguyên chung (Model, Middleware)
│   └── types/types.go            <-- Struct Request & Response
├── model/
│   └── ordermodel.go             <-- Tầng CSDL (CRUD)
├── schema.sql                    <-- Script tạo bảng PostgreSQL
├── order-hub.api                 <-- Bản thiết kế API DSL duy nhất
└── orderhub.go                   <-- Entrypoint khởi động server
```

---

## Hướng dẫn chạy thử nghiệm

### Cách 1: Chạy bằng Docker Compose (Khuyên dùng - 1 lệnh duy nhất)
Không cần cài Go, không cần cài PostgreSQL. Docker sẽ tự động bật cả CSDL PostgreSQL (chạy sẵn `schema.sql`) và Go-Zero API:
```bash
# Khởi động toàn bộ hệ thống
docker compose up -d

# Xem log thời gian thực
docker compose logs -f

# Tắt hệ thống khi xong
docker compose down
```

### Cách 2: Chạy thủ công với Go
```powershell
go run orderhub.go -f etc/orderhub-api.yaml
```

### 2. Tạo đơn hàng (POST)
Yêu cầu có Header `X-Role: admin` (do Middleware kiểm soát) và trường `email` hợp lệ:
```powershell
$body = '{"customer_name": "Nguyen Van A", "email": "nguyenvana@gmail.com", "product_code": "MACBOOK-PRO", "quantity": 1, "amount": 2500.0}'
Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders" -Method Post -Headers @{"X-Role"="admin"} -ContentType "application/json" -Body $body
```

### 3. Kiểm tra tính năng Gửi Email (Queue Worker)
Sau khi gửi lệnh POST, hãy nhìn vào cửa sổ Terminal đang chạy server:
- **Ngay lập tức:** Server phản hồi kết quả cho client (không bắt client chờ).
- **Sau đúng 1 giây:** Background Queue Worker in ra dòng log gửi email thành công:
  ```text
  [QUEUE WORKER - ASYNC] Da gui email hoa don toi: nguyenvana@gmail.com (Don hang: ORD-xxx, Khach hang: Nguyen Van A)!
  ```

### 4. Xem đơn hàng (GET)
Mở trình duyệt web hoặc chạy Terminal:
```powershell
Invoke-RestMethod -Uri "http://localhost:8888/api/v1/orders/<MÃ_ĐƠN_HÀNG>" -Method Get
```
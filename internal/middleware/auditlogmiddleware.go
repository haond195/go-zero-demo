package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuditLogMiddleware struct{}

func NewAuditLogMiddleware() *AuditLogMiddleware {
	return &AuditLogMiddleware{}
}

func (m *AuditLogMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Ghi log luc bat dau request voi TraceID tu dong
		logx.WithContext(r.Context()).Infof("[MIDDLEWARE - START] %s %s tu Remote IP: %s", r.Method, r.URL.Path, r.RemoteAddr)

		// 2. Kiem tra quyen: Tu choi neu thieu Header X-Role: admin khi goi POST
		if r.Method == http.MethodPost && r.Header.Get("X-Role") != "admin" {
			logx.WithContext(r.Context()).Errorf("[MIDDLEWARE - BLOCKED] Tu choi request tu IP %s: Thieu quyen admin", r.RemoteAddr)
			http.Error(w, "Tu choi: Yeu cau Header X-Role: admin de tao don hang!", http.StatusForbidden)
			return // Dung tai day, khong cho vao Handler va CSDL
		}

		// 3. Chuyen tiep request vao Handler & Logic
		next(w, r)

		// 3. Do latency va ghi log hoan tat
		duration := time.Since(start)
		logx.WithContext(r.Context()).Infof("[MIDDLEWARE - FINISH] %s %s hoan tat trong: %v", r.Method, r.URL.Path, duration)
	}
}
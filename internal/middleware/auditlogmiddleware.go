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

		// 2. Chuyen tiep request vao Handler & Logic
		next(w, r)

		// 3. Do latency va ghi log hoan tat
		duration := time.Since(start)
		logx.WithContext(r.Context()).Infof("[MIDDLEWARE - FINISH] %s %s hoan tat trong: %v", r.Method, r.URL.Path, duration)
	}
}
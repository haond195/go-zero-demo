package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type RateLimitMiddleware struct {
	limiter *limit.PeriodLimit
}

func NewRateLimitMiddleware(limiter *limit.PeriodLimit) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: limiter,
	}
}

func (m *RateLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.limiter == nil {
			next(w, r)
			return
		}

		ip := httpx.GetRemoteAddr(r)
		code, err := m.limiter.Take(ip)
		if err != nil {
			logx.Errorf("[RATE LIMIT] Loi kiem tra quota cho IP %s: %v", ip, err)
			next(w, r)
			return
		}

		switch code {
		case limit.OverQuota:
			logx.Errorf("[RATE LIMIT] IP %s da vuot qua han muc (OverQuota)!", ip)
			httpx.WriteJsonCtx(r.Context(), w, http.StatusTooManyRequests, map[string]interface{}{
				"error":   "Too Many Requests",
				"message": "He thong gioi han toi da 3 request / 10 giay. Vui long thu lai sau!",
			})
			return
		case limit.HitQuota:
			logx.Infof("[RATE LIMIT] IP %s vua dat tran han muc (HitQuota)", ip)
		}

		next(w, r)
	}
}

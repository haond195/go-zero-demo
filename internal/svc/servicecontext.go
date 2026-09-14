package svc

import (
	"net/http"
	"os"
	"strings"

	"order-hub/internal/config"
	"order-hub/internal/middleware"
	"order-hub/model"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config      config.Config
	AuditLog    rest.Middleware
	RateLimit   rest.Middleware
	OrderModel  *model.OrderModel
	KafkaPusher *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	dataSource := c.Postgres.DataSource
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" && dataSource != "" {
		dataSource = strings.Replace(dataSource, "127.0.0.1", dbHost, 1)
		dataSource = strings.Replace(dataSource, "localhost", dbHost, 1)
	}

	cacheConf := c.Cache
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" && len(cacheConf) > 0 {
		cacheConf[0].Host = redisHost
	}

	var cachedConn sqlc.CachedConn
	hasCache := false
	if dataSource != "" && len(cacheConf) > 0 {
		sqlConn := sqlx.NewSqlConn("postgres", dataSource)
		cachedConn = sqlc.NewConn(sqlConn, cacheConf)
		hasCache = true
		logx.Infof("[SQLC] Khoi tao ket noi Postgres + Redis Cache: %s (Redis: %s)", dataSource, cacheConf[0].Host)
	}

	// Khoi tao Kafka Pusher
	kafkaBrokers := c.Kafka.Brokers
	if kfEnv := os.Getenv("KAFKA_BROKERS"); kfEnv != "" {
		kafkaBrokers = strings.Split(kfEnv, ",")
	}

	var kafkaPusher *kq.Pusher
	if len(kafkaBrokers) > 0 && c.Kafka.Topic != "" {
		kafkaPusher = kq.NewPusher(kafkaBrokers, c.Kafka.Topic)
		logx.Infof("[KAFKA PUSHER] Khoi tao ket noi Kafka Broker: %v (Topic: %s)", kafkaBrokers, c.Kafka.Topic)
	}

	var rateLimitMiddleware rest.Middleware = func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
	if len(cacheConf) > 0 {
		redisClient := redis.MustNewRedis(redis.RedisConf{
			Host: cacheConf[0].Host,
			Type: "node",
		})
		limiter := limit.NewPeriodLimit(10, 3, redisClient, "rate_limit:orders:")
		rateLimitMiddleware = middleware.NewRateLimitMiddleware(limiter).Handle
		logx.Infof("[RATE LIMIT] Khoi tao PeriodLimit: 3 requests / 10s (Redis: %s)", cacheConf[0].Host)
	}

	return &ServiceContext{
		Config:      c,
		AuditLog:    middleware.NewAuditLogMiddleware().Handle,
		RateLimit:   rateLimitMiddleware,
		OrderModel:  model.NewOrderModel(cachedConn, hasCache),
		KafkaPusher: kafkaPusher,
	}
}
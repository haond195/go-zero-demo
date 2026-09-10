package svc

import (
	"os"
	"strings"

	"order-hub/internal/config"
	"order-hub/internal/middleware"
	"order-hub/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config     config.Config
	AuditLog   rest.Middleware
	OrderModel *model.OrderModel
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

	return &ServiceContext{
		Config:     c,
		AuditLog:   middleware.NewAuditLogMiddleware().Handle,
		OrderModel: model.NewOrderModel(cachedConn, hasCache),
	}
}
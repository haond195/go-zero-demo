package svc

import (
	"os"
	"strings"

	"order-hub/internal/config"
	"order-hub/internal/middleware"
	"order-hub/model"

	"github.com/zeromicro/go-zero/core/logx"
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
	// Neu chay trong Docker thi doi host sang "postgres"
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" && dataSource != "" {
		dataSource = strings.Replace(dataSource, "127.0.0.1", dbHost, 1)
		dataSource = strings.Replace(dataSource, "localhost", dbHost, 1)
	}

	var conn sqlx.SqlConn
	if dataSource != "" {
		conn = sqlx.NewSqlConn("postgres", dataSource)
		logx.Infof("[POSTGRES] Khoi tao ket noi Database: %s", dataSource)
	}

	return &ServiceContext{
		Config:     c,
		AuditLog:   middleware.NewAuditLogMiddleware().Handle,
		OrderModel: model.NewOrderModel(conn),
	}
}
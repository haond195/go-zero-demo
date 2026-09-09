package svc

import (
	"order-hub/internal/config"
	"order-hub/internal/middleware"
	"order-hub/model"

	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config     config.Config
	AuditLog   rest.Middleware
	OrderModel *model.OrderModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		AuditLog:   middleware.NewAuditLogMiddleware().Handle,
		OrderModel: model.NewOrderModel(),
	}
}
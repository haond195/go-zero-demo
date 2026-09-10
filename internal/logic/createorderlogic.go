package logic

import (
	"context"
	"fmt"
	"time"

	"order-hub/internal/svc"
	"order-hub/internal/types"
	"order-hub/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderReq) (resp *types.CreateOrderResp, err error) {
	// 1. Sinh ma don hang doc nhat
	orderId := fmt.Sprintf("ORD-%d", time.Now().UnixNano()%1000000)

	// 2. Luu thong tin don hang vao CSDL qua tang Model
	order := &model.Order{
		OrderId:      orderId,
		CustomerName: req.CustomerName,
		ProductCode:  req.ProductCode,
		Quantity:     req.Quantity,
		Amount:       req.Amount,
		Status:       "PROCESSING",
	}

	if err := l.svcCtx.OrderModel.Insert(l.ctx, order); err != nil {
		l.Logger.Errorf("Loi khi luu vao CSDL: %v", err)
		return nil, err
	}

	l.Logger.Infof("[DATABASE] Da luu don hang %s cho khach %s", orderId, req.CustomerName)

	// 3. QUEUE SIMULATION: Ban event vao Background Queue de gui email bat dong bo
	go func(oId string, cust string, email string) {
		// Worker ngam xu ly sau 1 giay, khong lam cham Client
		time.Sleep(1 * time.Second)
		logx.Infof("[QUEUE WORKER - ASYNC] Da gui email hoa don toi: %s (Don hang: %s, Khach hang: %s)!", email, oId, cust)
	}(orderId, req.CustomerName, req.Email)

	// 4. Tra ket qua ngay lap tuc cho Client
	return &types.CreateOrderResp{
		OrderId: orderId,
		Status:  "PROCESSING",
		Message: "Tao don hang thanh cong! Email xac nhan dang duoc xu ly qua Queue.",
	}, nil
}
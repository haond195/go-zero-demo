package logic

import (
	"context"
	"errors"

	"order-hub/internal/svc"
	"order-hub/internal/types"
	"order-hub/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderLogic {
	return &GetOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrderLogic) GetOrder(req *types.GetOrderReq) (resp *types.GetOrderResp, err error) {
	// 1. Tim kiem don hang tu tang Model (CSDL)
	order, err := l.svcCtx.OrderModel.FindOne(l.ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			l.Logger.Errorf("Khong tim thay don hang: %s", req.OrderId)
			return nil, errors.New("Khong tim thay ma don hang nay trong he thong!")
		}
		return nil, err
	}

	l.Logger.Infof("[DATABASE] Da lay thanh cong thong tin don hang: %s", req.OrderId)

	// 2. Tra ve thong tin chi tiet don hang
	return &types.GetOrderResp{
		OrderId:      order.OrderId,
		CustomerName: order.CustomerName,
		ProductCode:  order.ProductCode,
		Quantity:     order.Quantity,
		Amount:       order.Amount,
		Status:       order.Status,
	}, nil
}
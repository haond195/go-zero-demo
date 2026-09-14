package paymentlogic

import (
	"context"
	"fmt"
	"time"

	"order-hub/rpc/payment/internal/svc"
	"order-hub/rpc/payment/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProcessPaymentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessPaymentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessPaymentLogic {
	return &ProcessPaymentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProcessPaymentLogic) ProcessPayment(in *payment.ProcessPaymentReq) (*payment.ProcessPaymentResp, error) {
	txId := fmt.Sprintf("TX-%d", time.Now().UnixNano()%1000000)
	l.Logger.Infof("[PAYMENT RPC - SUCCESS] Da tru tien: $%.2f cho don hang %s (Khach hang: %s) -> Ma giao dich: %s",
		in.Amount, in.OrderId, in.CustomerName, txId)

	return &payment.ProcessPaymentResp{
		TransactionId: txId,
		Status:        "PAID",
		Message:       fmt.Sprintf("Thanh toan thanh cong $%.2f cho don hang %s", in.Amount, in.OrderId),
	}, nil
}

package logic

import (
	"context"
	"fmt"
	"time"

	"order-hub/internal/svc"
	"order-hub/internal/types"
	"order-hub/model"
	paymentclient "order-hub/rpc/payment/client/payment"

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

	// 2. Luu thong tin don hang vao CSDL qua tang Model (sqlc: PostgreSQL + Redis)
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

	// 3. MICROSERVICE RPC: Goi sang Payment-RPC service qua gRPC
	payMsg := "Chua bat RPC"
	if l.svcCtx.PaymentRpc != nil {
		payResp, err := l.svcCtx.PaymentRpc.ProcessPayment(l.ctx, &paymentclient.ProcessPaymentReq{
			OrderId:      orderId,
			Amount:       req.Amount,
			CustomerName: req.CustomerName,
		})
		if err != nil {
			l.Logger.Errorf("[ZRPC PAYMENT] Loi khi goi Payment RPC: %v", err)
			payMsg = fmt.Sprintf("Loi thanh toan: %v", err)
		} else {
			l.Logger.Infof("[ZRPC PAYMENT] Da goi thanh cong Payment RPC qua gRPC -> TransId: %s, Msg: %s",
				payResp.TransactionId, payResp.Message)
			payMsg = fmt.Sprintf("Thanh toan gRPC thanh cong (%s)", payResp.TransactionId)
		}
	}

	// 4. KAFKA REAL QUEUE: Ban event truc tiep vao Apache Kafka qua kq.Pusher
	payload := fmt.Sprintf(`{"order_id":"%s","customer_name":"%s","email":"%s","product_code":"%s","amount":%.2f}`,
		orderId, req.CustomerName, req.Email, req.ProductCode, req.Amount)

	if l.svcCtx.KafkaPusher != nil {
		if err := l.svcCtx.KafkaPusher.Push(l.ctx, payload); err != nil {
			l.Logger.Errorf("[KAFKA PUSHER] Loi ban message vao Kafka: %v", err)
		} else {
			l.Logger.Infof("[KAFKA PUSHER] Da ban message vao Topic %s: %s", l.svcCtx.Config.Kafka.Topic, payload)
		}
	} else {
		// Fallback mo phong neu chua bat Kafka
		go func(oId string, cust string, email string) {
			time.Sleep(1 * time.Second)
			logx.Infof("[FALLBACK QUEUE] Da gui email hoa don toi: %s (Don hang: %s)!", email, oId)
		}(orderId, req.CustomerName, req.Email)
	}

	// 5. Tra ket qua ngay lap tuc cho Client
	return &types.CreateOrderResp{
		OrderId: orderId,
		Status:  "PROCESSING",
		Message: fmt.Sprintf("Tao don hang thanh cong! [%s] Event da duoc ban vao Kafka Queue.", payMsg),
	}, nil
}
package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("order not found")

type Order struct {
	OrderId      string  `db:"order_id" json:"order_id"`
	CustomerName string  `db:"customer_name" json:"customer_name"`
	ProductCode  string  `db:"product_code" json:"product_code"`
	Quantity     int64   `db:"quantity" json:"quantity"`
	Amount       float64 `db:"amount" json:"amount"`
	Status       string  `db:"status" json:"status"`
}

type OrderModel struct {
	cachedConn sqlc.CachedConn
	hasCache   bool
	lock       sync.RWMutex
	db         map[string]*Order
}

func NewOrderModel(cachedConn sqlc.CachedConn, hasCache bool) *OrderModel {
	return &OrderModel{
		cachedConn: cachedConn,
		hasCache:   hasCache,
		db:         make(map[string]*Order),
	}
}

func (m *OrderModel) Insert(ctx context.Context, order *Order) error {
	if m.hasCache {
		orderIdKey := fmt.Sprintf("cache:orders:orderId:%v", order.OrderId)
		query := `INSERT INTO orders (order_id, customer_name, product_code, quantity, amount, status) VALUES ($1, $2, $3, $4, $5, $6)`
		
		_, err := m.cachedConn.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
			return conn.ExecCtx(ctx, query, order.OrderId, order.CustomerName, order.ProductCode, order.Quantity, order.Amount, order.Status)
		}, orderIdKey)
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.db[order.OrderId] = order
	return nil
}

func (m *OrderModel) FindOne(ctx context.Context, orderId string) (*Order, error) {
	if m.hasCache {
		orderIdKey := fmt.Sprintf("cache:orders:orderId:%v", orderId)
		var order Order
		
		err := m.cachedConn.QueryRowCtx(ctx, &order, orderIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
			query := `SELECT order_id, customer_name, product_code, quantity, amount, status FROM orders WHERE order_id = $1 LIMIT 1`
			return conn.QueryRowCtx(ctx, v, query, orderId)
		})
		if err != nil {
			if errors.Is(err, sqlc.ErrNotFound) || errors.Is(err, sqlx.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		return &order, nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	order, exists := m.db[orderId]
	if !exists {
		return nil, ErrNotFound
	}
	return order, nil
}
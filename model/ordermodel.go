package model

import (
	"context"
	"errors"
	"sync"

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
	conn sqlx.SqlConn
	lock sync.RWMutex
	db   map[string]*Order // Fallback in-memory neu khong co Postgres
}

func NewOrderModel(conn sqlx.SqlConn) *OrderModel {
	return &OrderModel{
		conn: conn,
		db:   make(map[string]*Order),
	}
}

func (m *OrderModel) Insert(ctx context.Context, order *Order) error {
	// Neu co ket noi PostgreSQL -> Thuc thi SQL query that
	if m.conn != nil {
		query := `INSERT INTO orders (order_id, customer_name, product_code, quantity, amount, status) VALUES ($1, $2, $3, $4, $5, $6)`
		_, err := m.conn.ExecCtx(ctx, query, order.OrderId, order.CustomerName, order.ProductCode, order.Quantity, order.Amount, order.Status)
		return err
	}

	// Fallback in-memory
	m.lock.Lock()
	defer m.lock.Unlock()
	m.db[order.OrderId] = order
	return nil
}

func (m *OrderModel) FindOne(ctx context.Context, orderId string) (*Order, error) {
	// Neu co ket noi PostgreSQL -> Query tu CSDL
	if m.conn != nil {
		var order Order
		query := `SELECT order_id, customer_name, product_code, quantity, amount, status FROM orders WHERE order_id = $1 LIMIT 1`
		err := m.conn.QueryRowCtx(ctx, &order, query, orderId)
		if err != nil {
			if errors.Is(err, sqlx.ErrNotFound) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		return &order, nil
	}

	// Fallback in-memory
	m.lock.RLock()
	defer m.lock.RUnlock()
	order, exists := m.db[orderId]
	if !exists {
		return nil, ErrNotFound
	}
	return order, nil
}
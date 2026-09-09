package model

import (
	"context"
	"errors"
	"sync"
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
	lock sync.RWMutex
	// Kho luu tru gia lap san sang cho PostgreSQL
	db map[string]*Order
}

func NewOrderModel() *OrderModel {
	return &OrderModel{
		db: make(map[string]*Order),
	}
}

func (m *OrderModel) Insert(ctx context.Context, order *Order) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.db[order.OrderId] = order
	return nil
}

func (m *OrderModel) FindOne(ctx context.Context, orderId string) (*Order, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	order, exists := m.db[orderId]
	if !exists {
		return nil, ErrNotFound
	}
	return order, nil
}
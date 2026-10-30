package repository

import (
	"errors"

	"example.com/orders/internal/model"
)

var ErrNotFound = errors.New("заказ не найден")

type Memory struct {
	orders map[int]model.Order
}

func NewMemory() *Memory {
	return &Memory{orders: make(map[int]model.Order)}
}

func (m *Memory) Save(order model.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *Memory) ByID(id int) (model.Order, error) {
	order, ok := m.orders[id]
	if !ok {
		return model.Order{}, ErrNotFound
	}
	return order, nil
}

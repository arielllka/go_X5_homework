package memory

import (
	"context"
	"sync"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	mu     sync.Mutex
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value.ID = r.nextID
	r.nextID++
	r.items = append(r.items, cloneOrder(value))
	return cloneOrder(value), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, value := range r.items {
		if value.ID == id {
			return cloneOrder(value), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	values := make([]order.Order, len(r.items))
	for i, value := range r.items {
		values[i] = cloneOrder(value)
	}
	return values, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.items {
		if r.items[i].ID == value.ID {
			r.items[i] = cloneOrder(value)
			return cloneOrder(value), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func cloneOrder(value order.Order) order.Order {
	value.Items = append([]order.Item(nil), value.Items...)
	value.History = append([]order.StatusChange(nil), value.History...)
	return value
}

package service

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
	statusMu   sync.Mutex
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, order.ErrInvalidOrder
	}
	var total int64
	for _, item := range input.Items {
		if strings.TrimSpace(item.Name) == "" || item.Quantity <= 0 || item.Price <= 0 {
			return order.Order{}, order.ErrInvalidOrder
		}
		total += item.Price * int64(item.Quantity)
	}
	now := time.Now().UTC()
	value := order.Order{
		Items:      input.Items,
		TotalPrice: total,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History:    []order.StatusChange{{Status: order.StatusNew, ChangedAt: now}},
	}
	return s.repository.Create(ctx, value)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	if id <= 0 {
		return order.Order{}, order.ErrInvalidOrder
	}
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	if id <= 0 {
		return order.Order{}, order.ErrInvalidOrder
	}
	if !validStatus(status) {
		return order.Order{}, order.ErrInvalidStatus
	}

	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	value, err := s.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}
	if !allowedTransition(value.Status, status) {
		return order.Order{}, order.ErrInvalidStatusTransition
	}
	value.Status = status
	value.History = append(value.History, order.StatusChange{Status: status, ChangedAt: time.Now().UTC()})
	return s.repository.Update(ctx, value)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	value, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	history := make([]order.StatusChange, len(value.History))
	copy(history, value.History)
	sort.SliceStable(history, func(i, j int) bool {
		return history[i].ChangedAt.Before(history[j].ChangedAt)
	})
	return history, nil
}

func validStatus(status order.Status) bool {
	switch status {
	case order.StatusNew, order.StatusProcessing, order.StatusCompleted, order.StatusCancelled:
		return true
	default:
		return false
	}
}

func allowedTransition(from, to order.Status) bool {
	switch from {
	case order.StatusNew:
		return to == order.StatusProcessing || to == order.StatusCancelled
	case order.StatusProcessing:
		return to == order.StatusCompleted || to == order.StatusCancelled
	default:
		return false
	}
}

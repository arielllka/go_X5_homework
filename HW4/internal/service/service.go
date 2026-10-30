package service

import (
	"errors"

	"example.com/orders/internal/model"
)

var ErrEmptyCustomer = errors.New("имя клиента не задано")

type Repository interface {
	Save(order model.Order) error
	ByID(id int) (model.Order, error)
}

type Service struct {
	repo   Repository
	nextID int
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo, nextID: 1}
}

func (s *Service) Create(customer string) (model.Order, error) {
	if customer == "" {
		return model.Order{}, ErrEmptyCustomer
	}
	order := model.Order{
		ID:       s.nextID,
		Customer: customer,
		Status:   "new",
	}
	if err := s.repo.Save(order); err != nil {
		return model.Order{}, err
	}
	s.nextID++
	return order, nil
}

package handler

import "example.com/orders/internal/model"

type Creator interface {
	Create(customer string) (model.Order, error)
}

type Handler struct {
	orders Creator
}

func NewHandler(orders Creator) *Handler {
	return &Handler{orders: orders}
}

func (h *Handler) Create(customer string) (model.Order, error) {
	return h.orders.Create(customer)
}

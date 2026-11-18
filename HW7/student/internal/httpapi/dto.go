package httpapi

import "example.com/lesson9practice/internal/model"

type CreateOrderRequest struct {
	Customer string `json:"customer" example:"Алиса" minLength:"1" binding:"required"`
}

type OrderResponse struct {
	ID       int    `json:"id" example:"1" minimum:"1" binding:"required"`
	Customer string `json:"customer" example:"Алиса" binding:"required"`
	Status   string `json:"status" example:"new" enums:"new,cancelled" binding:"required"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"customer is required" binding:"required"`
	Code  string `json:"code" example:"VALIDATION_ERROR" binding:"required"`
}

type StatusResponse struct {
	Status string `json:"status" example:"ok" binding:"required"`
}

func orderResponse(order model.Order) OrderResponse {
	return OrderResponse{ID: order.ID, Customer: order.Customer, Status: order.Status}
}

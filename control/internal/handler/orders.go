package handler

import (
	"net/http"

	"example.com/order-service/internal/order"
)

// @Summary Создать заказ
// @Accept json
// @Produce json
// @Param order body order.CreateOrderInput true "Позиции заказа"
// @Success 201 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [post]
func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput
	if !decodeJSON(w, r, &input) {
		return
	}
	value, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, value)
}

// @Summary Получить список заказов
// @Produce json
// @Success 200 {array} order.Order
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [get]
func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.List(r.Context())
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, values)
}

// @Summary Получить заказ
// @Produce json
// @Param id path int64 true "ID заказа" minimum(1)
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id} [get]
func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := orderID(w, r)
	if !ok {
		return
	}
	value, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

// @Summary Изменить статус заказа
// @Description new можно изменить на processing или cancelled. processing можно изменить на completed или cancelled. Конечные статусы изменять нельзя.
// @Accept json
// @Produce json
// @Param id path int64 true "ID заказа" minimum(1)
// @Param status body order.UpdateStatusInput true "Новый статус"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := orderID(w, r)
	if !ok {
		return
	}
	var input order.UpdateStatusInput
	if !decodeJSON(w, r, &input) {
		return
	}
	value, err := h.service.UpdateStatus(r.Context(), id, input.Status)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, value)
}

// @Summary Получить историю статусов
// @Description Записи возвращаются в хронологическом порядке.
// @Produce json
// @Param id path int64 true "ID заказа" minimum(1)
// @Success 200 {array} order.StatusChange
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) orderHistory(w http.ResponseWriter, r *http.Request) {
	id, ok := orderID(w, r)
	if !ok {
		return
	}
	history, err := h.service.History(r.Context(), id)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, history)
}

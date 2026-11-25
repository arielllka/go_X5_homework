package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"example.com/lesson9practice/internal/model"
	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

type Orders interface {
	Create(string) (model.Order, error)
	ByID(int) (model.Order, error)
	List(int, int, string) ([]model.Order, error)
	Cancel(int) (model.Order, error)
}
type API struct {
	orders Orders
	logger *slog.Logger
}

func New(orders Orders, logger *slog.Logger) *API { return &API{orders: orders, logger: logger} }

// @Summary Статус сервиса
// @Tags status
// @Produce json
// @Success 200 {object} StatusResponse
// @Router /status [get]
func (a *API) status(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, StatusResponse{Status: "ok"})
}

// @Summary Создать заказ
// @Tags orders
// @Accept json
// @Produce json
// @Description Тело содержит один JSON-объект с непустым customer, максимум 1024 байта. Неизвестные поля запрещены.
// @Param order body CreateOrderRequest true "Данные заказа"
// @Success 201 {object} OrderResponse
// @Header 201 {string} Location "Путь к созданному заказу"
// @Failure 400 {object} ErrorResponse
// @Failure 413 {object} ErrorResponse
// @Failure 415 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders [post]
func (a *API) createOrder(w http.ResponseWriter, r *http.Request) {
	var in *CreateOrderRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, err.status, err.message, err.code)
		return
	}
	if in == nil {
		writeError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	order, err := a.orders.Create(in.Customer)
	if errors.Is(err, service.ErrEmptyCustomer) {
		writeError(w, 400, "customer is required", "VALIDATION_ERROR")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/orders/"+strconv.Itoa(order.ID))
	writeJSON(w, 201, orderResponse(order))
}

// @Summary Получить заказ
// @Tags orders
// @Produce json
// @Param id path int true "Положительный идентификатор заказа" minimum(1)
// @Success 200 {object} OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders/{id} [get]
func (a *API) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, 400, "id must be a positive integer", "INVALID_ID")
		return
	}
	order, err := a.orders.ByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, 404, "order not found", "ORDER_NOT_FOUND")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	writeJSON(w, 200, orderResponse(order))
}

// @Summary Список заказов
// @Tags orders
// @Produce json
// @Description Сначала применяется фильтр status, затем сортировка по id и пагинация. Без status возвращаются все заказы. Пустая страница возвращает []. Числовые параметры допускают только десятичные цифры. Пустые, повторные, неизвестные параметры и неверные значения возвращают 400 INVALID_QUERY.
// @Param limit query int false "Размер страницы" minimum(1) maximum(100) default(20)
// @Param offset query int false "Количество пропускаемых заказов" minimum(0) default(0)
// @Param status query string false "Статус заказа" enums(new,cancelled)
// @Success 200 {array} OrderResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /orders [get]
func (a *API) listOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset, status, ok := parseListQuery(r.URL.RawQuery)
	if !ok {
		writeError(w, 400, "invalid query parameters", "INVALID_QUERY")
		return
	}
	orders, err := a.orders.List(limit, offset, status)
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	out := make([]OrderResponse, len(orders))
	for i, order := range orders {
		out[i] = orderResponse(order)
	}
	writeJSON(w, 200, out)
}

func parseListQuery(rawQuery string) (int, int, string, bool) {
	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return 0, 0, "", false
	}
	limit, offset := 20, 0
	status := ""
	for name, values := range query {
		if (name != "limit" && name != "offset" && name != "status") || len(values) != 1 || values[0] == "" {
			return 0, 0, "", false
		}
		if name == "status" {
			if values[0] != "new" && values[0] != "cancelled" {
				return 0, 0, "", false
			}
			status = values[0]
			continue
		}
		for _, digit := range values[0] {
			if digit < '0' || digit > '9' {
				return 0, 0, "", false
			}
		}
		value, err := strconv.Atoi(values[0])
		if err != nil {
			return 0, 0, "", false
		}
		if name == "limit" {
			if value < 1 || value > 100 {
				return 0, 0, "", false
			}
			limit = value
		} else {
			offset = value
		}
	}
	return limit, offset, status, true
}

// @Summary Отменить заказ
// @Tags orders
// @Accept json
// @Produce json
// @Description Разрешён только переход new в cancelled. Проверка и изменение статуса выполняются атомарно. Повторная отмена возвращает 409 ORDER_ALREADY_CANCELLED.
// @Param id path int true "Положительный идентификатор заказа" minimum(1)
// @Param status body UpdateOrderStatusRequest true "Новый статус: cancelled"
// @Success 200 {object} OrderResponse
// @Failure 400 {object} ErrorResponse "INVALID_ID, INVALID_JSON или VALIDATION_ERROR"
// @Failure 404 {object} ErrorResponse "ORDER_NOT_FOUND"
// @Failure 409 {object} ErrorResponse "ORDER_ALREADY_CANCELLED"
// @Failure 413 {object} ErrorResponse "BODY_TOO_LARGE"
// @Failure 415 {object} ErrorResponse "UNSUPPORTED_MEDIA_TYPE"
// @Failure 500 {object} ErrorResponse
// @Router /orders/{id}/status [patch]
func (a *API) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, 400, "id must be a positive integer", "INVALID_ID")
		return
	}
	var in *UpdateOrderStatusRequest
	if err := decodeJSON(w, r, &in); err != nil {
		writeError(w, err.status, err.message, err.code)
		return
	}
	if in == nil {
		writeError(w, 400, "body must be a JSON object", "INVALID_JSON")
		return
	}
	if in.Status != "cancelled" {
		writeError(w, 400, "status must be cancelled", "VALIDATION_ERROR")
		return
	}
	order, err := a.orders.Cancel(id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, 404, "order not found", "ORDER_NOT_FOUND")
		return
	}
	if errors.Is(err, repository.ErrAlreadyCancelled) {
		writeError(w, 409, "order already cancelled", "ORDER_ALREADY_CANCELLED")
		return
	}
	if err != nil {
		a.internalError(w, r, err)
		return
	}
	writeJSON(w, 200, orderResponse(order))
}

func (a *API) internalError(w http.ResponseWriter, r *http.Request, err error) {
	a.logger.ErrorContext(r.Context(), "order operation failed", "error", err)
	writeError(w, 500, "internal server error", "INTERNAL_ERROR")
}

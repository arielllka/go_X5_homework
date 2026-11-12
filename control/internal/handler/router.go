package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	_ "example.com/order-service/docs"
	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", h.createOrder)
		r.Get("/", h.listOrders)
		r.Get("/{id}", h.getOrder)
		r.Patch("/{id}/status", h.updateStatus)
		r.Get("/{id}/history", h.orderHistory)
	})
	router.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "route not found", Code: "NOT_FOUND"})
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed", Code: "METHOD_NOT_ALLOWED"})
	})

	return router
}

// @Summary Проверка доступности
// @Description Возвращает status со значением ok.
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(target)
	var extra any
	if err != nil || decoder.Decode(&extra) != io.EOF {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body", Code: "INVALID_JSON"})
		return false
	}
	return true
}

func orderID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "order id must be an integer", Code: "INVALID_ID"})
		return 0, false
	}
	return id, true
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusInternalServerError
	body := errorResponse{Error: "internal server error", Code: "INTERNAL_ERROR"}
	switch {
	case errors.Is(err, order.ErrInvalidOrder):
		status, body = http.StatusBadRequest, errorResponse{Error: "invalid order", Code: "INVALID_ORDER"}
	case errors.Is(err, order.ErrInvalidStatus):
		status, body = http.StatusBadRequest, errorResponse{Error: "invalid status", Code: "INVALID_STATUS"}
	case errors.Is(err, order.ErrOrderNotFound):
		status, body = http.StatusNotFound, errorResponse{Error: "order not found", Code: "ORDER_NOT_FOUND"}
	case errors.Is(err, order.ErrInvalidStatusTransition):
		status, body = http.StatusConflict, errorResponse{Error: "status transition is not allowed", Code: "INVALID_STATUS_TRANSITION"}
	default:
		h.logger.ErrorContext(r.Context(), "request failed", "error", err, "request_id", middleware.GetReqID(r.Context()))
	}
	writeJSON(w, status, body)
}

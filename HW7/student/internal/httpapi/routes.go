package httpapi

import (
	"net/http"

	_ "example.com/lesson9practice/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/status", a.status)
	mux.HandleFunc("POST /api/v1/orders", a.createOrder)
	mux.HandleFunc("GET /api/v1/orders", a.listOrders)
	mux.HandleFunc("GET /api/v1/orders/{id}", a.getOrder)
	mux.HandleFunc("PATCH /api/v1/orders/{id}/status", a.updateOrderStatus)
	mux.HandleFunc("/api/v1/status", methodNotAllowed("GET, HEAD"))
	mux.HandleFunc("/api/v1/orders", methodNotAllowed("GET, HEAD, POST"))
	mux.HandleFunc("/api/v1/orders/{id}", methodNotAllowed("GET, HEAD"))
	mux.HandleFunc("/api/v1/orders/{id}/status", methodNotAllowed("PATCH"))
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) { writeError(w, 404, "route not found", "ROUTE_NOT_FOUND") })
	mux.Handle("GET /swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
	return mux
}
func methodNotAllowed(allow string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", allow)
		writeError(w, 405, "method not allowed", "METHOD_NOT_ALLOWED")
	}
}

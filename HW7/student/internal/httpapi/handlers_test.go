package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

func newHandler() http.Handler {
	return New(service.New(repository.New()), slog.New(slog.NewTextHandler(io.Discard, nil))).Handler()
}

func request(handler http.Handler, method, path, body, contentType string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestCreateAndReadOrder(t *testing.T) {
	handler := newHandler()
	created := request(handler, "POST", "/api/v1/orders", `{"customer":"  Алиса  "}`, "application/json")
	if created.Code != 201 {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body)
	}
	if got := created.Header().Get("Location"); got != "/api/v1/orders/1" {
		t.Fatalf("Location=%q", got)
	}
	var order OrderResponse
	if err := json.Unmarshal(created.Body.Bytes(), &order); err != nil {
		t.Fatal(err)
	}
	if order.ID != 1 || order.Customer != "Алиса" || order.Status != "new" {
		t.Fatalf("unexpected order: %+v", order)
	}
	read := request(handler, "GET", created.Header().Get("Location"), "", "")
	if read.Code != 200 || read.Body.String() != created.Body.String() {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body)
	}
	again := request(handler, "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
	if again.Code != 201 || again.Header().Get("Location") != "/api/v1/orders/2" {
		t.Fatalf("second create status=%d Location=%q", again.Code, again.Header().Get("Location"))
	}
}

func TestErrorsKeepJSONContract(t *testing.T) {
	tests := []struct {
		name, method, path, body, contentType string
		status                                int
		code                                  string
	}{
		{"invalid id", "GET", "/api/v1/orders/abc", "", "", 400, "INVALID_ID"},
		{"zero id", "GET", "/api/v1/orders/0", "", "", 400, "INVALID_ID"},
		{"missing order", "GET", "/api/v1/orders/123", "", "", 404, "ORDER_NOT_FOUND"},
		{"missing route", "GET", "/api/v1/missing", "", "", 404, "ROUTE_NOT_FOUND"},
		{"wrong method", "PUT", "/api/v1/orders/1", "", "", 405, "METHOD_NOT_ALLOWED"},
		{"empty customer", "POST", "/api/v1/orders", `{"customer":" "}`, "application/json", 400, "VALIDATION_ERROR"},
		{"null body", "POST", "/api/v1/orders", `null`, "application/json", 400, "INVALID_JSON"},
		{"wrong type", "POST", "/api/v1/orders", `{"customer":42}`, "application/json", 400, "INVALID_JSON"},
		{"unknown field", "POST", "/api/v1/orders", `{"customer":"Алиса","status":"paid"}`, "application/json", 400, "INVALID_JSON"},
		{"second JSON", "POST", "/api/v1/orders", `{"customer":"Алиса"} {}`, "application/json", 400, "INVALID_JSON"},
		{"wrong media type", "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "text/plain", 415, "UNSUPPORTED_MEDIA_TYPE"},
		{"large body", "POST", "/api/v1/orders", `{"customer":"` + strings.Repeat("a", 1100) + `"}`, "application/json", 413, "BODY_TOO_LARGE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := request(newHandler(), tt.method, tt.path, tt.body, tt.contentType)
			if w.Code != tt.status {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tt.status, w.Body)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type=%q", got)
			}
			var out ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Code != tt.code || out.Error == "" {
				t.Fatalf("unexpected error: %+v", out)
			}
		})
	}
}

func TestUnsupportedMethodDescribesAllowedMethods(t *testing.T) {
	handler := newHandler()
	for _, tt := range []struct{ method, path, allow string }{
		{"DELETE", "/api/v1/orders/1", "GET, HEAD"},
		{"POST", "/api/v1/status", "GET, HEAD"},
		{"DELETE", "/api/v1/orders", "GET, HEAD, POST"},
		{"GET", "/api/v1/orders/1/status", "PATCH"},
	} {
		w := request(handler, tt.method, tt.path, "", "")
		if w.Code != 405 || w.Header().Get("Allow") != tt.allow {
			t.Errorf("%s %s: status=%d Allow=%q", tt.method, tt.path, w.Code, w.Header().Get("Allow"))
		}
	}
}

func TestStatus(t *testing.T) {
	w := request(newHandler(), "GET", "/api/v1/status", "", "")
	if w.Code != 200 || strings.TrimSpace(w.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("status=%d body=%s", w.Code, w.Body)
	}
}

func TestSwaggerRoutes(t *testing.T) {
	handler := newHandler()
	ui := request(handler, "GET", "/swagger/index.html", "", "")
	if ui.Code != 200 || !strings.Contains(ui.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("Swagger UI status=%d body=%s", ui.Code, ui.Body)
	}
	doc := request(handler, "GET", "/swagger/doc.json", "", "")
	if doc.Code != 200 {
		t.Fatalf("Swagger document status=%d body=%s", doc.Code, doc.Body)
	}
	var spec struct {
		BasePath string                                `json:"basePath"`
		Paths    map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(doc.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	if spec.BasePath != "/api/v1" || spec.Paths["/status"]["get"] == nil ||
		spec.Paths["/orders"]["get"] == nil || spec.Paths["/orders"]["post"] == nil ||
		spec.Paths["/orders/{id}"]["get"] == nil || spec.Paths["/orders/{id}/status"]["patch"] == nil {
		t.Fatalf("incomplete Swagger specification: %+v", spec)
	}
}

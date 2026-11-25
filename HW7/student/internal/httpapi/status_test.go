package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCancelOrder(t *testing.T) {
	handler := newHandler()
	created := request(handler, "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
	if created.Code != 201 {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body)
	}
	cancelled := request(handler, "PATCH", "/api/v1/orders/1/status", `{"status":"cancelled"}`, "application/json")
	if cancelled.Code != 200 || cancelled.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("cancel status=%d body=%s", cancelled.Code, cancelled.Body)
	}
	var order OrderResponse
	if err := json.Unmarshal(cancelled.Body.Bytes(), &order); err != nil {
		t.Fatal(err)
	}
	if order.ID != 1 || order.Customer != "Алиса" || order.Status != "cancelled" {
		t.Fatalf("unexpected order: %+v", order)
	}
	read := request(handler, "GET", "/api/v1/orders/1", "", "")
	if read.Code != 200 || read.Body.String() != cancelled.Body.String() {
		t.Fatalf("read status=%d body=%s", read.Code, read.Body)
	}
	again := request(handler, "PATCH", "/api/v1/orders/1/status", `{"status":"cancelled"}`, "application/json")
	var out ErrorResponse
	if err := json.Unmarshal(again.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if again.Code != 409 || out.Code != "ORDER_ALREADY_CANCELLED" {
		t.Fatalf("repeated cancel status=%d body=%s", again.Code, again.Body)
	}
}

func TestCancelOrderErrors(t *testing.T) {
	for _, tt := range []struct {
		name, id, body, contentType, code string
		status                            int
	}{
		{"missing", "999", `{"status":"cancelled"}`, "application/json", "ORDER_NOT_FOUND", 404},
		{"invalid id", "abc", `{"status":"cancelled"}`, "application/json", "INVALID_ID", 400},
		{"zero id", "0", `{"status":"cancelled"}`, "application/json", "INVALID_ID", 400},
		{"negative id", "-1", `{"status":"cancelled"}`, "application/json", "INVALID_ID", 400},
		{"overflow id", "18446744073709551616", `{"status":"cancelled"}`, "application/json", "INVALID_ID", 400},
		{"empty body", "1", ``, "application/json", "INVALID_JSON", 400},
		{"null body", "1", `null`, "application/json", "INVALID_JSON", 400},
		{"malformed JSON", "1", `{`, "application/json", "INVALID_JSON", 400},
		{"missing status", "1", `{}`, "application/json", "VALIDATION_ERROR", 400},
		{"empty status", "1", `{"status":""}`, "application/json", "VALIDATION_ERROR", 400},
		{"null status", "1", `{"status":null}`, "application/json", "VALIDATION_ERROR", 400},
		{"new status", "1", `{"status":"new"}`, "application/json", "VALIDATION_ERROR", 400},
		{"unknown status", "1", `{"status":"paid"}`, "application/json", "VALIDATION_ERROR", 400},
		{"wrong type", "1", `{"status":42}`, "application/json", "INVALID_JSON", 400},
		{"unknown field", "1", `{"status":"cancelled","other":1}`, "application/json", "INVALID_JSON", 400},
		{"second JSON", "1", `{"status":"cancelled"} {}`, "application/json", "INVALID_JSON", 400},
		{"wrong media type", "1", `{"status":"cancelled"}`, "text/plain", "UNSUPPORTED_MEDIA_TYPE", 415},
		{"large body", "1", `{"status":"` + strings.Repeat("a", 1100) + `"}`, "application/json", "BODY_TOO_LARGE", 413},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := newHandler()
			created := request(handler, "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
			if created.Code != 201 {
				t.Fatalf("create status=%d body=%s", created.Code, created.Body)
			}
			w := request(handler, "PATCH", "/api/v1/orders/"+tt.id+"/status", tt.body, tt.contentType)
			if w.Code != tt.status || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tt.status, w.Body)
			}
			var out ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Code != tt.code || out.Error == "" {
				t.Fatalf("unexpected error: %+v", out)
			}
			read := request(handler, "GET", "/api/v1/orders/1", "", "")
			if read.Code != 200 || read.Body.String() != created.Body.String() {
				t.Fatalf("failed request changed the order: %s", read.Body)
			}
		})
	}
}

func TestConcurrentCancelOrder(t *testing.T) {
	handler := newHandler()
	created := request(handler, "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
	if created.Code != 201 {
		t.Fatalf("create status=%d body=%s", created.Code, created.Body)
	}
	start := make(chan struct{})
	results := make(chan *httptest.ResponseRecorder, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			results <- request(handler, "PATCH", "/api/v1/orders/1/status", `{"status":"cancelled"}`, "application/json")
		}()
	}
	close(start)
	statuses := make(map[int]int)
	for i := 0; i < 2; i++ {
		w := <-results
		statuses[w.Code]++
		if w.Code == 409 {
			var out ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Code != "ORDER_ALREADY_CANCELLED" {
				t.Fatalf("unexpected conflict: %+v", out)
			}
		}
	}
	if statuses[200] != 1 || statuses[409] != 1 {
		t.Fatalf("parallel cancellation results: %v", statuses)
	}
}

package httpapi

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestListEmptyOrders(t *testing.T) {
	w := request(newHandler(), "GET", "/api/v1/orders", "", "")
	if w.Code != 200 || w.Body.String() != "[]\n" {
		t.Fatalf("status=%d body=%s", w.Code, w.Body)
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type=%q", got)
	}
}

func TestListOrderPages(t *testing.T) {
	handler := newHandler()
	for i := 0; i < 25; i++ {
		w := request(handler, "POST", "/api/v1/orders", `{"customer":"Алиса"}`, "application/json")
		if w.Code != 201 {
			t.Fatalf("create status=%d body=%s", w.Code, w.Body)
		}
	}
	maxInt := strconv.Itoa(int(^uint(0) >> 1))
	tests := []struct {
		name, query  string
		first, count int
	}{
		{"defaults", "", 1, 20},
		{"limit only", "?limit=2", 1, 2},
		{"offset only", "?offset=20", 21, 5},
		{"first page", "?limit=2&offset=0", 1, 2},
		{"second page", "?limit=2&offset=2", 3, 2},
		{"single order", "?limit=1&offset=1", 2, 1},
		{"last page", "?limit=2&offset=24", 25, 1},
		{"maximum limit", "?limit=100", 1, 25},
		{"decimal leading zeros", "?limit=02&offset=01", 2, 2},
		{"end", "?limit=2&offset=25", 0, 0},
		{"beyond end", "?offset=1000", 0, 0},
		{"maximum offset", "?limit=100&offset=" + maxInt, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := request(handler, "GET", "/api/v1/orders"+tt.query, "", "")
			if w.Code != 200 {
				t.Fatalf("status=%d body=%s", w.Code, w.Body)
			}
			var orders []OrderResponse
			if err := json.Unmarshal(w.Body.Bytes(), &orders); err != nil {
				t.Fatal(err)
			}
			if orders == nil || len(orders) != tt.count {
				t.Fatalf("orders=%v want count=%d body=%s", orders, tt.count, w.Body)
			}
			for i, order := range orders {
				if order.ID != tt.first+i || order.Customer != "Алиса" || order.Status != "new" {
					t.Fatalf("order[%d]=%+v", i, order)
				}
			}
		})
	}
}

func TestListRejectsInvalidQuery(t *testing.T) {
	for _, query := range []string{
		"limit=", "offset=", "limit", "offset",
		"limit=abc", "offset=abc", "limit=-1", "offset=-1",
		"limit=0", "limit=101", "limit=1.5", "offset=1.5",
		"limit=+1", "offset=%2B1", "limit=%201", "offset=1%20",
		"limit=1e1", "offset=0x10", "limit=%D9%A1", "offset=%EF%BC%91",
		"limit=18446744073709551616", "offset=18446744073709551616",
		"limit=1&limit=2", "limit=1&limit=1", "offset=0&offset=1", "limit=1&%6cimit=1",
		"unknown=1", "status=new", "=1", "limit=1;offset=0", "limit=%zz",
	} {
		t.Run(query, func(t *testing.T) {
			w := request(newHandler(), "GET", "/api/v1/orders?"+query, "", "")
			if w.Code != 400 {
				t.Fatalf("status=%d body=%s", w.Code, w.Body)
			}
			if got := w.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type=%q", got)
			}
			var out ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Code != "INVALID_QUERY" || out.Error == "" {
				t.Fatalf("unexpected error: %+v", out)
			}
		})
	}
}

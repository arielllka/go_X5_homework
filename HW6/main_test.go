package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// До реализации health этот тест должен быть красным.
func TestHealth(t *testing.T) {
	h := newRouter(NewStore())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d, want 200", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type=%q", rec.Header().Get("Content-Type"))
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ok" {
		t.Fatalf("status in body=%q", body.Status)
	}
}

func serveRequest(h http.Handler, method, path, body, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func decodeResponse(t *testing.T, rec *httptest.ResponseRecorder, target any) {
	t.Helper()
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type=%q, want application/json", got)
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing X-Request-ID")
	}
	decoder := json.NewDecoder(rec.Body)
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("response must contain exactly one JSON value: %v", err)
	}
}

func checkErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, status int) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status=%d, want %d; body=%s", rec.Code, status, rec.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	decodeResponse(t, rec, &body)
	if body.Error == "" {
		t.Fatal("missing error message")
	}
}

func TestCreateAndReadOrders(t *testing.T) {
	h := newRouter(NewStore())
	limitBody := `{"customer":"Лимит"}`
	tests := []struct {
		name     string
		body     string
		customer string
	}{
		{"trim_name", `{"customer":"  Анна \t"}`, "Анна"},
		{"next_order", `{"customer":"Борис"}`, "Борис"},
		{"1024_bytes_with_trailing_spaces", limitBody + strings.Repeat(" ", 1024-len(limitBody)), "Лимит"},
	}
	seenRequestIDs := make(map[string]bool)
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveRequest(h, http.MethodPost, "/orders", tt.body, "application/json; charset=utf-8")
			if rec.Code != http.StatusCreated {
				t.Fatalf("status=%d, want 201; body=%s", rec.Code, rec.Body.String())
			}
			var created Order
			decodeResponse(t, rec, &created)
			want := Order{ID: i + 1, Customer: tt.customer, Status: "new"}
			if created != want {
				t.Fatalf("created=%+v, want %+v", created, want)
			}
			location := fmt.Sprintf("/orders/%d", created.ID)
			if got := rec.Header().Get("Location"); got != location {
				t.Fatalf("Location=%q, want %q", got, location)
			}
			read := serveRequest(h, http.MethodGet, location, "", "")
			if read.Code != http.StatusOK {
				t.Fatalf("read status=%d, want 200", read.Code)
			}
			var fetched Order
			decodeResponse(t, read, &fetched)
			if fetched != created {
				t.Fatalf("fetched=%+v, want %+v", fetched, created)
			}
			for _, response := range []*httptest.ResponseRecorder{rec, read} {
				id := response.Header().Get("X-Request-ID")
				if seenRequestIDs[id] {
					t.Fatalf("request id reused: %q", id)
				}
				seenRequestIDs[id] = true
			}
		})
	}
}

func TestCreateOrderInvalidBody(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{"empty", "", 400},
		{"whitespace", " \t\n", 400},
		{"broken_json", `{"customer":`, 400},
		{"number_customer", `{"customer":123}`, 400},
		{"boolean_customer", `{"customer":true}`, 400},
		{"missing_customer", `{}`, 400},
		{"empty_name", `{"customer":""}`, 400},
		{"whitespace_name", `{"customer":" \t\n "}`, 400},
		{"null_customer", `{"customer":null}`, 400},
		{"null_object", `null`, 400},
		{"array", `[]`, 400},
		{"unknown_field", `{"customer":"Анна","extra":1}`, 400},
		{"client_id", `{"customer":"Анна","id":99}`, 400},
		{"client_status", `{"customer":"Анна","status":"done"}`, 400},
		{"second_object", `{"customer":"Анна"} {}`, 400},
		{"second_null", `{"customer":"Анна"} null`, 400},
		{"trailing_garbage", `{"customer":"Анна"} garbage`, 400},
		{"oversized_value", `{"customer":"` + strings.Repeat("a", 1024) + `"}`, 413},
		{"oversized_trailing_spaces", `{"customer":"Анна"}` + strings.Repeat(" ", 1024), 413},
		{"oversized_invalid_json", strings.Repeat("x", 1025), 413},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			existing := store.Create("Первый клиент")
			rec := serveRequest(newRouter(store), http.MethodPost, "/orders", tt.body, "application/json")
			checkErrorResponse(t, rec, tt.status)
			if len(store.orders) != 1 || store.nextID != 2 || store.orders[existing.ID] != existing {
				t.Fatal("invalid request changed the store")
			}
		})
	}
}

func TestCreateOrderContentType(t *testing.T) {
	for _, contentType := range []string{"", "text/plain", "application/json; broken"} {
		t.Run(contentType, func(t *testing.T) {
			store := NewStore()
			rec := serveRequest(newRouter(store), http.MethodPost, "/orders", `{"customer":"Анна"}`, contentType)
			checkErrorResponse(t, rec, http.StatusUnsupportedMediaType)
			if len(store.orders) != 0 || store.nextID != 1 {
				t.Fatal("unsupported Content-Type created an order")
			}
		})
	}
}

func TestGetOrderErrors(t *testing.T) {
	tests := []struct {
		id     string
		status int
	}{
		{"abc", 400},
		{"0", 400},
		{"-1", 400},
		{"1.5", 400},
		{strings.Repeat("9", 40), 400},
		{"42", 404},
	}
	store := NewStore()
	store.Create("Анна")
	h := newRouter(store)
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			rec := serveRequest(h, http.MethodGet, "/orders/"+tt.id, "", "")
			checkErrorResponse(t, rec, tt.status)
		})
	}
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	previous := slog.Default()
	buffer := new(bytes.Buffer)
	slog.SetDefault(slog.New(slog.NewJSONHandler(buffer, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return buffer
}

func readLogs(t *testing.T, buffer *bytes.Buffer) []map[string]any {
	t.Helper()
	decoder := json.NewDecoder(buffer)
	var entries []map[string]any
	for {
		var entry map[string]any
		if err := decoder.Decode(&entry); err == io.EOF {
			return entries
		} else if err != nil {
			t.Fatalf("invalid JSON log: %v", err)
		}
		entries = append(entries, entry)
	}
}

func TestAccessLog(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
	}{
		{"health", http.MethodGet, "/health", "", 200},
		{"created", http.MethodPost, "/orders", `{"customer":"private-customer-marker"}`, 201},
		{"bad_request", http.MethodPost, "/orders", `{"customer":""}`, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := captureLogs(t)
			req := httptest.NewRequest(tt.method, tt.path+"?token=private-token-marker", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer private-token-marker")
			rec := httptest.NewRecorder()
			newRouter(NewStore()).ServeHTTP(rec, req)
			if rec.Code != tt.status {
				t.Fatalf("status=%d, want %d", rec.Code, tt.status)
			}
			if text := buffer.String(); strings.Contains(text, "private-customer-marker") || strings.Contains(text, "private-token-marker") {
				t.Fatal("private data leaked into logs")
			}
			entries := readLogs(t, buffer)
			if len(entries) != 1 {
				t.Fatalf("log entries=%d, want 1", len(entries))
			}
			entry := entries[0]
			id := rec.Header().Get("X-Request-ID")
			if entry["method"] != tt.method || entry["path"] != tt.path || entry["status"] != float64(tt.status) || id == "" || entry["request_id"] != id {
				t.Fatalf("unexpected access log: %v", entry)
			}
			if duration, ok := entry["duration_ms"].(float64); !ok || duration < 0 {
				t.Fatalf("invalid duration_ms: %v", entry["duration_ms"])
			}
		})
	}
}

func TestRequestIDContext(t *testing.T) {
	var contextID string
	h := requestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = requestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	previousID := ""
	for i := 0; i < 2; i++ {
		rec := serveRequest(h, http.MethodGet, "/context", "", "")
		id := rec.Header().Get("X-Request-ID")
		if id == "" || id == previousID || contextID != id {
			t.Fatalf("header=%q, context=%q, previous=%q", id, contextID, previousID)
		}
		previousID = id
	}
}

func TestRecovery(t *testing.T) {
	buffer := captureLogs(t)
	var contextID string
	broken := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = requestIDFromContext(r.Context())
		panic("private-panic-marker")
	})
	h := requestID(accessLog(recovery(broken)))
	rec := serveRequest(h, http.MethodGet, "/test-only-panic", "", "")
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d, want 500", rec.Code)
	}
	var body struct {
		Error string `json:"error"`
	}
	decodeResponse(t, rec, &body)
	if body.Error != "internal server error" || strings.Contains(buffer.String(), "private-panic-marker") {
		t.Fatal("panic must produce a generic error without leaking its contents")
	}
	id := rec.Header().Get("X-Request-ID")
	if contextID != id {
		t.Fatalf("context request_id=%q, header=%q", contextID, id)
	}
	entries := readLogs(t, buffer)
	if len(entries) != 2 {
		t.Fatalf("log entries=%d, want panic and access records", len(entries))
	}
	foundAccessLog := false
	for _, entry := range entries {
		if entry["request_id"] != id {
			t.Fatalf("request id changed between records: %v", entry)
		}
		if entry["msg"] == "http request" {
			foundAccessLog = true
			if entry["status"] != float64(http.StatusInternalServerError) {
				t.Fatalf("access log must contain status 500: %v", entry)
			}
		}
	}
	if !foundAccessLog {
		t.Fatal("missing access log after panic")
	}
}

func TestRouting(t *testing.T) {
	h := newRouter(NewStore())
	for _, tt := range []struct {
		method string
		path   string
		status int
	}{
		{http.MethodGet, "/panic", http.StatusNotFound},
		{http.MethodPost, "/health", http.StatusMethodNotAllowed},
	} {
		t.Run(tt.method+tt.path, func(t *testing.T) {
			rec := serveRequest(h, tt.method, tt.path, "", "")
			if rec.Code != tt.status || rec.Header().Get("X-Request-ID") == "" {
				t.Fatalf("status=%d, want %d; request_id=%q", rec.Code, tt.status, rec.Header().Get("X-Request-ID"))
			}
		})
	}
}

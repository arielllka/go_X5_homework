package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

func newRouter(store *Store) http.Handler {
	router := chi.NewRouter()
	router.Use(requestID, accessLog, recovery)
	router.Get("/health", health)
	router.Post("/orders", createOrder(store))
	router.Get("/orders/{id}", getOrder(store))
	return router
}

func health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func createOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "Content-Type must be application/json"})
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeRequestError(w, err)
			return
		}
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		var input struct {
			Customer string `json:"customer"`
		}
		if err := decoder.Decode(&input); err != nil {
			writeRequestError(w, err)
			return
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			writeRequestError(w, err)
			return
		}
		customer := strings.TrimSpace(input.Customer)
		if customer == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "customer must not be empty"})
			return
		}

		order := store.Create(customer)
		w.Header().Set("Location", fmt.Sprintf("/orders/%d", order.ID))
		writeJSON(w, http.StatusCreated, order)
	}
}

func getOrder(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order id must be a positive integer"})
			return
		}
		order, ok := store.ByID(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}
		writeJSON(w, http.StatusOK, order)
	}
}

func writeRequestError(w http.ResponseWriter, err error) {
	var sizeError *http.MaxBytesError
	if errors.As(err, &sizeError) {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body must not exceed 1024 bytes"})
		return
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": "expected one JSON object with a string customer field"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		slog.Error("encode response", "request_id", w.Header().Get("X-Request-ID"))
		status = http.StatusInternalServerError
		data = []byte(`{"error":"internal server error"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(data, '\n')); err != nil {
		slog.Error("write response", "request_id", w.Header().Get("X-Request-ID"))
	}
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	server := &http.Server{
		Addr: "127.0.0.1:8080", Handler: newRouter(NewStore()),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

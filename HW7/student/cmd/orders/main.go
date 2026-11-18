package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"example.com/lesson9practice/internal/httpapi"
	"example.com/lesson9practice/internal/repository"
	"example.com/lesson9practice/internal/service"
)

// @title Orders API
// @version 1.0
// @BasePath /api/v1
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	api := httpapi.New(service.New(repository.New()), logger)
	server := &http.Server{Addr: "127.0.0.1:8080", Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second}
	logger.Info("orders API", "url", "http://127.0.0.1:8080/api/v1/status")
	if err := server.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

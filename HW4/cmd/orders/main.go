package main

import (
	"fmt"

	"example.com/orders/internal/config"
	"example.com/orders/internal/handler"
	"example.com/orders/internal/repository"
	"example.com/orders/internal/service"
)

func main() {
	cfg := config.Load()
	repo := repository.NewMemory()
	orders := service.NewService(repo)
	orderHandler := handler.NewHandler(orders)

	order, err := orderHandler.Create("Покупатель")
	if err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	fmt.Printf("сервис слушал бы %s, создан заказ: %+v\n", cfg.HTTPAddr, order)
}

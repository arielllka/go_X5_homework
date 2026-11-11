package service

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidWeight    = errors.New("некорректный вес")
	ErrInvalidDistance  = errors.New("некорректная дистанция")
	ErrDistanceTooLarge = errors.New("слишком большая дистанция")
	ErrUnknownTariff    = errors.New("неизвестный тариф")
)

type OrderPriceError struct {
	OrderID string
	Tariff  string
	Err     error
}

func (e *OrderPriceError) Error() string {
	return fmt.Sprintf("заказ %s, тариф %s: %v", e.OrderID, e.Tariff, e.Err)
}

func (e *OrderPriceError) Unwrap() error {
	return e.Err
}

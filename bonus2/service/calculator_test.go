package service_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"example.com/delivery/model"
	"example.com/delivery/repository"
	"example.com/delivery/service"
)

func TestTariffPrices(t *testing.T) {
	tests := []struct {
		name     string
		pricer   service.DeliveryPricer
		weight   int
		distance int
		want     int64
	}{
		{"courier_minimum", service.NewCourier(), 1, 1, 26_500},
		{"courier_exact_kg", service.NewCourier(), 1000, 10, 40_000},
		{"courier_round_up", service.NewCourier(), 1001, 10, 45_000},
		{"courier_max_distance", service.NewCourier(), 2500, 100, 185_000},
		{"pickup_minimum", service.NewPickup(), 1, 1, 12_500},
		{"pickup_exact_kg", service.NewPickup(), 1000, 10, 17_000},
		{"pickup_round_up", service.NewPickup(), 1001, 10, 19_000},
		{"pickup_max_distance", service.NewPickup(), 2500, 50, 41_000},
		{"express_minimum", service.NewExpress(), 1, 1, 63_000},
		{"express_exact_kg", service.NewExpress(), 1000, 10, 90_000},
		{"express_round_up", service.NewExpress(), 1001, 10, 100_000},
		{"express_max_distance", service.NewExpress(), 2500, 300, 980_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pricer.Price(tt.weight, tt.distance)
			if err != nil || got != tt.want {
				t.Fatalf("Price(%d, %d) = %d, %v; want %d, nil", tt.weight, tt.distance, got, err, tt.want)
			}
		})
	}
}

func TestTariffValidation(t *testing.T) {
	tariffs := []struct {
		name        string
		pricer      service.DeliveryPricer
		maxDistance int
	}{
		{"courier", service.NewCourier(), 100},
		{"pickup", service.NewPickup(), 50},
		{"express", service.NewExpress(), 300},
	}
	for _, tariff := range tariffs {
		t.Run(tariff.name, func(t *testing.T) {
			for _, tt := range []struct {
				name     string
				weight   int
				distance int
				wantErr  error
			}{
				{"zero_weight", 0, 1, service.ErrInvalidWeight},
				{"negative_weight", -1, 1, service.ErrInvalidWeight},
				{"zero_distance", 1000, 0, service.ErrInvalidDistance},
				{"negative_distance", 1000, -1, service.ErrInvalidDistance},
				{"above_max_distance", 1000, tariff.maxDistance + 1, service.ErrDistanceTooLarge},
				{"weight_checked_first", 0, 0, service.ErrInvalidWeight},
			} {
				t.Run(tt.name, func(t *testing.T) {
					price, err := tariff.pricer.Price(tt.weight, tt.distance)
					if price != 0 || !errors.Is(err, tt.wantErr) {
						t.Fatalf("Price() = %d, %v; want 0, %v", price, err, tt.wantErr)
					}
				})
			}
		})
	}
}

func TestCalculatorResults(t *testing.T) {
	calculator := service.NewCalculator(repository.Default())
	tests := []struct {
		name     string
		tariff   string
		weight   int
		distance int
		want     int64
		wantErr  error
	}{
		{"courier", "courier", 2500, 10, 50_000, nil},
		{"pickup", "pickup", 2500, 10, 21_000, nil},
		{"express", "express", 2500, 10, 110_000, nil},
		{"unknown_tariff", "drone", 1000, 10, 0, service.ErrUnknownTariff},
		{"invalid_weight", "courier", 0, 10, 0, service.ErrInvalidWeight},
		{"invalid_distance", "pickup", 1000, 0, 0, service.ErrInvalidDistance},
		{"distance_too_large", "express", 1000, 301, 0, service.ErrDistanceTooLarge},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := model.Order{ID: "test-order", Tariff: tt.tariff, WeightGrams: tt.weight, DistanceKM: tt.distance}
			got, err := calculator.Calculate(order)
			if got != tt.want {
				t.Fatalf("Calculate() = %d, want %d", got, tt.want)
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Calculate() error = %v, want %v", err, tt.wantErr)
			}
			var priceErr *service.OrderPriceError
			if !errors.As(err, &priceErr) {
				t.Fatalf("expected *OrderPriceError, got %T", err)
			}
			if priceErr.OrderID != order.ID || priceErr.Tariff != order.Tariff {
				t.Fatalf("wrong error context: %+v", priceErr)
			}
		})
	}
}

func TestOrderPriceErrorChain(t *testing.T) {
	cause := fmt.Errorf("проверка: %w", service.ErrInvalidDistance)
	priceErr := &service.OrderPriceError{OrderID: "order-7", Tariff: "courier", Err: cause}
	for _, fragment := range []string{"order-7", "courier", cause.Error()} {
		if !strings.Contains(priceErr.Error(), fragment) {
			t.Fatalf("error %q must contain %q", priceErr.Error(), fragment)
		}
	}
	if priceErr.Unwrap() != cause {
		t.Fatal("Unwrap lost the original error")
	}
	wrapped := fmt.Errorf("расчёт: %w", priceErr)
	if !errors.Is(wrapped, service.ErrInvalidDistance) {
		t.Fatal("errors.Is lost the cause after multiple wrappers")
	}
	var target *service.OrderPriceError
	if !errors.As(wrapped, &target) || target != priceErr {
		t.Fatal("errors.As did not find the original *OrderPriceError")
	}
}

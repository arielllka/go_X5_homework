package repository_test

import (
	"errors"
	"testing"

	"example.com/delivery/repository"
	"example.com/delivery/service"
)

func TestDefaultTariffs(t *testing.T) {
	memory := repository.Default()
	for _, tt := range []struct {
		name string
		want int64
	}{
		{"courier", 40_000},
		{"pickup", 17_000},
		{"express", 90_000},
	} {
		t.Run(tt.name, func(t *testing.T) {
			pricer, err := memory.Find(tt.name)
			if err != nil || pricer == nil {
				t.Fatalf("Find(%q) = %v, %v", tt.name, pricer, err)
			}
			price, err := pricer.Price(1000, 10)
			if err != nil || price != tt.want {
				t.Fatalf("wrong tariff: price=%d, error=%v, want %d", price, err, tt.want)
			}
		})
	}
}

func TestMemoryUnknownTariff(t *testing.T) {
	for _, name := range []string{"", "drone", "Courier"} {
		t.Run(name, func(t *testing.T) {
			pricer, err := repository.Default().Find(name)
			if pricer != nil || !errors.Is(err, service.ErrUnknownTariff) || err == service.ErrUnknownTariff {
				t.Fatalf("Find(%q) = %v, %v; want nil and wrapped ErrUnknownTariff", name, pricer, err)
			}
		})
	}
}

func TestMemoryCopiesInputMap(t *testing.T) {
	pricers := map[string]service.DeliveryPricer{"courier": service.NewCourier()}
	memory := repository.NewMemory(pricers)
	delete(pricers, "courier")
	pricers["pickup"] = service.NewPickup()
	if pricer, err := memory.Find("courier"); err != nil || pricer == nil {
		t.Fatal("changing the source map removed the stored tariff")
	}
	if _, err := memory.Find("pickup"); !errors.Is(err, service.ErrUnknownTariff) {
		t.Fatal("changing the source map added a tariff to the repository")
	}
	if _, err := repository.NewMemory(nil).Find("courier"); !errors.Is(err, service.ErrUnknownTariff) {
		t.Fatal("an empty repository must return ErrUnknownTariff")
	}
}

package repository

import (
	"fmt"

	"example.com/delivery/service"
)

type Memory struct {
	pricers map[string]service.DeliveryPricer
}

func NewMemory(pricers map[string]service.DeliveryPricer) *Memory {
	copyOfPricers := make(map[string]service.DeliveryPricer, len(pricers))
	for name, pricer := range pricers {
		copyOfPricers[name] = pricer
	}
	return &Memory{pricers: copyOfPricers}
}

func Default() *Memory {
	return NewMemory(map[string]service.DeliveryPricer{
		"courier": service.NewCourier(),
		"pickup":  service.NewPickup(),
		"express": service.NewExpress(),
	})
}

func (m *Memory) Find(name string) (service.DeliveryPricer, error) {
	pricer, ok := m.pricers[name]
	if !ok {
		return nil, fmt.Errorf("тариф %q: %w", name, service.ErrUnknownTariff)
	}
	return pricer, nil
}

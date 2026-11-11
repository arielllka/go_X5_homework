package service

import "example.com/delivery/model"

type TariffRepository interface {
	Find(name string) (DeliveryPricer, error)
}

type Calculator struct {
	repository TariffRepository
}

func NewCalculator(repository TariffRepository) *Calculator {
	return &Calculator{repository: repository}
}

func (c *Calculator) Calculate(order model.Order) (int64, error) {
	pricer, err := c.repository.Find(order.Tariff)
	if err != nil {
		return 0, &OrderPriceError{OrderID: order.ID, Tariff: order.Tariff, Err: err}
	}
	price, err := pricer.Price(order.WeightGrams, order.DistanceKM)
	if err != nil {
		return 0, &OrderPriceError{OrderID: order.ID, Tariff: order.Tariff, Err: err}
	}
	return price, nil
}
